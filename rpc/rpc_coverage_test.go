// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestCLI_AddMirror_UnexpectedID(t *testing.T) {
	c := &CLI{}

	m, err := MirrorToRPC(&mirrors.Mirror{ID: 1, StateSince: mirrors.Time{}, LastSync: mirrors.Time{}, LastSuccessfulSync: mirrors.Time{}, LastModTime: mirrors.Time{}})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	_, err = c.AddMirror(context.Background(), m)
	if err == nil {
		t.Fatalf("Expected error for unexpected ID")
	}
}

func TestCLI_AddMirror_InvalidURL(t *testing.T) {
	c := &CLI{}

	now := timestampNow()
	m := &Mirror{
		ID:             0,
		Name:           "test",
		HttpURL:        "://invalid",
		StateSince:     now,
		LastSync:       now,
		LastSuccessfulSync: now,
		LastModTime:    now,
	}

	_, err := c.AddMirror(context.Background(), m)
	if err == nil {
		t.Fatalf("Expected error for invalid URL")
	}
}

func TestCLI_UpdateMirror_FromRPCError(t *testing.T) {
	c := &CLI{}

	// Mirror with nil timestamps will cause MirrorFromRPC to fail
	m := &Mirror{ID: 1}

	_, err := c.UpdateMirror(context.Background(), m)
	if err == nil {
		t.Fatalf("Expected error from MirrorFromRPC with nil timestamps")
	}
}

func TestCLI_AddMirror_FromRPCError(t *testing.T) {
	c := &CLI{}

	m := &Mirror{ID: 0}

	_, err := c.AddMirror(context.Background(), m)
	if err == nil {
		t.Fatalf("Expected error from MirrorFromRPC with nil timestamps")
	}
}

func TestCLI_MirrorInfo_FromRPCError(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}

	// This should fail because Connect returns ErrUnreachable
	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 1})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_RefreshRepository_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}

	_, err := c.RefreshRepository(context.Background(), &RefreshRepositoryRequest{Rehash: false})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestMirrorToRPC_Error(t *testing.T) {
	// Create a mirror with an invalid time that causes TimestampProto to fail
	m := &mirrors.Mirror{
		ID:         1,
		StateSince: mirrors.Time{},
	}
	m.StateSince.Time = m.StateSince.Time.AddDate(-10000, 0, 0)

	_, err := MirrorToRPC(m)
	if err == nil {
		t.Fatalf("Expected error from TimestampProto for invalid time")
	}
}

func TestMirrorFromRPC_Error(t *testing.T) {
	m := &Mirror{
		ID:         1,
		StateSince: &timestamp.Timestamp{Seconds: -1, Nanos: -1},
	}

	_, err := MirrorFromRPC(m)
	if err == nil {
		t.Fatalf("Expected error from Timestamp for invalid timestamp")
	}
}

func TestCLI_SetMethods(t *testing.T) {
	c := &CLI{}
	c.SetDatabase(nil)
	c.SetCache(nil)
	c.SetSignals(nil)
}

func TestCLI_Ping_WithContext(t *testing.T) {
	c := &CLI{}
	resp, err := c.Ping(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if resp == nil {
		t.Fatalf("Expected non-nil response")
	}
}

func TestConfiguration_Default(t *testing.T) {
	SetConfiguration(&Configuration{RedisAddress: ""})
	if GetConfig().RedisAddress != "" {
		t.Fatalf("Expected empty RedisAddress")
	}
}
