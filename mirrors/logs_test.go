// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"errors"
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/core"
)

func TestTypeToInstance(t *testing.T) {
	tests := []struct {
		typ     LogType
		wantNil bool
	}{
		{LOGTYPE_ERROR, false},
		{LOGTYPE_ADDED, false},
		{LOGTYPE_EDITED, false},
		{LOGTYPE_ENABLED, false},
		{LOGTYPE_DISABLED, false},
		{LOGTYPE_STATECHANGED, false},
		{LOGTYPE_SCANSTARTED, false},
		{LOGTYPE_SCANCOMPLETED, false},
		{LogType(999), true},
	}
	for _, tt := range tests {
		r := typeToInstance(tt.typ)
		if (r == nil) != tt.wantNil {
			t.Fatalf("typeToInstance(%d): expected nil=%v, got nil=%v", tt.typ, tt.wantNil, r == nil)
		}
	}
}

func TestLogCommonActionGetters(t *testing.T) {
	ts := time.Now()
	l := LogCommonAction{Type: LOGTYPE_ADDED, MirrorID: 5, Timestamp: ts}
	if l.GetType() != LOGTYPE_ADDED {
		t.Fatal("GetType mismatch")
	}
	if l.GetMirrorID() != 5 {
		t.Fatal("GetMirrorID mismatch")
	}
	if !l.GetTimestamp().Equal(ts) {
		t.Fatal("GetTimestamp mismatch")
	}
}

func TestLogError(t *testing.T) {
	l := NewLogError(1, errors.New("test error"))
	if l.GetType() != LOGTYPE_ERROR {
		t.Fatal("Type mismatch")
	}
	if l.GetMirrorID() != 1 {
		t.Fatal("MirrorID mismatch")
	}
	if l.GetOutput() != "Error: test error" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogAdded(t *testing.T) {
	l := NewLogAdded(2)
	if l.GetType() != LOGTYPE_ADDED {
		t.Fatal("Type mismatch")
	}
	if l.GetMirrorID() != 2 {
		t.Fatal("MirrorID mismatch")
	}
	if l.GetOutput() != "Mirror added" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogEdited(t *testing.T) {
	l := NewLogEdited(3)
	if l.GetType() != LOGTYPE_EDITED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() != "Mirror edited" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogEnabled(t *testing.T) {
	l := NewLogEnabled(4)
	if l.GetType() != LOGTYPE_ENABLED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() != "Mirror enabled" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogDisabled(t *testing.T) {
	l := NewLogDisabled(5)
	if l.GetType() != LOGTYPE_DISABLED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() != "Mirror disabled" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogStateChangedUp(t *testing.T) {
	l := NewLogStateChanged(6, true, "")
	if l.GetType() != LOGTYPE_STATECHANGED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() != "Mirror is up" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogStateChangedDown(t *testing.T) {
	l := NewLogStateChanged(7, false, "connection refused")
	if l.GetType() != LOGTYPE_STATECHANGED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() != "Mirror is down: connection refused" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogStateChangedDownNoReason(t *testing.T) {
	l := NewLogStateChanged(8, false, "")
	if l.GetOutput() != "Mirror is down" {
		t.Fatalf("Output mismatch: %s", l.GetOutput())
	}
}

func TestLogScanStarted(t *testing.T) {
	tests := []struct {
		typ  core.ScannerType
		want string
	}{
		{core.RSYNC, "RSYNC scan started"},
		{core.FTP, "FTP scan started"},
		{core.ScannerType(99), "Scan started using a unknown protocol"},
	}
	for _, tt := range tests {
		l := NewLogScanStarted(8, tt.typ)
		if l.GetType() != LOGTYPE_SCANSTARTED {
			t.Fatal("Type mismatch")
		}
		if l.GetOutput() != tt.want {
			t.Fatalf("Expected %q, got %q", tt.want, l.GetOutput())
		}
	}
}

func TestLogScanCompleted(t *testing.T) {
	l := NewLogScanCompleted(9, 100, 90, 5, int64(3600000))
	if l.GetType() != LOGTYPE_SCANCOMPLETED {
		t.Fatal("Type mismatch")
	}
	if l.GetOutput() == "" {
		t.Fatal("Expected non-empty output")
	}
}
