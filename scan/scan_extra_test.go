// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/rafaeljusto/redigomock"
)

func TestScannerAddFile(t *testing.T) {
	mock := redigomock.NewConn()
	s := &scan{
		conn:        mock,
		mirrorid:    1,
		filesTmpKey: "FILES_TMP_1",
	}

	fd := filesystem.FileData{
		Path:    "/test/file.iso",
		Size:    1024,
		ModTime: time.Now(),
	}

	s.ScannerAddFile(fd)
	if s.count != 1 {
		t.Fatalf("Expected count 1, got %d", s.count)
	}
}

func TestScannerDiscard(t *testing.T) {
	mock := redigomock.NewConn()
	mock.GenericCommand("DISCARD").Expect("OK")
	s := &scan{conn: mock}
	s.ScannerDiscard()
}

func TestScannerCommit(t *testing.T) {
	mock := redigomock.NewConn()
	mock.GenericCommand("EXEC").Expect([]interface{}{})
	s := &scan{conn: mock}
	err := s.ScannerCommit()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestScanResultFields(t *testing.T) {
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
}
