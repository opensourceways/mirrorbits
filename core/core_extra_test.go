// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package core

import (
	"os"
	"testing"
)

func TestParseflags_Daemon(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"mirrorbits", "daemon", "-config", "/etc/mirrorbits.conf", "-monitor", "true"}
	Parseflags()

	if !Daemon {
		t.Fatalf("Expected Daemon=true")
	}
	if ConfigFile != "/etc/mirrorbits.conf" {
		t.Fatalf("Expected ConfigFile, got %s", ConfigFile)
	}
	if !Monitor {
		t.Fatalf("Expected Monitor=true")
	}
}

func TestFlagsDefaults(t *testing.T) {
	if RPCPort != 3390 {
		t.Fatalf("Expected default RPCPort 3390, got %d", RPCPort)
	}
	if RPCHost != "localhost" {
		t.Fatalf("Expected default RPCHost 'localhost', got %s", RPCHost)
	}
}
