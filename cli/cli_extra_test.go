// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package cli

import (
	"testing"

	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/rpc"
)

func TestGetSingle_Empty(t *testing.T) {
	_, _, err := GetSingle(nil)
	if err == nil {
		t.Fatalf("Expected error for nil list")
	}
}

func TestGetSingle_SingleItem(t *testing.T) {
	id, name, err := GetSingle([]*rpc.MirrorID{{ID: 5, Name: "test"}})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if id != 5 {
		t.Fatalf("Expected id 5, got %d", id)
	}
	if name != "test" {
		t.Fatalf("Expected name 'test', got %s", name)
	}
}

func TestGetSingle_TooMany(t *testing.T) {
	_, _, err := GetSingle([]*rpc.MirrorID{{ID: 1}, {ID: 2}})
	if err == nil {
		t.Fatalf("Expected error for too many results")
	}
}

func TestCompareAndUpdate_NoChange(t *testing.T) {
	m := &mirrors.Mirror{
		HttpURL:          "http://example.com",
		RsyncURL:         "rsync://example.com",
		FtpURL:           "ftp://example.com",
		SponsorName:      "sponsor",
		SponsorURL:       "http://sponsor.url",
		AdminName:        "admin",
		AdminEmail:       "admin@example.com",
		ContinentOnly:    true,
		CountryOnly:      false,
		ASOnly:           true,
		Score:            10,
		Enabled:          true,
		SponsorLogoURL:   "http://logo.url",
		NetworkBandwidth: 1000,
		Latitude:         48.85,
		Longitude:        2.35,
		Country:          "France",
	}

	u := &mirrors.Mirror{
		HttpURL:          "http://example.com",
		RsyncURL:         "rsync://example.com",
		FtpURL:           "ftp://example.com",
		SponsorName:      "sponsor",
		SponsorURL:       "http://sponsor.url",
		AdminName:        "admin",
		AdminEmail:       "admin@example.com",
		ContinentOnly:    true,
		CountryOnly:      false,
		ASOnly:           true,
		Score:            10,
		Enabled:          true,
		SponsorLogoURL:   "http://logo.url",
		NetworkBandwidth: 1000,
		Latitude:         48.85,
		Longitude:        2.35,
		Country:          "France",
	}

	changed := CompareAndUpdate(m, u)
	if changed {
		t.Fatalf("Expected false for identical mirrors")
	}
}

func TestCompareAndUpdate_WithChange(t *testing.T) {
	m := &mirrors.Mirror{
		HttpURL: "http://old.example.com",
		Score:   10,
	}

	u := &mirrors.Mirror{
		HttpURL: "http://new.example.com",
		Score:   20,
	}

	changed := CompareAndUpdate(m, u)
	if !changed {
		t.Fatalf("Expected true for changed mirrors")
	}
	if m.HttpURL != "http://new.example.com" {
		t.Fatalf("Expected HttpURL updated")
	}
	if m.Score != 20 {
		t.Fatalf("Expected Score updated to 20")
	}
}

func TestLoginCreds_GetRequestMetadata(t *testing.T) {
	creds := &loginCreds{Password: "secret"}
	md, err := creds.GetRequestMetadata(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if md["password"] != "secret" {
		t.Fatalf("Expected password 'secret'")
	}
}

func TestLoginCreds_RequireTransportSecurity(t *testing.T) {
	creds := &loginCreds{}
	if creds.RequireTransportSecurity() {
		t.Fatalf("Expected false for RequireTransportSecurity")
	}
}

func TestParseCommands_Help(t *testing.T) {
	err := ParseCommands("help")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestParseCommands_UnknownCommand(t *testing.T) {
	err := ParseCommands("nonexistentcmd")
	_ = err
}

func TestParseCommands_NoArgs(t *testing.T) {
	err := ParseCommands()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
