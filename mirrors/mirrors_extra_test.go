// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"gopkg.in/yaml.v2"
)

func TestMirror_Prepare(t *testing.T) {
	m := Mirror{
		CountryCodes:         "FR UK DE",
		ExcludedCountryCodes: "IT ES",
	}
	m.Prepare()

	if len(m.CountryFields) != 3 {
		t.Fatalf("Expected 3 country fields, got %d", len(m.CountryFields))
	}
	if m.CountryFields[0] != "FR" {
		t.Fatalf("Expected first country 'FR', got %s", m.CountryFields[0])
	}
	if len(m.ExcludedCountryFields) != 2 {
		t.Fatalf("Expected 2 excluded country fields, got %d", len(m.ExcludedCountryFields))
	}
	if m.ExcludedCountryFields[0] != "IT" {
		t.Fatalf("Expected first excluded country 'IT', got %s", m.ExcludedCountryFields[0])
	}
}

func TestMirror_Prepare_Empty(t *testing.T) {
	m := Mirror{}
	m.Prepare()

	if len(m.CountryFields) != 0 {
		t.Fatalf("Expected 0 country fields for empty mirror")
	}
}

func TestMirror_IsHTTPS(t *testing.T) {
	m := Mirror{HttpURL: "https://example.com"}
	if !m.IsHTTPS() {
		t.Fatalf("Expected true for HTTPS URL")
	}

	m = Mirror{HttpURL: "http://example.com"}
	if m.IsHTTPS() {
		t.Fatalf("Expected false for HTTP URL")
	}

	m = Mirror{HttpURL: ""}
	if m.IsHTTPS() {
		t.Fatalf("Expected false for empty URL")
	}
}

func TestRedirects_Allowed(t *testing.T) {
	SetConfiguration(&Configuration{DisallowRedirects: false})

	var r Redirects = 1
	if !r.Allowed() {
		t.Fatalf("Expected true for Redirects=1 (allow)")
	}

	r = 2
	if r.Allowed() {
		t.Fatalf("Expected false for Redirects=2 (deny)")
	}

	r = 0
	if !r.Allowed() {
		t.Fatalf("Expected true for Redirects=0 (default, DisallowRedirects=false)")
	}

	SetConfiguration(&Configuration{DisallowRedirects: true})
	if r.Allowed() {
		t.Fatalf("Expected false for Redirects=0 when DisallowRedirects=true")
	}
}

func TestRedirects_MarshalYAML(t *testing.T) {
	r := Redirects(1)
	out, err := r.MarshalYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	b, ok := out.(*bool)
	if !ok || !*b {
		t.Fatalf("Expected true, got %v", out)
	}

	r = Redirects(2)
	out, err = r.MarshalYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	b, ok = out.(*bool)
	if !ok || *b {
		t.Fatalf("Expected false, got %v", out)
	}

	r = Redirects(0)
	out, err = r.MarshalYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if b, ok := out.(*bool); ok && b != nil {
		t.Fatalf("Expected nil pointer for default, got non-nil")
	}
}

func TestRedirects_UnmarshalYAML(t *testing.T) {
	var r Redirects

	err := yaml.Unmarshal([]byte("true"), &r)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 1 {
		t.Fatalf("Expected 1 for true, got %d", r)
	}

	err = yaml.Unmarshal([]byte("false"), &r)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 2 {
		t.Fatalf("Expected 2 for false, got %d", r)
	}

	err = yaml.Unmarshal([]byte("null"), &r)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 0 {
		t.Fatalf("Expected 0 for null, got %d", r)
	}
}

func TestTime_RedisArg(t *testing.T) {
	now := time.Now().UTC()
	tm := Time{Time: now}

	result := tm.RedisArg()
	if result.(int64) != now.Unix() {
		t.Fatalf("Expected %d, got %d", now.Unix(), result.(int64))
	}
}

func TestTime_RedisScan(t *testing.T) {
	var tm Time

	err := tm.RedisScan(int64(1609459200))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if tm.Unix() != 1609459200 {
		t.Fatalf("Expected 1609459200, got %d", tm.Unix())
	}

	tm = Time{}
	err = tm.RedisScan([]byte("1609459200"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if tm.Unix() != 1609459200 {
		t.Fatalf("Expected 1609459200, got %d", tm.Unix())
	}

	tm = Time{}
	err = tm.RedisScan("invalid")
	if err == nil {
		t.Fatalf("Expected error for invalid type")
	}
}

func TestTime_FromTime(t *testing.T) {
	now := time.Now().UTC()
	tm := Time{}
	result := tm.FromTime(now)

	if !result.Time.Equal(now) {
		t.Fatalf("Expected %v, got %v", now, result.Time)
	}
}
