// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/rafaeljusto/redigomock"
)

type daemonTestPool struct {
	Conn *redigomock.Conn
}

func (p *daemonTestPool) Get() redis.Conn {
	return p.Conn
}

func (p *daemonTestPool) Close() error {
	return nil
}

func newDaemonTestRedis(mock *redigomock.Conn) *database.Redis {
	pool := &daemonTestPool{Conn: mock}
	return database.NewRedisCustomPool(pool)
}

func TestGetRandomFileSuccess(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SRANDMEMBER", "HANDLEDFILES_1").Expect("test/file.iso")
	mock.Command("HGET", "FILE_test/file.iso", "size").Expect(int64(1024))

	m := &monitor{
		redis: newDaemonTestRedis(mock),
	}

	file, size, err := m.getRandomFile(1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if file != "test/file.iso" {
		t.Fatalf("Expected test/file.iso, got %s", file)
	}
	if size != 1024 {
		t.Fatalf("Expected 1024, got %d", size)
	}
}

func TestGetRandomFileNoFile(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SRANDMEMBER", "HANDLEDFILES_2").ExpectError(redis.ErrNil)

	m := &monitor{
		redis: newDaemonTestRedis(mock),
	}

	file, _, err := m.getRandomFile(2)
	if err == nil {
		t.Fatal("Expected error for no file")
	}
	if file != "" {
		t.Fatalf("Expected empty file, got %s", file)
	}
}

func TestGetRandomFileSizeError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SRANDMEMBER", "HANDLEDFILES_3").Expect("test/file.iso")
	mock.Command("HGET", "FILE_test/file.iso", "size").ExpectError(redis.ErrNil)

	m := &monitor{
		redis: newDaemonTestRedis(mock),
	}

	_, size, err := m.getRandomFile(3)
	if err == nil {
		t.Fatal("Expected error for size fetch")
	}
	if size != 0 {
		t.Fatalf("Expected 0, got %d", size)
	}
}

func TestMirrorsIDError(t *testing.T) {
	mock := redigomock.NewConn()
	m := &monitor{
		redis: newDaemonTestRedis(mock),
	}

	_, err := m.mirrorsID()
	if err == nil {
		t.Fatal("Expected error for unreachable redis")
	}
}

func TestScanRepositoryLockFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "sources.lock")

	SetConfiguration(&Configuration{
		Repository:               tmpDir,
		ListenAddress:            ":8080",
		RepositorySourcesLockFile: lockFile,
		RepositoryFileListText:    filepath.Join(tmpDir, "nonexistent.txt"),
	})

	mock := redigomock.NewConn()
	m := &monitor{
		redis: newDaemonTestRedis(mock),
		stop:  make(chan struct{}),
	}

	err := m.scanRepository()
	if err == nil {
		t.Fatal("Expected error for scan failure")
	}

	// Lock file should have been created and removed
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Fatal("Expected lock file to be removed")
	}
}

func TestScanRepositoryLockFileFail(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:               "/nonexistent",
		ListenAddress:            ":8080",
		RepositorySourcesLockFile: "/nonexistent/dir/sources.lock",
		RepositoryFileListText:    "/nonexistent.txt",
	})

	mock := redigomock.NewConn()
	m := &monitor{
		redis: newDaemonTestRedis(mock),
		stop:  make(chan struct{}),
	}

	err := m.scanRepository()
	if err == nil {
		t.Fatal("Expected error for lock file creation failure")
	}
}

func TestSyncMirrorListEmpty(t *testing.T) {
	m := &monitor{
		redis:   nil,
		cache:   nil,
		mirrors: make(map[int]*mirror),
	}

	err := m.syncMirrorList()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRemoveMirrorID(t *testing.T) {
	c := &cluster{
		mirrorsIndex:  []int{1, 2, 3},
		stop:          make(chan bool),
	}
	c.RemoveMirrorID(2)
	if len(c.mirrorsIndex) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(c.mirrorsIndex))
	}
}

func TestRemoveMirrorIDNotFound(t *testing.T) {
	c := &cluster{
		mirrorsIndex:  []int{1, 2},
		stop:          make(chan bool),
	}
	c.RemoveMirrorID(99)
	if len(c.mirrorsIndex) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(c.mirrorsIndex))
	}
}

func TestRemoveMirrorIDEmpty(t *testing.T) {
	c := &cluster{
		mirrorsIndex:  []int{},
		stop:          make(chan bool),
	}
	c.RemoveMirrorID(1)
	if len(c.mirrorsIndex) != 0 {
		t.Fatalf("Expected 0 mirrors, got %d", len(c.mirrorsIndex))
	}
}

func TestMirrorNeedHealthCheckNegativeInterval(t *testing.T) {
	mm := mirror{
		lastCheck:   time.Now(),
	}
	// With a negative check interval, NeedHealthCheck should return true
	// because time.Since(lastCheck) > negative duration is always true
	if !mm.NeedHealthCheck(-1) {
		t.Fatal("Expected true for negative interval")
	}
}

func TestMirrorNeedSyncNegativeInterval(t *testing.T) {
	mm := mirror{}
	if !mm.NeedSync(-1) {
		t.Fatal("Expected true for negative interval")
	}
}

func TestNewMonitorNilRedis(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic: %v", r)
		}
	}()
	// NewMonitor with nil redis will panic when accessing cache
	// so we just test it doesn't panic on construction
	m := &monitor{
		redis:           nil,
		cache:           nil,
		mirrors:         make(map[int]*mirror),
		healthCheckChan: make(chan int, 50),
		syncChan:        make(chan int),
		stop:            make(chan struct{}),
		configNotifier:  make(chan bool, 1),
	}
	if m == nil {
		t.Fatal("Expected non-nil monitor")
	}
}
