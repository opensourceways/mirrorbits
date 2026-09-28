// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestNewContext_Standard(t *testing.T) {
	SetConfiguration(&Configuration{
		SchemaStrictMatch: false,
	})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.Type() != STANDARD {
		t.Fatalf("Expected STANDARD, got %d", c.Type())
	}
	if c.IsMirrorlist() {
		t.Fatalf("Expected false for IsMirrorlist")
	}
	if c.IsFileStats() {
		t.Fatalf("Expected false for IsFileStats")
	}
	if c.IsMirrorStats() {
		t.Fatalf("Expected false for IsMirrorStats")
	}
	if c.IsChecksum() {
		t.Fatalf("Expected false for IsChecksum")
	}
}

func TestNewContext_Mirrorlist(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?mirrorlist", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.Type() != MIRRORLIST {
		t.Fatalf("Expected MIRRORLIST, got %d", c.Type())
	}
	if !c.IsMirrorlist() {
		t.Fatalf("Expected true for IsMirrorlist")
	}
}

func TestNewContext_FileStats(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?stats", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.Type() != FILESTATS {
		t.Fatalf("Expected FILESTATS, got %d", c.Type())
	}
	if !c.IsFileStats() {
		t.Fatalf("Expected true for IsFileStats")
	}
}

func TestNewContext_MirrorStats(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.Type() != MIRRORSTATS {
		t.Fatalf("Expected MIRRORSTATS, got %d", c.Type())
	}
	if !c.IsMirrorStats() {
		t.Fatalf("Expected true for IsMirrorStats")
	}
}

func TestNewContext_Checksum(t *testing.T) {
	SetConfiguration(&Configuration{})

	for _, param := range []string{"md5", "sha1", "sha256"} {
		req := httptest.NewRequest("GET", "/test/file.txt?"+param, nil)
		w := httptest.NewRecorder()

		c := NewContext(w, req, Templates{})

		if c.Type() != CHECKSUM {
			t.Fatalf("Expected CHECKSUM for param %s, got %d", param, c.Type())
		}
		if !c.IsChecksum() {
			t.Fatalf("Expected true for IsChecksum with param %s", param)
		}
	}
}

func TestNewContext_Pretty(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?pretty", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if !c.IsPretty() {
		t.Fatalf("Expected true for IsPretty")
	}
}

func TestNewContext_SecureOption_TLS(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.SecureOption() != WITHTLS {
		t.Fatalf("Expected WITHTLS, got %d", c.SecureOption())
	}
}

func TestNewContext_SecureOption_NoTLS(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?https=0", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.SecureOption() != WITHOUTTLS {
		t.Fatalf("Expected WITHOUTTLS, got %d", c.SecureOption())
	}
}

func TestNewContext_SecureOption_ExplicitTLS(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test/file.txt?https=1", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.SecureOption() != WITHTLS {
		t.Fatalf("Expected WITHTLS, got %d", c.SecureOption())
	}
}

func TestContext_Request(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.Request() != req {
		t.Fatalf("Request mismatch")
	}
}

func TestContext_ResponseWriter(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.ResponseWriter() != w {
		t.Fatalf("ResponseWriter mismatch")
	}
}

func TestContext_QueryParam(t *testing.T) {
	SetConfiguration(&Configuration{})

	req := httptest.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()

	c := NewContext(w, req, Templates{})

	if c.QueryParam("foo") != "bar" {
		t.Fatalf("Expected 'bar', got %s", c.QueryParam("foo"))
	}
	if c.QueryParam("baz") != "qux" {
		t.Fatalf("Expected 'qux', got %s", c.QueryParam("baz"))
	}
	if c.QueryParam("nonexistent") != "" {
		t.Fatalf("Expected '', got %s", c.QueryParam("nonexistent"))
	}
}

func TestNewGzipHandler_Disabled(t *testing.T) {
	SetConfiguration(&Configuration{Gzip: false})

	called := false
	fn := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

	h := NewGzipHandler(fn)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	h(w, req)

	if !called {
		t.Fatalf("Handler should be called when gzip is disabled")
	}
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("Should not set gzip encoding when disabled")
	}
}

func TestNewGzipHandler_NoAcceptEncoding(t *testing.T) {
	SetConfiguration(&Configuration{Gzip: true})

	called := false
	fn := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

	h := NewGzipHandler(fn)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h(w, req)

	if !called {
		t.Fatalf("Handler should be called when no Accept-Encoding")
	}
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("Should not set gzip encoding when not accepted")
	}
}

func TestNewGzipHandler_Enabled(t *testing.T) {
	SetConfiguration(&Configuration{Gzip: true})

	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}

	h := NewGzipHandler(fn)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	h(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Expected gzip encoding, got %s", w.Header().Get("Content-Encoding"))
	}
}
