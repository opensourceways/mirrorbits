// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package utils

import (
	"testing"
	"time"
)

func TestFuzzyTimeStr_UpToDate(t *testing.T) {
	r := FuzzyTimeStr(0)
	if r != "up-to-date" {
		t.Fatalf("Expected 'up-to-date', got %s", r)
	}
}

func TestFuzzyTimeStr_Future(t *testing.T) {
	r := FuzzyTimeStr(-5 * time.Minute)
	if r != "in the future" {
		t.Fatalf("Expected 'in the future', got %s", r)
	}
}

func TestFuzzyTimeStr_Minutes(t *testing.T) {
	r := FuzzyTimeStr(90 * time.Second)
	if r != "1 minute ago" {
		t.Fatalf("Expected '1 minute ago', got %s", r)
	}

	r = FuzzyTimeStr(2 * time.Minute)
	if r != "2 minutes ago" {
		t.Fatalf("Expected '2 minutes ago', got %s", r)
	}
}

func TestFuzzyTimeStr_Hours(t *testing.T) {
	r := FuzzyTimeStr(1 * time.Hour)
	if r != "1 hour ago" {
		t.Fatalf("Expected '1 hour ago', got %s", r)
	}

	r = FuzzyTimeStr(2 * time.Hour)
	if r != "2 hours ago" {
		t.Fatalf("Expected '2 hours ago', got %s", r)
	}
}

func TestFuzzyTimeStr_Days(t *testing.T) {
	r := FuzzyTimeStr(24 * time.Hour)
	if r != "1 day ago" {
		t.Fatalf("Expected '1 day ago', got %s", r)
	}

	r = FuzzyTimeStr(48 * time.Hour)
	if r != "2 days ago" {
		t.Fatalf("Expected '2 days ago', got %s", r)
	}
}

func TestFuzzyTimeStr_Years(t *testing.T) {
	r := FuzzyTimeStr(365 * 24 * 2 * time.Hour)
	if r != "2 years ago" {
		t.Fatalf("Expected '2 years ago', got %s", r)
	}
}

func TestSanitizeLocationCodes(t *testing.T) {
	r := SanitizeLocationCodes("fr, de, uk")
	if r != "FR DE UK" {
		t.Fatalf("Expected 'FR DE UK', got %s", r)
	}

	r = SanitizeLocationCodes("CN")
	if r != "CN" {
		t.Fatalf("Expected 'CN', got %s", r)
	}

	r = SanitizeLocationCodes("")
	if r != "" {
		t.Fatalf("Expected '', got %s", r)
	}

	r = SanitizeLocationCodes("  us  jp  ")
	if r != "US JP" {
		t.Fatalf("Expected 'US JP', got %s", r)
	}
}

func TestFormattedDateUTC(t *testing.T) {
	tm := time.Date(2020, 1, 15, 12, 30, 45, 0, time.UTC)
	r := FormattedDateUTC(tm)

	if r == "" {
		t.Fatalf("Expected non-empty result")
	}
}

func TestNormalizeURL_EdgeCases(t *testing.T) {
	r := NormalizeURL("")
	if r != "" {
		t.Fatalf("Expected '', got %s", r)
	}

	r = NormalizeURL("http://test.com")
	if r != "http://test.com/" {
		t.Fatalf("Expected 'http://test.com/', got %s", r)
	}
}

func TestReadableSize_EdgeCases(t *testing.T) {
	r := ReadableSize(1073741824)
	if r == "" {
		t.Fatalf("Expected non-empty result for 1GB")
	}

	r = ReadableSize(1099511627776)
	if r == "" {
		t.Fatalf("Expected non-empty result for 1TB")
	}

	r = ReadableSize(0)
	if r != "0.0 B" {
		t.Fatalf("Expected '0.0 B', got %s", r)
	}
}

func TestConcatURL_EdgeCases(t *testing.T) {
	r := ConcatURL("http://test.com/", "/file.txt")
	if r != "http://test.com/file.txt" {
		t.Fatalf("Expected 'http://test.com/file.txt', got %s", r)
	}
}

func TestGetDistanceKm_SamePoint(t *testing.T) {
	r := GetDistanceKm(0, 0, 0, 0)
	if r != 0 {
		t.Fatalf("Expected 0 for same point, got %f", r)
	}
}

func TestMin_Equal(t *testing.T) {
	if r := Min(5, 5); r != 5 {
		t.Fatalf("Expected 5, got %d", r)
	}
}

func TestMax_Equal(t *testing.T) {
	if r := Max(5, 5); r != 5 {
		t.Fatalf("Expected 5, got %d", r)
	}
}

func TestMax_Greater(t *testing.T) {
	if r := Max(10, 5); r != 10 {
		t.Fatalf("Expected 10, got %d", r)
	}
}

func TestIsInSlice_Empty(t *testing.T) {
	if IsInSlice("a", []string{}) {
		t.Fatalf("Expected false for empty list")
	}
}

func TestElapsedSec_Boundary(t *testing.T) {
	now := time.Now().UTC().Unix()

	if !ElapsedSec(now-200, 100) {
		t.Fatalf("Expected true for elapsed time")
	}

	if ElapsedSec(now, 100) {
		t.Fatalf("Expected false for not yet elapsed")
	}
}

func TestPlural_Float(t *testing.T) {
	r := Plural(1.5)
	if r != "" {
		t.Fatalf("Expected '' for float, got %s", r)
	}

	r = Plural("test")
	if r != "" {
		t.Fatalf("Expected '' for string, got %s", r)
	}
}
