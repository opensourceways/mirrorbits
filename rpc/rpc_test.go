// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func timestampNow() *timestamp.Timestamp {
	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	return ts
}

func TestMirrorToRPC(t *testing.T) {
	now := time.Now().UTC()
	m := &mirrors.Mirror{
		ID:             1,
		Name:           "test-mirror",
		HttpURL:        "https://example.com",
		RsyncURL:       "rsync://example.com",
		FtpURL:         "ftp://example.com",
		SponsorName:    "sponsor",
		SponsorURL:     "http://sponsor.url",
		SponsorLogoURL: "http://sponsor.logo",
		AdminName:      "admin",
		AdminEmail:     "admin@example.com",
		CustomData:     "custom",
		ContinentOnly:  true,
		CountryOnly:    false,
		ASOnly:         true,
		Score:          10,
		Latitude:       48.85,
		Longitude:      2.35,
		ContinentCode:  "EU",
		CountryCodes:   "FR DE",
		Asnum:          1234,
		Comment:        "comment",
		Enabled:        true,
		Up:             true,
		ExcludeReason:  "none",
		StateSince:     mirrors.Time{}.FromTime(now),
		LastSync:       mirrors.Time{}.FromTime(now),
		LastSuccessfulSync: mirrors.Time{}.FromTime(now),
		LastModTime:    mirrors.Time{}.FromTime(now),
		NetworkBandwidth: 1000,
	}

	rpcMirror, err := MirrorToRPC(m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rpcMirror == nil {
		t.Fatalf("Expected non-nil result")
	}
	if rpcMirror.ID != 1 {
		t.Fatalf("Expected ID 1, got %d", rpcMirror.ID)
	}
	if rpcMirror.Name != "test-mirror" {
		t.Fatalf("Expected Name 'test-mirror', got %s", rpcMirror.Name)
	}
	if rpcMirror.HttpURL != "https://example.com" {
		t.Fatalf("Expected HttpURL 'https://example.com', got %s", rpcMirror.HttpURL)
	}
	if rpcMirror.ContinentOnly != true {
		t.Fatalf("Expected ContinentOnly true")
	}
	if rpcMirror.Score != 10 {
		t.Fatalf("Expected Score 10, got %d", rpcMirror.Score)
	}
	if rpcMirror.Latitude != 48.85 {
		t.Fatalf("Expected Latitude 48.85, got %f", rpcMirror.Latitude)
	}
	if rpcMirror.Asnum != 1234 {
		t.Fatalf("Expected Asnum 1234, got %d", rpcMirror.Asnum)
	}
	if rpcMirror.Enabled != true {
		t.Fatalf("Expected Enabled true")
	}
	if rpcMirror.NetworkBandwidth != 1000 {
		t.Fatalf("Expected NetworkBandwidth 1000, got %d", rpcMirror.NetworkBandwidth)
	}
}

func TestMirrorFromRPC(t *testing.T) {
	now := time.Now().UTC()
	rpcMirror := &Mirror{
		ID:             2,
		Name:           "rpc-mirror",
		HttpURL:        "http://example.com",
		RsyncURL:       "rsync://example.com",
		FtpURL:         "ftp://example.com",
		SponsorName:    "sponsor",
		SponsorURL:     "http://sponsor.url",
		SponsorLogoURL: "http://sponsor.logo",
		AdminName:      "admin",
		AdminEmail:     "admin@example.com",
		CustomData:     "custom",
		ContinentOnly:  false,
		CountryOnly:    true,
		ASOnly:         false,
		Score:          20,
		Latitude:       51.50,
		Longitude:      -0.12,
		ContinentCode:  "NA",
		CountryCodes:   "UK US",
		Asnum:          5678,
		Comment:        "rpc comment",
		Enabled:        false,
		Up:             false,
		ExcludeReason:  "down",
		NetworkBandwidth: 500,
		StateSince:           timestampNow(),
		LastSync:             timestampNow(),
		LastSuccessfulSync:   timestampNow(),
		LastModTime:          timestampNow(),
	}

	m, err := MirrorFromRPC(rpcMirror)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if m == nil {
		t.Fatalf("Expected non-nil result")
	}
	if m.ID != 2 {
		t.Fatalf("Expected ID 2, got %d", m.ID)
	}
	if m.Name != "rpc-mirror" {
		t.Fatalf("Expected Name 'rpc-mirror', got %s", m.Name)
	}
	if m.HttpURL != "http://example.com" {
		t.Fatalf("Expected HttpURL 'http://example.com', got %s", m.HttpURL)
	}
	if m.CountryOnly != true {
		t.Fatalf("Expected CountryOnly true")
	}
	if m.Score != 20 {
		t.Fatalf("Expected Score 20, got %d", m.Score)
	}
	if m.Latitude != 51.50 {
		t.Fatalf("Expected Latitude 51.50, got %f", m.Latitude)
	}
	if m.Asnum != 5678 {
		t.Fatalf("Expected Asnum 5678, got %d", m.Asnum)
	}
	if m.Enabled != false {
		t.Fatalf("Expected Enabled false")
	}
	if m.NetworkBandwidth != 500 {
		t.Fatalf("Expected NetworkBandwidth 500, got %d", m.NetworkBandwidth)
	}

	_ = now // keep reference
}

func TestMirrorToRPC_RoundTrip(t *testing.T) {
	m := &mirrors.Mirror{
		ID:           42,
		Name:         "roundtrip",
		HttpURL:      "https://roundtrip.example.com",
		ContinentCode: "AS",
		CountryCodes:  "JP KR",
		Enabled:       true,
		Up:            true,
		NetworkBandwidth: 2000,
	}

	rpcMirror, err := MirrorToRPC(m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	m2, err := MirrorFromRPC(rpcMirror)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if m2.ID != m.ID {
		t.Fatalf("ID mismatch: %d != %d", m2.ID, m.ID)
	}
	if m2.Name != m.Name {
		t.Fatalf("Name mismatch: %s != %s", m2.Name, m.Name)
	}
	if m2.HttpURL != m.HttpURL {
		t.Fatalf("HttpURL mismatch: %s != %s", m2.HttpURL, m.HttpURL)
	}
	if m2.ContinentCode != m.ContinentCode {
		t.Fatalf("ContinentCode mismatch")
	}
	if m2.CountryCodes != m.CountryCodes {
		t.Fatalf("CountryCodes mismatch")
	}
	if m2.Enabled != m.Enabled {
		t.Fatalf("Enabled mismatch")
	}
	if m2.NetworkBandwidth != m.NetworkBandwidth {
		t.Fatalf("NetworkBandwidth mismatch")
	}
}
