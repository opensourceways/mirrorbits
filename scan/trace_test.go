// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func newScanRedis(pool *mockPool) *database.Redis {
	return database.NewRedisCustomPool(pool)
}

func TestNewTraceHandler(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	stop := make(chan struct{})

	t1 := NewTraceHandler(r, stop)
	if t1 == nil {
		t.Fatalf("Expected non-nil trace")
	}
}

func TestGetLastUpdate_NoTraceFile(t *testing.T) {
	SetConfiguration(&Configuration{TraceFileLocation: ""})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	stop := make(chan struct{})

	t1 := NewTraceHandler(r, stop)

	err := t1.GetLastUpdate(mirrors.Mirror{ID: 1, Name: "m1"})
	if err != ErrNoTrace {
		t.Fatalf("Expected ErrNoTrace, got %v", err)
	}
}

func TestGetLastUpdate_HTTPError(t *testing.T) {
	SetConfiguration(&Configuration{TraceFileLocation: "/trace"})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	stop := make(chan struct{})

	t1 := NewTraceHandler(r, stop)

	err := t1.GetLastUpdate(mirrors.Mirror{ID: 1, Name: "m1", HttpURL: "http://nonexistent.invalid"})
	if err == nil {
		t.Fatalf("Expected HTTP error")
	}
}

func TestGetLastUpdate_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "1609459200")
	}))
	defer ts.Close()

	SetConfiguration(&Configuration{TraceFileLocation: "/trace"})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	stop := make(chan struct{})

	t1 := NewTraceHandler(r, stop)
	t1.httpClient = resty.New().RemoveProxy().SetHeader(userAgentName, userAgent).SetDoNotParseResponse(true)

	mock.Command("HSET", "MIRROR_1", "lastModTime", int64(1609459200)).Expect(int64(1))
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")

	err := t1.GetLastUpdate(mirrors.Mirror{ID: 1, Name: "m1", HttpURL: ts.URL})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestGetLastUpdate_InvalidTimestamp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not-a-number")
	}))
	defer ts.Close()

	SetConfiguration(&Configuration{TraceFileLocation: "/trace"})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	stop := make(chan struct{})

	t1 := NewTraceHandler(r, stop)
	t1.httpClient = resty.New().RemoveProxy().SetHeader(userAgentName, userAgent).SetDoNotParseResponse(true)

	err := t1.GetLastUpdate(mirrors.Mirror{ID: 1, Name: "m1", HttpURL: ts.URL})
	if err == nil {
		t.Fatalf("Expected error for invalid timestamp")
	}
}

func TestErrNoTrace(t *testing.T) {
	if ErrNoTrace == nil {
		t.Fatalf("Expected non-nil error")
	}
	if ErrNoTrace.Error() != "No trace file" {
		t.Fatalf("Unexpected message: %s", ErrNoTrace.Error())
	}
}

func TestScan_HGETError(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)

	mock.Command("HGET", "MIRRORS", 1).ExpectError(redis.Error("ERR not found"))

	_, err := Scan(core.HTTP, r, nil, "http://example.com", 1, nil)
	if err == nil {
		t.Fatalf("Expected error from HGET")
	}
}

func TestScan_AlreadyInProgress(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)

	mock.Command("HGET", "MIRRORS", 1).Expect("mirror1")
	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).ExpectError(redis.ErrNil)

	_, err := Scan(core.HTTP, r, nil, "http://example.com", 1, nil)
	if err != ErrScanInProgress {
		t.Fatalf("Expected ErrScanInProgress, got %v", err)
	}
}

func TestScan_LockError(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)

	mock.Command("HGET", "MIRRORS", 1).Expect("mirror1")
	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).ExpectError(redis.Error("ERR fail"))

	_, err := Scan(core.HTTP, r, nil, "http://example.com", 1, nil)
	if err == nil {
		t.Fatalf("Expected error from lock")
	}
}

func TestHttpScanner_Scan_NotHTTPS(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	s := &scan{conn: mock, redis: r, mirrorid: 1, filesTmpKey: "MIRRORFILESTMP_1"}
	scanner := &HttpScanner{scan: s}

	repoVersion := []*filesystem.LayerFile{
		{Dir: "test", Name: "file.iso"},
	}
	_, _, err := scanner.Scan("http://example.com", "mirror1", repoVersion, nil)
	if err == nil {
		t.Fatalf("Expected error for non-HTTPS URL")
	}
}

func TestHttpScanner_Scan_Stopped(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock := redigomock.NewConn()
	pool := &mockPool{conn: mock}
	r := newScanRedis(pool)
	s := &scan{conn: mock, redis: r, mirrorid: 1, filesTmpKey: "MIRRORFILESTMP_1"}
	scanner := &HttpScanner{scan: s}

	stop := make(chan struct{})
	close(stop)

	repoVersion := []*filesystem.LayerFile{
		{Dir: "test", Name: "file.iso"},
	}
	_, _, err := scanner.Scan("https://example.com", "mirror1", repoVersion, stop)
	if err != ErrScanAborted {
		t.Fatalf("Expected ErrScanAborted, got %v", err)
	}
}
