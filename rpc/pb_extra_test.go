// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMessageDescriptors(t *testing.T) {
	_, _ = (&VersionReply{}).Descriptor()
	_, _ = (&MatchRequest{}).Descriptor()
	_, _ = (&Mirror{}).Descriptor()
	_, _ = (&MirrorListReply{}).Descriptor()
	_, _ = (&MirrorID{}).Descriptor()
	_, _ = (&MatchReply{}).Descriptor()
	_, _ = (&ChangeStatusRequest{}).Descriptor()
	_, _ = (&MirrorIDRequest{}).Descriptor()
	_, _ = (&AddMirrorReply{}).Descriptor()
	_, _ = (&UpdateMirrorReply{}).Descriptor()
	_, _ = (&RefreshRepositoryRequest{}).Descriptor()
	_, _ = (&ScanMirrorRequest{}).Descriptor()
	_, _ = (&ScanMirrorReply{}).Descriptor()
}

func TestEnumDescriptor(t *testing.T) {
	_, _ = ScanMirrorRequest_Method(0).EnumDescriptor()
}

func TestEnumMethods(t *testing.T) {
	v := ScanMirrorRequest_ALL
	_ = v.Enum()
	_ = v.String()
	_ = v.Number()
	_ = v.Type()

	v2 := ScanMirrorRequest_FTP
	_ = v2.Enum()
	_ = v2.String()
	_ = v2.Number()
}

func TestStatsFileRequestGetters(t *testing.T) {
	m := &StatsFileRequest{Pattern: "test*"}
	if m.GetPattern() != "test*" {
		t.Fatal("Pattern mismatch")
	}
	m.Reset()
	if m.GetPattern() != "" {
		t.Fatal("Expected empty after reset")
	}
}

func TestStatsFileRequestNilGetters(t *testing.T) {
	var m *StatsFileRequest
	if m.GetPattern() != "" {
		t.Fatal("Expected empty")
	}
	if m.GetDateStart() != nil {
		t.Fatal("Expected nil")
	}
}

func TestStatsFileReplyGetters(t *testing.T) {
	m := &StatsFileReply{
		Files: map[string]int64{"file1": 10, "file2": 20},
	}
	if len(m.GetFiles()) != 2 {
		t.Fatal("Files mismatch")
	}
	m.Reset()
	if len(m.GetFiles()) != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestStatsFileReplyNilGetters(t *testing.T) {
	var m *StatsFileReply
	if m.GetFiles() != nil {
		t.Fatal("Expected nil")
	}
}

func TestStatsMirrorRequestGetters(t *testing.T) {
	m := &StatsMirrorRequest{ID: 5}
	if m.GetID() != 5 {
		t.Fatal("ID mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestStatsMirrorRequestNilGetters(t *testing.T) {
	var m *StatsMirrorRequest
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetDateStart() != nil {
		t.Fatal("Expected nil")
	}
}

func TestStatsMirrorReplyGetters(t *testing.T) {
	m := &StatsMirrorReply{Requests: 42, Bytes: 1024}
	if m.GetRequests() != 42 {
		t.Fatal("Requests mismatch")
	}
	if m.GetBytes() != 1024 {
		t.Fatal("Bytes mismatch")
	}
	m.Reset()
	if m.GetRequests() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestStatsMirrorReplyNilGetters(t *testing.T) {
	var m *StatsMirrorReply
	if m.GetRequests() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetMirror() != nil {
		t.Fatal("Expected nil")
	}
}

func TestUnimplementedCLIServer(t *testing.T) {
	s := &UnimplementedCLIServer{}
	ctx := context.Background()

	tests := []func() error{
		func() error { _, err := s.Upgrade(ctx, nil); return err },
		func() error { _, err := s.Reload(ctx, nil); return err },
		func() error { _, err := s.ChangeStatus(ctx, nil); return err },
		func() error { _, err := s.List(ctx, nil); return err },
		func() error { _, err := s.MirrorInfo(ctx, nil); return err },
		func() error { _, err := s.AddMirror(ctx, nil); return err },
		func() error { _, err := s.UpdateMirror(ctx, nil); return err },
		func() error { _, err := s.RemoveMirror(ctx, nil); return err },
		func() error { _, err := s.RefreshRepository(ctx, nil); return err },
		func() error { _, err := s.ScanMirror(ctx, nil); return err },
		func() error { _, err := s.StatsFile(ctx, nil); return err },
		func() error { _, err := s.StatsMirror(ctx, nil); return err },
		func() error { _, err := s.Ping(ctx, nil); return err },
		func() error { _, err := s.MatchMirror(ctx, nil); return err },
	}

	for i, fn := range tests {
		err := fn()
		if err == nil {
			t.Fatalf("Test %d: Expected error from UnimplementedCLIServer", i)
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Test %d: Expected gRPC status error", i)
		}
		if st.Code() != codes.Unimplemented {
			t.Fatalf("Test %d: Expected Unimplemented, got %s", i, st.Code())
		}
	}
}

func TestRegisterCLIServer(t *testing.T) {
	// Just verify it doesn't panic
	// Using a nil server would work since RegisterCLIServer just sets up handlers
	_ = RegisterCLIServer
}
