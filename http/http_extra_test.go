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
	"github.com/opensourceways/mirrorbits/network"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type redisPoolMock struct {
	Conn *redigomock.Conn
}

func (r *redisPoolMock) Get() redis.Conn {
	return r.Conn
}

func (r *redisPoolMock) Close() error {
	return nil
}

func TestStats_CountDownload_EmptyMirror(t *testing.T) {
	s := &Stats{
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	m := mirrors.Mirror{}
	fi := filesystem.FileInfo{Path: "/test/file.txt"}

	err := s.CountDownload(m, fi)
	if err != errUnknownMirror {
		t.Fatalf("Expected errUnknownMirror, got %v", err)
	}
}

func TestStats_CountDownload_EmptyPath(t *testing.T) {
	s := &Stats{
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	m := mirrors.Mirror{Name: "test"}
	fi := filesystem.FileInfo{}

	err := s.CountDownload(m, fi)
	if err != errEmptyFileError {
		t.Fatalf("Expected errEmptyFileError, got %v", err)
	}
}

func TestStats_CountDownload_Success(t *testing.T) {
	s := &Stats{
		countChan: make(chan countItem, 100),
		mapStats:  make(map[string]int64),
		stop:      make(chan bool),
	}

	m := mirrors.Mirror{Name: "test", ID: 1}
	fi := filesystem.FileInfo{Path: "/test/file.txt", Size: 100}

	err := s.CountDownload(m, fi)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestJSONRenderer_Type(t *testing.T) {
	r := JSONRenderer{}
	if r.Type() != "JSON" {
		t.Fatalf("Expected 'JSON', got %s", r.Type())
	}
}

func TestRedirectRenderer_Type(t *testing.T) {
	r := RedirectRenderer{}
	if r.Type() != "REDIRECT" {
		t.Fatalf("Expected 'REDIRECT', got %s", r.Type())
	}
}

func TestMirrorListRenderer_Type(t *testing.T) {
	r := MirrorListRenderer{}
	if r.Type() != "MIRRORLIST" {
		t.Fatalf("Expected 'MIRRORLIST', got %s", r.Type())
	}
}

func TestJSONRenderer_Write(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	r := JSONRenderer{}
	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
		IP:       "127.0.0.1",
	}

	status, err := r.Write(c, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status != http.StatusOK {
		t.Fatalf("Expected 200, got %d", status)
	}
}

func TestJSONRenderer_Write_Pretty(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?pretty", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	r := JSONRenderer{}
	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
	}

	status, err := r.Write(c, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status != http.StatusOK {
		t.Fatalf("Expected 200, got %d", status)
	}
}

func TestRedirectRenderer_Write_NoMirrors(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	r := RedirectRenderer{}
	results := &mirrors.Results{}

	status, err := r.Write(c, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", status)
	}
}

func TestRedirectRenderer_Write_WithMirrors(t *testing.T) {
	SetConfiguration(&Configuration{MaxLinkHeaders: 5})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	r := RedirectRenderer{}
	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
		MirrorList: mirrors.Mirrors{
			mirrors.Mirror{
				ID:      1,
				Name:    "m1",
				HttpURL: "http://m1.example.com",
			},
		},
	}

	status, err := r.Write(c, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status != http.StatusFound {
		t.Fatalf("Expected 302, got %d", status)
	}
}

func TestRedirectRenderer_Write_MultipleMirrors(t *testing.T) {
	SetConfiguration(&Configuration{MaxLinkHeaders: 5})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})

	r := RedirectRenderer{}
	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test/file.txt"},
		MirrorList: mirrors.Mirrors{
			mirrors.Mirror{
				ID:            1,
				Name:          "m1",
				HttpURL:       "http://m1.example.com",
				CountryFields: []string{"FR"},
			},
			mirrors.Mirror{
				ID:            2,
				Name:          "m2",
				HttpURL:       "http://m2.example.com",
				CountryFields: []string{"UK"},
			},
			mirrors.Mirror{
				ID:            3,
				Name:          "m3",
				HttpURL:       "http://m3.example.com",
			},
		},
	}

	status, err := r.Write(c, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status != http.StatusFound {
		t.Fatalf("Expected 302, got %d", status)
	}
}

func TestDefaultEngine_Selection_InvalidURL(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:      1,
			Name:    "bad",
			HttpURL: "ftp://bad.example.com",
			Enabled: true,
			Up:      true,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected mirrors")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror, got %d", len(excluded))
	}
	if excluded[0].ExcludeReason != "Invalid URL" {
		t.Fatalf("Expected 'Invalid URL', got %s", excluded[0].ExcludeReason)
	}
}

func TestDefaultEngine_Selection_Disabled(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:      1,
			Name:    "disabled",
			HttpURL: "http://disabled.example.com",
			Enabled: false,
			Up:      true,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected mirrors")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror")
	}
	if excluded[0].ExcludeReason != "Disabled" {
		t.Fatalf("Expected 'Disabled', got %s", excluded[0].ExcludeReason)
	}
}

func TestDefaultEngine_Selection_Down(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:      1,
			Name:    "down",
			HttpURL: "http://down.example.com",
			Enabled: true,
			Up:      false,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected mirrors")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror")
	}
	if excluded[0].ExcludeReason != "Down" {
		t.Fatalf("Expected 'Down', got %s", excluded[0].ExcludeReason)
	}
}

func TestDefaultEngine_Selection_ValidMirror(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:       1,
			Name:     "valid",
			HttpURL:  "http://valid.example.com",
			Enabled:  true,
			Up:       true,
			Distance: 100,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected mirror, got %d", len(selected))
	}
	if len(excluded) != 0 {
		t.Fatalf("Expected 0 excluded mirrors")
	}
}

func TestDefaultEngine_Selection_ContinentOnly(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	clientInfo := network.GeoIPRecord{
		CountryCode:   "FR",
		ContinentCode: "EU",
	}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:            1,
			Name:          "NA only",
			HttpURL:       "http://na.example.com",
			Enabled:       true,
			Up:            true,
			ContinentOnly: true,
			ContinentCode: "NA",
			Distance:      500,
		},
		mirrors.Mirror{
			ID:            2,
			Name:          "EU only",
			HttpURL:       "http://eu.example.com",
			Enabled:       true,
			Up:            true,
			ContinentOnly: true,
			ContinentCode: "EU",
			Distance:      300,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, clientInfo, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected mirror, got %d", len(selected))
	}
	if selected[0].ID != 2 {
		t.Fatalf("Expected EU mirror to be selected")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror")
	}
}

func TestDefaultEngine_Selection_FileSizeMismatch(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt", Size: 100}

	mirrorInfo := &filesystem.FileInfo{Size: 200}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:       1,
			Name:     "mismatch",
			HttpURL:  "http://mismatch.example.com",
			Enabled:  true,
			Up:       true,
			Distance: 100,
			FileInfo: mirrorInfo,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected mirrors")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror")
	}
	if excluded[0].ExcludeReason != "File size mismatch" {
		t.Fatalf("Expected 'File size mismatch', got %s", excluded[0].ExcludeReason)
	}
}

func TestDefaultEngine_Selection_TLSRestriction(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: true})

	engine := DefaultEngine{}

	ctx := &Context{secureOption: WITHTLS}
	fi := &filesystem.FileInfo{Path: "/test.txt"}

	mirrorsList := mirrors.Mirrors{
		mirrors.Mirror{
			ID:       1,
			Name:     "http-only",
			HttpURL:  "http://http-only.example.com",
			Enabled:  true,
			Up:       true,
			Distance: 100,
		},
	}

	selected, excluded, err := engine.Selection(ctx, fi, network.GeoIPRecord{}, mirrorsList, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected mirrors")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded mirror")
	}
	if excluded[0].ExcludeReason != "Not HTTPS" {
		t.Fatalf("Expected 'Not HTTPS', got %s", excluded[0].ExcludeReason)
	}
}
