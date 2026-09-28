// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type mockPool struct {
	conn *redigomock.Conn
}

func (m *mockPool) Get() redis.Conn { return m.conn }
func (m *mockPool) Close() error { return nil }

func TestIsScanning(t *testing.T) {
	mock := redigomock.NewConn()

	mock.Command("EXISTS", "SCANNING_1").Expect(int64(1))
	mock.Command("EXISTS", "SCANNING_2").Expect(int64(0))

	scanning, err := IsScanning(mock, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !scanning {
		t.Fatalf("Expected true for existing scan")
	}

	scanning, err = IsScanning(mock, 2)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if scanning {
		t.Fatalf("Expected false for non-existing scan")
	}
}

func TestErrScanAborted(t *testing.T) {
	if ErrScanAborted == nil {
		t.Fatalf("ErrScanAborted should not be nil")
	}
	if ErrScanAborted.Error() != "scan aborted" {
		t.Fatalf("Expected 'scan aborted', got %s", ErrScanAborted.Error())
	}
}

func TestErrScanInProgress(t *testing.T) {
	if ErrScanInProgress == nil {
		t.Fatalf("ErrScanInProgress should not be nil")
	}
	if ErrScanInProgress.Error() != "scan already in progress" {
		t.Fatalf("Expected 'scan already in progress', got %s", ErrScanInProgress.Error())
	}
}

func TestErrNoSyncMethod(t *testing.T) {
	if ErrNoSyncMethod == nil {
		t.Fatalf("ErrNoSyncMethod should not be nil")
	}
	if ErrNoSyncMethod.Error() != "no suitable URL for the scan" {
		t.Fatalf("Expected 'no suitable URL for the scan', got %s", ErrNoSyncMethod.Error())
	}
}
