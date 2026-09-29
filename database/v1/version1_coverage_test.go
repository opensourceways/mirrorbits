// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package v1

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func TestRenameKeys_WithFiles(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("SMEMBERS", "MIRROR_mirror1_FILES").Expect([]interface{}{
		[]byte("file1.iso"),
		[]byte("file2.iso"),
	})
	mock.Command("SMEMBERS", "FILES").Expect([]interface{}{
		[]byte("file1.iso"),
		[]byte("file2.iso"),
	})
	mock.Command("SMEMBERS", "FILEMIRRORS_file1.iso").Expect([]interface{}{
		[]byte("mirror1"),
	})
	mock.Command("SMEMBERS", "FILEMIRRORS_file2.iso").Expect([]interface{}{
		[]byte("mirror1"),
	})
	mock.Command("SADD", "V1_FILEMIRRORS_file1.iso", 1).Expect("OK")
	mock.Command("SADD", "V1_FILEMIRRORS_file2.iso", 1).Expect("OK")
	mock.Command("FLUSH").Expect("OK")

	err := v.RenameKeys(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(a.rename) == 0 {
		t.Fatalf("Expected rename map to have entries")
	}
}

func TestRenameKeys_SMembersError(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("SMEMBERS", "MIRROR_mirror1_FILES").ExpectError(redis.Error("ERR connection"))

	err := v.RenameKeys(a, m)
	if err == nil {
		t.Fatalf("Expected error for SMEMBERS failure")
	}
}

func TestRenameKeys_FilesSMembersError(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("SMEMBERS", "MIRROR_mirror1_FILES").ExpectError(redis.ErrNil)
	mock.Command("SMEMBERS", "FILES").ExpectError(redis.Error("ERR connection"))

	err := v.RenameKeys(a, m)
	if err == nil {
		t.Fatalf("Expected error for FILES SMEMBERS failure")
	}
}

func TestRenameKeys_FileMirrorsError(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("SMEMBERS", "MIRROR_mirror1_FILES").ExpectError(redis.ErrNil)
	mock.Command("SMEMBERS", "FILES").Expect([]interface{}{
		[]byte("file1.iso"),
	})
	mock.Command("SMEMBERS", "FILEMIRRORS_file1.iso").ExpectError(redis.Error("ERR connection"))

	err := v.RenameKeys(a, m)
	if err == nil {
		t.Fatalf("Expected error for FILEMIRRORS SMEMBERS failure")
	}
}

func TestCreateMirrorIndex_WithMirrors(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}

	mock.Command("LRANGE", "MIRRORS", "0", "-1").Expect([]interface{}{
		[]byte("mirror1"),
	})
	mock.Command("INCR", "LAST_MID").Expect(int64(1))
	mock.Command("HSET", "V1_MIRRORS", 1, "mirror1").Expect(int64(1))

	m, err := v.CreateMirrorIndex(a)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(m) != 1 {
		t.Fatalf("Expected 1 mirror, got %d", len(m))
	}
}

func TestCreateMirrorIndex_Error(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}

	mock.Command("LRANGE", "MIRRORS", "0", "-1").ExpectError(redis.Error("ERR connection"))

	_, err := v.CreateMirrorIndex(a)
	if err == nil {
		t.Fatalf("Expected error for LRANGE failure")
	}
}

func TestCreateMirrorIndex_INCRError(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}

	mock.Command("LRANGE", "MIRRORS", "0", "-1").Expect([]interface{}{
		[]byte("mirror1"),
	})
	mock.Command("INCR", "LAST_MID").ExpectError(redis.Error("ERR connection"))

	_, err := v.CreateMirrorIndex(a)
	if err == nil {
		t.Fatalf("Expected error for INCR failure")
	}
}

func TestFixMirrorID_CopyKey(t *testing.T) {
	mock := redigomock.NewConn()
	r := &mockRedis{conn: mock}
	v := NewUpgraderV1(r)

	a := &actions{rename: make(map[string]string)}
	m := map[int]string{1: "mirror1"}

	mock.Command("DUMP", "MIRROR_mirror1").Expect([]byte("dumped"))
	mock.Command("RESTORE", "V1_MIRROR_1", redigomock.NewAnyInt(), redigomock.NewAnyData(), "REPLACE").Expect("OK")
	mock.Command("HMSET", "V1_MIRROR_1", "ID", 1, "name", "mirror1").Expect("OK")
	mock.Command("FLUSH").Expect("OK")

	err := v.FixMirrorID(a, m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(a.rename) == 0 {
		t.Fatalf("Expected rename entries")
	}
}
