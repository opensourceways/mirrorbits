// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"errors"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/rafaeljusto/redigomock"
)

func TestSetLastSyncError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
	mock := redigomock.NewConn()
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_3", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").ExpectError(errors.New("EXEC error"))
	mock.Command("PUBLISH", "MIRROR_UPDATE", "3").Expect("OK")

	s := &scan{conn: mock, mirrorid: 3}
	err := s.setLastSync(mock, 3, core.HTTP, 0, false)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestAdjustTZOffsetWithCacheDisabled(t *testing.T) {
	SetConfiguration(&Configuration{
		FixTimezoneOffsets: false,
	})
	mock := redigomock.NewConn()
	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.Command("PUBLISH", "MIRROR_UPDATE", "1").Expect("OK")

	s := &scan{
		conn:     mock,
		mirrorid: 1,
		cache:    &mirrors.Cache{},
	}
	_, err := s.adjustTZOffset("test", core.Precision(time.Second))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestHttpScannerStruct(t *testing.T) {
	s := &scan{mirrorid: 1}
	scanner := &HttpScanner{scan: s}
	if scanner.scan != s {
		t.Fatal("Struct mismatch")
	}
}

func TestScannerAddFileMultiple(t *testing.T) {
	mock := redigomock.NewConn()
	s := &scan{
		conn:        mock,
		mirrorid:    1,
		filesTmpKey: "FILES_TMP_1",
	}

	for i := 0; i < 3; i++ {
		fd := filesystem.FileData{
			Path:    "/test/file" + string(rune('0'+i)) + ".iso",
			Size:    int64(1024 * (i + 1)),
			ModTime: time.Now(),
		}
		s.ScannerAddFile(fd)
	}
	if s.count != 3 {
		t.Fatalf("Expected count 3, got %d", s.count)
	}
}

func TestScannerCommitError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.GenericCommand("EXEC").ExpectError(errors.New("EXEC failed"))
	s := &scan{conn: mock}
	err := s.ScannerCommit()
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestScannerDiscardError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.GenericCommand("DISCARD").ExpectError(errors.New("DISCARD failed"))
	s := &scan{conn: mock}
	s.ScannerDiscard()
}

func TestScanResultFull(t *testing.T) {
	r := ScanResult{
		MirrorID:     42,
		MirrorName:   "mirror.example.com",
		FilesIndexed: 1000,
		KnownIndexed: 900,
		Removed:      100,
		TZOffsetMs:   7200000,
	}
	if r.MirrorID != 42 || r.MirrorName != "mirror.example.com" {
		t.Fatal("Fields mismatch")
	}
	if r.FilesIndexed != 1000 || r.KnownIndexed != 900 || r.Removed != 100 {
		t.Fatal("Fields mismatch")
	}
	if r.TZOffsetMs != 7200000 {
		t.Fatal("Fields mismatch")
	}
}

func TestErrMessages(t *testing.T) {
	if ErrScanAborted.Error() == "" {
		t.Fatal("Expected non-empty")
	}
	if ErrScanInProgress.Error() == "" {
		t.Fatal("Expected non-empty")
	}
	if ErrNoSyncMethod.Error() == "" {
		t.Fatal("Expected non-empty")
	}
	if ErrNoTrace.Error() == "" {
		t.Fatal("Expected non-empty")
	}
}

func TestTraceStruct(t *testing.T) {
	tr := &Trace{
		redis: nil,
		stop:  make(chan struct{}),
	}
	if tr.redis != nil {
		t.Fatal("Expected nil redis")
	}
}
