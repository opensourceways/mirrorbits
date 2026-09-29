// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package database

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestRedis_auth_WithPassword(t *testing.T) {
	SetConfiguration(&Configuration{RedisPassword: "secret"})
	mock := redigomock.NewConn()
	r := &Redis{pool: &redisPoolMock{Conn: mock}}

	mock.Command("AUTH", "secret").Expect("OK")

	err := r.auth(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedis_auth_NoPassword(t *testing.T) {
	SetConfiguration(&Configuration{RedisPassword: ""})
	mock := redigomock.NewConn()
	r := &Redis{pool: &redisPoolMock{Conn: mock}}

	err := r.auth(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedis_selectDB(t *testing.T) {
	SetConfiguration(&Configuration{RedisDB: 3})
	mock := redigomock.NewConn()
	r := &Redis{pool: &redisPoolMock{Conn: mock}}

	mock.Command("SELECT", 3).Expect("OK")

	err := r.selectDB(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedis_askRole(t *testing.T) {
	mock := redigomock.NewConn()
	r := &Redis{pool: &redisPoolMock{Conn: mock}}

	mock.Command("ROLE").Expect([]interface{}{[]byte("master")})

	role, err := r.askRole(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if role != "master" {
		t.Fatalf("Expected 'master', got %s", role)
	}
}

func TestRedis_askRole_Error(t *testing.T) {
	mock := redigomock.NewConn()
	r := &Redis{pool: &redisPoolMock{Conn: mock}}

	mock.Command("ROLE").ExpectError(redis.Error("ERR bad"))

	_, err := r.askRole(mock)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestRedis_setFailureState(t *testing.T) {
	r := &Redis{}

	r.setFailureState(true)
	if !r.Failure() {
		t.Fatalf("Expected failure=true")
	}

	r.setFailureState(false)
	if r.Failure() {
		t.Fatalf("Expected failure=false")
	}
}

func TestRedis_logError_Failure(t *testing.T) {
	SetConfiguration(&Configuration{RedisAddress: ""})
	r := &Redis{}
	r.setFailureState(true)

	r.logError("test %s", "msg")
}

func TestRedis_logError_NoFailure(t *testing.T) {
	SetConfiguration(&Configuration{RedisAddress: ""})
	r := &Redis{}
	r.setFailureState(false)

	r.logError("test %s", "msg")
}

func TestRedis_printConnectedMaster(t *testing.T) {
	r := &Redis{}

	r.printConnectedMaster("host1:6379")
	r.printConnectedMaster("host1:6379")
	r.printConnectedMaster("host2:6379")
}

func TestRedis_Upgrade_AlreadyCurrent(t *testing.T) {
	SetConfiguration(&Configuration{RedisDB: 0, RedisAddress: ""})
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(1))

	err := conn.Upgrade()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedis_Upgrade_UnsupportedVersion(t *testing.T) {
	SetConfiguration(&Configuration{RedisDB: 0, RedisAddress: ""})
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(99))

	err := conn.Upgrade()
	if err != ErrUnsupportedVersion {
		t.Fatalf("Expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestRedis_Upgrade_Locked(t *testing.T) {
	SetConfiguration(&Configuration{RedisDB: 0, RedisAddress: ""})
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(0))
	mock.Command("SET", "LOCK_upgrade", redigomock.NewAnyData(), "NX", "PX", "5000").ExpectError(redis.ErrNil)

	err := conn.Upgrade()
	if err != ErrAlreadyLocked {
		t.Fatalf("Expected ErrAlreadyLocked, got %v", err)
	}
}

func TestLock_isValid_Owner(t *testing.T) {
	mock, conn := prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "myvalue",
		held:  true,
	}

	mock.Command("GET", "LOCK_test").Expect("myvalue")

	valid, err := lock.isValid()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !valid {
		t.Fatalf("Expected valid=true")
	}
}

func TestLock_isValid_NotOwner(t *testing.T) {
	mock, conn := prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "myvalue",
		held:  true,
	}

	mock.Command("GET", "LOCK_test").Expect("othervalue")

	valid, err := lock.isValid()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if valid {
		t.Fatalf("Expected valid=false")
	}
}

func TestLock_isValid_ErrNil(t *testing.T) {
	mock, conn := prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "myvalue",
		held:  true,
	}

	mock.Command("GET", "LOCK_test").ExpectError(redis.ErrNil)

	valid, err := lock.isValid()
	if err != nil {
		t.Fatalf("Unexpected error for ErrNil: %s", err)
	}
	if valid {
		t.Fatalf("Expected valid=false for ErrNil")
	}
}

func TestLock_Release_Held_Owner(t *testing.T) {
	mock, conn := prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "myvalue",
		held:  true,
	}

	mock.Command("GET", "LOCK_test").Expect("myvalue")
	mock.Command("DEL", "LOCK_test").Expect(int64(1))

	lock.Release()
	if lock.Held() {
		t.Fatalf("Expected lock released")
	}
}

func TestLock_Release_Held_NotOwner(t *testing.T) {
	mock, conn := prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "myvalue",
		held:  true,
	}

	mock.Command("GET", "LOCK_test").Expect("othervalue")

	lock.Release()
	if lock.Held() {
		t.Fatalf("Expected lock released")
	}
}

func TestLock_Held(t *testing.T) {
	lock := &Lock{held: true}
	if !lock.Held() {
		t.Fatalf("Expected true")
	}

	lock.held = false
	if lock.Held() {
		t.Fatalf("Expected false")
	}
}

func TestRedis_UnblockedGet(t *testing.T) {
	_, conn := prepareTestRedis()

	c := conn.UnblockedGet()
	if c == nil {
		t.Fatalf("Expected non-nil conn")
	}
	c.Close()
}

func TestRedis_Get_Ready(t *testing.T) {
	_, conn := prepareTestRedis()

	c := conn.Get()
	if c == nil {
		t.Fatalf("Expected non-nil conn")
	}
	c.Close()
}
