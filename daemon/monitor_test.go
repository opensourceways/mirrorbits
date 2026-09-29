// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestMirrorNeedHealthCheck(t *testing.T) {
	m := mirror{
		lastCheck: time.Now(),
	}
	if m.NeedHealthCheck(30) {
		t.Fatal("Expected false for recent check")
	}

	m.lastCheck = time.Now().Add(-time.Hour)
	if !m.NeedHealthCheck(30) {
		t.Fatal("Expected true for old check")
	}
}

func TestMirrorNeedSync(t *testing.T) {
	m := mirror{}
	if !m.NeedSync(60) {
		t.Fatal("Expected true for zero time (sync needed)")
	}

	m.LastSync = mirrors.Time{}.FromTime(time.Now())
	if m.NeedSync(60) {
		t.Fatal("Expected false for recent sync")
	}

	m.LastSync = mirrors.Time{}.FromTime(time.Now().Add(-2 * time.Hour))
	if !m.NeedSync(60) {
		t.Fatal("Expected true for old sync")
	}
}

func TestMirrorIsScanning(t *testing.T) {
	m := mirror{scanning: true}
	if !m.IsScanning() {
		t.Fatal("Expected true")
	}
	m.scanning = false
	if m.IsScanning() {
		t.Fatal("Expected false")
	}
}

func TestMirrorIsChecking(t *testing.T) {
	m := mirror{checking: true}
	if !m.IsChecking() {
		t.Fatal("Expected true")
	}
	m.checking = false
	if m.IsChecking() {
		t.Fatal("Expected false")
	}
}

func TestCheckRedirectAllowed(t *testing.T) {
	r := mirrors.Redirects(1)
	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, r)
	req := &http.Request{
		Method: "GET",
	}
	req = req.WithContext(ctx)
	err := checkRedirect(req, nil)
	if err != nil {
		t.Fatalf("Expected nil, got %s", err)
	}
}

func TestCheckRedirectNotAllowed(t *testing.T) {
	r := mirrors.Redirects(2)
	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, r)
	req := &http.Request{}
	req = req.WithContext(ctx)
	err := checkRedirect(req, []*http.Request{{}})
	if err == nil {
		t.Fatal("Expected error for not allowed redirect")
	}
}
