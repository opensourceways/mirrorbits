// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package http

import (
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
)

func TestSelection_CountryOnly_Excluded(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU"}

	m := mirrors.Mirror{
		ID:            1,
		Name:          "country-only",
		HttpURL:       "http://mirror.example.com",
		Enabled:       true,
		Up:            true,
		CountryOnly:   true,
		ContinentCode: "NA",
	}
	m.Prepare()

	selected, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected")
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason != "Country only" {
		t.Fatalf("Expected 'Country only' exclusion")
	}
}

func TestSelection_CountryOnly_Match(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU"}

	m := mirrors.Mirror{
		ID:            1,
		Name:          "country-only",
		HttpURL:       "http://mirror.example.com",
		Enabled:       true,
		Up:            true,
		CountryOnly:   true,
		CountryCodes:  "FR DE",
		ContinentCode: "EU",
	}
	m.Prepare()

	selected, _, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected, got %d", len(selected))
	}
}

func TestSelection_ASOnly_Excluded(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR", ASNum: 9999}

	m := mirrors.Mirror{
		ID:      1,
		Name:    "as-only",
		HttpURL: "http://mirror.example.com",
		Enabled: true,
		Up:      true,
		ASOnly:  true,
		Asnum:   1234,
	}
	m.Prepare()

	selected, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected")
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason != "AS only" {
		t.Fatalf("Expected 'AS only' exclusion")
	}
}

func TestSelection_ExcludedCountry(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "CN"}

	m := mirrors.Mirror{
		ID:                   1,
		Name:                 "excluded-country",
		HttpURL:              "http://mirror.example.com",
		Enabled:              true,
		Up:                   true,
		ExcludedCountryCodes: "CN",
	}
	m.Prepare()

	selected, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected")
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason != "User's country restriction" {
		t.Fatalf("Expected 'User's country restriction', got %s", excluded[0].ExcludeReason)
	}
}

func TestSelection_ModTimeMismatch(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	now := time.Now()
	fi := &filesystem.FileInfo{Path: "/test.txt", Size: 100, ModTime: now}
	client := network.GeoIPRecord{CountryCode: "FR"}

	m := mirrors.Mirror{
		ID:      1,
		Name:    "modtime-mismatch",
		HttpURL: "http://mirror.example.com",
		Enabled: true,
		Up:      true,
		FileInfo: &filesystem.FileInfo{
			Size:    100,
			ModTime: now.Add(2 * time.Hour),
		},
	}
	m.Prepare()

	selected, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 0 {
		t.Fatalf("Expected 0 selected")
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason == "" {
		t.Fatalf("Expected exclusion with reason")
	}
}

func TestSelection_InvalidClient_Shuffle(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED, typ: STANDARD}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{}

	mlist := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.example.com", Enabled: true, Up: true, Distance: 100},
		{ID: 2, Name: "m2", HttpURL: "http://m2.example.com", Enabled: true, Up: true, Distance: 200},
		{ID: 3, Name: "m3", HttpURL: "http://m3.example.com", Enabled: true, Up: true, Distance: 300},
	}

	selected, _, err := engine.Selection(ctx, fi, client, mlist, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) == 0 {
		t.Fatalf("Expected some selected mirrors")
	}
	if len(selected) > 5 {
		t.Fatalf("Expected at most 5 selected for non-mirrorlist")
	}
}

func TestSelection_InvalidClient_Mirrorlist(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED, isMirrorList: true}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{}

	mlist := mirrors.Mirrors{
		{ID: 1, Name: "m1", HttpURL: "http://m1.example.com", Enabled: true, Up: true},
		{ID: 2, Name: "m2", HttpURL: "http://m2.example.com", Enabled: true, Up: true},
		{ID: 3, Name: "m3", HttpURL: "http://m3.example.com", Enabled: true, Up: true},
		{ID: 4, Name: "m4", HttpURL: "http://m4.example.com", Enabled: true, Up: true},
		{ID: 5, Name: "m5", HttpURL: "http://m5.example.com", Enabled: true, Up: true},
		{ID: 6, Name: "m6", HttpURL: "http://m6.example.com", Enabled: true, Up: true},
	}

	selected, _, err := engine.Selection(ctx, fi, client, mlist, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 6 {
		t.Fatalf("Expected all 6 selected for mirrorlist, got %d", len(selected))
	}
}

func TestSelection_ScoreComputation(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED, typ: STANDARD}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU", ASNum: 1234}

	mlist := mirrors.Mirrors{
		{ID: 1, Name: "near", HttpURL: "http://near.example.com", Enabled: true, Up: true, Distance: 50, Score: 10, ContinentCode: "EU", CountryCodes: "FR", Asnum: 1234},
		{ID: 2, Name: "far", HttpURL: "http://far.example.com", Enabled: true, Up: true, Distance: 500, Score: 5, ContinentCode: "EU", CountryCodes: "DE", Asnum: 5678},
	}
	for i := range mlist {
		mlist[i].Prepare()
	}

	selected, _, err := engine.Selection(ctx, fi, client, mlist, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 2 {
		t.Fatalf("Expected 2 selected, got %d", len(selected))
	}
}

func TestSelection_DownWithReason(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR"}

	m := mirrors.Mirror{
		ID:            1,
		Name:          "down-with-reason",
		HttpURL:       "http://mirror.example.com",
		Enabled:       true,
		Up:            false,
		ExcludeReason: "maintenance",
	}
	m.Prepare()

	_, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(excluded) != 1 {
		t.Fatalf("Expected 1 excluded")
	}
	if excluded[0].ExcludeReason != "maintenance" {
		t.Fatalf("Expected 'maintenance', got %s", excluded[0].ExcludeReason)
	}
}

func TestSelection_FileSizeMismatch(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	fi := &filesystem.FileInfo{Path: "/test.txt", Size: 100}
	client := network.GeoIPRecord{CountryCode: "FR"}

	m := mirrors.Mirror{
		ID:       1,
		Name:     "size-mismatch",
		HttpURL:  "http://mirror.example.com",
		Enabled:  true,
		Up:       true,
		FileInfo: &filesystem.FileInfo{Size: 200},
	}
	m.Prepare()

	_, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason != "File size mismatch" {
		t.Fatalf("Expected 'File size mismatch'")
	}
}

func TestSelection_ModTimeMatch(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: false, FixTimezoneOffsets: false})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: UNDEFINED}
	now := time.Now().Truncate(time.Second)
	fi := &filesystem.FileInfo{Path: "/test.txt", Size: 100, ModTime: now}
	client := network.GeoIPRecord{CountryCode: "FR"}

	m := mirrors.Mirror{
		ID:       1,
		Name:     "modtime-match",
		HttpURL:  "http://mirror.example.com",
		Enabled:  true,
		Up:       true,
		FileInfo: &filesystem.FileInfo{Size: 100, ModTime: now},
	}
	m.Prepare()

	selected, _, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(selected) != 1 {
		t.Fatalf("Expected 1 selected, got %d", len(selected))
	}
}

func TestSelection_TLSRestriction_HTTP(t *testing.T) {
	SetConfiguration(&Configuration{SchemaStrictMatch: true})
	engine := DefaultEngine{}
	ctx := &Context{secureOption: WITHOUTTLS}
	fi := &filesystem.FileInfo{Path: "/test.txt"}
	client := network.GeoIPRecord{CountryCode: "FR"}

	m := mirrors.Mirror{
		ID:      1,
		Name:    "https-mirror",
		HttpURL: "https://mirror.example.com",
		Enabled: true,
		Up:      true,
	}
	m.Prepare()

	_, excluded, err := engine.Selection(ctx, fi, client, mirrors.Mirrors{m}, GetConfig())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(excluded) != 1 || excluded[0].ExcludeReason != "Not HTTP" {
		t.Fatalf("Expected 'Not HTTP' exclusion, got %s", excluded[0].ExcludeReason)
	}
}
