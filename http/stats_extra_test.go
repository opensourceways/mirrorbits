// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/rafaeljusto/redigomock"
	"github.com/opensourceways/mirrorbits/database"
)

func newTestRedis(pool *redisPoolMock) *database.Redis {
	return database.NewRedisCustomPool(pool)
}

func TestGzipResponseWriter_Write_ContentTypeDetection(t *testing.T) {
	w := httptest.NewRecorder()
	gzw := &gzipResponseWriter{Writer: w, ResponseWriter: w}

	_, err := gzw.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if w.Header().Get("Content-Type") == "" {
		t.Fatalf("Expected Content-Type to be set")
	}
	if !gzw.typeGuessed {
		t.Fatalf("Expected typeGuessed=true")
	}
}

func TestGzipResponseWriter_Write_ExistingContentType(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "text/plain")
	gzw := &gzipResponseWriter{Writer: w, ResponseWriter: w}

	_, err := gzw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !gzw.typeGuessed {
		t.Fatalf("Expected typeGuessed=true even with existing CT")
	}
}

func TestHandlerRes_JSONMode(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "json"})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
	}

	handlerRes(w, req, c, results, GetConfig())
}

func TestHandlerRes_RedirectMode(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "redirect"})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
		MirrorList: mirrors.Mirrors{
			mirrors.Mirror{ID: 1, Name: "m1", HttpURL: "http://m1.example.com"},
		},
	}

	handlerRes(w, req, c, results, GetConfig())
}

func TestHandlerRes_AutoMode_JSON(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "auto"})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
	}

	handlerRes(w, req, c, results, GetConfig())
}

func TestHandlerRes_AutoMode_Redirect(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "auto"})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
		MirrorList: mirrors.Mirrors{
			mirrors.Mirror{ID: 1, Name: "m1", HttpURL: "http://m1.example.com"},
		},
	}

	handlerRes(w, req, c, results, GetConfig())
}

func TestHandlerRes_MirrorlistMode(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "json"})

	req := httptest.NewRequest("GET", "/test/file.txt?mirrorlist", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
	}

	handlerRes(w, req, c, results, GetConfig())
}

func TestHandlerRes_DefaultMode_Error(t *testing.T) {
	SetConfiguration(&Configuration{OutputMode: "invalid"})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
	}

	handlerRes(w, req, c, results, GetConfig())
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500 for invalid mode, got %d", w.Code)
	}
}

func TestStats_PushStats_Empty(t *testing.T) {
	s := &Stats{
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}
	s.pushStats()
}

func TestStats_PushStats_WithConnError(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	rdb := newTestRedis(pool)

	s := &Stats{
		r:         rdb,
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	s.mapStats["x2020_01_02|/file.txt"] = 1

	s.pushStats()
}

func TestStats_PushStats_WithStats(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	rdb := newTestRedis(pool)

	s := &Stats{
		r:         rdb,
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	s.mapStats["f2020_01_02|/file.txt"] = 1
	s.mapStats["m2020_01_02|1"] = 1
	s.mapStats["s2020_01_02|1"] = 100

	mock.Command("MULTI").Expect("OK")
	mock.Command("HINCRBY", "STATS_FILE_2020_01_02", "/file.txt", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_FILE_2020_01", "/file.txt", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_FILE_2020", "/file.txt", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_FILE", "/file.txt", int64(1)).Expect(int64(1))
	mock.Command("INCRBY", "STATS_TOTAL", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_MIRROR_2020_01_02", "1", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_MIRROR_2020_01", "1", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_MIRROR_2020", "1", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_MIRROR", "1", int64(1)).Expect(int64(1))
	mock.Command("HINCRBY", "STATS_MIRROR_BYTES_2020_01_02", "1", int64(100)).Expect(int64(100))
	mock.Command("HINCRBY", "STATS_MIRROR_BYTES_2020_01", "1", int64(100)).Expect(int64(100))
	mock.Command("HINCRBY", "STATS_MIRROR_BYTES_2020", "1", int64(100)).Expect(int64(100))
	mock.Command("HINCRBY", "STATS_MIRROR_BYTES", "1", int64(100)).Expect(int64(100))
	mock.Command("EXEC").Expect("OK")

	s.pushStats()
}

func TestStats_Terminate(t *testing.T) {
	s := &Stats{
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	go func() {
		s.processCountDownload()
	}()

	s.Terminate()
}

func TestStats_NewStats(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redisPoolMock{Conn: mock}
	rdb := newTestRedis(pool)

	s := NewStats(rdb)
	if s == nil {
		t.Fatalf("Expected non-nil stats")
	}
	s.Terminate()
}

func TestMirrorStatsSlice_Len(t *testing.T) {
	s := mirrorStatsSlice{{}, {}, {}}
	if s.Len() != 3 {
		t.Fatalf("Expected 3, got %d", s.Len())
	}
}

func TestMirrorStatsSlice_Swap(t *testing.T) {
	s := mirrorStatsSlice{{ID: 1}, {ID: 2}}
	s.Swap(0, 1)
	if s[0].ID != 2 {
		t.Fatalf("Expected 2, got %d", s[0].ID)
	}
}

func TestByDownloadNumbers_Less(t *testing.T) {
	b := byDownloadNumbers{
		mirrorStatsSlice{
			{Downloads: 100},
			{Downloads: 50},
		},
	}
	if !b.Less(0, 1) {
		t.Fatalf("Expected true for 100 > 50")
	}
	if b.Less(1, 0) {
		t.Fatalf("Expected false for 50 < 100")
	}
}
