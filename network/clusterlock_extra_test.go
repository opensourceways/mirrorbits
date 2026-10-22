// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package network

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/rafaeljusto/redigomock"
)

func newNetworkTestRedis(mock *redigomock.Conn) *database.Redis {
	pool := &networkTestPool{Conn: mock}
	return database.NewRedisCustomPool(pool)
}

func TestClusterLockGetSuccess(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SET", "test_key", 1, "NX", "EX", 10).Expect("OK")

	r := newNetworkTestRedis(mock)
	l := NewClusterLock(r, "test_key", "test_id")

	done, err := l.Get()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if done == nil {
		t.Fatal("Expected non-nil done channel")
	}

	// Clean up
	l.Release()
	time.Sleep(50 * time.Millisecond)
}

func TestClusterLockGetNotAcquired(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SET", "test_key2", 1, "NX", "EX", 10).ExpectError(redis.ErrNil)

	r := newNetworkTestRedis(mock)
	l := NewClusterLock(r, "test_key2", "test_id")

	done, err := l.Get()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if done != nil {
		t.Fatal("Expected nil done channel for not acquired lock")
	}
}

func TestClusterLockGetError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("SET", "test_key3", 1, "NX", "EX", 10).ExpectError(errTestNetError)

	r := newNetworkTestRedis(mock)
	l := NewClusterLock(r, "test_key3", "test_id")

	_, err := l.Get()
	if err == nil {
		t.Fatal("Expected error for SET failure")
	}
}

func TestClusterLockRelease(t *testing.T) {
	l := &ClusterLock{
		done: make(chan struct{}),
	}
	l.Release()

	// Verify done is closed
	select {
	case <-l.done:
		// Expected: done is closed
	default:
		t.Fatal("Expected done channel to be closed")
	}
}

func TestClusterLockReleaseNilDone(t *testing.T) {
	l := &ClusterLock{
		done: nil,
	}
	defer func() {
		if r := recover(); r != nil {
			// Expected: close(nil) panics
		}
	}()
	l.Release()
}

func TestExtractRemoteIPWithComma(t *testing.T) {
	ip := ExtractRemoteIP("8.8.8.8, 10.0.0.1")
	if ip != "8.8.8.8" {
		t.Fatalf("Expected 8.8.8.8, got %s", ip)
	}
}

func TestExtractRemoteIPSingle(t *testing.T) {
	ip := ExtractRemoteIP("8.8.8.8")
	if ip != "8.8.8.8" {
		t.Fatalf("Expected 8.8.8.8, got %s", ip)
	}
}

func TestGetRecordEmptyIP(t *testing.T) {
	geoip := NewGeoIP()
	rec := geoip.GetRecord("")
	if rec.CountryCode != "" {
		t.Fatalf("Expected empty country code, got %s", rec.CountryCode)
	}
}

func TestGetRecordLocalhost(t *testing.T) {
	geoip := NewGeoIP()
	rec := geoip.GetRecord("127.0.0.1")
	_ = rec
}

type testNetError struct{}

func (e *testNetError) Error() string { return "network error" }

var errTestNetError = &testNetError{}
