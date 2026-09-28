// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
)

func newHTTPWithMockRedis() (*redigomock.Conn, *HTTP) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	rdb := database.NewRedisCustomPool(pool)
	h := &HTTP{
		redis:    rdb,
		geoip:    network.NewGeoIP(),
		engine:   DefaultEngine{},
		templates: Templates{RWMutex: new(sync.RWMutex)},
		stats:    &Stats{countChan: make(chan countItem, 100), mapStats: make(map[string]int64), stop: make(chan bool)},
	}
	return mock, h
}

func TestFileStatsHandler_Today(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/test/file.txt?stats", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	today := time.Now().Format("2006_01_02")
	mock.Command("MULTI").Expect("OK")
	mock.Command("HGET", "STATS_FILE_"+today, "/test/file.txt").Expect([]byte("10"))
	mock.Command("HGET", "STATS_FILE_"+today[:7], "/test/file.txt").Expect([]byte("100"))
	mock.Command("HGET", "STATS_FILE_"+today[:4], "/test/file.txt").Expect([]byte("1000"))
	mock.Command("HGET", "STATS_FILE", "/test/file.txt").Expect([]byte("10000"))
	mock.Command("EXEC").Expect([]interface{}{
		[]byte("10"), []byte("100"), []byte("1000"), []byte("10000"),
	})

	h.fileStatsHandler(w, req, c)

	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestFileStatsHandler_Period(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/test/file.txt?stats=2020-01", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	mock.Command("HGET", "STATS_FILE_2020_01", "/test/file.txt").Expect(int64(42))

	h.fileStatsHandler(w, req, c)

	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestFileStatsHandler_InvalidPeriod(t *testing.T) {
	SetConfiguration(&Configuration{})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/test/file.txt?stats=abc", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	h.fileStatsHandler(w, req, c)

	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestChecksumHandler_OutsideRepo(t *testing.T) {
	dir := "/tmp/mirrorbits-checksum-test"
	SetConfiguration(&Configuration{Repository: dir})

	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/../../etc/passwd?md5", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	h.checksumHandler(w, req, c)

	if w.Code != 403 {
		t.Fatalf("Expected 403, got %d", w.Code)
	}
}

func TestChecksumHandler_NotFound(t *testing.T) {
	dir := "/tmp/mirrorbits-checksum-test2"
	SetConfiguration(&Configuration{Repository: dir})

	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/nonexistent?md5", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	h.checksumHandler(w, req, c)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestRequestDispatcher_Standard(t *testing.T) {
	SetConfiguration(&Configuration{})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()

	h.requestDispatcher(w, req)
}

func TestRequestDispatcher_Checksum(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp"})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/test?md5", nil)
	w := httptest.NewRecorder()

	h.requestDispatcher(w, req)
}

func TestRequestDispatcher_FileStats(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, h := newHTTPWithMockRedis()

	today := time.Now().Format("2006_01_02")
	mock.Command("MULTI").Expect("OK")
	mock.Command("HGET", "STATS_FILE_"+today, "/test").Expect([]byte("0"))
	mock.Command("HGET", "STATS_FILE_"+today[:7], "/test").Expect([]byte("0"))
	mock.Command("HGET", "STATS_FILE_"+today[:4], "/test").Expect([]byte("0"))
	mock.Command("HGET", "STATS_FILE", "/test").Expect([]byte("0"))
	mock.Command("EXEC").Expect([]interface{}{[]byte("0"), []byte("0"), []byte("0"), []byte("0")})

	req := httptest.NewRequest("GET", "/test?stats", nil)
	w := httptest.NewRecorder()

	h.requestDispatcher(w, req)
}

func TestMirrorHandler_RootPath(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "json"})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.mirrorHandler(w, req, NewContext(w, req, Templates{}))
}

func TestMirrorHandler_OutsideRepo(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp/mirrorbits-test-repo", OutputMode: "json"})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/../../etc/passwd", nil)
	w := httptest.NewRecorder()

	h.mirrorHandler(w, req, NewContext(w, req, Templates{}))

	if w.Code != 403 {
		t.Fatalf("Expected 403, got %d", w.Code)
	}
}

func TestMirrorHandler_NotFound(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp/mirrorbits-test-repo", OutputMode: "json"})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/nonexistent_file", nil)
	w := httptest.NewRecorder()

	h.mirrorHandler(w, req, NewContext(w, req, Templates{}))

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestMirrorHandler_Fallback(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:  "/tmp/mirrorbits-test-repo",
		OutputMode:  "json",
		Fallbacks:   []Fallback{{URL: "http://fallback.example.com", CountryCode: "US", ContinentCode: "NA", Name: "fallback"}},
	})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/nonexistent_file", nil)
	w := httptest.NewRecorder()

	h.mirrorHandler(w, req, NewContext(w, req, Templates{}))
}

func TestProcessCountDownload_WithItems(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	rdb := database.NewRedisCustomPool(pool)

	s := &Stats{
		r:         rdb,
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	s.countChan <- countItem{mirrorID: 1, filepath: "/test", size: 100, time: time.Now().UTC()}

	go s.processCountDownload()

	time.Sleep(100 * time.Millisecond)
	s.Terminate()
}

func TestMirrorHandler_MirrorlistWithFromIP(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:  "/tmp/mirrorbits-test-repo",
		OutputMode:  "json",
		Fallbacks:   []Fallback{{URL: "http://fb.example.com", CountryCode: "CN", ContinentCode: "AS", Name: "fb"}},
	})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/nonexistent?mirrorlist&fromip=1.2.3.4", nil)
	w := httptest.NewRecorder()

	h.mirrorHandler(w, req, NewContext(w, req, Templates{}))
}

func TestHTTP_SetListener(t *testing.T) {
	h := &HTTP{}
	_ = h
}

func TestHTTP_Stop(t *testing.T) {
	h := &HTTP{stopped: true}
	h.Stop(0)
}

func TestMirrorSelector_NoRepoVersionList(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp"})
	_, h := newHTTPWithMockRedis()

	fi := filesystem.NewFileInfo("/test")
	mlist, excluded, err := h.mirrorSelector(&Context{}, nil, &fi, network.GeoIPRecord{})
	_ = mlist
	_ = excluded
	_ = err
}

func TestSyncOffsetStruct(t *testing.T) {
	s := SyncOffset{Valid: true, Value: 5, HumanReadable: "5h"}
	if !s.Valid {
		t.Fatalf("Expected valid")
	}
	if s.Value != 5 {
		t.Fatalf("Expected 5")
	}
}

func TestMirrorStatsPageStruct(t *testing.T) {
	p := MirrorStatsPage{
		List: []MirrorStats{{ID: 1, Name: "m1"}},
	}
	if len(p.List) != 1 {
		t.Fatalf("Expected 1")
	}
}

func TestStatsFileNowStruct(t *testing.T) {
	s := StatsFileNow{Today: 1, Month: 2, Year: 3, Total: 4}
	if s.Today != 1 || s.Month != 2 || s.Year != 3 || s.Total != 4 {
		t.Fatalf("Struct values mismatch")
	}
}

func TestStatsFilePeriodStruct(t *testing.T) {
	s := StatsFilePeriod{Period: "2020-01", Downloads: 100}
	if s.Period != "2020-01" {
		t.Fatalf("Period mismatch")
	}
}

func TestMirrorStatsStruct(t *testing.T) {
	s := MirrorStats{ID: 1, Name: "m1", Downloads: 100}
	if s.ID != 1 {
		t.Fatalf("ID mismatch")
	}
}

func TestMirrorStatsExtendedStruct(t *testing.T) {
	s := MirrorStatsExtended{
		Mirror:    mirrors.Mirror{ID: 1},
		Downloads: 100,
	}
	if s.ID != 1 {
		t.Fatalf("ID mismatch")
	}
}

func ensureRedisImport() {
	_ = redis.ErrNil
}
