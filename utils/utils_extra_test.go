// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package utils

import (
	"testing"
	"time"
)

func TestFormattedDateUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	tests := []struct {
		input    time.Time
		expected string
	}{
		{time.Date(2023, 6, 15, 12, 30, 45, 0, time.UTC), "Thu, 15 Jun 2023 12:30:45 UTC"},
		{time.Date(2023, 1, 1, 0, 0, 0, 0, loc), "Sun, 01 Jan 2023 05:00:00 UTC"},
	}
	for _, tt := range tests {
		if r := FormattedDateUTC(tt.input); r != tt.expected {
			t.Fatalf("Expected %q, got %q", tt.expected, r)
		}
	}
}

func TestFuzzyTimeStr(t *testing.T) {
	tests := []struct {
		dur      time.Duration
		expected string
	}{
		{0, "up-to-date"},
		{-5 * time.Minute, "in the future"},
		{30 * time.Minute, "30 minutes ago"},
		{1 * time.Minute, "1 minute ago"},
		{2 * time.Hour, "2 hours ago"},
		{1 * time.Hour, "1 hour ago"},
		{48 * time.Hour, "2 days ago"},
		{365 * 24 * 2 * time.Hour, "2 years ago"},
	}
	for _, tt := range tests {
		r := FuzzyTimeStr(tt.dur)
		if tt.expected != "" && r != tt.expected {
			t.Fatalf("Expected %q, got %q", tt.expected, r)
		}
	}
}

func TestSanitizeLocationCodes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fr,de,uk", "FR DE UK"},
		{"fr de uk", "FR DE UK"},
		{"FR", "FR"},
		{"", ""},
		{"fr, de , uk", "FR DE UK"},
	}
	for _, tt := range tests {
		if r := SanitizeLocationCodes(tt.input); r != tt.expected {
			t.Fatalf("Input %q: expected %q, got %q", tt.input, tt.expected, r)
		}
	}
}

func TestReadableSizeLargeValues(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TiB"},
		{1024*1024*1024*1024 + 512*1024*1024*1024, "1.5 TiB"},
	}
	for _, tt := range tests {
		if r := ReadableSize(tt.input); r != tt.expected {
			t.Fatalf("Expected %q, got %q", tt.expected, r)
		}
	}
}

func TestConcatURLEdgeCases(t *testing.T) {
	if r := ConcatURL("http://test/", "/file.bin"); r != "http://test/file.bin" {
		t.Fatalf("Expected http://test/file.bin, got %s", r)
	}
	if r := ConcatURL("http://test", "file.bin"); r != "http://test/file.bin" {
		t.Fatalf("Expected http://test/file.bin, got %s", r)
	}
	if r := ConcatURL("http://test/", "file.bin"); r != "http://test/file.bin" {
		t.Fatalf("Expected http://test/file.bin, got %s", r)
	}
	if r := ConcatURL("http://test", "/file.bin"); r != "http://test/file.bin" {
		t.Fatalf("Expected http://test/file.bin, got %s", r)
	}
}

func TestNormalizeURLEmpty(t *testing.T) {
	if r := NormalizeURL(""); r != "" {
		t.Fatalf("Expected empty string, got %s", r)
	}
}

func TestPluralFloat(t *testing.T) {
	if r := Plural(1.5); r != "" {
		t.Fatalf("Expected empty for float, got %q", r)
	}
}

func TestGetDistanceSamePoint(t *testing.T) {
	if r := GetDistanceKm(0, 0, 0, 0); r != 0 {
		t.Fatalf("Expected 0, got %f", r)
	}
}
