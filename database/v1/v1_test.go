// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package v1

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type mockPool struct {
	Conn *redigomock.Conn
}

func (r *mockPool) Get() redis.Conn {
	return r.Conn
}

func (r *mockPool) Close() error {
	return nil
}

func TestIsErrNoSuchKey(t *testing.T) {
	if !IsErrNoSuchKey(errors.New("ERR no such key")) {
		t.Fatal("Expected true for 'ERR no such key'")
	}
	if IsErrNoSuchKey(errors.New("some other error")) {
		t.Fatal("Expected false for other errors")
	}
	if IsErrNoSuchKey(nil) {
		t.Fatal("Expected false for nil")
	}
}

func TestNewUpgraderV1(t *testing.T) {
	// Just verify it doesn't panic
	u := NewUpgraderV1(nil)
	if u == nil {
		t.Fatal("Expected non-nil upgrader")
	}
}

func TestVersion1Upgrade(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Clear()
	mock.GenericCommand("EVAL").Expect([]interface{}{})
	mock.GenericCommand("LRANGE").Expect([]interface{}{})
	mock.GenericCommand("SMEMBERS").ExpectError(redis.ErrNil)
	mock.GenericCommand("KEYS").ExpectError(redis.ErrNil)
	mock.GenericCommand("MULTI")
	mock.GenericCommand("SET").Expect("OK")
	mock.GenericCommand("EXEC").Expect("OK")
	mock.GenericCommand("DEL").Expect(int64(1))
	mock.GenericCommand("RENAME").Expect("OK")
	mock.GenericCommand("DUMP").Expect("")
	mock.GenericCommand("RESTORE").Expect("OK")
	mock.GenericCommand("HMSET").Expect("OK")
	mock.GenericCommand("HSET").Expect(int64(1))
	mock.GenericCommand("HGETALL").ExpectMap(map[string]string{})
	mock.GenericCommand("INCR").Expect(int64(1))
	mock.GenericCommand("SADD").Expect(int64(1))
	mock.GenericCommand("FLUSH").Expect("OK")

	r := &mockRedis{conn: mock}
	u := &Version1{Redis: r}
	err := u.Upgrade()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCopyKey(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("DUMP", "src").Expect("dumped")
	mock.Command("RESTORE", "dst", 0, "dumped", "REPLACE").Expect("OK")
	r := &mockRedis{conn: mock}
	err := CopyKey(r.UnblockedGet(), "src", "dst")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCopyKeyError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("DUMP", "src").ExpectError(errors.New("dump error"))
	r := &mockRedis{conn: mock}
	err := CopyKey(r.UnblockedGet(), "src", "dst")
	if err == nil {
		t.Fatal("Expected error")
	}
}

type mockRedis struct {
	conn *redigomock.Conn
}

func (m *mockRedis) Get() redis.Conn {
	return m.conn
}

func (m *mockRedis) UnblockedGet() redis.Conn {
	return m.conn
}
