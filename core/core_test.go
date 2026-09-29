// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package core

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGetVersionInfo(t *testing.T) {
	VERSION = "1.0.0"
	BUILD = "42"
	DEV = "dev"
	defer func() {
		VERSION = ""
		BUILD = ""
		DEV = ""
	}()

	info := GetVersionInfo()

	if info.Version != "1.0.0" {
		t.Fatalf("Expected version 1.0.0, got %s", info.Version)
	}
	if info.Build != "42dev" {
		t.Fatalf("Expected build 42dev, got %s", info.Build)
	}
	if info.GoVersion != runtime.Version() {
		t.Fatalf("Expected GoVersion %s, got %s", runtime.Version(), info.GoVersion)
	}
	if info.OS != runtime.GOOS {
		t.Fatalf("Expected OS %s, got %s", runtime.GOOS, info.OS)
	}
	if info.Arch != runtime.GOARCH {
		t.Fatalf("Expected Arch %s, got %s", runtime.GOARCH, info.Arch)
	}
	if info.GoMaxProcs != runtime.GOMAXPROCS(0) {
		t.Fatalf("Expected GoMaxProcs %d, got %d", runtime.GOMAXPROCS(0), info.GoMaxProcs)
	}
}

func TestGetVersionInfo_Empty(t *testing.T) {
	VERSION = ""
	BUILD = ""
	DEV = ""

	info := GetVersionInfo()

	if info.Version != "" {
		t.Fatalf("Expected empty version, got %s", info.Version)
	}
	if info.Build != "" {
		t.Fatalf("Expected empty build, got %s", info.Build)
	}
}

func TestPrintVersion(t *testing.T) {
	info := VersionInfo{
		Version:    "1.0.0",
		Build:      "42",
		GoVersion:  "go1.20",
		OS:         "linux",
		Arch:       "amd64",
		GoMaxProcs: 4,
	}

	PrintVersion(info)
}

func TestPrecisionDuration(t *testing.T) {
	p := Precision(5 * time.Second)
	if p.Duration() != 5*time.Second {
		t.Fatalf("Expected 5s, got %v", p.Duration())
	}

	p = Precision(time.Millisecond * 500)
	if p.Duration() != 500*time.Millisecond {
		t.Fatalf("Expected 500ms, got %v", p.Duration())
	}

	p = Precision(0)
	if p.Duration() != 0 {
		t.Fatalf("Expected 0, got %v", p.Duration())
	}
}

func TestScanConstants(t *testing.T) {
	if RSYNC != 0 {
		t.Fatalf("Expected RSYNC=0, got %d", RSYNC)
	}
	if FTP != 1 {
		t.Fatalf("Expected FTP=1, got %d", FTP)
	}
	if HTTP != 2 {
		t.Fatalf("Expected HTTP=2, got %d", HTTP)
	}
}

func TestDatabaseConstants(t *testing.T) {
	if RedisMinimumVersion != "3.2.0" {
		t.Fatalf("Expected RedisMinimumVersion 3.2.0, got %s", RedisMinimumVersion)
	}
	if DBVersion != 1 {
		t.Fatalf("Expected DBVersion 1, got %d", DBVersion)
	}
	if DBVersionKey != "MIRRORBITS_DB_VERSION" {
		t.Fatalf("Expected DBVersionKey MIRRORBITS_DB_VERSION, got %s", DBVersionKey)
	}
}

func TestContextKeys(t *testing.T) {
	if ContextAllowRedirects != 0 {
		t.Fatalf("Expected ContextAllowRedirects=0, got %d", ContextAllowRedirects)
	}
	if ContextMirrorID != 1 {
		t.Fatalf("Expected ContextMirrorID=1, got %d", ContextMirrorID)
	}
	if ContextMirrorName != 2 {
		t.Fatalf("Expected ContextMirrorName=2, got %d", ContextMirrorName)
	}
}

func TestBannerNotEmpty(t *testing.T) {
	if len(Banner) == 0 {
		t.Fatalf("Banner should not be empty")
	}
	if !strings.Contains(Banner, "%s") {
		t.Fatalf("Banner should contain format placeholder %%s")
	}
}
