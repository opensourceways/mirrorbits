// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package http

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/network"
)

func TestLoadTemplates_Success(t *testing.T) {
	dir, err := os.MkdirTemp("", "templates-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	baseContent := `{{define "base"}}<html><body>{{template "content" .}}</body></html>{{end}}`
	listContent := `{{define "content"}}mirrorlist{{end}}`
	statsContent := `{{define "content"}}mirrorstats{{end}}`

	if err := os.WriteFile(filepath.Join(dir, "base.html"), []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write base.html: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mirrorlist.html"), []byte(listContent), 0644); err != nil {
		t.Fatalf("Failed to write mirrorlist.html: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mirrorstats.html"), []byte(statsContent), 0644); err != nil {
		t.Fatalf("Failed to write mirrorstats.html: %s", err)
	}

	SetConfiguration(&Configuration{Templates: dir})

	h := &HTTP{}
	h.templates.RWMutex = new(sync.RWMutex)

	t1, err := h.LoadTemplates("mirrorlist")
	if err != nil {
		t.Fatalf("LoadTemplates mirrorlist failed: %s", err)
	}
	if t1 == nil {
		t.Fatalf("Expected non-nil template")
	}

	t2, err := h.LoadTemplates("mirrorstats")
	if err != nil {
		t.Fatalf("LoadTemplates mirrorstats failed: %s", err)
	}
	if t2 == nil {
		t.Fatalf("Expected non-nil template")
	}
}

func TestHTTP_Reload(t *testing.T) {
	dir, err := os.MkdirTemp("", "templates-reload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	baseContent := `{{define "base"}}<html>{{template "content" .}}</html>{{end}}`
	listContent := `{{define "content"}}list{{end}}`
	statsContent := `{{define "content"}}stats{{end}}`

	os.WriteFile(filepath.Join(dir, "base.html"), []byte(baseContent), 0644)
	os.WriteFile(filepath.Join(dir, "mirrorlist.html"), []byte(listContent), 0644)
	os.WriteFile(filepath.Join(dir, "mirrorstats.html"), []byte(statsContent), 0644)

	SetConfiguration(&Configuration{Templates: dir})

	h := &HTTP{}
	h.templates.RWMutex = new(sync.RWMutex)
	h.geoip = network.NewGeoIP()

	h.Reload()
}
