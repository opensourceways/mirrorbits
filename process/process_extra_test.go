// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package process

import (
	"os"
	"testing"

	"github.com/opensourceways/mirrorbits/core"
)

func TestRecover_NoEnv(t *testing.T) {
	os.Unsetenv("OLD_FD")
	os.Unsetenv("OLD_NAME")
	os.Unsetenv("OLD_PPID")

	_, _, err := Recover()
	if err == nil {
		t.Fatalf("Expected error when no env vars set")
	}
}

func TestRecover_InvalidFd(t *testing.T) {
	os.Setenv("OLD_FD", "abc")
	os.Setenv("OLD_NAME", "tcp:test->")
	os.Setenv("OLD_PPID", "123")
	defer os.Unsetenv("OLD_FD")
	defer os.Unsetenv("OLD_NAME")
	defer os.Unsetenv("OLD_PPID")

	_, _, err := Recover()
	if err == nil {
		t.Fatalf("Expected error for invalid fd")
	}
}

func TestKillParent(t *testing.T) {
	_ = KillParent(999999)
}

func TestWritePidFile_AndRemove(t *testing.T) {
	dir, err := os.MkdirTemp("", "process-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	core.PidFile = dir + "/test.pid"
	defer func() { core.PidFile = "" }()

	WritePidFile()

	pid := GetRemoteProcPid()
	if pid != os.Getpid() {
		t.Fatalf("Expected pid %d, got %d", os.Getpid(), pid)
	}

	RemovePidFile()

	if _, err := os.Stat(core.PidFile); !os.IsNotExist(err) {
		t.Fatalf("Expected pid file to be removed")
	}
}

func TestGetPidLocation_DefaultPath(t *testing.T) {
	core.PidFile = ""
	old := defaultPidFile
	defaultPidFile = ""
	defer func() { defaultPidFile = old; core.PidFile = "" }()

	loc := GetPidLocation()
	if loc != "/run/mirrorbits/mirrorbits.pid" {
		t.Fatalf("Expected default path, got %s", loc)
	}
}
