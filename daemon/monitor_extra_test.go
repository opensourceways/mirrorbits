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

func TestMonitorRetryImmediateSuccess(t *testing.T) {
	m := &monitor{
		stop: make(chan struct{}),
	}

	called := false
	m.retry(func(i uint) error {
		called = true
		return nil
	}, time.Second)

	if !called {
		t.Fatal("Expected function to be called")
	}
}

func TestMonitorRetryStopChannel(t *testing.T) {
	m := &monitor{
		stop: make(chan struct{}),
	}

	close(m.stop)
	callCount := 0
	m.retry(func(i uint) error {
		callCount++
		return errMirrorNotScanned
	}, 100*time.Millisecond)

	if callCount != 1 {
		t.Fatalf("Expected 1 call, got %d", callCount)
	}
}

func TestMonitorStop(t *testing.T) {
	m := &monitor{
		stop:    make(chan struct{}),
		cluster: &cluster{stop: make(chan bool)},
	}
	m.Stop()
	m.Stop()
}

func TestMonitorWait(t *testing.T) {
	m := &monitor{}
	m.Wait()
}

func TestMirrorNeedHealthCheckZero(t *testing.T) {
	mm := mirror{}
	if !mm.NeedHealthCheck(0) {
		t.Fatal("Expected true for zero interval")
	}
}

func TestMirrorNeedSyncZero(t *testing.T) {
	mm := mirror{}
	if !mm.NeedSync(0) {
		t.Fatal("Expected true for zero interval")
	}
}

func TestCheckRedirectNoContextValue(t *testing.T) {
	req := &http.Request{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for missing context value")
		}
	}()
	_ = checkRedirect(req, nil)
}

func TestCheckRedirectAllowedNilVia(t *testing.T) {
	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, mirrors.Redirects(1))
	req := &http.Request{}
	req = req.WithContext(ctx)
	err := checkRedirect(req, nil)
	if err != nil {
		t.Fatalf("Expected nil: %v", err)
	}
}

func TestCheckRedirectNotAllowedWithVia(t *testing.T) {
	ctx := context.WithValue(context.Background(), core.ContextAllowRedirects, mirrors.Redirects(2))
	req := &http.Request{}
	req = req.WithContext(ctx)
	err := checkRedirect(req, []*http.Request{{}})
	if err == nil {
		t.Fatal("Expected error for not allowed redirect")
	}
}
