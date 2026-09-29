// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package network

import (
	"strings"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestLookupMirrorIP_Localhost(t *testing.T) {
	ip, err := LookupMirrorIP("localhost")
	if err != nil {
		if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "lookup") {
			t.Skip("DNS resolution not available in test environment")
		}
		if err != ErrMultipleAddresses {
			t.Fatalf("Unexpected error: %s", err)
		}
	}
	if ip == "" {
		t.Fatalf("Expected non-empty IP")
	}
}

func TestLookupMirrorIP_InvalidHost(t *testing.T) {
	_, err := LookupMirrorIP("this-host-does-not-exist.invalid.tld")
	if err == nil {
		t.Fatalf("Expected error for invalid host")
	}
}

func TestExtractRemoteIP_Multiple(t *testing.T) {
	result := ExtractRemoteIP("1.2.3.4, 5.6.7.8, 9.10.11.12")
	if result != "1.2.3.4" {
		t.Fatalf("Expected '1.2.3.4', got %s", result)
	}
}

func TestExtractRemoteIP_WithSpaces(t *testing.T) {
	result := ExtractRemoteIP("  192.168.1.1  , 10.0.0.1")
	if result != "192.168.1.1" {
		t.Fatalf("Expected '192.168.1.1', got %s", result)
	}
}

func TestRemoteIPFromAddr_IPv6(t *testing.T) {
	result := RemoteIPFromAddr("[::1]:8080")
	if result != "[::1]" {
		t.Fatalf("Expected '[::1]', got %s", result)
	}
}

func TestGeoIP_LoadGeoIP_WithFallbacks(t *testing.T) {
	SetConfiguration(&Configuration{
		GeoipDatabasePath: "/nonexistent",
		Fallbacks:         []Fallback{{URL: "http://fallback.example.com"}},
	})
	g := NewGeoIP()
	_ = g.LoadGeoIP()
}

func TestGetRecord_NoDB(t *testing.T) {
	g := NewGeoIP()
	rec := g.GetRecord("")
	if rec.IsValid() {
		t.Fatalf("Expected invalid record for empty IP with no DB")
	}
}
