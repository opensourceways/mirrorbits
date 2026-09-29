// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package core

import (
	"strings"
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	vi := GetVersionInfo()
	if vi.Version != VERSION {
		t.Fatalf("Expected %s, got %s", VERSION, vi.Version)
	}
	if vi.Build != BUILD+DEV {
		t.Fatalf("Expected %s, got %s", BUILD+DEV, vi.Build)
	}
	if vi.GoVersion == "" {
		t.Fatal("Expected non-empty GoVersion")
	}
	if vi.OS == "" {
		t.Fatal("Expected non-empty OS")
	}
	if vi.Arch == "" {
		t.Fatal("Expected non-empty Arch")
	}
	if vi.GoMaxProcs <= 0 {
		t.Fatalf("Expected GoMaxProcs > 0, got %d", vi.GoMaxProcs)
	}
}

func TestPrintVersion(t *testing.T) {
	vi := GetVersionInfo()
	PrintVersion(vi)
}

func TestBannerContainsVersion(t *testing.T) {
	if !strings.Contains(Banner, "%s") {
		t.Fatal("Expected Banner to contain format specifier")
	}
}

func TestParseflags(t *testing.T) {
	Parseflags()
	if NArg < 0 {
		t.Fatal("NArg should be >= 0")
	}
}

func TestContextKeys(t *testing.T) {
	if ContextAllowRedirects != 0 {
		t.Fatalf("Expected ContextAllowRedirects to be 0, got %d", ContextAllowRedirects)
	}
	if ContextMirrorID != 1 {
		t.Fatalf("Expected ContextMirrorID to be 1, got %d", ContextMirrorID)
	}
	if ContextMirrorName != 2 {
		t.Fatalf("Expected ContextMirrorName to be 2, got %d", ContextMirrorName)
	}
}
