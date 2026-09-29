// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/rafaeljusto/redigomock"
)

func TestMirrorHandlerWithIsoPath(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/test/ISO", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerFileStats(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("MULTI").Expect("OK")
	mock.GenericCommand("HGET").Expect([]byte("10"))
	mock.GenericCommand("HGET").Expect([]byte("20"))
	mock.GenericCommand("HGET").Expect([]byte("30"))
	mock.GenericCommand("HGET").Expect([]byte("40"))
	mock.GenericCommand("EXEC").Expect([]interface{}{[]byte("10"), []byte("20"), []byte("30"), []byte("40")})

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestFileStatsHandlerWithPeriodFullDate(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "STATS_FILE_2024_01_15", "/test.txt").Expect(int64(50))

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats=2024-01-15", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)
}

func TestFileStatsHandlerWithPeriodYearMonth(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "STATS_FILE_2024_06", "/test.txt").Expect(int64(30))

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats=2024-06", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)
}

func TestFileStatsHandlerWithPeriodYearOnly(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "STATS_FILE_2024", "/test.txt").Expect(int64(200))

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats=2024", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)
}

func TestFileStatsHandlerWithPeriodError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "STATS_FILE_2024_01", "/test.txt").ExpectError(redigomockError)

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats=2024-01", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500 for HGET error, got %d", w.Code)
	}
}

func TestChecksumHandlerSha1(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt?sha1", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.checksumHandler(w, req, ctx)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestChecksumHandlerSha256(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt?sha256", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.checksumHandler(w, req, ctx)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestRequestDispatcherChecksumMd5(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/file.txt?md5", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestRequestDispatcherChecksumSha1(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/file.txt?sha1", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)
}

func TestRequestDispatcherChecksumSha256(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/file.txt?sha256", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)
}

func TestHandlerResRedirectWithMaxLinkHeaders(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:      "/tmp/test",
		ListenAddress:   ":8080",
		OutputMode:      "redirect",
		MaxLinkHeaders:  10,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
		MirrorList: []mirrors.Mirror{
			{HttpURL: "http://mirror1.com/"},
			{HttpURL: "http://mirror2.com/"},
		},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Code != http.StatusFound {
		t.Fatalf("Expected 302, got %d", w.Code)
	}
}

func TestHandlerResAutoJSONAccept(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "auto",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	handlerRes(w, req, ctx, results, GetConfig())
}

func TestMirrorHandlerRootPathMirrorlist(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/?mirrorlist", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)

	if !strings.Contains(w.Header().Get("Server"), "Mirrorbits/") {
		t.Fatal("Expected Server header")
	}
}

func TestMirrorHandlerWithForwardedForMultipleIPs(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	req.Header.Set("X-Forwarded-For", "8.8.8.8, 10.0.0.1, 192.168.1.1")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerNoFallbacks(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
		Fallbacks:     []Fallback{},
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusNotFound {
		t.Fatalf("Expected 503 or 404, got %d", w.Code)
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }

var redigomockError = &testError{}
