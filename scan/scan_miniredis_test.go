package scan

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/gomodule/redigo/redis"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
)

func setupMiniredisScan(t *testing.T) (*miniredis.Miniredis, *database.Redis, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %s", err)
	}

	mr.Server().Register("ROLE", func(c *server.Peer, cmd string, args []string) {
		c.WriteLen(3)
		c.WriteBulk("master")
		c.WriteInt(0)
		c.WriteLen(0)
	})

	SetConfiguration(&Configuration{
		RedisAddress: mr.Addr(),
		RedisDB:      0,
	})

	r := database.NewRedis()
	r.ConnectPubsub()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn := r.Get()
		if _, ok := conn.(*database.NotReadyError); !ok {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cleanup := func() {
		r.Close()
		mr.Close()
	}
	return mr, r, cleanup
}

func TestIsScanning_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	mr.Set("SCANNING_1", "mirror1")

	scanning, err := IsScanning(conn, 1)
	if err != nil {
		t.Fatalf("IsScanning failed: %s", err)
	}
	if !scanning {
		t.Fatal("Expected scanning=true")
	}

	scanning, err = IsScanning(conn, 2)
	if err != nil {
		t.Fatalf("IsScanning failed: %s", err)
	}
	if scanning {
		t.Fatal("Expected scanning=false")
	}
}

func TestScanSource_NoRepo(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	SetConfiguration(&Configuration{
		RedisAddress:          GetConfig().RedisAddress,
		Repository:            "/nonexistent/path/",
		RepositoryFileListText: "/nonexistent/filelist.txt",
		RepositorySourcesLockFile: "/tmp/test-lock-file.txt",
	})

	stop := make(chan struct{})
	defer close(stop)
	err := ScanSource(r, false, stop)
	if err == nil {
		t.Fatal("Expected error for nonexistent repository")
	}
}

func TestScanSource_NoFileList(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	repoDir, _ := os.MkdirTemp("", "testrepo")
	defer os.RemoveAll(repoDir)

	SetConfiguration(&Configuration{
		RedisAddress:          GetConfig().RedisAddress,
		Repository:            repoDir,
		RepositoryFileListText: filepath.Join(repoDir, "nonexistent.txt"),
		RepositorySourcesLockFile: filepath.Join(repoDir, "lock.txt"),
	})

	stop := make(chan struct{})
	defer close(stop)
	err := ScanSource(r, false, stop)
	if err == nil {
		t.Fatal("Expected error for nonexistent file list")
	}
}

func TestScanSource_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	repoDir, _ := os.MkdirTemp("", "testrepo")
	defer os.RemoveAll(repoDir)

	fileListPath := filepath.Join(repoDir, "filelist.txt")
	lockPath := filepath.Join(repoDir, "lock.txt")

	fileListContent := "drwxrwxrwx          4,096 2024/08/08 11:01:29 .\n"
	fileListContent += "-rwxrwxrwx          1,024 2024/08/08 11:01:29 testfile.txt\n"
	os.WriteFile(fileListPath, []byte(fileListContent), 0644)

	os.WriteFile(filepath.Join(repoDir, "testfile.txt"), []byte("test content"), 0644)

	SetConfiguration(&Configuration{
		RedisAddress:          GetConfig().RedisAddress,
		Repository:            repoDir,
		RepositoryFileListText: fileListPath,
		RepositorySourcesLockFile: lockPath,
	})

	stop := make(chan struct{})
	defer close(stop)
	err := ScanSource(r, false, stop)
	_ = err

	filesExist := mr.Exists("FILES")
	_ = filesExist
}

func TestScan_SetLastSync_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRROR_1", "ID", "1")
	mr.HSet("MIRROR_1", "name", "mirror1")
	mr.HSet("MIRROR_1", "http", "http://mirror1.example.com/")

	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis:    r,
		mirrorid: 1,
		conn:     conn,
	}

	err := s.setLastSync(conn, 1, core.HTTP, core.Precision(0), true)
	if err != nil {
		t.Fatalf("setLastSync failed: %s", err)
	}

	lastSync := mr.HGet("MIRROR_1", "lastSync")
	if lastSync == "" {
		t.Fatal("Expected lastSync to be set")
	}

	lastSuccessfulSync := mr.HGet("MIRROR_1", "lastSuccessfulSync")
	if lastSuccessfulSync == "" {
		t.Fatal("Expected lastSuccessfulSync to be set")
	}
}

func TestScan_ScannerAddFile_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis:       r,
		mirrorid:    1,
		conn:        conn,
		filesTmpKey: "MIRRORFILESTMP_1",
	}

	fd := filesystem.FileData{
		Path:    "/test/file.txt",
		Size:    1024,
		ModTime: time.Now(),
	}

	conn.Send("MULTI")
	conn.Send("DEL", s.filesTmpKey)

	s.ScannerAddFile(fd)

	_, err := conn.Do("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed: %s", err)
	}

	filesTmp, _ := mr.SMembers("MIRRORFILESTMP_1")
	if len(filesTmp) != 1 || filesTmp[0] != "/test/file.txt" {
		t.Fatalf("Expected file in MIRRORFILESTMP_1, got %v", filesTmp)
	}

	fileMirrors, _ := mr.SMembers("FILEMIRRORS_/test/file.txt")
	if len(fileMirrors) != 1 || fileMirrors[0] != "1" {
		t.Fatalf("Expected mirror ID in FILEMIRRORS, got %v", fileMirrors)
	}
}

func TestScan_ScannerCommit_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis: r,
		conn:  conn,
	}

	conn.Send("MULTI")
	conn.Send("SET", "testkey", "testvalue")

	err := s.ScannerCommit()
	if err != nil {
		t.Fatalf("ScannerCommit failed: %s", err)
	}
}

func TestScan_ScannerDiscard_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis: r,
		conn:  conn,
	}

	conn.Send("MULTI")
	conn.Send("SET", "testkey", "testvalue")

	s.ScannerDiscard()

	_, err := conn.Do("PING")
	if err != nil {
		t.Fatalf("PING after DISCARD failed: %s", err)
	}
}

func TestScan_adjustTZOffset_NoCache(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis:    r,
		mirrorid: 1,
		conn:     conn,
		cache:    nil,
	}

	_, err := s.adjustTZOffset("mirror1", core.Precision(0))
	if err != nil {
		t.Fatalf("adjustTZOffset with nil cache should not return error: %s", err)
	}
}

func TestScan_adjustTZOffset_FixTimezoneDisabled(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	SetConfiguration(&Configuration{
		RedisAddress:       GetConfig().RedisAddress,
		FixTimezoneOffsets: false,
	})

	cache := mirrors.NewCache(r)

	s := &scan{
		redis:    r,
		mirrorid: 1,
		conn:     conn,
		cache:    cache,
	}

	_, err := s.adjustTZOffset("mirror1", core.Precision(0))
	if err != nil {
		t.Fatalf("adjustTZOffset with disabled timezone fix should not return error: %s", err)
	}
}

func TestScan_Scan_EmptyURL(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRROR_1", "name", "mirror1")
	mr.HSet("MIRROR_1", "http", "")

	cache := mirrors.NewCache(r)

	stop := make(chan struct{})
	defer close(stop)

	_, err := Scan(core.HTTP, r, cache, "", 1, stop)
	_ = err
}

func TestScan_Scan_LockAcquired(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRROR_1", "name", "mirror1")
	mr.HSet("MIRROR_1", "http", "http://mirror1.example.com/")

	// Pre-set the SCANNING lock to simulate scan in progress
	mr.Set("SCANNING_1", "mirror1")

	cache := mirrors.NewCache(r)

	stop := make(chan struct{})
	defer close(stop)

	_, err := Scan(core.HTTP, r, cache, "http://mirror1.example.com/", 1, stop)
	if err == nil {
		t.Fatal("Expected error for scan in progress")
	}
}

func TestScan_Scan_HttpScanner(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRROR_1", "name", "mirror1")
	mr.HSet("MIRROR_1", "http", "http://mirror1.example.com/")

	cache := mirrors.NewCache(r)

	stop := make(chan struct{})
	defer close(stop)

	// This will try to scan via HTTP which will fail (unreachable mirror)
	_, err := Scan(core.HTTP, r, cache, "http://mirror1.example.com/", 1, stop)
	_ = err
}

func TestScan_walkSource(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	s := &sourcescanner{}

	fd := &filesystem.FileData{
		Path:    "test/file.txt",
		Size:    1024,
		ModTime: time.Now(),
	}

	result := s.walkSource(conn, fd)
	_ = result
}

func TestClusterLock_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")

	lock := network.NewClusterLock(r, "TEST_LOCK", "test")

	done, err := lock.Get()
	if err != nil {
		t.Fatalf("Lock.Get failed: %s", err)
	}
	if done == nil {
		t.Fatal("Expected non-nil done")
	}

	lock.Release()
}

func TestScan_GetMiniredisConn(t *testing.T) {
	_, r, cleanup := setupMiniredisScan(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	_ = err
}
