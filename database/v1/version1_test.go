// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package v1

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

func TestNewUpgraderV1(t *testing.T) {
	r := &mockRedis{conn: redigomock.NewConn()}
	u := NewUpgraderV1(r)
	if u == nil {
		t.Fatalf("Expected non-nil upgrader")
	}
	if u.Redis != r {
		t.Fatalf("Redis field not set correctly")
	}
}
