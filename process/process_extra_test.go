// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package process

import (
	"os"
	"testing"

	"github.com/opensourceways/mirrorbits/core"
)

func TestRecoverNoEnv(t *testing.T) {
	oldFD := os.Getenv("OLD_FD")
	defer os.Setenv("OLD_FD", oldFD)
	os.Unsetenv("OLD_FD")

	_, _, err := Recover()
	if err == nil {
		t.Fatal("Expected error for no OLD_FD env")
	}
}

func TestRecoverInvalidEnv(t *testing.T) {
	oldFD := os.Getenv("OLD_FD")
	defer os.Setenv("OLD_FD", oldFD)
	os.Setenv("OLD_FD", "invalid")

	_, _, err := Recover()
	if err == nil {
		t.Fatal("Expected error for invalid OLD_FD")
	}
}

func TestKillParentInvalidPid(t *testing.T) {
	err := KillParent(999999)
	if err == nil {
		t.Fatal("Expected error for invalid pid")
	}
}

func TestRelaunchInvalidListener(t *testing.T) {
	err := Relaunch(nil)
	if err != ErrInvalidfd {
		t.Fatalf("Expected ErrInvalidfd, got %v", err)
	}
}

func TestGetPidLocationFromDefaultPidFile(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = ""

	oldEnv := os.Getenv("XDG_RUNTIME_DIR")
	defer os.Setenv("XDG_RUNTIME_DIR", oldEnv)
	os.Unsetenv("XDG_RUNTIME_DIR")

	p := GetPidLocation()
	if p == "" {
		t.Fatal("Expected non-empty path")
	}
}

func TestGetRemoteProcPidInvalidFile(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = "/nonexistent/path/to/pid"

	pid := GetRemoteProcPid()
	if pid != -1 {
		t.Fatalf("Expected -1, got %d", pid)
	}
}

func TestGetRemoteProcPidInvalidContent(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pid-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	pidPath := dir + "/mirrorbits.pid"
	os.WriteFile(pidPath, []byte("not-a-number"), 0644)

	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = pidPath

	pid := GetRemoteProcPid()
	if pid != -1 {
		t.Fatalf("Expected -1 for invalid content, got %d", pid)
	}
}

func TestWritePidFileToDirWithoutPermission(t *testing.T) {
	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = "/nonexistent_root/no_perm/mirrorbits.pid"

	WritePidFile()
}

func TestRemovePidFileNotOurs(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pid-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	pidPath := dir + "/mirrorbits.pid"
	os.WriteFile(pidPath, []byte("999999"), 0644)

	oldPidFile := core.PidFile
	defer func() { core.PidFile = oldPidFile }()
	core.PidFile = pidPath

	RemovePidFile()
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Fatal("File should not be removed (not our pid)")
	}
}
