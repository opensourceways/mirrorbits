// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"errors"
	"testing"

	"github.com/rafaeljusto/redigomock"
)

func TestIsScanning(t *testing.T) {
	mock := redigomock.NewConn()

	mock.Command("EXISTS", "SCANNING_1").Expect(int64(1))

	scanning, err := IsScanning(mock, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !scanning {
		t.Fatal("Expected true")
	}

	mock.Command("EXISTS", "SCANNING_2").Expect(int64(0))
	scanning, err = IsScanning(mock, 2)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if scanning {
		t.Fatal("Expected false")
	}
}

func TestIsScanningError(t *testing.T) {
	mock := redigomock.NewConn()
	mock.Command("EXISTS", "SCANNING_3").ExpectError(errors.New("connection error"))
	_, err := IsScanning(mock, 3)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestScanErrors(t *testing.T) {
	if ErrScanAborted == nil {
		t.Fatal("ErrScanAborted should not be nil")
	}
	if ErrScanInProgress == nil {
		t.Fatal("ErrScanInProgress should not be nil")
	}
	if ErrNoSyncMethod == nil {
		t.Fatal("ErrNoSyncMethod should not be nil")
	}
}
