// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestStatsCountDownloadValidAndTerminate(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: "m1", ID: 1}
	fi := filesystem.FileInfo{Path: "/test.txt", Size: 100}
	err := s.CountDownload(m, fi)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestStatsPushStatsEmpty(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()
	s.pushStats()
}

func TestStatsPushStatsWithData(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: "m1", ID: 1}
	fi := filesystem.FileInfo{Path: "/test.txt", Size: 100}
	_ = s.CountDownload(m, fi)

	time.Sleep(10 * time.Millisecond)
}

func TestStatsCountDownloadNilMirror(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: ""}
	fi := filesystem.FileInfo{Path: "/test.txt", Size: 100}
	err := s.CountDownload(m, fi)
	if err == nil {
		t.Fatal("Expected error for nil mirror")
	}
}

func TestStatsCountDownloadEmptyPathExtra(t *testing.T) {
	s := NewStats(nil)
	defer s.Terminate()

	m := mirrors.Mirror{Name: "m1"}
	fi := filesystem.FileInfo{Path: "", Size: 100}
	err := s.CountDownload(m, fi)
	if err == nil {
		t.Fatal("Expected error for empty path")
	}
}

func TestStatsTerminateMultipleTimes(t *testing.T) {
	s := NewStats(nil)
	s.Terminate()
}
