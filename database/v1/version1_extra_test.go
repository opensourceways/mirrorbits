// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package v1

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func TestIsErrNoSuchKey(t *testing.T) {
	if !IsErrNoSuchKey(redis.Error("ERR no such key")) {
		t.Fatalf("Expected true for 'ERR no such key'")
	}
	if IsErrNoSuchKey(redis.Error("ERR other")) {
		t.Fatalf("Expected false for other error")
	}
	if IsErrNoSuchKey(nil) {
		t.Fatalf("Expected false for nil")
	}
}

func TestCopyKey_Success(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}

	mock.Command("DUMP", "src").Expect("dumpdata")
	mock.Command("RESTORE", "dst", 0, "dumpdata", "REPLACE").Expect("OK")

	err := CopyKey(r.Get(), "src", "dst")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCopyKey_DumpError(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}

	mock.Command("DUMP", "src").ExpectError(redis.Error("ERR dump fail"))

	err := CopyKey(r.Get(), "src", "dst")
	if err == nil {
		t.Fatalf("Expected error from DUMP")
	}
}

func TestCreateMirrorIndex(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}

	mock.Command("LRANGE", "MIRRORS", "0", "-1").Expect([]interface{}{
		[]byte("mirror1"),
		[]byte("mirror2"),
	})
	mock.Command("INCR", "LAST_MID").Expect(int64(1)).Expect(int64(2))
	mock.Command("HSET", "V1_MIRRORS", 1, "mirror1").Expect(int64(1))
	mock.Command("HSET", "V1_MIRRORS", 2, "mirror2").Expect(int64(1))

	m, err := v.CreateMirrorIndex(a)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(m) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(m))
	}
	if m[1] != "mirror1" {
		t.Fatalf("Expected mirror1 at id 1")
	}
	if a.rename["V1_MIRRORS"] != "MIRRORS" {
		t.Fatalf("Expected rename V1_MIRRORS -> MIRRORS")
	}
}

func TestCreateMirrorIndex_Empty(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}

	mock.Command("LRANGE", "MIRRORS", "0", "-1").Expect([]interface{}{})

	m, err := v.CreateMirrorIndex(a)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(m) != 0 {
		t.Fatalf("Expected 0 mirrors, got %d", len(m))
	}
}

func TestRenameKeys_NoFiles(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("SMEMBERS", "MIRROR_mirror1_FILES").ExpectError(redis.ErrNil)
	mock.Command("SMEMBERS", "FILES").ExpectError(redis.ErrNil)

	err := v.RenameKeys(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestFixMirrorID(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("DUMP", "MIRROR_mirror1").Expect("dump")
	mock.Command("RESTORE", "V1_MIRROR_1", 0, "dump", "REPLACE").Expect("OK")
	mock.Command("HMSET", "V1_MIRROR_1", "ID", 1, "name", "mirror1").Expect("OK")

	err := v.FixMirrorID(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if a.rename["V1_MIRROR_1"] != "MIRROR_1" {
		t.Fatalf("Expected rename V1_MIRROR_1 -> MIRROR_1")
	}
}

func TestRenameStats_NoKeys(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("KEYS", "STATS_MIRROR_*").ExpectError(redis.ErrNil)

	err := v.RenameStats(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRenameStats_WithKeys(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("KEYS", "STATS_MIRROR_*").Expect([]interface{}{
		[]byte("STATS_MIRROR_2020"),
	})
	mock.Command("HGETALL", "STATS_MIRROR_2020").ExpectMap(map[string]string{
		"mirror1": "100",
	})
	mock.Command("HSET", "V1_STATS_MIRROR_2020", 1, "100").Expect(int64(1))

	err := v.RenameStats(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestUpgrade_Full(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	mock.Command("EVAL", redigomock.NewAnyData(), 0, "V1_*").Expect([]interface{}{})
	mock.Command("LRANGE", "MIRRORS", "0", "-1").Expect([]interface{}{})
	mock.Command("SMEMBERS", "FILES").ExpectError(redis.ErrNil)
	mock.Command("KEYS", "STATS_MIRROR_*").ExpectError(redis.ErrNil)
	mock.Command("MULTI").Expect("OK")
	mock.Command("RENAME", "V1_MIRRORS", "MIRRORS").Expect("OK")
	mock.Command("SET", "MIRRORBITS_DB_VERSION", 1).Expect("OK")
	mock.Command("EXEC").Expect("OK")

	err := v.Upgrade()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
