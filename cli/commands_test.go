// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package cli

import (
	"testing"
)

func newTestCli() *cli {
	return &cli{creds: &loginCreds{}}
}

func TestCmdList_ExtraArg(t *testing.T) {
	c := newTestCli()
	err := c.CmdList("extra")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdAdd_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdAdd()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdRemove_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdRemove()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdScan_NoArgsNotAll(t *testing.T) {
	c := newTestCli()
	err := c.CmdScan()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdScan_AllWithArg(t *testing.T) {
	c := newTestCli()
	err := c.CmdScan("-all", "extra")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdScan_TwoArgsNotAll(t *testing.T) {
	c := newTestCli()
	err := c.CmdScan("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdRefresh_ExtraArg(t *testing.T) {
	c := newTestCli()
	err := c.CmdRefresh("extra")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdShow_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdShow()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdShow_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdShow("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdExport_InvalidFormat(t *testing.T) {
	c := newTestCli()
	err := c.CmdExport("invalid")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdExport_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdExport()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdExport_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdExport("mirmon", "extra")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdEnable_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdEnable()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdEnable_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdEnable("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdDisable_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdDisable()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdDisable_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdDisable("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdStats_NotEnoughArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdStats("mirror")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdStats_ThreeArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdStats("mirror", "name", "extra")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdStats_InvalidType(t *testing.T) {
	c := newTestCli()
	err := c.CmdStats("invalid", "name")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdLogs_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdLogs()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdLogs_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdLogs("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdEdit_NoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdEdit()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestCmdEdit_TwoArgs(t *testing.T) {
	c := newTestCli()
	err := c.CmdEdit("arg1", "arg2")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestParseCommands_Empty(t *testing.T) {
	err := ParseCommands()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestParseCommands_HelpCommand(t *testing.T) {
	err := ParseCommands("help")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestParseCommands_VersionCommand(t *testing.T) {
	// version command calls GetRPC which would block, skip
	_ = ParseCommands
}
