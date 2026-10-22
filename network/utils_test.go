// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package network

import (
	"testing"
)

func TestRemoteIPFromAddr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1:1234", "10.0.0.1"},
		{"127.0.0.1:8080", "127.0.0.1"},
		{"[::1]:8080", "[::1]"},
	}
	for _, tt := range tests {
		if r := RemoteIPFromAddr(tt.input); r != tt.expected {
			t.Fatalf("Input %q: expected %q, got %q", tt.input, tt.expected, r)
		}
	}
}

func TestRemoteIPFromAddrEmpty(t *testing.T) {
	r := RemoteIPFromAddr(":8080")
	if r != "" {
		t.Fatalf("Expected empty, got %s", r)
	}
}

func TestExtractRemoteIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3.4, 5.6.7.8", "1.2.3.4"},
		{"1.2.3.4", "1.2.3.4"},
		{"  1.2.3.4  , 5.6.7.8", "1.2.3.4"},
		{"", ""},
	}
	for _, tt := range tests {
		if r := ExtractRemoteIP(tt.input); r != tt.expected {
			t.Fatalf("Input %q: expected %q, got %q", tt.input, tt.expected, r)
		}
	}
}

func TestNewGeoIPNotNil(t *testing.T) {
	g := NewGeoIP()
	if g == nil {
		t.Fatal("Expected non-nil GeoIP")
	}
}

func TestGeoIPGetRecordInvalidIP(t *testing.T) {
	g := NewGeoIP()
	r := g.GetRecord("not-an-ip")
	if r.CountryCode != "" {
		t.Fatalf("Expected empty CountryCode for invalid IP, got %s", r.CountryCode)
	}
}

func TestGeoIPGetRecordNoDB(t *testing.T) {
	g := NewGeoIP()
	r := g.GetRecord("127.0.0.1")
	if r.CountryCode != "" {
		t.Fatalf("Expected empty CountryCode with no DB loaded, got %s", r.CountryCode)
	}
	if r.City != "" {
		t.Fatalf("Expected empty City with no DB loaded, got %s", r.City)
	}
	if r.Latitude != 0 {
		t.Fatalf("Expected 0 Latitude with no DB, got %f", r.Latitude)
	}
}

func TestGeoIPErrorIsFatal(t *testing.T) {
	e := GeoIPError{}
	if !e.IsFatal() {
		t.Fatal("Expected IsFatal to be true when loaded == 0 == len(Errors)")
	}
	e.loaded = 1
	e.Errors = []error{nil, nil}
	if e.IsFatal() {
		t.Fatal("Expected IsFatal to be false when loaded < len(Errors)")
	}
}

func TestGeoIPErrorString(t *testing.T) {
	e := GeoIPError{}
	if e.Error() != "One or more GeoIP database could not be loaded" {
		t.Fatalf("Unexpected error message: %s", e.Error())
	}
}

func TestGeoIPRecordIsValidEmpty(t *testing.T) {
	var r GeoIPRecord
	if r.IsValid() {
		t.Fatal("Expected empty record to be invalid")
	}
}

func TestGeoIPRecordIsValidWithCountryCode(t *testing.T) {
	r := GeoIPRecord{CountryCode: "US"}
	if !r.IsValid() {
		t.Fatal("Expected record with CountryCode to be valid")
	}
}
