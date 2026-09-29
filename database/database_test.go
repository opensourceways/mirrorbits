// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package database

import (
	"errors"
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"

	. "github.com/opensourceways/mirrorbits/config"
)

type redisPoolMock struct {
	Conn *redigomock.Conn
}

func (r *redisPoolMock) Get() redis.Conn {
	return r.Conn
}

func (r *redisPoolMock) Close() error {
	return nil
}

func prepareTestRedis() (*redigomock.Conn, *Redis) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	return mock, NewRedisCustomPool(pool)
}

func TestMain(m *testing.M) {
	SetConfiguration(&Configuration{
		RedisDB:      0,
		RedisAddress: "",
	})
	os.Exit(m.Run())
}

func TestRedisIsLoading(t *testing.T) {
	err := errors.New("LOADING Redis is loading the dataset in memory")
	if !RedisIsLoading(err) {
		t.Fatalf("Expected true for LOADING error")
	}

	err = errors.New("some other error")
	if RedisIsLoading(err) {
		t.Fatalf("Expected false for non-LOADING error")
	}

	if RedisIsLoading(nil) {
		t.Fatalf("Expected false for nil error")
	}
}

func TestParseVersion(t *testing.T) {
	if parseVersion("3.2.0") <= 0 {
		t.Fatalf("Expected positive version for 3.2.0")
	}

	if parseVersion("4.0.0") <= parseVersion("3.2.0") {
		t.Fatalf("Expected 4.0.0 > 3.2.0")
	}

	if parseVersion("3.2.10") <= parseVersion("3.2.0") {
		t.Fatalf("Expected 3.2.10 > 3.2.0")
	}

	if parseVersion("invalid") != -1 {
		t.Fatalf("Expected -1 for invalid version")
	}
}

func TestParseInfo(t *testing.T) {
	info := "redis_version:3.2.0\r\n# Comment\r\nredis_mode:standalone\r\n"
	m, err := parseInfo(info, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if m["redis_version"] != "3.2.0" {
		t.Fatalf("Expected redis_version 3.2.0, got %s", m["redis_version"])
	}
	if m["redis_mode"] != "standalone" {
		t.Fatalf("Expected redis_mode standalone, got %s", m["redis_mode"])
	}
}

func TestParseInfo_Error(t *testing.T) {
	_, err := parseInfo(nil, errors.New("test error"))
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestParseInfo_EmptyLines(t *testing.T) {
	info := "\r\n\r\nredis_version:4.0.0\r\n"
	m, err := parseInfo(info, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if m["redis_version"] != "4.0.0" {
		t.Fatalf("Expected redis_version 4.0.0, got %s", m["redis_version"])
	}
}

func TestNewNetTemporaryError(t *testing.T) {
	err := NewNetTemporaryError()
	if err.Timeout() != false {
		t.Fatalf("Expected Timeout false")
	}
	if err.Temporary() != true {
		t.Fatalf("Expected Temporary true")
	}
	if err.Error() != "database not ready" {
		t.Fatalf("Expected 'database not ready', got %s", err.Error())
	}
}

func TestNotReadyError_Close(t *testing.T) {
	e := &NotReadyError{}
	err := e.Close()
	if err == nil {
		t.Fatalf("Expected error from Close")
	}
}

func TestNotReadyError_Err(t *testing.T) {
	e := &NotReadyError{}
	err := e.Err()
	if err == nil {
		t.Fatalf("Expected error from Err")
	}
}

func TestNotReadyError_Do(t *testing.T) {
	e := &NotReadyError{}
	_, err := e.Do("GET", "key")
	if err == nil {
		t.Fatalf("Expected error from Do")
	}
}

func TestNotReadyError_Send(t *testing.T) {
	e := &NotReadyError{}
	err := e.Send("GET", "key")
	if err == nil {
		t.Fatalf("Expected error from Send")
	}
}

func TestNotReadyError_Flush(t *testing.T) {
	e := &NotReadyError{}
	err := e.Flush()
	if err == nil {
		t.Fatalf("Expected error from Flush")
	}
}

func TestNotReadyError_Receive(t *testing.T) {
	e := &NotReadyError{}
	_, err := e.Receive()
	if err == nil {
		t.Fatalf("Expected error from Receive")
	}
}

func TestGetListOfMirrors_Unreachable(t *testing.T) {
	_, conn := prepareTestRedis()

	// RedisAddress is empty so Connect() returns ErrUnreachable
	_, err := conn.GetListOfMirrors()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestPublish(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	cmd := mock.Command("PUBLISH", string(MIRROR_UPDATE), "1").Expect("1")

	err := Publish(conn.Get(), MIRROR_UPDATE, "1")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if mock.Stats(cmd) < 1 {
		t.Fatalf("PUBLISH not executed")
	}
}

func TestSendPublish(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	mock.Command("PUBLISH", string(FILE_UPDATE), "file1")

	err := SendPublish(conn.Get(), FILE_UPDATE, "file1")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestNewPubsub(t *testing.T) {
	_, conn := 	prepareTestRedis()
	p := NewPubsub(conn)
	if p == nil {
		t.Fatalf("Expected non-nil Pubsub")
	}
	p.Close()
}

func TestPubsub_SubscribeEvent(t *testing.T) {
	_, conn := 	prepareTestRedis()
	p := NewPubsub(conn)
	defer p.Close()

	ch := make(chan string, 1)
	p.SubscribeEvent(MIRROR_UPDATE, ch)

	p.extSubscribersLock.RLock()
	defer p.extSubscribersLock.RUnlock()
	if len(p.extSubscribers[string(MIRROR_UPDATE)]) != 1 {
		t.Fatalf("Expected 1 subscriber")
	}
}

func TestPubsub_HandleMessage(t *testing.T) {
	_, conn := 	prepareTestRedis()
	p := NewPubsub(conn)
	defer p.Close()

	ch := make(chan string, 1)
	p.SubscribeEvent(MIRROR_UPDATE, ch)

	p.handleMessage(string(MIRROR_UPDATE), []byte("test-data"))

	select {
	case msg := <-ch:
		if msg != "test-data" {
			t.Fatalf("Expected 'test-data', got %s", msg)
		}
	default:
		t.Fatalf("Expected to receive message")
	}
}

func TestPubsub_HandleMessage_NoSubscriber(t *testing.T) {
	_, conn := 	prepareTestRedis()
	p := NewPubsub(conn)
	defer p.Close()

	p.handleMessage("nonexistent", []byte("data"))
}

func TestRedis_Get_NotReady(t *testing.T) {
	r := NewRedisCustomPool(nil)
	// With nil pool, pool.Get() will panic, but before that,
	// the ready channel is not closed (since pool is nil but not mock),
	// so it should return NotReadyError
	conn := r.Get()
	if conn == nil {
		t.Fatalf("Expected non-nil conn")
	}
	_, err := conn.Do("PING")
	if err == nil {
		t.Fatalf("Expected error from NotReadyError")
	}
}

func TestRedis_Failure(t *testing.T) {
	_, conn := 	prepareTestRedis()

	if conn.Failure() != false {
		t.Fatalf("Expected false for initial failure state")
	}
}

func TestRedis_CheckVersion(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("INFO", "server").Expect("redis_version:5.0.0\r\n")

	err := conn.CheckVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedis_CheckVersion_Unsupported(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("INFO", "server").Expect("redis_version:2.0.0\r\n")

	err := conn.CheckVersion()
	if err != ErrRedisUpgradeRequired {
		t.Fatalf("Expected ErrRedisUpgradeRequired, got %v", err)
	}
}

func TestRedis_GetDBFormatVersion(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(1))

	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

func TestRedis_GetDBFormatVersion_NotSet(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").ExpectError(redis.ErrNil)
	mock.Command("EXISTS", "MIRRORS").Expect(int64(0))
	mock.Command("SET", "MIRRORBITS_DB_VERSION", 1).Expect("OK")

	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

func TestRedis_GetDBFormatVersion_NotSetButHasMirrors(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").ExpectError(redis.ErrNil)
	mock.Command("EXISTS", "MIRRORS").Expect(int64(1))

	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0, got %d", version)
	}
}

func TestRedis_UpgradeNeeded(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(1))

	needed, err := conn.UpgradeNeeded()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if needed {
		t.Fatalf("Expected false for current version")
	}
}

func TestRedis_UpgradeNeeded_Needed(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(0))

	needed, err := conn.UpgradeNeeded()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !needed {
		t.Fatalf("Expected true for old version")
	}
}

func TestRedis_UpgradeNeeded_Unsupported(t *testing.T) {
	mock, conn := prepareTestRedis()

	mock.Command("GET", "MIRRORBITS_DB_VERSION").Expect(int64(99))

	_, err := conn.UpgradeNeeded()
	if err != ErrUnsupportedVersion {
		t.Fatalf("Expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestRedis_ConnectPubsub(t *testing.T) {
	_, conn := 	prepareTestRedis()

	conn.ConnectPubsub()
	if conn.Pubsub == nil {
		t.Fatalf("Expected non-nil Pubsub after ConnectPubsub")
	}

	conn.ConnectPubsub()
}

func TestRedis_Close(t *testing.T) {
	_, conn := 	prepareTestRedis()
	conn.ConnectPubsub()
	conn.Close()
}

func TestAcquireLock_InvalidName(t *testing.T) {
	_, conn := 	prepareTestRedis()

	_, err := conn.AcquireLock("")
	if err != ErrInvalidLockName {
		t.Fatalf("Expected ErrInvalidLockName, got %v", err)
	}
}

func TestAcquireLock_AlreadyLocked(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	mock.Command("SET", "LOCK_test", redigomock.NewAnyData(), "NX", "PX", "5000").ExpectError(redis.ErrNil)

	_, err := conn.AcquireLock("test")
	if err != ErrAlreadyLocked {
		t.Fatalf("Expected ErrAlreadyLocked, got %v", err)
	}
}

func TestAcquireLock_Success(t *testing.T) {
	mock, conn := 	prepareTestRedis()

	mock.Command("SET", "LOCK_test", redigomock.NewAnyData(), "NX", "PX", "5000").Expect("OK")

	lock, err := conn.AcquireLock("test")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if lock == nil {
		t.Fatalf("Expected non-nil lock")
	}
	if !lock.Held() {
		t.Fatalf("Expected lock to be held")
	}
}

func TestLock_Release_NotHeld(t *testing.T) {
	_, conn := 	prepareTestRedis()

	lock := &Lock{
		redis: conn,
		name:  "LOCK_test",
		value: "test",
		held:  false,
	}
	lock.Release()
}
