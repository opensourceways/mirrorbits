// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package database

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/rafaeljusto/redigomock"
)

type testPool struct {
	Conn *redigomock.Conn
}

func (r *testPool) Get() redis.Conn {
	return r.Conn
}

func (r *testPool) Close() error {
	return nil
}

func prepareTest() (*redigomock.Conn, *Redis) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	conn := NewRedisCustomPool(pool)
	return mock, conn
}

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"3.2.0", 30200},
		{"4.0.0", 40000},
		{"5.0.14", 50014},
		{"invalid", -1},
		{"", 0},
	}
	for _, tt := range tests {
		if got := parseVersion(tt.input); got != tt.expected {
			t.Fatalf("parseVersion(%q): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestParseInfo(t *testing.T) {
	info := "redis_version:5.0.14\r\nredis_mode:standalone\r\n# Server\r\nrun_id:abc123\r\n"
	m, err := parseInfo(info, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if m["redis_version"] != "5.0.14" {
		t.Fatalf("Expected 5.0.14, got %s", m["redis_version"])
	}
	if m["redis_mode"] != "standalone" {
		t.Fatalf("Expected standalone, got %s", m["redis_mode"])
	}
	if _, ok := m["# Server"]; ok {
		t.Fatal("Expected comment line to be skipped")
	}
}

func TestParseInfoError(t *testing.T) {
	_, err := parseInfo(nil, errors.New("test error"))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRedisIsLoading(t *testing.T) {
	if !RedisIsLoading(errors.New("LOADING Redis is loading the dataset in memory")) {
		t.Fatal("Expected true for LOADING error")
	}
	if RedisIsLoading(errors.New("some other error")) {
		t.Fatal("Expected false for non-LOADING error")
	}
	if RedisIsLoading(nil) {
		t.Fatal("Expected false for nil error")
	}
}

func TestNotReadyError(t *testing.T) {
	var e NotReadyError
	if e.Err() == nil {
		t.Fatal("Expected non-nil error")
	}
	if _, err := e.Do("GET", "key"); err == nil {
		t.Fatal("Expected error from Do")
	}
	if err := e.Send("GET", "key"); err == nil {
		t.Fatal("Expected error from Send")
	}
	if err := e.Flush(); err == nil {
		t.Fatal("Expected error from Flush")
	}
	if _, err := e.Receive(); err == nil {
		t.Fatal("Expected error from Receive")
	}
	if err := e.Close(); err == nil {
		t.Fatal("Expected error from Close")
	}
}

func TestNetReadyError(t *testing.T) {
	e := NewNetTemporaryError()
	if e.Timeout() != false {
		t.Fatal("Expected Timeout to be false")
	}
	if e.Temporary() != true {
		t.Fatal("Expected Temporary to be true")
	}
	if e.Error() != "database not ready" {
		t.Fatalf("Expected 'database not ready', got %s", e.Error())
	}
}

func TestNewRedisCustomPool(t *testing.T) {
	r := NewRedisCustomPool(nil)
	if r == nil {
		t.Fatal("Expected non-nil Redis")
	}
}

func TestRedisGetWithMock(t *testing.T) {
	_, conn := prepareTest()
	c := conn.Get()
	if c == nil {
		t.Fatal("Expected non-nil connection")
	}
}

func TestRedisUnblockedGetWithMock(t *testing.T) {
	_, conn := prepareTest()
	c := conn.UnblockedGet()
	if c == nil {
		t.Fatal("Expected non-nil connection")
	}
}

func TestRedisCheckVersion(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("INFO", "server").Expect("redis_version:5.0.14\r\n")
	err := conn.CheckVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestRedisCheckVersionUnsupported(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("INFO", "server").Expect("redis_version:2.0.0\r\n")
	err := conn.CheckVersion()
	if err != ErrRedisUpgradeRequired {
		t.Fatalf("Expected ErrRedisUpgradeRequired, got %v", err)
	}
}

func TestRedisCheckVersionNilConn(t *testing.T) {
	conn2 := &Redis{}
	err := conn2.checkVersion(nil)
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestRedisFailure(t *testing.T) {
	_, conn := prepareTest()
	if conn.Failure() {
		t.Fatal("Expected failure to be false initially")
	}
	conn.setFailureState(true)
	if !conn.Failure() {
		t.Fatal("Expected failure to be true after set")
	}
	conn.setFailureState(false)
}

func TestRedisConnectPubsub(t *testing.T) {
	_, conn := prepareTest()
	conn.ConnectPubsub()
	if conn.Pubsub == nil {
		t.Fatal("Expected non-nil Pubsub")
	}
}

func TestGetDBFormatVersionExisting(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).Expect(int64(1))
	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != 1 {
		t.Fatalf("Expected 1, got %d", version)
	}
}

func TestGetDBFormatVersionNotFoundWithMirrors(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).ExpectError(redis.ErrNil)
	mock.Command("EXISTS", "MIRRORS").Expect(int64(1))
	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != 0 {
		t.Fatalf("Expected 0, got %d", version)
	}
}

func TestGetDBFormatVersionNotFoundNoMirrors(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).ExpectError(redis.ErrNil)
	mock.Command("EXISTS", "MIRRORS").Expect(int64(0))
	mock.Command("SET", core.DBVersionKey, core.DBVersion).Expect("OK")
	version, err := conn.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if version != core.DBVersion {
		t.Fatalf("Expected %d, got %d", core.DBVersion, version)
	}
}

func TestUpgradeNeededCurrent(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).Expect(int64(core.DBVersion))
	needed, err := conn.UpgradeNeeded()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if needed {
		t.Fatal("Expected false, upgrade not needed")
	}
}

func TestUpgradeNeededOlder(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).Expect(int64(0))
	needed, err := conn.UpgradeNeeded()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !needed {
		t.Fatal("Expected true, upgrade needed")
	}
}

func TestUpgradeNeededNewerUnsupported(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("GET", core.DBVersionKey).Expect(int64(core.DBVersion + 1))
	_, err := conn.UpgradeNeeded()
	if err != ErrUnsupportedVersion {
		t.Fatalf("Expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestLockAcquireInvalidName(t *testing.T) {
	_, conn := prepareTest()
	_, err := conn.AcquireLock("")
	if err != ErrInvalidLockName {
		t.Fatalf("Expected ErrInvalidLockName, got %v", err)
	}
}

func TestLockAcquireSuccess(t *testing.T) {
	mock, conn := prepareTest()
	mock.GenericCommand("SET").Expect("OK")
	lock, err := conn.AcquireLock("test")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !lock.Held() {
		t.Fatal("Expected lock to be held")
	}
	lock.Release()
}

func TestLockAcquireAlreadyLocked(t *testing.T) {
	mock, conn := prepareTest()
	mock.GenericCommand("SET").ExpectError(redis.ErrNil)
	_, err := conn.AcquireLock("test")
	if err != ErrAlreadyLocked {
		t.Fatalf("Expected ErrAlreadyLocked, got %v", err)
	}
}

func TestLockReleaseNotHeld(t *testing.T) {
	_, conn := prepareTest()
	l := &Lock{
		redis: conn,
		name:  "LOCK_test",
		held:  false,
	}
	l.Release()
}

func TestPubsubSubscribeEvent(t *testing.T) {
	_, conn := prepareTest()
	conn.ConnectPubsub()

	ch := make(chan string, 1)
	conn.Pubsub.SubscribeEvent(FILE_UPDATE, ch)

	conn.Pubsub.handleMessage(string(FILE_UPDATE), []byte("test_file"))

	select {
	case msg := <-ch:
		if msg != "test_file" {
			t.Fatalf("Expected 'test_file', got %s", msg)
		}
	default:
		t.Fatal("Expected to receive message")
	}
}

func TestPubsubHandleMessageNoListeners(t *testing.T) {
	_, conn := prepareTest()
	conn.ConnectPubsub()
	conn.Pubsub.handleMessage("_nonexistent", []byte("data"))
}

func TestPublish(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("PUBLISH", string(FILE_UPDATE), "test_msg").Expect(1)
	c := conn.UnblockedGet()
	defer c.Close()
	err := Publish(c, FILE_UPDATE, "test_msg")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSendPublish(t *testing.T) {
	mock, conn := prepareTest()
	mock.Command("PUBLISH", string(MIRROR_UPDATE), "test_mirror").Expect(1)
	c := conn.UnblockedGet()
	defer c.Close()
	err := SendPublish(c, MIRROR_UPDATE, "test_mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

