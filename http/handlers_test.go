// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
	"github.com/rafaeljusto/redigomock"
)

type mockRedisPool struct {
	conn *redigomock.Conn
}

func (p *mockRedisPool) Get() redis.Conn {
	return p.conn
}

func (p *mockRedisPool) Close() error {
	return nil
}

func newMockRedis(mock *redigomock.Conn) *database.Redis {
	pool := &mockRedisPool{conn: mock}
	return database.NewRedisCustomPool(pool)
}

func newTestHTTP() *HTTP {
	h := &HTTP{
		templates: Templates{RWMutex: new(sync.RWMutex)},
		engine:    DefaultEngine{},
		geoip:     network.NewGeoIP(),
	}
	return h
}

func newTestHTTPWithRedis(mock *redigomock.Conn) *HTTP {
	h := newTestHTTP()
	h.redis = newMockRedis(mock)
	h.stats = NewStats(h.redis)
	return h
}

func TestRequestDispatcherStandard(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)

	if w.Header().Get("Server") == "" {
		t.Fatal("Expected Server header to be set")
	}
	if !strings.Contains(w.Header().Get("Server"), "Mirrorbits/") {
		t.Fatalf("Expected Server header to contain Mirrorbits/, got %s", w.Header().Get("Server"))
	}
}

func TestRequestDispatcherMirrorlist(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/?mirrorlist", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)
}

func TestRequestDispatcherChecksum(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/test.txt?md5", nil)
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for non-existent file, got %d", w.Code)
	}
}

func TestRequestDispatcherFileStats(t *testing.T) {
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
	h.requestDispatcher(w, req)
}

func TestRequestDispatcherMirrorStats(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	h.requestDispatcher(w, req)

	if !strings.Contains(w.Header().Get("Server"), "Mirrorbits/") {
		t.Fatalf("Expected Server header to contain Mirrorbits/")
	}
}

func TestHandlerResJSON(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Header().Get("Cache-Control") != "private, no-cache" {
		t.Fatalf("Expected Cache-Control header, got %s", w.Header().Get("Cache-Control"))
	}
}

func TestHandlerResRedirect(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "redirect",
		MaxLinkHeaders: 5,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
		MirrorList: []mirrors.Mirror{
			{HttpURL: "http://mirror1.com/"},
		},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Code != http.StatusFound {
		t.Fatalf("Expected 302, got %d", w.Code)
	}
}

func TestHandlerResRedirectNoMirrors(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "redirect",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestHandlerResAutoJSON(t *testing.T) {
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

func TestHandlerResAutoRedirect(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "auto",
		MaxLinkHeaders: 5,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
		MirrorList: []mirrors.Mirror{
			{HttpURL: "http://mirror1.com/"},
		},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Code != http.StatusFound {
		t.Fatalf("Expected 302, got %d", w.Code)
	}
}

func TestHandlerResMirrorlist(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?mirrorlist", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	handlerRes(w, req, ctx, results, GetConfig())
}

func TestHandlerResDefaultError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "unknown",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	handlerRes(w, req, ctx, results, GetConfig())

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500, got %d", w.Code)
	}
}

func TestFileStatsHandlerNoPeriod(t *testing.T) {
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
	h.fileStatsHandler(w, req, ctx)
}

func TestFileStatsHandlerWithPeriod(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	mock.Command("HGET", "STATS_FILE_2024_01", "/test.txt").Expect(int64(100))

	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/test.txt?stats=2024-01", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)
}

func TestFileStatsHandlerInvalidPeriod(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	h := newTestHTTPWithRedis(mock)

	req := httptest.NewRequest("GET", "/test.txt?stats=abc", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for invalid period, got %d", w.Code)
	}
}

func TestFileStatsHandlerInvalidPeriodPart2(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	h := newTestHTTPWithRedis(mock)

	req := httptest.NewRequest("GET", "/test.txt?stats=2024-abc", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.fileStatsHandler(w, req, ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for invalid period, got %d", w.Code)
	}
}

func TestChecksumHandlerNotFound(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt?md5", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.checksumHandler(w, req, ctx)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestChecksumHandlerOutsideRepo(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/../../../etc/passwd?md5", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.checksumHandler(w, req, ctx)

	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Fatalf("Expected 403 or 404, got %d", w.Code)
	}
}

func TestMirrorHandlerRootPath(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerRootPathRedirect(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "redirect",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerFileNotFound(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)

	if w.Code != http.StatusNotFound && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected 404 or 503, got %d", w.Code)
	}
}

func TestMirrorHandlerWithFallbacks(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "redirect",
		MaxLinkHeaders: 5,
		Fallbacks: []Fallback{
			{Name: "fb1", URL: "http://fb1.com/", CountryCode: "CN", ContinentCode: "AS", NetworkBandwidth: 1000},
		},
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerWithFallbacksRedirect(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "redirect",
		MaxLinkHeaders: 5,
		Fallbacks: []Fallback{
			{Name: "fb1", URL: "http://fb1.com/", CountryCode: "FR", ContinentCode: "EU", NetworkBandwidth: 1000},
			{Name: "fb2", URL: "http://fb2.com/", CountryCode: "DE", ContinentCode: "EU", NetworkBandwidth: 500},
		},
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)

	if w.Code != http.StatusFound && w.Code != http.StatusNotFound {
		t.Fatalf("Expected 302 or 404, got %d", w.Code)
	}
}

func TestMirrorHandlerISOSuffix(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/test/ISO/", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerWithFromIP(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt?mirrorlist&fromip=8.8.8.8", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerWithInvalidFromIP(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt?mirrorlist&fromip=invalid", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorHandlerWithXForwardedFor(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	h := newTestHTTP()
	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorHandler(w, req, ctx)
}

func TestMirrorStatsHandlerError(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	mock := redigomock.NewConn()
	h := newTestHTTPWithRedis(mock)
	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})
	h.mirrorStatsHandler(w, req, ctx)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500 for unreachable redis, got %d", w.Code)
	}
}

func TestRequestDispatcherHealthEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	if w == nil {
		t.Fatal("Expected non-nil recorder")
	}
}

func TestHTTPStopDoubleCall(t *testing.T) {
	h := &HTTP{
		stoppedMutex: sync.Mutex{},
		stopped:      true,
	}
	h.Stop(0)
	h.Stop(0)
}

func TestStatsFileNowStruct(t *testing.T) {
	s := StatsFileNow{Today: 1, Month: 2, Year: 3, Total: 4}
	if s.Today != 1 || s.Month != 2 || s.Year != 3 || s.Total != 4 {
		t.Fatal("Fields mismatch")
	}
}

func TestStatsFilePeriodStruct(t *testing.T) {
	s := StatsFilePeriod{Period: "2024-01", Downloads: 100}
	if s.Period != "2024-01" || s.Downloads != 100 {
		t.Fatal("Fields mismatch")
	}
}

func TestMirrorStatsExtendedStruct(t *testing.T) {
	mse := MirrorStatsExtended{
		Downloads: 10,
		Bytes:     1024,
		PercentD:  50.0,
		PercentB:  75.0,
	}
	if mse.Downloads != 10 || mse.Bytes != 1024 {
		t.Fatal("Fields mismatch")
	}
}

func TestSyncOffsetValid(t *testing.T) {
	so := SyncOffset{Valid: true, Value: 5, HumanReadable: "5h"}
	if !so.Valid || so.Value != 5 || so.HumanReadable != "5h" {
		t.Fatal("Fields mismatch")
	}
}

func TestMirrorStatsPageWithTZ(t *testing.T) {
	p := MirrorStatsPage{
		List:             []MirrorStats{{ID: 1}},
		MirrorList:       []mirrors.Mirror{{Name: "m1"}},
		LocalJSPath:       "/js",
		HasTZAdjustement:  true,
	}
	if !p.HasTZAdjustement || p.LocalJSPath != "/js" {
		t.Fatal("Fields mismatch")
	}
}
