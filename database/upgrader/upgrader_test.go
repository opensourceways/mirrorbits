// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package upgrader

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type mockRedis struct {
	conn *redigomock.Conn
}

func (m *mockRedis) Get() redis.Conn { return m.conn }
func (m *mockRedis) UnblockedGet() redis.Conn { return m.conn }

func newMockRedis() *mockRedis {
	return &mockRedis{conn: redigomock.NewConn()}
}

func TestGetUpgrader_Version1(t *testing.T) {
	r := newMockRedis()
	u := GetUpgrader(r, 1)
	if u == nil {
		t.Fatalf("Expected non-nil upgrader for version 1")
	}
}

func TestGetUpgrader_UnknownVersion(t *testing.T) {
	r := newMockRedis()
	u := GetUpgrader(r, 99)
	if u != nil {
		t.Fatalf("Expected nil for unknown version")
	}
}

func TestGetUpgrader_Version0(t *testing.T) {
	r := newMockRedis()
	u := GetUpgrader(r, 0)
	if u != nil {
		t.Fatalf("Expected nil for version 0")
	}
}
