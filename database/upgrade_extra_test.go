// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package database

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/rafaeljusto/redigomock"
)

func TestGetDBFormatVersionGenericError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).ExpectError(errors.New("connection error"))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	version, err := r.GetDBFormatVersion()
	if err == nil {
		t.Fatal("Expected error")
	}
	if version != -1 {
		t.Fatalf("Expected -1, got %d", version)
	}
}

func TestGetDBFormatVersionExistsError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).ExpectError(redis.ErrNil)
	mock.Command("EXISTS", "MIRRORS").ExpectError(errors.New("exists error"))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	version, err := r.GetDBFormatVersion()
	if err == nil {
		t.Fatal("Expected error")
	}
	if version != -1 {
		t.Fatalf("Expected -1, got %d", version)
	}
}

func TestUpgradeCurrentVersion(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).Expect(int64(core.DBVersion))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	err := r.Upgrade()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestUpgradeUnsupportedVersion(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).Expect(int64(999))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	err := r.Upgrade()
	if err != ErrUnsupportedVersion {
		t.Fatalf("Expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestUpgradeGetVersionError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).ExpectError(errors.New("conn error"))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	err := r.Upgrade()
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestUpgradeNeededGetVersionError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("GET", core.DBVersionKey).ExpectError(errors.New("conn error"))

	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	_, err := r.UpgradeNeeded()
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestNewRedisCustomPoolWithMock(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &testPool{Conn: mock}
	r := NewRedisCustomPool(pool)

	select {
	case <-r.ready:
	default:
		t.Fatal("Expected ready channel to be closed for mock pool")
	}
}

func TestRedisIsLoadingNonLoadingErr(t *testing.T) {
	err := errors.New("not a loading error")
	if RedisIsLoading(err) {
		t.Fatal("Expected false for non-loading error")
	}
}

func TestRedisIsLoadingNilErr(t *testing.T) {
	if RedisIsLoading(nil) {
		t.Fatal("Expected false for nil error")
	}
}

func TestNetReadyErrorTimeoutTemporary(t *testing.T) {
	e := NewNetTemporaryError()
	if e.Timeout() {
		t.Fatal("Expected false for Timeout")
	}
	if !e.Temporary() {
		t.Fatal("Expected true for Temporary")
	}
}

func TestNotReadyErrorFlush(t *testing.T) {
	e := &NotReadyError{}
	err := e.Flush()
	if err == nil {
		t.Fatal("Expected error from NotReadyError.Flush")
	}
}

func TestNotReadyErrorErr(t *testing.T) {
	e := &NotReadyError{}
	err := e.Err()
	if err == nil {
		t.Fatal("Expected error from NotReadyError.Err")
	}
}

func TestNotReadyErrorClose(t *testing.T) {
	e := &NotReadyError{}
	err := e.Close()
	if err == nil {
		t.Fatal("Expected error from NotReadyError.Close")
	}
}

func TestErrUnsupportedVersion(t *testing.T) {
	if ErrUnsupportedVersion == nil {
		t.Fatal("Expected non-nil")
	}
	if ErrUnsupportedVersion.Error() == "" {
		t.Fatal("Expected non-empty message")
	}
}
