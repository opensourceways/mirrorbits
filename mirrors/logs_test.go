// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/core"
)

func parseTestTime() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestLogError_GetOutput(t *testing.T) {
	l := &LogError{
		Err: "test error",
	}
	if l.GetOutput() != "Error: test error" {
		t.Fatalf("Expected 'Error: test error', got %s", l.GetOutput())
	}
}

func TestLogAdded_GetOutput(t *testing.T) {
	l := &LogAdded{}
	if l.GetOutput() != "Mirror added" {
		t.Fatalf("Expected 'Mirror added', got %s", l.GetOutput())
	}
}

func TestLogEdited_GetOutput(t *testing.T) {
	l := &LogEdited{}
	if l.GetOutput() != "Mirror edited" {
		t.Fatalf("Expected 'Mirror edited', got %s", l.GetOutput())
	}
}

func TestLogEnabled_GetOutput(t *testing.T) {
	l := &LogEnabled{}
	if l.GetOutput() != "Mirror enabled" {
		t.Fatalf("Expected 'Mirror enabled', got %s", l.GetOutput())
	}
}

func TestLogDisabled_GetOutput(t *testing.T) {
	l := &LogDisabled{}
	if l.GetOutput() != "Mirror disabled" {
		t.Fatalf("Expected 'Mirror disabled', got %s", l.GetOutput())
	}
}

func TestLogStateChanged_GetOutput_Up(t *testing.T) {
	l := &LogStateChanged{Up: true}
	if l.GetOutput() != "Mirror is up" {
		t.Fatalf("Expected 'Mirror is up', got %s", l.GetOutput())
	}
}

func TestLogStateChanged_GetOutput_DownNoReason(t *testing.T) {
	l := &LogStateChanged{Up: false, Reason: ""}
	if l.GetOutput() != "Mirror is down" {
		t.Fatalf("Expected 'Mirror is down', got %s", l.GetOutput())
	}
}

func TestLogStateChanged_GetOutput_DownWithReason(t *testing.T) {
	l := &LogStateChanged{Up: false, Reason: "timeout"}
	if l.GetOutput() != "Mirror is down: timeout" {
		t.Fatalf("Expected 'Mirror is down: timeout', got %s", l.GetOutput())
	}
}

func TestLogScanStarted_GetOutput(t *testing.T) {
	l := &LogScanStarted{Typ: core.RSYNC}
	if l.GetOutput() != "RSYNC scan started" {
		t.Fatalf("Expected 'RSYNC scan started', got %s", l.GetOutput())
	}

	l = &LogScanStarted{Typ: core.FTP}
	if l.GetOutput() != "FTP scan started" {
		t.Fatalf("Expected 'FTP scan started', got %s", l.GetOutput())
	}

	l = &LogScanStarted{Typ: core.HTTP}
	if l.GetOutput() != "Scan started using a unknown protocol" {
		t.Fatalf("Expected unknown protocol message, got %s", l.GetOutput())
	}
}

func TestLogScanCompleted_GetOutput(t *testing.T) {
	l := &LogScanCompleted{
		FilesIndexed: 100,
		KnownIndexed: 90,
		Removed:      5,
	}
	expected := "Scan completed: 100 files (90 known), 5 removed"
	if l.GetOutput() != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, l.GetOutput())
	}
}

func TestLogScanCompleted_GetOutput_WithTZOffset(t *testing.T) {
	l := &LogScanCompleted{
		FilesIndexed: 100,
		KnownIndexed: 90,
		Removed:      5,
		TZOffset:     3600000,
	}
	output := l.GetOutput()
	if output == "" {
		t.Fatalf("Expected non-empty output")
	}
}

func TestNewLogError(t *testing.T) {
	l := NewLogError(1, errTest("some error"))
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetMirrorID() != 1 {
		t.Fatalf("Expected mirror ID 1, got %d", l.GetMirrorID())
	}
	if l.GetType() != LOGTYPE_ERROR {
		t.Fatalf("Expected LOGTYPE_ERROR, got %d", l.GetType())
	}
}

func TestNewLogAdded(t *testing.T) {
	l := NewLogAdded(1)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetMirrorID() != 1 {
		t.Fatalf("Expected mirror ID 1, got %d", l.GetMirrorID())
	}
	if l.GetType() != LOGTYPE_ADDED {
		t.Fatalf("Expected LOGTYPE_ADDED, got %d", l.GetType())
	}
}

func TestNewLogEdited(t *testing.T) {
	l := NewLogEdited(1)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_EDITED {
		t.Fatalf("Expected LOGTYPE_EDITED, got %d", l.GetType())
	}
}

func TestNewLogEnabled(t *testing.T) {
	l := NewLogEnabled(1)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_ENABLED {
		t.Fatalf("Expected LOGTYPE_ENABLED, got %d", l.GetType())
	}
}

func TestNewLogDisabled(t *testing.T) {
	l := NewLogDisabled(1)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_DISABLED {
		t.Fatalf("Expected LOGTYPE_DISABLED, got %d", l.GetType())
	}
}

func TestNewLogStateChanged(t *testing.T) {
	l := NewLogStateChanged(1, true, "")
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_STATECHANGED {
		t.Fatalf("Expected LOGTYPE_STATECHANGED, got %d", l.GetType())
	}
}

func TestNewLogScanStarted(t *testing.T) {
	l := NewLogScanStarted(1, core.RSYNC)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_SCANSTARTED {
		t.Fatalf("Expected LOGTYPE_SCANSTARTED, got %d", l.GetType())
	}
}

func TestNewLogScanCompleted(t *testing.T) {
	l := NewLogScanCompleted(1, 10, 8, 2, 0)
	if l == nil {
		t.Fatalf("Expected non-nil")
	}
	if l.GetType() != LOGTYPE_SCANCOMPLETED {
		t.Fatalf("Expected LOGTYPE_SCANCOMPLETED, got %d", l.GetType())
	}
}

func TestTypeToInstance(t *testing.T) {
	tests := []struct {
		typ     LogType
		isNil   bool
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
		result := typeToInstance(tt.typ)
		if tt.isNil && result != nil {
			t.Fatalf("Expected nil for type %d, got non-nil", tt.typ)
		}
		if !tt.isNil && result == nil {
			t.Fatalf("Expected non-nil for type %d, got nil", tt.typ)
		}
	}
}

func TestLogCommonAction_Getters(t *testing.T) {
	l := LogCommonAction{
		Type:      LOGTYPE_ADDED,
		MirrorID:  42,
		Timestamp: parseTestTime(),
	}
	if l.GetType() != LOGTYPE_ADDED {
		t.Fatalf("Expected LOGTYPE_ADDED, got %d", l.GetType())
	}
	if l.GetMirrorID() != 42 {
		t.Fatalf("Expected 42, got %d", l.GetMirrorID())
	}
	if l.GetTimestamp() != parseTestTime() {
		t.Fatalf("Timestamp mismatch")
	}
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}
