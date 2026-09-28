// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestMirror_NeedHealthCheck(t *testing.T) {
	m := &mirror{lastCheck: time.Now().Add(-10 * time.Minute)}
	if !m.NeedHealthCheck(5) {
		t.Fatalf("Expected true for check needed")
	}
	if m.NeedHealthCheck(15) {
		t.Fatalf("Expected false for check not needed")
	}
}

func TestMirror_NeedSync(t *testing.T) {
	m := &mirror{
		Mirror: mirrors.Mirror{
			LastSync: mirrors.Time{}.FromTime(time.Now().Add(-10 * time.Minute)),
		},
	}
	if !m.NeedSync(5) {
		t.Fatalf("Expected true for sync needed")
	}
	if m.NeedSync(15) {
		t.Fatalf("Expected false for sync not needed")
	}
}

func TestMirror_IsScanning(t *testing.T) {
	m := &mirror{scanning: true}
	if !m.IsScanning() {
		t.Fatalf("Expected true")
	}
	m.scanning = false
	if m.IsScanning() {
		t.Fatalf("Expected false")
	}
}

func TestMirror_IsChecking(t *testing.T) {
	m := &mirror{checking: true}
	if !m.IsChecking() {
		t.Fatalf("Expected true")
	}
	m.checking = false
	if m.IsChecking() {
		t.Fatalf("Expected false")
	}
}

func TestCheckRedirect_Allowed(t *testing.T) {
	SetConfiguration(&Configuration{})
	redirects := mirrors.Redirects(1)

	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, redirects)
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req = req.WithContext(ctx)

	err := checkRedirect(req, []*http.Request{})
	if err != nil {
		t.Fatalf("Expected nil for allowed redirects: %s", err)
	}
}

func TestCheckRedirect_NotAllowed(t *testing.T) {
	SetConfiguration(&Configuration{})
	redirects := mirrors.Redirects(2)

	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, redirects)
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req = req.WithContext(ctx)

	err := checkRedirect(req, []*http.Request{req})
	if err != errRedirect {
		t.Fatalf("Expected errRedirect, got %v", err)
	}
}

func TestMonitor_retry_Success(t *testing.T) {
	m := &monitor{stop: make(chan struct{})}
	defer close(m.stop)

	calls := 0
	m.retry(func(i uint) error {
		calls++
		if i < 2 {
			return errRedirect
		}
		return nil
	}, 10*time.Millisecond)

	if calls != 3 {
		t.Fatalf("Expected 3 calls, got %d", calls)
	}
}

func TestMonitor_retry_ImmediateSuccess(t *testing.T) {
	m := &monitor{stop: make(chan struct{})}
	defer close(m.stop)

	called := false
	m.retry(func(i uint) error {
		called = true
		return nil
	}, 10*time.Millisecond)

	if !called {
		t.Fatalf("Expected fn to be called")
	}
}

func TestMonitor_retry_StopsWhenChannelClosed(t *testing.T) {
	m := &monitor{stop: make(chan struct{})}

	var calls uint
	done := make(chan struct{})
	go func() {
		m.retry(func(i uint) error {
			calls++
			return errRedirect
		}, 50*time.Millisecond)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	close(m.stop)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("retry did not stop after channel closed")
	}
}

func TestErrRedirect(t *testing.T) {
	if errRedirect == nil {
		t.Fatalf("Expected non-nil error")
	}
}

func TestErrMirrorNotScanned(t *testing.T) {
	if errMirrorNotScanned == nil {
		t.Fatalf("Expected non-nil error")
	}
}
