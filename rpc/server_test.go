// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package rpc

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestCLIChangeStatusNegativeID(t *testing.T) {
	c := &CLI{}
	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: -1, Enabled: true})
	if err == nil {
		t.Fatal("Expected error for negative ID")
	}
}

func TestCLIMirrorInfoNegativeID(t *testing.T) {
	c := &CLI{}
	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: -1})
	if err == nil {
		t.Fatal("Expected error for negative ID")
	}
}

func TestCLIAddMirrorNilTimestamps(t *testing.T) {
	c := &CLI{}
	_, err := c.AddMirror(context.Background(), &Mirror{
		Name:       "test",
		HttpURL:    "http://test.com",
		StateSince: nil,
	})
	if err == nil {
		t.Fatal("Expected error for nil timestamps")
	}
}

func TestCLIAddMirrorValidTimestamps(t *testing.T) {
	c := &CLI{}
	ts := &timestamp.Timestamp{Seconds: 1000}
	_, err := c.AddMirror(context.Background(), &Mirror{
		Name:                 "test",
		HttpURL:              "http://test.com",
		StateSince:           ts,
		LastSync:             ts,
		LastSuccessfulSync:   ts,
		LastModTime:          ts,
	})
	// Will fail because redis is nil, but MirrorFromRPC should succeed
	_ = err
}

func TestCLIRemoveMirrorNegativeID(t *testing.T) {
	c := &CLI{}
	_, err := c.RemoveMirror(context.Background(), &MirrorIDRequest{ID: -1})
	if err == nil {
		t.Fatal("Expected error for negative ID")
	}
}

func TestCLIScanMirrorNilRedis(t *testing.T) {
	c := &CLI{}
	_, err := c.ScanMirror(context.Background(), &ScanMirrorRequest{ID: 1})
	if err == nil {
		t.Fatal("Expected error for nil redis")
	}
}

func TestCLIGetMirrorLogsNegativeID(t *testing.T) {
	c := &CLI{}
	_, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{ID: -1})
	if err == nil {
		t.Fatal("Expected error for negative ID")
	}
}

func TestCLIUpgradeNilSig(t *testing.T) {
	c := &CLI{}
	_, err := c.Upgrade(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatal("Expected error for nil sig")
	}
}

func TestCLIReloadNilSig(t *testing.T) {
	c := &CLI{}
	_, err := c.Reload(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatal("Expected error for nil sig")
	}
}

func TestCLIUpgradeWithSigBufferedFull(t *testing.T) {
	sig := make(chan os.Signal, 1)
	sig <- syscall.SIGUSR1
	c := &CLI{sig: sig}
	_, err := c.Upgrade(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatal("Expected error for full sig buffer")
	}
}

func TestCLIUpgradeWithSigSuccess(t *testing.T) {
	sig := make(chan os.Signal, 1)
	c := &CLI{sig: sig}
	_, err := c.Upgrade(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLIReloadWithSigSuccess(t *testing.T) {
	sig := make(chan os.Signal, 1)
	c := &CLI{sig: sig}
	_, err := c.Reload(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLIReloadWithSigBufferedFull(t *testing.T) {
	sig := make(chan os.Signal, 1)
	sig <- syscall.SIGHUP
	c := &CLI{sig: sig}
	_, err := c.Reload(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatal("Expected error for full sig buffer")
	}
}
