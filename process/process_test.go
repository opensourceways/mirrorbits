// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package process

import (
	"os"
	"testing"

	"github.com/opensourceways/mirrorbits/core"
)

func TestGetPidLocation_Runtime(t *testing.T) {
	core.PidFile = ""
	old := defaultPidFile
	defaultPidFile = ""
	defer func() { defaultPidFile = old }()

	loc := GetPidLocation()
	if loc == "" {
		t.Fatalf("Expected non-empty pid location")
	}
}

func TestGetPidLocation_DefaultPidFile(t *testing.T) {
	core.PidFile = ""
	old := defaultPidFile
	defaultPidFile = "/var/run/mirrorbits.pid"
	defer func() { defaultPidFile = old }()

	loc := GetPidLocation()
	if loc != "/var/run/mirrorbits.pid" {
		t.Fatalf("Expected '/var/run/mirrorbits.pid', got %s", loc)
	}
}

func TestGetPidLocation_XDGRuntimeDir(t *testing.T) {
	core.PidFile = ""
	os.Setenv("XDG_RUNTIME_DIR", "/tmp/xdg-test")
	defer os.Unsetenv("XDG_RUNTIME_DIR")

	loc := GetPidLocation()
	if loc != "/tmp/xdg-test/mirrorbits.pid" {
		t.Fatalf("Expected '/tmp/xdg-test/mirrorbits.pid', got %s", loc)
	}
}

func TestGetPidLocation_CorePidFile(t *testing.T) {
	core.PidFile = "/custom/path.pid"

	loc := GetPidLocation()
	if loc != "/custom/path.pid" {
		t.Fatalf("Expected '/custom/path.pid', got %s", loc)
	}

	core.PidFile = ""
}

func TestWritePidFile(t *testing.T) {
	core.PidFile = ""
	os.Setenv("XDG_RUNTIME_DIR", "/tmp/xdg-runtime")
	defer os.Unsetenv("XDG_RUNTIME_DIR")
	os.MkdirAll("/tmp/xdg-runtime", 0755)
	defer os.RemoveAll("/tmp/xdg-runtime")

	WritePidFile()

	pidFile := GetPidLocation()
	if _, err := os.Stat(pidFile); err != nil {
		t.Fatalf("Expected pid file to exist: %s", err)
	}

	pid := GetRemoteProcPid()
	if pid != os.Getpid() {
		t.Fatalf("Expected pid %d, got %d", os.Getpid(), pid)
	}

	RemovePidFile()
}

func TestGetRemoteProcPid_NotExists(t *testing.T) {
	core.PidFile = "/nonexistent/path/pid"
	defer func() { core.PidFile = "" }()

	pid := GetRemoteProcPid()
	if pid != -1 {
		t.Fatalf("Expected -1 for non-existent pid file, got %d", pid)
	}
}

func TestGetRemoteProcPid_InvalidContent(t *testing.T) {
	dir, err := os.MkdirTemp("", "pid-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	pidFile := dir + "/test.pid"
	os.WriteFile(pidFile, []byte("invalid"), 0644)

	core.PidFile = pidFile
	defer func() { core.PidFile = "" }()

	pid := GetRemoteProcPid()
	if pid != -1 {
		t.Fatalf("Expected -1 for invalid content, got %d", pid)
	}
}

func TestRemovePidFile_NotExists(t *testing.T) {
	core.PidFile = "/nonexistent/path/pid"
	defer func() { core.PidFile = "" }()

	RemovePidFile()
}

func TestRemovePidFile_OtherProcess(t *testing.T) {
	dir, err := os.MkdirTemp("", "pid-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	pidFile := dir + "/test.pid"
	os.WriteFile(pidFile, []byte("99999999"), 0644)

	core.PidFile = pidFile
	defer func() { core.PidFile = "" }()

	RemovePidFile()

	if _, err := os.Stat(pidFile); err != nil {
		t.Fatalf("Expected pid file to still exist (other process)")
	}
}

func TestErrInvalidfd(t *testing.T) {
	if ErrInvalidfd == nil {
		t.Fatalf("ErrInvalidfd should not be nil")
	}
	if ErrInvalidfd.Error() != "invalid file descriptor" {
		t.Fatalf("Expected 'invalid file descriptor', got %s", ErrInvalidfd.Error())
	}
}
