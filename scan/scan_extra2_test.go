// Copyright (c) 2014-2019 Ludovic Fauzet
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

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestSetLastSync(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.Command("HMSET", "MIRROR_1", "lastSuccessfulSync", redigomock.NewAnyData(), "lastSuccessfulSyncProtocol", core.HTTP, "lastSuccessfulSyncPrecision", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	mock.Command("PUBLISH", "MIRROR_UPDATE", "1").Expect("OK")

	s := &scan{conn: mock, mirrorid: 1}
	err := s.setLastSync(mock, 1, core.HTTP, core.Precision(time.Second), true)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSetLastSyncUnsuccessful(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_1", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	mock.Command("PUBLISH", "MIRROR_UPDATE", "1").Expect("OK")

	s := &scan{conn: mock, mirrorid: 1}
	err := s.setLastSync(mock, 1, core.HTTP, 0, false)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSetLastSyncPrecisionZero(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("MULTI").Expect("OK")
	mock.Command("HSET", "MIRROR_2", "lastSync", redigomock.NewAnyData()).Expect("OK")
	mock.Command("HMSET", "MIRROR_2", "lastSuccessfulSync", redigomock.NewAnyData(), "lastSuccessfulSyncProtocol", core.RSYNC, "lastSuccessfulSyncPrecision", redigomock.NewAnyData()).Expect("OK")
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	mock.Command("PUBLISH", "MIRROR_UPDATE", "2").Expect("OK")

	s := &scan{conn: mock, mirrorid: 2}
	err := s.setLastSync(mock, 2, core.RSYNC, 0, true)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestAdjustTZOffsetNilCache(t *testing.T) {
	SetConfiguration(&Configuration{
		FixTimezoneOffsets: true,
	})
	mock := redigomock.NewConn()
	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.GenericCommand("PUBLISH").Expect("OK")

	s := &scan{conn: mock, mirrorid: 1, cache: nil}
	_, err := s.adjustTZOffset("test", core.Precision(time.Second))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestAdjustTZOffsetDisabled(t *testing.T) {
	SetConfiguration(&Configuration{
		FixTimezoneOffsets: false,
	})
	mock := redigomock.NewConn()
	mock.Command("HMSET", "MIRROR_1", "tzoffset", int64(0)).Expect("OK")
	mock.GenericCommand("PUBLISH").Expect("OK")

	s := &scan{conn: mock, mirrorid: 1, cache: nil}
	_, err := s.adjustTZOffset("test", core.Precision(time.Second))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestWalkSourceNil(t *testing.T) {
	mock := redigomock.NewConn()
	s := &sourcescanner{}
	result := s.walkSource(mock, nil)
	if result != nil {
		t.Fatal("Expected nil for nil input")
	}
}

func TestWalkSource(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("HMGET", "FILE_/test/path", "size", "modTime", "sha256").Expect([]interface{}{
		[]byte("1024"), []byte("2024-08-08 11:01:29"), []byte("abc123"),
	})

	s := &sourcescanner{}
	fd := &filesystem.FileData{Path: "/test/path", Size: 1024, ModTime: time.Now(), Sha256: "abc123"}
	result := s.walkSource(mock, fd)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestWalkSourceRehash(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("HMGET", "FILE_/test/new", "size", "modTime", "sha256").ExpectError(errors.New("not found"))

	s := &sourcescanner{}
	fd := &filesystem.FileData{Path: "/test/new", Size: 2048, ModTime: time.Now(), Sha256: "newsha"}
	result := s.walkSource(mock, fd)
	if result != nil {
		t.Fatal("Expected nil result on error")
	}
}

func TestHttpScannerScanNonHTTPS(t *testing.T) {
	s := &scan{mirrorid: 1}
	scanner := &HttpScanner{scan: s}
	repoVersion := []*filesystem.LayerFile{
		{Dir: "test", Name: "file.iso"},
	}
	_, _, err := scanner.Scan("http://example.com", "test", repoVersion, nil)
	if err == nil {
		t.Fatal("Expected error for non-https URL")
	}
}

func TestHttpScannerScanStopped(t *testing.T) {
	s := &scan{mirrorid: 1}
	scanner := &HttpScanner{scan: s}
	repoVersion := []*filesystem.LayerFile{
		{Dir: "test", Name: "file.iso"},
	}
	stop := make(chan struct{})
	close(stop)
	_, _, err := scanner.Scan("https://example.com", "test", repoVersion, stop)
	if err != ErrScanAborted {
		t.Fatalf("Expected ErrScanAborted, got %v", err)
	}
}

func TestNewTraceHandler(t *testing.T) {
	stop := make(chan struct{})
	tr := NewTraceHandler(nil, stop)
	if tr == nil {
		t.Fatal("Expected non-nil Trace")
	}
}

func TestTraceGetLastUpdateNoTraceFile(t *testing.T) {
	SetConfiguration(&Configuration{
		TraceFileLocation: "",
	})
	tr := &Trace{redis: nil, stop: nil}
	err := tr.GetLastUpdate(mirrors.Mirror{})
	if err != ErrNoTrace {
		t.Fatalf("Expected ErrNoTrace, got %v", err)
	}
}

func TestScanResultStruct(t *testing.T) {
	r := ScanResult{
		MirrorID:     1,
		MirrorName:   "test",
		FilesIndexed: 100,
		KnownIndexed: 90,
		Removed:      10,
		TZOffsetMs:   3600000,
	}
	if r.MirrorID != 1 || r.MirrorName != "test" || r.FilesIndexed != 100 {
		t.Fatal("Fields mismatch")
	}
	if r.KnownIndexed != 90 || r.Removed != 10 || r.TZOffsetMs != 3600000 {
		t.Fatal("Fields mismatch")
	}
}
