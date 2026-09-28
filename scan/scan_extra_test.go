// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/filesystem"
)

func newScanWithMock() (*redigomock.Conn, *scan) {
	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	s := &scan{
		conn:        mock,
		mirrorid:    1,
		filesTmpKey: "MIRRORFILESTMP_1",
	}
	_ = pool
	return mock, s
}

func TestScannerAddFile(t *testing.T) {
	mock, s := newScanWithMock()

	fd := filesystem.FileData{
		Path:    "/test/file.txt",
		Size:    100,
		ModTime: time.Now(),
	}

	mock.Command("SADD", "MIRRORFILESTMP_1", "/test/file.txt").Expect("OK")
	mock.Command("SADD", "FILEMIRRORS_/test/file.txt", 1).Expect("OK")
	mock.Command("HMSET", "FILEINFO_1_/test/file.txt", "size", int64(100), "modTime", fd.ModTime).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_file_update", "1 /test/file.txt").Expect("OK")

	s.ScannerAddFile(fd)

	if s.count != 1 {
		t.Fatalf("Expected count 1, got %d", s.count)
	}
}

func TestScannerCommit(t *testing.T) {
	mock, s := newScanWithMock()

	mock.Command("EXEC").Expect("OK")

	err := s.ScannerCommit()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestScannerDiscard(t *testing.T) {
	mock, s := newScanWithMock()

	mock.Command("DISCARD").Expect("OK")

	s.ScannerDiscard()
}

func TestSetLastSync_Success(t *testing.T) {
	mock, s := newScanWithMock()
	SetConfiguration(&Configuration{})

	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyInt()).Expect("OK")
	mock.Command("HMSET", "MIRROR_1", "lastSuccessfulSync", redigomock.NewAnyInt(), "lastSuccessfulSyncProtocol", core.HTTP, "lastSuccessfulSyncPrecision", core.Precision(time.Second)).Expect("OK")
	mock.Command("EXEC").Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	err := s.setLastSync(mock, 1, core.HTTP, core.Precision(time.Second), true)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSetLastSync_NotSuccessful(t *testing.T) {
	mock, s := newScanWithMock()
	SetConfiguration(&Configuration{})

	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyInt()).Expect("OK")
	mock.Command("EXEC").Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	err := s.setLastSync(mock, 1, core.HTTP, 0, false)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestWalkSource_NilData(t *testing.T) {
	mock, _ := newScanWithMock()
	ss := &sourcescanner{}

	result := ss.walkSource(mock, nil)
	if result != nil {
		t.Fatalf("Expected nil for nil input")
	}
}

func TestWalkSource_Rehash(t *testing.T) {
	mock, _ := newScanWithMock()
	ss := &sourcescanner{}

	fd := &filesystem.FileData{
		Path:    "/test/file.txt",
		Size:    200,
		ModTime: time.Now(),
		Sha256:  "newhash",
	}

	mock.Command("HMGET", "FILE_/test/file.txt", "size", "modTime", "sha256").Expect([]interface{}{
		[]byte("100"),
		[]byte(""),
		[]byte("oldhash"),
	})

	result := ss.walkSource(mock, fd)
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
}

func TestWalkSource_NoPreviousData(t *testing.T) {
	mock, _ := newScanWithMock()
	ss := &sourcescanner{}

	fd := &filesystem.FileData{
		Path:    "/test/new.txt",
		Size:    200,
		ModTime: time.Now(),
		Sha256:  "hash",
	}

	mock.Command("HMGET", "FILE_/test/new.txt", "size", "modTime", "sha256").ExpectError(redis.ErrNil)

	result := ss.walkSource(mock, fd)
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
}

func TestAdjustTZOffset_NoCache(t *testing.T) {
	mock, s := newScanWithMock()
	SetConfiguration(&Configuration{FixTimezoneOffsets: true})

	s.cache = nil

	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	ms, err := s.adjustTZOffset("test", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if ms != 0 {
		t.Fatalf("Expected 0, got %d", ms)
	}
}

func TestAdjustTZOffset_FixDisabled(t *testing.T) {
	mock, s := newScanWithMock()
	SetConfiguration(&Configuration{FixTimezoneOffsets: false})

	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	ms, err := s.adjustTZOffset("test", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if ms != 0 {
		t.Fatalf("Expected 0, got %d", ms)
	}
}

func TestAdjustTZOffset_WithOffset(t *testing.T) {
	mock, s := newScanWithMock()
	SetConfiguration(&Configuration{FixTimezoneOffsets: true})

	mock.Command("SRANDMEMBER", fmt.Sprintf("HANDLEDFILES_%d", s.mirrorid), 100).ExpectError(redis.ErrNil)
	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	ms, err := s.adjustTZOffset("test", core.Precision(time.Second))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	_ = ms
}
