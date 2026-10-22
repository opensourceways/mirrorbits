// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package network

import (
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type networkTestPool struct {
	Conn *redigomock.Conn
}

func (r *networkTestPool) Get() redis.Conn {
	return r.Conn
}

func (r *networkTestPool) Close() error {
	return nil
}

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestNewClusterLock(t *testing.T) {
	l := NewClusterLock(nil, "test_key", "test_id")
	if l == nil {
		t.Fatal("Expected non-nil lock")
	}
	if l.key != "test_key" {
		t.Fatalf("Expected test_key, got %s", l.key)
	}
}

func TestClusterLockGetAlreadyInUse(t *testing.T) {
	l := &ClusterLock{
		redis:      nil,
		key:        "test",
		identifier: "test",
		done:       make(chan struct{}),
	}
	_, err := l.Get()
	if err == nil {
		t.Fatal("Expected error for already in use")
	}
}

func TestRemoteIPFromAddrNoColon(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for string without colon")
		}
	}()
	_ = RemoteIPFromAddr("nocolon")
}

func TestExtractRemoteIPEmpty(t *testing.T) {
	r := ExtractRemoteIP("")
	if r != "" {
		t.Fatalf("Expected empty, got %s", r)
	}
}

func TestLookupMirrorIPLocalhost(t *testing.T) {
	_, err := LookupMirrorIP("localhost")
	_ = err
}

func TestLookupMirrorIPInvalidHost(t *testing.T) {
	_, err := LookupMirrorIP("invalid.invalid.invalid")
	if err == nil {
		t.Fatal("Expected error for invalid host")
	}
}

func TestGeoIPErrorMethodsExtra(t *testing.T) {
	e := GeoIPError{
		Errors: []error{},
		loaded: 0,
	}
	if !e.IsFatal() {
		t.Fatal("Expected true when loaded == len(Errors)")
	}

	e.loaded = 1
	e.Errors = []error{nil, nil}
	if e.IsFatal() {
		t.Fatal("Expected false when loaded < len(Errors)")
	}

	if e.Error() == "" {
		t.Fatal("Expected non-empty error string")
	}
}

func TestGeoIPRecordIsValidEmptyCountryCode(t *testing.T) {
	r := GeoIPRecord{ContinentCode: "EU"}
	if r.IsValid() {
		t.Fatal("Expected false for empty CountryCode")
	}
}

func TestGeoIPLoadGeoIPNoFiles(t *testing.T) {
	SetConfiguration(&Configuration{
		GeoipDatabasePath: "/nonexistent/path",
	})
	g := NewGeoIP()
	err := g.LoadGeoIP()
	if err == nil {
		t.Fatal("Expected error for non-existent GeoIP files")
	}
}

func TestGeoIPLoadGeoIPEmptyPath(t *testing.T) {
	SetConfiguration(&Configuration{
		GeoipDatabasePath: "",
	})
	g := NewGeoIP()
	err := g.LoadGeoIP()
	if err == nil {
		t.Fatal("Expected error for empty path")
	}
}
