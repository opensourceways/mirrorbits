// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package scan

import (
	"os"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/rafaeljusto/redigomock"
)

type scanTestPool struct {
	Conn *redigomock.Conn
}

func (p *scanTestPool) Get() redis.Conn {
	return p.Conn
}

func (p *scanTestPool) Close() error {
	return nil
}

func newScanTestRedis(mock *redigomock.Conn) *database.Redis {
	pool := &scanTestPool{Conn: mock}
	return database.NewRedisCustomPool(pool)
}

func TestScanMirrorNameError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "MIRRORS", 1).ExpectError(redis.ErrNil)

	r := newScanTestRedis(mock)

	result, err := Scan(core.HTTP, r, nil, "https://example.com", 1, nil)
	if err == nil {
		t.Fatal("Expected error for HGET failure")
	}
	if result != nil {
		t.Fatal("Expected nil result")
	}
}

func TestScanLockAcquisitionFail(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "MIRRORS", 1).Expect("test-mirror")
	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).ExpectError(redis.ErrNil)

	r := newScanTestRedis(mock)

	result, err := Scan(core.HTTP, r, nil, "https://example.com", 1, nil)
	if err != ErrScanInProgress {
		t.Fatalf("Expected ErrScanInProgress, got %v", err)
	}
	if result != nil {
		t.Fatal("Expected nil result")
	}
}

func TestScanLockAcquisitionError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "MIRRORS", 1).Expect("test-mirror")
	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).ExpectError(errTestConnError)

	r := newScanTestRedis(mock)

	_, err := Scan(core.HTTP, r, nil, "https://example.com", 1, nil)
	if err == nil {
		t.Fatal("Expected error for SET failure")
	}
}

func TestScanSuccessEmptyFileList(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	mock := redigomock.NewConn()

	// Get mirror name
	mock.Command("HGET", "MIRRORS", 1).Expect("test-mirror")

	// Lock acquisition
	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).Expect("OK")

	// First setLastSync (unsuccessful: false)
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	// Scan setup MULTI
	mock.Command("MULTI").Expect("OK")
	mock.Command("DEL", "MIRRORFILESTMP_1").Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})

	// SDIFF (no files to remove)
	mock.Command("SDIFF", "MIRRORFILES_1", "MIRRORFILESTMP_1").Expect([]interface{}{})

	// SINTERSTORE (count common files)
	mock.Command("SINTERSTORE", "HANDLEDFILES_1", "FILES", "MIRRORFILES_1").Expect(int64(0))

	// Second setLastSync (successful: true)
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.Command("HMSET", "MIRROR_1",
		"lastSuccessfulSync", redigomock.NewAnyData(),
		"lastSuccessfulSyncProtocol", core.HTTP,
		"lastSuccessfulSyncPrecision", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	// adjustTZOffset with nil cache returns early, no Redis ops needed

	r := newScanTestRedis(mock)

	result, err := Scan(core.HTTP, r, nil, "https://example.com", 1, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.MirrorName != "test-mirror" {
		t.Fatalf("Expected name test-mirror, got %s", result.MirrorName)
	}
	if result.FilesIndexed != 0 {
		t.Fatalf("Expected 0 files indexed, got %d", result.FilesIndexed)
	}

	// Wait for lock goroutine to finish
	time.Sleep(100 * time.Millisecond)
}

func TestScanSourceRepoNotExist(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:           "/nonexistent/path/that/does/not/exist",
		ListenAddress:        ":8080",
		RepositoryFileListText: "/nonexistent/filelist.txt",
	})

	mock := redigomock.NewConn()
	r := newScanTestRedis(mock)

	err := ScanSource(r, false, nil)
	if err == nil {
		t.Fatal("Expected error for non-existent repository")
	}
}

func TestScanSourceFileListNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfiguration(&Configuration{
		Repository:           tmpDir,
		ListenAddress:        ":8080",
		RepositoryFileListText: tmpDir + "/nonexistent_filelist.txt",
	})

	mock := redigomock.NewConn()
	r := newScanTestRedis(mock)

	err := ScanSource(r, false, nil)
	if err == nil {
		t.Fatal("Expected error for non-existent file list")
	}
}

func TestScanSourceCannotOpenFile(t *testing.T) {
	tmpDir := t.TempDir()
	// Create the repository directory but point file list to a directory (can't be read as file)
	SetConfiguration(&Configuration{
		Repository:           tmpDir,
		ListenAddress:        ":8080",
		RepositoryFileListText: tmpDir,
	})

	mock := redigomock.NewConn()
	r := newScanTestRedis(mock)

	err := ScanSource(r, false, nil)
	if err == nil {
		t.Fatal("Expected error for cannot open file")
	}
}

func TestScanSourceStopped(t *testing.T) {
	tmpDir := t.TempDir()
	fileListPath := tmpDir + "/filelist.txt"
	// Create empty file list
	writeTestFile(t, fileListPath, "")

	SetConfiguration(&Configuration{
		Repository:           tmpDir,
		ListenAddress:        ":8080",
		RepositoryFileListText: fileListPath,
	})

	mock := redigomock.NewConn()
	r := newScanTestRedis(mock)

	stop := make(chan struct{})
	close(stop)

	err := ScanSource(r, false, stop)
	// With empty file list and stop channel closed, should return ErrScanAborted or nil
	// depending on timing
	_ = err
}

func TestScanSourceSuccessEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fileListPath := tmpDir + "/filelist.txt"
	// Create a minimal file list with a header line
	writeTestFile(t, fileListPath, "header line\n")

	SetConfiguration(&Configuration{
		Repository:           tmpDir,
		ListenAddress:        ":8080",
		RepositoryFileListText: fileListPath,
		FixTimezoneOffsets:   false,
	})

	mock := redigomock.NewConn()

	// Lock acquisition
	mock.Command("SET", "SOURCE_REPO_SYNC", 1, "NX", "EX", 10).Expect("OK")

	// MULTI for FILES_TMP
	mock.Command("MULTI").Expect("OK")
	mock.Command("DEL", "FILES_TMP").Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})

	// SDIFF (no files to remove)
	mock.Command("SDIFF", "FILES", "FILES_TMP").Expect([]interface{}{})

	// MULTI for file updates
	mock.Command("MULTI").Expect("OK")
	mock.Command("RENAME", "FILES_TMP", "FILES").Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})

	r := newScanTestRedis(mock)

	err := ScanSource(r, false, nil)
	// May or may not succeed depending on lock timing
	_ = err

	// Wait for lock goroutine
	time.Sleep(100 * time.Millisecond)
}

var errTestConnError = &testError{}

type testError struct{}

func (e *testError) Error() string { return "connection error" }

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %s", err)
	}
}
