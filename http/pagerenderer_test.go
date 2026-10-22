// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestJSONRendererType(t *testing.T) {
	r := &JSONRenderer{}
	if r.Type() != "JSON" {
		t.Fatalf("Expected JSON, got %s", r.Type())
	}
}

func TestJSONRendererWrite(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	r := &JSONRenderer{}
	code, err := r.Write(ctx, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", code)
	}
}

func TestJSONRendererWritePretty(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		OutputMode:    "json",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?pretty", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	r := &JSONRenderer{}
	code, err := r.Write(ctx, results)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", code)
	}
}

func TestRedirectRendererType(t *testing.T) {
	r := &RedirectRenderer{}
	if r.Type() != "REDIRECT" {
		t.Fatalf("Expected REDIRECT, got %s", r.Type())
	}
}

func TestRedirectRendererWriteEmpty(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	r := &RedirectRenderer{}
	code, _ := r.Write(ctx, results)
	if code != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", code)
	}
}

func TestRedirectRendererWriteWithMirror(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
		MaxLinkHeaders: 5,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
		MirrorList: []mirrors.Mirror{
			{HttpURL: "http://mirror1.com/"},
		},
	}

	r := &RedirectRenderer{}
	code, _ := r.Write(ctx, results)
	if code != http.StatusFound {
		t.Fatalf("Expected 302, got %d", code)
	}
}

func TestMirrorListRendererType(t *testing.T) {
	r := &MirrorListRenderer{}
	if r.Type() != "MIRRORLIST" {
		t.Fatalf("Expected MIRRORLIST, got %s", r.Type())
	}
}

func TestMirrorListRendererWriteNoTemplates(t *testing.T) {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, req, Templates{})

	results := &mirrors.Results{
		FileInfo: filesystem.FileInfo{Path: "/test"},
	}

	r := &MirrorListRenderer{}
	code, err := r.Write(ctx, results)
	if err == nil {
		t.Fatal("Expected error for nil templates")
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("Expected 500, got %d", code)
	}
}

func TestErrTemplatesNotFound(t *testing.T) {
	if ErrTemplatesNotFound == nil {
		t.Fatal("Expected non-nil")
	}
}
