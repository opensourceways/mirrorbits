// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package cli

import (
	"testing"

	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/rpc"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestSubCmd(t *testing.T) {
	fs := SubCmd("test", "test <arg>", "test description")
	if fs == nil {
		t.Fatal("Expected non-nil FlagSet")
	}
	if fs.Name() != "test" {
		t.Fatalf("Expected test, got %s", fs.Name())
	}
}

func TestGetSingle(t *testing.T) {
	list := []*rpc.MirrorID{
		{ID: 1, Name: "m1"},
	}
	id, name, err := GetSingle(list)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if id != 1 {
		t.Fatalf("Expected 1, got %d", id)
	}
	if name != "m1" {
		t.Fatalf("Expected m1, got %s", name)
	}

	empty := []*rpc.MirrorID{}
	_, _, err = GetSingle(empty)
	if err == nil {
		t.Fatal("Expected error for empty list")
	}

	multiple := []*rpc.MirrorID{
		{ID: 1, Name: "m1"},
		{ID: 2, Name: "m2"},
	}
	_, _, err = GetSingle(multiple)
	if err == nil {
		t.Fatal("Expected error for multiple items")
	}
}

func TestCompareAndUpdate(t *testing.T) {
	m := &mirrors.Mirror{
		Name:     "test",
		HttpURL:  "http://test.com",
		Enabled:  true,
		Asnum:    123,
		Latitude: 48.85,
	}

	update := &mirrors.Mirror{
		Name:     "test",
		HttpURL:  "http://updated.com",
		Enabled:  true,
		Asnum:    123,
		Latitude: 48.85,
	}

	changed := CompareAndUpdate(m, update)
	if !changed {
		t.Fatal("Expected true for changed HttpURL")
	}
	if m.HttpURL != "http://updated.com" {
		t.Fatalf("Expected http://updated.com, got %s", m.HttpURL)
	}

	sameUpdate := &mirrors.Mirror{
		HttpURL:  "http://updated.com",
		Enabled:  true,
		Asnum:    123,
		Latitude: 48.85,
	}
	changed = CompareAndUpdate(m, sameUpdate)
	if changed {
		t.Fatal("Expected false for no changes")
	}
}

func TestByDate(t *testing.T) {
	t1 := &rpc.Mirror{StateSince: &timestamp.Timestamp{Seconds: 100}}
	t2 := &rpc.Mirror{StateSince: &timestamp.Timestamp{Seconds: 50}}
	list := []*rpc.Mirror{t1, t2}
	bd := ByDate(list)

	if bd.Len() != 2 {
		t.Fatalf("Expected 2, got %d", bd.Len())
	}
	if !bd.Less(0, 1) {
		t.Fatal("Expected Less(0,1) to be true")
	}
	bd.Swap(0, 1)
	if bd[0] != t2 {
		t.Fatal("Expected swap to change order")
	}
}

func TestGetMethod(t *testing.T) {
	c := &cli{}
	m, ok := c.getMethod("Help")
	if !ok {
		t.Fatal("Expected to find Help method")
	}
	if m.Name != "CmdHelp" {
		t.Fatalf("Expected CmdHelp, got %s", m.Name)
	}

	_, ok = c.getMethod("Nonexistent")
	if ok {
		t.Fatal("Expected false for nonexistent method")
	}
}
