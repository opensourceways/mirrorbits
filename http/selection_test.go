// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
)

func TestDefaultEngineSelectionBasic(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:            "/tmp/test",
		SchemaStrictMatch:     false,
		WeightDistributionRange: 1.5,
	})

	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{
		Path:    "/test.txt",
		Size:    1000,
		ModTime: time.Now(),
	}

	clientInfo := network.GeoIPRecord{
		CountryCode:   "FR",
		ContinentCode: "EU",
		Latitude:      48.85,
		Longitude:     2.35,
	}

	pMirrors := mirrors.Mirrors{
		{
			ID:            1,
			Name:          "m1",
			HttpURL:       "http://m1.example.com/",
			Enabled:       true,
			Up:            true,
			Latitude:      48.85,
			Longitude:     2.35,
			Distance:      0,
		},
		{
			ID:            2,
			Name:          "m2",
			HttpURL:       "http://m2.example.com/",
			Enabled:       true,
			Up:            true,
			Latitude:      51.50,
			Longitude:     -0.12,
			Distance:      334,
		},
		{
			ID:            3,
			Name:          "m3",
			HttpURL:       "invalid-url",
			Enabled:       true,
			Up:            true,
		},
	}

	engine := DefaultEngine{}
	selected, excluded, err := engine.Selection(ctx, fileInfo, clientInfo, pMirrors, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 2 {
		t.Fatalf("Expected 2 selected, got %d", len(selected))
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded, got %d", len(excluded))
	}
}

func TestDefaultEngineSelectionDisabledMirror(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:        "/tmp/test",
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{Path: "/test.txt", Size: 1000}

	pMirrors := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.com/", Enabled: false, Up: true},
	}

	engine := DefaultEngine{}
	selected, excluded, _ := engine.Selection(ctx, fileInfo, network.GeoIPRecord{}, pMirrors, GetConfig())
	if len(selected) != 0 {
		t.Fatal("Expected 0 selected for disabled mirror")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded, got %d", len(excluded))
	}
}

func TestDefaultEngineSelectionDownMirror(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:        "/tmp/test",
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{Path: "/test.txt", Size: 1000}

	pMirrors := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.com/", Enabled: true, Up: false},
	}

	engine := DefaultEngine{}
	selected, excluded, _ := engine.Selection(ctx, fileInfo, network.GeoIPRecord{}, pMirrors, GetConfig())
	if len(selected) != 0 {
		t.Fatal("Expected 0 selected for down mirror")
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded, got %d", len(excluded))
	}
}

func TestDefaultEngineSelectionNoClientInfo(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:        "/tmp/test",
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test.txt?mirrorlist", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{Path: "/test.txt", Size: 1000}

	pMirrors := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.com/", Enabled: true, Up: true, Distance: 100},
		{ID: 2, Name: "m2", HttpURL: "http://m2.com/", Enabled: true, Up: true, Distance: 200},
		{ID: 3, Name: "m3", HttpURL: "http://m3.com/", Enabled: true, Up: true, Distance: 300},
	}

	engine := DefaultEngine{}
	selected, _, err := engine.Selection(ctx, fileInfo, network.GeoIPRecord{}, pMirrors, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 3 {
		t.Fatalf("Expected 3 selected, got %d", len(selected))
	}
}

func TestDefaultEngineSelectionContinentOnly(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:        "/tmp/test",
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{Path: "/test.txt", Size: 1000}

	clientInfo := network.GeoIPRecord{
		CountryCode:   "FR",
		ContinentCode: "EU",
		Latitude:      48.85,
		Longitude:     2.35,
	}

	pMirrors := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.com/", Enabled: true, Up: true, ContinentOnly: true, ContinentCode: "AS", Distance: 100},
		{ID: 2, Name: "m2", HttpURL: "http://m2.com/", Enabled: true, Up: true, ContinentOnly: true, ContinentCode: "EU", Distance: 200},
	}

	engine := DefaultEngine{}
	selected, excluded, _ := engine.Selection(ctx, fileInfo, clientInfo, pMirrors, GetConfig())
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected, got %d", len(selected))
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded, got %d", len(excluded))
	}
}

func TestDefaultEngineSelectionCountryOnly(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:        "/tmp/test",
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, Templates{})

	fileInfo := &filesystem.FileInfo{Path: "/test.txt", Size: 1000}

	clientInfo := network.GeoIPRecord{
		CountryCode: "FR",
		Latitude:    48.85,
		Longitude:   2.35,
	}

	m := mirrors.Mirror{
		ID:            1,
		Name:          "m1",
		HttpURL:       "http://m1.com/",
		Enabled:       true,
		Up:            true,
		CountryOnly:   true,
		CountryCodes:  "FR DE",
		Distance:      100,
	}
	m.Prepare()

	pMirrors := mirrors.Mirrors{m}

	engine := DefaultEngine{}
	selected, _, _ := engine.Selection(ctx, fileInfo, clientInfo, pMirrors, GetConfig())
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected, got %d", len(selected))
	}
}

func TestStatsCountDownloadEmptyMirror(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: ""}
	fi := filesystem.FileInfo{Path: "/test.txt", Size: 100}
	err := s.CountDownload(m, fi)
	if err != errUnknownMirror {
		t.Fatalf("Expected errUnknownMirror, got %v", err)
	}
}

func TestStatsCountDownloadEmptyPath(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: "m1"}
	fi := filesystem.FileInfo{Path: "", Size: 100}
	err := s.CountDownload(m, fi)
	if err != errEmptyFileError {
		t.Fatalf("Expected errEmptyFileError, got %v", err)
	}
}

func TestStatsCountDownloadValid(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: "m1", ID: 1}
	fi := filesystem.FileInfo{Path: "/test.txt", Size: 100}
	err := s.CountDownload(m, fi)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
