// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package cli

import (
	"testing"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opensourceways/mirrorbits/rpc"
)

func rpcTimestampPtr(seconds int64) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: seconds}
}

func TestSubCmd(t *testing.T) {
	flags := SubCmd("test", "[options]", "Test command")
	if flags == nil {
		t.Fatalf("Expected non-nil FlagSet")
	}
	if flags.Name() != "test" {
		t.Fatalf("Expected name 'test', got %s", flags.Name())
	}
}

func TestSubCmd_Usage(t *testing.T) {
	flags := SubCmd("testcmd", "<arg>", "A test subcommand")
	if flags == nil {
		t.Fatalf("Expected non-nil FlagSet")
	}
}

func TestByDate_Len(t *testing.T) {
	d := ByDate{
		&rpc.Mirror{ID: 1},
		&rpc.Mirror{ID: 2},
		&rpc.Mirror{ID: 3},
	}
	if d.Len() != 3 {
		t.Fatalf("Expected 3, got %d", d.Len())
	}

	d = ByDate{}
	if d.Len() != 0 {
		t.Fatalf("Expected 0, got %d", d.Len())
	}
}

func TestByDate_Swap(t *testing.T) {
	d := ByDate{
		&rpc.Mirror{ID: 1, Name: "first"},
		&rpc.Mirror{ID: 2, Name: "second"},
	}

	d.Swap(0, 1)

	if d[0].ID != 2 || d[1].ID != 1 {
		t.Fatalf("Swap didn't work correctly")
	}
}

func TestByDate_Less(t *testing.T) {
	d := ByDate{
		&rpc.Mirror{ID: 1, StateSince: rpcTimestampPtr(100)},
		&rpc.Mirror{ID: 2, StateSince: rpcTimestampPtr(200)},
	}

	if !d.Less(1, 0) {
		t.Fatalf("Expected true for larger Seconds")
	}

	if d.Less(0, 1) {
		t.Fatalf("Expected false for smaller Seconds")
	}
}

func TestCmdHelp(t *testing.T) {
	c := &cli{}
	err := c.CmdHelp()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestGetMethod(t *testing.T) {
	c := &cli{}
	method, exists := c.getMethod("help")
	if !exists {
		t.Fatalf("Expected method 'help' to exist")
	}
	_ = method

	_, exists = c.getMethod("nonexistent")
	if exists {
		t.Fatalf("Expected method 'nonexistent' to not exist")
	}
}

func TestGetMethod_Capitalization(t *testing.T) {
	c := &cli{}

	method, exists := c.getMethod("list")
	if !exists {
		t.Fatalf("Expected method 'list' to exist")
	}
	_ = method

	method, exists = c.getMethod("LIST")
	_ = method
}
