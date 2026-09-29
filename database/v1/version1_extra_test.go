// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package v1

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type testPoolV1 struct {
	Conn *redigomock.Conn
}

func (r *testPoolV1) Get() redis.Conn {
	return r.Conn
}

func (r *testPoolV1) Close() error {
	return nil
}

func TestIsErrNoSuchKeyExtra(t *testing.T) {
	if !IsErrNoSuchKey(errors.New("ERR no such key")) {
		t.Fatal("Expected true")
	}
	if IsErrNoSuchKey(errors.New("other error")) {
		t.Fatal("Expected false")
	}
	if IsErrNoSuchKey(nil) {
		t.Fatal("Expected false for nil")
	}
}

func TestNewUpgraderV1Extra(t *testing.T) {
	v := NewUpgraderV1(nil)
	if v == nil {
		t.Fatal("Expected non-nil")
	}
}

func TestCopyKeyExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("DUMP", "src").Expect("dumpdata")
	mock.Command("RESTORE", "dst", 0, "dumpdata", "REPLACE").Expect("OK")

	err := CopyKey(mock, "src", "dst")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCopyKeyErrorExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("DUMP", "src").ExpectError(errors.New("not found"))

	err := CopyKey(mock, "src", "dst")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestActionsStructExtra(t *testing.T) {
	a := &actions{
		delete: []string{"key1", "key2"},
		rename: map[string]string{"old": "new"},
	}
	if len(a.delete) != 2 {
		t.Fatal("Expected 2 deletes")
	}
	if a.rename["old"] != "new" {
		t.Fatal("Rename mismatch")
	}
}
