// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"testing"
	"time"
)

func TestMirrorStatsStruct(t *testing.T) {
	ms := MirrorStats{
		ID:         1,
		Name:       "test",
		Downloads:  100,
		Bytes:      1024,
		PercentD:   50.0,
		PercentB:   75.0,
		SyncOffset: SyncOffset{Valid: true, Value: 5, HumanReadable: "5h"},
		TZOffset:   time.Hour,
	}
	if ms.ID != 1 || ms.Name != "test" {
		t.Fatal("Fields mismatch")
	}
	if ms.Downloads != 100 || ms.Bytes != 1024 {
		t.Fatal("Fields mismatch")
	}
}

func TestSyncOffsetStruct(t *testing.T) {
	so := SyncOffset{Valid: true, Value: 3, HumanReadable: "3 hours"}
	if !so.Valid || so.Value != 3 || so.HumanReadable != "3 hours" {
		t.Fatal("Fields mismatch")
	}
}

func TestMirrorStatsPageStruct(t *testing.T) {
	p := MirrorStatsPage{
		List:        []MirrorStats{{ID: 1}},
		LocalJSPath: "/js",
	}
	if len(p.List) != 1 || p.LocalJSPath != "/js" {
		t.Fatal("Fields mismatch")
	}
}

func TestMirrorStatsSliceSort(t *testing.T) {
	s := mirrorStatsSlice{
		{ID: 1, Downloads: 50},
		{ID: 2, Downloads: 100},
		{ID: 3, Downloads: 75},
	}
	if s.Len() != 3 {
		t.Fatalf("Expected 3, got %d", s.Len())
	}
	s.Swap(0, 1)
	if s[0].ID != 2 {
		t.Fatal("Swap failed")
	}
	s.Swap(0, 1)

	bd := byDownloadNumbers{s}
	if !bd.Less(1, 0) {
		t.Fatal("Expected Less(1,0) to be true (100 > 50)")
	}
	if bd.Less(0, 1) {
		t.Fatal("Expected Less(0,1) to be false (50 < 100)")
	}
}

func TestHTTPStruct(t *testing.T) {
	h := &HTTP{}
	if h.stopped {
		t.Fatal("Expected false")
	}
	h.stopped = true
	h.Stop(0)
	h.Stop(0)
}

func TestHTTPStopChan(t *testing.T) {
	h := &HTTP{}
	ch := h.StopChan()
	if ch != nil {
		t.Fatal("Expected nil for uninitialized StopChan")
	}
}

func TestHTTPSetListener(t *testing.T) {
	h := &HTTP{}
	if h.Listener != nil {
		t.Fatal("Expected nil")
	}
}

func TestHTTPTerminateNoServer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic: %v", r)
		}
	}()
	_ = NewStats(nil)
}
