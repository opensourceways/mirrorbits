// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package cli

import (
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/rpc"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestCmdHelpReturns(t *testing.T) {
	c := &cli{}
	err := c.CmdHelp()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestParseCommandsHelp(t *testing.T) {
	err := ParseCommands("help")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestParseCommandsUnknown(t *testing.T) {
	err := ParseCommands("nonexistentcommand")
	_ = err
}

func TestParseCommandsEmpty(t *testing.T) {
	err := ParseCommands()
	_ = err
}

func TestSubCmdWithArgs(t *testing.T) {
	fs := SubCmd("add", "<name>", "Add a new mirror")
	if fs == nil {
		t.Fatal("Expected non-nil FlagSet")
	}
}

func TestCompareAndUpdateMultipleFields(t *testing.T) {
	m := &mirrors.Mirror{
		Name:            "test",
		HttpURL:         "http://old.com",
		Enabled:         true,
		Latitude:        48.85,
		ContinentCode:   "EU",
		CountryCodes:    "FR",
		Score:           5,
		Comment:         "old comment",
	}

	update := &mirrors.Mirror{
		Name:            "test",
		HttpURL:         "http://new.com",
		Enabled:         false,
		Latitude:        35.68,
		ContinentCode:   "AS",
		CountryCodes:    "JP",
		Score:           10,
		Comment:         "new comment",
	}

	changed := CompareAndUpdate(m, update)
	if !changed {
		t.Fatal("Expected true for multiple changes")
	}
	if m.HttpURL != "http://new.com" {
		t.Fatal("HttpURL not updated")
	}
	if m.Enabled {
		t.Fatal("Enabled not updated")
	}
	if m.Latitude != 35.68 {
		t.Fatal("Latitude not updated")
	}
	if m.Score != 10 {
		t.Fatal("Score not updated")
	}
}

func TestCompareAndUpdateOnlyHttpURL(t *testing.T) {
	m := &mirrors.Mirror{
		Name:     "test",
		HttpURL:  "http://old.com",
		Enabled:  true,
	}

	update := &mirrors.Mirror{
		Name:    "test",
		HttpURL: "http://new.com",
		Enabled: true,
	}

	changed := CompareAndUpdate(m, update)
	if !changed {
		t.Fatal("Expected true for changed HttpURL")
	}
	if m.HttpURL != "http://new.com" {
		t.Fatalf("Expected http://new.com, got %s", m.HttpURL)
	}
}

func TestGetSingleMultiple(t *testing.T) {
	list := []*rpc.MirrorID{
		{ID: 1, Name: "m1"},
		{ID: 2, Name: "m2"},
		{ID: 3, Name: "m3"},
	}
	id, name, err := GetSingle(list)
	if err == nil {
		t.Fatal("Expected error for multiple items")
	}
	if id != -1 {
		t.Fatalf("Expected -1, got %d", id)
	}
	if name != "" {
		t.Fatalf("Expected empty, got %s", name)
	}
}

func TestDefaultRPCTimeout(t *testing.T) {
	if defaultRPCTimeout != time.Second*60 {
		t.Fatal("Expected 60 second timeout")
	}
}

func TestCommentSeparator(t *testing.T) {
	if commentSeparator == "" {
		t.Fatal("Expected non-empty separator")
	}
}

func TestByDateLessEqual(t *testing.T) {
	t1 := &rpc.Mirror{StateSince: &timestamp.Timestamp{Seconds: 100}}
	t2 := &rpc.Mirror{StateSince: &timestamp.Timestamp{Seconds: 100}}
	bd := ByDate([]*rpc.Mirror{t1, t2})
	if bd.Less(0, 1) {
		t.Fatal("Expected false for equal timestamps")
	}
}

func TestGetMethodLowercase(t *testing.T) {
	c := &cli{}
	m, ok := c.getMethod("help")
	if !ok {
		t.Fatal("Expected to find help method")
	}
	if m.Name != "CmdHelp" {
		t.Fatalf("Expected CmdHelp, got %s", m.Name)
	}
}

func TestGetMethodUppercase(t *testing.T) {
	c := &cli{}
	_, ok := c.getMethod("HELP")
	if !ok {
		t.Fatal("Expected to find HELP method")
	}
}
