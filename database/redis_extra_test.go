// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package database

import (
	"errors"
	"strings"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		RedisAddress:  "localhost:6379",
	})
}

func TestParseVersionExtra2(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"3.2.0", 30200},
		{"4.0.0", 40000},
		{"5.0.14", 50014},
		{"10.20.30", 102030},
		{"invalid", -1},
		{"", 0},
		{"1", 1},
		{"1.2", 102},
	}
	for _, tt := range tests {
		got := parseVersion(tt.input)
		if got != tt.expected {
			t.Fatalf("parseVersion(%q): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestParseInfoExtra(t *testing.T) {
	data := "redis_version:5.0.14\r\nredis_mode:standalone\r\nos:Linux\r\n"
	result, err := parseInfo(data, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result["redis_version"] != "5.0.14" {
		t.Fatalf("Expected 5.0.14, got %s", result["redis_version"])
	}
	if result["os"] != "Linux" {
		t.Fatalf("Expected Linux, got %s", result["os"])
	}
}

func TestParseInfoWithComments2(t *testing.T) {
	data := "# Server\r\nredis_version:5.0.14\r\n# Comment\r\nos:Linux\r\n"
	result, _ := parseInfo(data, nil)
	if result["redis_version"] != "5.0.14" {
		t.Fatalf("Expected 5.0.14, got %s", result["redis_version"])
	}
}

func TestParseInfoError2(t *testing.T) {
	_, err := parseInfo(nil, errors.New("test error"))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRedisIsLoadingExtra(t *testing.T) {
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

func TestCheckVersionNilConnExtra(t *testing.T) {
	r := &Redis{}
	err := r.checkVersion(nil)
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestCheckVersionWithMockExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("INFO", "server").Expect("redis_version:5.0.14\r\nredis_mode:standalone\r\n")

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	err := r.CheckVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCheckVersionOldRedisExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("INFO", "server").Expect("redis_version:2.4.0\r\nredis_mode:standalone\r\n")

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	err := r.CheckVersion()
	if err != ErrRedisUpgradeRequired {
		t.Fatalf("Expected ErrRedisUpgradeRequired, got %v", err)
	}
}

func TestGetNotReadyExtra(t *testing.T) {
	r := &Redis{
		stop:  make(chan bool),
		ready: make(chan struct{}),
	}
	conn := r.Get()
	if conn == nil {
		t.Fatal("Expected non-nil conn")
	}
	if conn.Err() == nil {
		t.Fatal("Expected error for not ready")
	}
}

func TestGetReadyExtra(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	conn := r.Get()
	if conn == nil {
		t.Fatal("Expected non-nil conn")
	}
}

func TestUnblockedGetExtra(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	conn := r.UnblockedGet()
	if conn == nil {
		t.Fatal("Expected non-nil conn")
	}
}

func TestFailureStateExtra(t *testing.T) {
	r := &Redis{}
	r.setFailureState(true)
	if !r.Failure() {
		t.Fatal("Expected true")
	}
	r.setFailureState(false)
	if r.Failure() {
		t.Fatal("Expected false")
	}
}

func TestLogErrorExtra(t *testing.T) {
	r := &Redis{}
	r.setFailureState(false)
	r.logError("test error: %s", "detail")
	r.setFailureState(true)
	r.logError("test error: %s", "detail")
}

func TestCloseExtra(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)
	r.ConnectPubsub()
	r.Close()
	r.Close()
}

func TestConnectPubsubExtra(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)
	r.ConnectPubsub()
	if r.Pubsub == nil {
		t.Fatal("Expected non-nil pubsub")
	}
}

func TestAuthNoPasswordExtra(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)
	err := r.auth(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSelectDBExtra(t *testing.T) {
	SetConfiguration(&Configuration{RedisDB: 42})
	mock := redigomock.NewConn()
	mock.Command("SELECT", 42).Expect("OK")
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)
	err := r.selectDB(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestAskRoleExtra(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("ROLE").Expect([]interface{}{[]byte("master")})
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)
	role, err := r.askRole(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if role != "master" {
		t.Fatalf("Expected master, got %s", role)
	}
}

func TestConnectNoConfigExtra(t *testing.T) {
	SetConfiguration(&Configuration{})
	r := &Redis{
		stop:  make(chan bool),
		ready: make(chan struct{}),
	}
	_, err := r.Connect()
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestPrintConnectedMasterExtra(t *testing.T) {
	r := &Redis{}
	r.printConnectedMaster("localhost:6379")
	r.printConnectedMaster("localhost:6379")
}

func TestNetReadyErrorExtra(t *testing.T) {
	e := NewNetTemporaryError()
	if e.Timeout() {
		t.Fatal("Expected false")
	}
	if !e.Temporary() {
		t.Fatal("Expected true")
	}
}

func TestNotReadyErrorExtra(t *testing.T) {
	e := &NotReadyError{}
	if e.Err() == nil {
		t.Fatal("Expected non-nil error")
	}
	if e.Flush() == nil {
		t.Fatal("Expected non-nil from Flush")
	}
}

func TestNotReadyErrorExtraDo(t *testing.T) {
	e := &NotReadyError{}
	_, err := e.Do("PING")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestNotReadyErrorExtraSend(t *testing.T) {
	e := &NotReadyError{}
	err := e.Send("PING")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestNotReadyErrorExtraReceive(t *testing.T) {
	e := &NotReadyError{}
	_, err := e.Receive()
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestRedisIsLoadingExtraPrefix(t *testing.T) {
	err := errors.New("LOADING something")
	if !RedisIsLoading(err) {
		t.Fatal("Expected true")
	}
	err2 := errors.New("ERR something else")
	if RedisIsLoading(err2) {
		t.Fatal("Expected false")
	}
}

func TestStringsImportExtra(t *testing.T) {
	_ = strings.Contains
	_ = redis.ErrNil
}
