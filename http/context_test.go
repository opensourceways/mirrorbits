// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
)

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestNewContextStandard(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if c.Type() != STANDARD {
		t.Fatalf("Expected STANDARD, got %d", c.Type())
	}
	if c.IsMirrorlist() {
		t.Fatal("Expected IsMirrorlist to be false")
	}
}

func TestNewContextMirrorlist(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?mirrorlist", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if !c.IsMirrorlist() {
		t.Fatal("Expected IsMirrorlist to be true")
	}
	if c.Type() != MIRRORLIST {
		t.Fatalf("Expected MIRRORLIST, got %d", c.Type())
	}
}

func TestNewContextFileStats(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?stats", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if !c.IsFileStats() {
		t.Fatal("Expected IsFileStats to be true")
	}
}

func TestNewContextMirrorStats(t *testing.T) {
	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if !c.IsMirrorStats() {
		t.Fatal("Expected IsMirrorStats to be true")
	}
}

func TestNewContextChecksum(t *testing.T) {
	tests := []string{"md5", "sha1", "sha256"}
	for _, param := range tests {
		req := httptest.NewRequest("GET", "/test.txt?"+param, nil)
		w := httptest.NewRecorder()
		c := NewContext(w, req, Templates{})
		if !c.IsChecksum() {
			t.Fatalf("Expected IsChecksum to be true for %s", param)
		}
	}
}

func TestNewContextPretty(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?pretty", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if !c.IsPretty() {
		t.Fatal("Expected IsPretty to be true")
	}
}

func TestNewContextWithTLS(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if c.SecureOption() != WITHTLS {
		t.Fatalf("Expected WITHTLS, got %d", c.SecureOption())
	}
}

func TestNewContextWithoutTLS(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?https=0", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if c.SecureOption() != WITHOUTTLS {
		t.Fatalf("Expected WITHOUTTLS, got %d", c.SecureOption())
	}
}

func TestNewContextWithTLSParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?https=1", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if c.SecureOption() != WITHTLS {
		t.Fatalf("Expected WITHTLS, got %d", c.SecureOption())
	}
}

func TestContextRequestParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt?custom=value", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if r := c.QueryParam("custom"); r != "value" {
		t.Fatalf("Expected 'value', got %s", r)
	}
	if r := c.QueryParam("nonexistent"); r != "" {
		t.Fatalf("Expected empty, got %s", r)
	}
}

func TestContextRequestAndResponseWriter(t *testing.T) {
	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, Templates{})
	if c.Request() != req {
		t.Fatal("Request mismatch")
	}
	if c.ResponseWriter() != w {
		t.Fatal("ResponseWriter mismatch")
	}
}

func TestNewGzipHandlerNoGzip(t *testing.T) {
	called := false
	handler := NewGzipHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("hello"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if !called {
		t.Fatal("Expected handler to be called")
	}
}

func TestNewGzipHandlerWithAcceptEncoding(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		Gzip:          true,
	})

	called := false
	handler := NewGzipHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("hello world"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler(w, req)
	if !called {
		t.Fatal("Expected handler to be called")
	}
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("Expected gzip content encoding")
	}
}

func TestNewGzipHandlerGzipDisabled(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		Gzip:          false,
	})

	called := false
	handler := NewGzipHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("hello"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler(w, req)
	if !called {
		t.Fatal("Expected handler to be called")
	}
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("Expected no gzip encoding when disabled")
	}
}
