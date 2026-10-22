// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"os"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	SetConfiguration(&Configuration{
		Repository:    "/tmp/test",
		ListenAddress: ":8080",
	})
}

func TestCreateDiff(t *testing.T) {
	m1 := &mirrors.Mirror{Name: "test", HttpURL: "http://old.com", Asnum: 100}
	m2 := &mirrors.Mirror{Name: "test", HttpURL: "http://new.com", Asnum: 200}
	diff := createDiff(m1, m2)
	if diff == "" {
		t.Fatal("Expected non-empty diff")
	}
}

func TestCreateDiffNoChange(t *testing.T) {
	m1 := &mirrors.Mirror{Name: "test", HttpURL: "http://same.com"}
	m2 := &mirrors.Mirror{Name: "test", HttpURL: "http://same.com"}
	diff := createDiff(m1, m2)
	if diff != "" {
		t.Fatalf("Expected empty diff, got %s", diff)
	}
}

func TestCLIPing(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLIGetVersion(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	reply, err := c.GetVersion(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if reply == nil {
		t.Fatal("Expected non-nil reply")
	}
	if reply.Version == "" && reply.GoVersion == "" {
		t.Fatal("Expected some version info")
	}
}

func TestCLIMatchMirrorNilRedis(t *testing.T) {
	c := &CLI{redis: nil}
	ctx := context.Background()
	_, err := c.MatchMirror(ctx, &MatchRequest{Pattern: "test"})
	if err == nil {
		t.Fatal("Expected error for nil redis")
	}
}

func TestCLIChangeStatusInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.ChangeStatus(ctx, &ChangeStatusRequest{ID: 0, Enabled: true})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.FailedPrecondition {
		t.Fatalf("Expected FailedPrecondition, got %v", err)
	}
}

func TestCLIMirrorInfoInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.MirrorInfo(ctx, &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestCLIScanMirrorInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.ScanMirror(ctx, &ScanMirrorRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestCLIRemoveMirrorInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.RemoveMirror(ctx, &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestCLIStatsMirrorInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.StatsMirror(ctx, &StatsMirrorRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestCLIGetMirrorLogsInvalidID(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.GetMirrorLogs(ctx, &GetMirrorLogsRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestCLISetCache(t *testing.T) {
	c := &CLI{}
	c.SetCache(nil)
}

func TestCLISetDatabase(t *testing.T) {
	c := &CLI{}
	c.SetDatabase(nil)
}

func TestCLISetSignals(t *testing.T) {
	c := &CLI{}
	sig := make(chan<- os.Signal)
	c.SetSignals(sig)
}

func TestMirrorToRPCError(t *testing.T) {
	m := &mirrors.Mirror{
		ID:            1,
		Name:          "test",
		LastSync:      mirrors.Time{},
		LastSuccessfulSync: mirrors.Time{},
		LastModTime:   mirrors.Time{},
		StateSince:    mirrors.Time{},
	}
	_, err := MirrorToRPC(m)
	_ = err
}

func TestScanMirrorReplyGettersExtra(t *testing.T) {
	m := &ScanMirrorReply{
		Enabled:       true,
		FilesIndexed:  42,
		KnownIndexed:  30,
		Removed:       5,
		TZOffsetMs:    3600000,
	}
	if m.GetKnownIndexed() != 30 {
		t.Fatal("KnownIndexed mismatch")
	}
	if m.GetRemoved() != 5 {
		t.Fatal("Removed mismatch")
	}
	if m.GetTZOffsetMs() != 3600000 {
		t.Fatal("TZOffsetMs mismatch")
	}
	m.Reset()
	if m.GetKnownIndexed() != 0 {
		t.Fatal("Expected 0 after reset")
	}
	if m.GetRemoved() != 0 {
		t.Fatal("Expected 0 after reset")
	}
	if m.GetTZOffsetMs() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestScanMirrorReplyNilGettersExtra(t *testing.T) {
	var m *ScanMirrorReply
	if m.GetKnownIndexed() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetRemoved() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetTZOffsetMs() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestStatsFileRequestGetDateEnd(t *testing.T) {
	ts := &timestamp.Timestamp{Seconds: 100}
	m := &StatsFileRequest{
		Pattern:   "test",
		DateStart: ts,
		DateEnd:   ts,
	}
	if m.GetDateEnd() != ts {
		t.Fatal("DateEnd mismatch")
	}
	if m.GetDateStart() != ts {
		t.Fatal("DateStart mismatch")
	}
	m.Reset()
	if m.GetDateEnd() != nil {
		t.Fatal("Expected nil after reset")
	}
}

func TestStatsFileReplyStringAndProto(t *testing.T) {
	m := &StatsFileReply{Files: map[string]int64{"f1": 1}}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
	m.Reset()
	if len(m.GetFiles()) != 0 {
		t.Fatal("Expected empty after reset")
	}
}

func TestStatsMirrorRequestGetDateEnd(t *testing.T) {
	ts := &timestamp.Timestamp{Seconds: 200}
	m := &StatsMirrorRequest{
		ID:        5,
		DateStart: ts,
		DateEnd:   ts,
	}
	if m.GetDateEnd() != ts {
		t.Fatal("DateEnd mismatch")
	}
	m.Reset()
	if m.GetDateEnd() != nil {
		t.Fatal("Expected nil after reset")
	}
}

func TestStatsMirrorReplyStringAndProto(t *testing.T) {
	m := &StatsMirrorReply{Requests: 10, Bytes: 1024}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
	m.Reset()
	if m.GetRequests() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestStatsFileRequestStringAndProto(t *testing.T) {
	m := &StatsFileRequest{Pattern: "test"}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
}

func TestStatsMirrorRequestStringAndProto(t *testing.T) {
	m := &StatsMirrorRequest{ID: 5}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
}

func TestGetMirrorLogsRequestGetters(t *testing.T) {
	m := &GetMirrorLogsRequest{ID: 3, MaxResults: 100}
	if m.GetID() != 3 {
		t.Fatal("ID mismatch")
	}
	if m.GetMaxResults() != 100 {
		t.Fatal("MaxResults mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
	if m.GetMaxResults() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestGetMirrorLogsRequestNilGetters(t *testing.T) {
	var m *GetMirrorLogsRequest
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetMaxResults() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestGetMirrorLogsReplyGetters(t *testing.T) {
	m := &GetMirrorLogsReply{Line: []string{"line1", "line2"}}
	if len(m.GetLine()) != 2 {
		t.Fatal("Line mismatch")
	}
	m.Reset()
	if len(m.GetLine()) != 0 {
		t.Fatal("Expected empty after reset")
	}
}

func TestGetMirrorLogsReplyNilGetters(t *testing.T) {
	var m *GetMirrorLogsReply
	if m.GetLine() != nil {
		t.Fatal("Expected nil")
	}
}

func TestGetMirrorLogsRequestStringAndProto(t *testing.T) {
	m := &GetMirrorLogsRequest{ID: 3, MaxResults: 100}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
}

func TestGetMirrorLogsReplyStringAndProto(t *testing.T) {
	m := &GetMirrorLogsReply{Line: []string{"line1"}}
	_ = m.String()
	m.ProtoMessage()
	m.ProtoReflect()
	_, _ = m.Descriptor()
}

func TestNewCLIClient(t *testing.T) {
	c := NewCLIClient(nil)
	if c == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestCLIServerDesc(t *testing.T) {
	if _CLI_serviceDesc.ServiceName == "" {
		t.Fatal("Expected non-empty service name")
	}
	if len(_CLI_serviceDesc.Methods) == 0 {
		t.Fatal("Expected methods")
	}
}
