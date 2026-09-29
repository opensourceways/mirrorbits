// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package process

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/opensourceways/mirrorbits/core"
)

func TestGetPidLocationDefault(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = ""

	oldEnv := os.Getenv("XDG_RUNTIME_DIR")
	defer os.Setenv("XDG_RUNTIME_DIR", oldEnv)
	os.Unsetenv("XDG_RUNTIME_DIR")

	p := GetPidLocation()
	if p != "/run/mirrorbits/mirrorbits.pid" && defaultPidFile == "" {
		t.Fatalf("Expected /run/mirrorbits/mirrorbits.pid, got %s", p)
	}
}

func TestGetPidLocationFromPidFile(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = "/custom/path.pid"

	p := GetPidLocation()
	if p != "/custom/path.pid" {
		t.Fatalf("Expected /custom/path.pid, got %s", p)
	}
}

func TestGetPidLocationFromXDG(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = ""

	oldEnv := os.Getenv("XDG_RUNTIME_DIR")
	defer os.Setenv("XDG_RUNTIME_DIR", oldEnv)
	os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")

	p := GetPidLocation()
	if p != "/run/user/1000/mirrorbits.pid" {
		t.Fatalf("Expected /run/user/1000/mirrorbits.pid, got %s", p)
	}
}

func TestWriteAndReadPidFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pid-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	pidPath := filepath.Join(dir, "mirrorbits.pid")

	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = pidPath

	WritePidFile()

	data, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("Expected pid file to exist: %s", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("Expected integer pid, got %s", string(data))
	}
	if pid != os.Getpid() {
		t.Fatalf("Expected %d, got %d", os.Getpid(), pid)
	}

	gotPid := GetRemoteProcPid()
	if gotPid != os.Getpid() {
		t.Fatalf("Expected %d, got %d", os.Getpid(), gotPid)
	}
}

func TestGetRemoteProcPidNonExistent(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = "/nonexistent/path/to/pid/file.pid"

	pid := GetRemoteProcPid()
	if pid != -1 {
		t.Fatalf("Expected -1, got %d", pid)
	}
}

func TestRemovePidFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pid-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	pidPath := filepath.Join(dir, "mirrorbits.pid")

	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = pidPath

	WritePidFile()

	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Fatal("Expected pid file to exist")
	}

	RemovePidFile()

	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatal("Expected pid file to be removed")
	}
}

func TestRemovePidFileNonExistent(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = "/nonexistent/path/to/pid/file.pid"

	RemovePidFile()
}
