// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"errors"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	. "github.com/opensourceways/mirrorbits/testing"
)

func TestMirrorPrepare(t *testing.T) {
	m := Mirror{CountryCodes: "FR DE UK", ExcludedCountryCodes: "US CN"}
	m.Prepare()
	if len(m.CountryFields) != 3 {
		t.Fatalf("Expected 3 country fields, got %d", len(m.CountryFields))
	}
	if m.CountryFields[0] != "FR" {
		t.Fatalf("Expected FR, got %s", m.CountryFields[0])
	}
	if len(m.ExcludedCountryFields) != 2 {
		t.Fatalf("Expected 2 excluded country fields, got %d", len(m.ExcludedCountryFields))
	}
}

func TestMirrorPrepareEmpty(t *testing.T) {
	m := Mirror{}
	m.Prepare()
	if len(m.CountryFields) != 0 {
		t.Fatal("Expected 0 country fields")
	}
}

func TestMirrorIsHTTPS(t *testing.T) {
	m := Mirror{HttpURL: "https://example.com"}
	if !m.IsHTTPS() {
		t.Fatal("Expected true for https URL")
	}
	m.HttpURL = "http://example.com"
	if m.IsHTTPS() {
		t.Fatal("Expected false for http URL")
	}
	m.HttpURL = "ftp://example.com"
	if m.IsHTTPS() {
		t.Fatal("Expected false for ftp URL")
	}
}

func TestRedirectsAllowed(t *testing.T) {
	SetConfiguration(&Configuration{DisallowRedirects: false})

	r1 := Redirects(1)
	if !r1.Allowed() {
		t.Fatal("Expected true for Redirects(1)")
	}
	r2 := Redirects(2)
	if r2.Allowed() {
		t.Fatal("Expected false for Redirects(2)")
	}
	r0 := Redirects(0)
	if !r0.Allowed() {
		t.Fatal("Expected true for Redirects(0) when DisallowRedirects is false")
	}

	SetConfiguration(&Configuration{DisallowRedirects: true})
	if r0.Allowed() {
		t.Fatal("Expected false for Redirects(0) when DisallowRedirects is true")
	}
}

func TestRedirectsMarshalYAML(t *testing.T) {
	r := Redirects(1)
	v, err := r.MarshalYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	bp, ok := v.(*bool)
	if !ok || bp == nil {
		t.Fatal("Expected *bool")
	}
	if !*bp {
		t.Fatal("Expected true")
	}

	r = Redirects(2)
	v, _ = r.MarshalYAML()
	bp, _ = v.(*bool)
	if *bp {
		t.Fatal("Expected false")
	}

	r = Redirects(0)
	v, _ = r.MarshalYAML()
	if v == nil {
		t.Fatal("Expected non-nil interface")
	}
	bp, ok = v.(*bool)
	if !ok {
		t.Fatal("Expected *bool")
	}
	if bp != nil {
		t.Fatal("Expected nil pointer for Redirects(0)")
	}
}

func TestRedirectsUnmarshalYAML(t *testing.T) {
	var r Redirects

	err := r.UnmarshalYAML(func(v interface{}) error {
		bp := v.(*(*bool))
		*bp = nil
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 0 {
		t.Fatalf("Expected 0, got %d", r)
	}

	err = r.UnmarshalYAML(func(v interface{}) error {
		bp := v.(*(*bool))
		t := true
		*bp = &t
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 1 {
		t.Fatalf("Expected 1, got %d", r)
	}

	err = r.UnmarshalYAML(func(v interface{}) error {
		bp := v.(*(*bool))
		f := false
		*bp = &f
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if r != 2 {
		t.Fatalf("Expected 2, got %d", r)
	}
}

func TestTimeRedisArg(t *testing.T) {
	tm := Time{Time: time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)}
	v := tm.RedisArg()
	if vi, ok := v.(int64); !ok {
		t.Fatalf("Expected int64, got %T", v)
	} else if vi != tm.UTC().Unix() {
		t.Fatalf("Expected %d, got %d", tm.UTC().Unix(), vi)
	}
}

func TestTimeRedisScan(t *testing.T) {
	var tm Time

	err := tm.RedisScan(int64(1686830400))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if tm.Unix() != 1686830400 {
		t.Fatalf("Expected 1686830400, got %d", tm.Unix())
	}

	err = tm.RedisScan([]byte("1686830400"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if tm.Unix() != 1686830400 {
		t.Fatalf("Expected 1686830400, got %d", tm.Unix())
	}

	err = tm.RedisScan("invalid")
	if err == nil {
		t.Fatal("Expected error for string type")
	}

	err = tm.RedisScan([]byte("not-a-number"))
	if err == nil {
		t.Fatal("Expected error for invalid byte string")
	}
}

func TestTimeFromTime(t *testing.T) {
	now := time.Now()
	t1 := Time{}
	t2 := t1.FromTime(now)
	if !t2.Time.Equal(now) {
		t.Fatal("Time mismatch")
	}
}

func TestPushLogAndReadLogs(t *testing.T) {
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	mock.GenericCommand("RPUSH").Expect(int64(1))

	la := NewLogError(1, errors.New("test error"))
	err := PushLog(conn, la)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	mock.GenericCommand("LRANGE").Expect([]interface{}{})

	_, err = ReadLogs(conn, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
