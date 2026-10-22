// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestAuthorizeNoMetadata(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	ctx := context.Background()
	err := authorize(ctx)
	if err == nil {
		t.Fatal("Expected error for no metadata")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("Expected Unauthenticated, got %s", st.Code())
	}
}

func TestAuthorizeWrongPassword(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "wrong")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := authorize(ctx)
	if err == nil {
		t.Fatal("Expected error for wrong password")
	}
}

func TestAuthorizeCorrectPassword(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "secret")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := authorize(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestUnaryInterceptor(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "secret")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	result, err := UnaryInterceptor(ctx, nil, nil, handler)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Expected 'ok', got %v", result)
	}
}

func TestUnaryInterceptorUnauthorized(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("Handler should not be called")
		return nil, nil
	}

	_, err := UnaryInterceptor(ctx, nil, nil, handler)
	if err == nil {
		t.Fatal("Expected error for unauthorized request")
	}
}

func TestMirrorToRPC(t *testing.T) {
	m := &mirrors.Mirror{
		ID:            1,
		Name:          "test",
		HttpURL:       "http://test.com",
		RsyncURL:      "rsync://test.com",
		FtpURL:        "ftp://test.com",
		SponsorName:   "sponsor",
		SponsorURL:    "http://sponsor.com",
		AdminName:     "admin",
		AdminEmail:    "admin@test.com",
		ContinentCode: "EU",
		CountryCodes:  "FR DE",
		Asnum:         123,
		Comment:       "comment",
		Enabled:       true,
		Up:            true,
	}

	rpcM, err := MirrorToRPC(m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rpcM.ID != 1 {
		t.Fatalf("Expected ID 1, got %d", rpcM.ID)
	}
	if rpcM.Name != "test" {
		t.Fatalf("Expected Name test, got %s", rpcM.Name)
	}
	if rpcM.HttpURL != "http://test.com" {
		t.Fatalf("Expected http://test.com, got %s", rpcM.HttpURL)
	}
}

func TestMirrorFromRPC(t *testing.T) {
	now := time.Now()
	ts, _ := ptypes.TimestampProto(now)

	rpcM := &Mirror{
		ID:            2,
		Name:          "test2",
		HttpURL:       "http://test2.com",
		RsyncURL:      "rsync://test2.com",
		FtpURL:        "ftp://test2.com",
		SponsorName:   "sponsor2",
		SponsorURL:    "http://sponsor2.com",
		AdminName:     "admin2",
		AdminEmail:    "admin2@test.com",
		ContinentCode: "AS",
		CountryCodes: "JP KR",
		Asnum:         456,
		Comment:       "comment2",
		Enabled:       false,
		Up:            false,
		StateSince:           ts,
		LastSync:              ts,
		LastSuccessfulSync:    ts,
		LastModTime:           ts,
	}

	m, err := MirrorFromRPC(rpcM)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if m.ID != 2 {
		t.Fatalf("Expected ID 2, got %d", m.ID)
	}
	if m.Name != "test2" {
		t.Fatalf("Expected Name test2, got %s", m.Name)
	}
	if m.HttpURL != "http://test2.com" {
		t.Fatalf("Expected http://test2.com, got %s", m.HttpURL)
	}
}

func TestMirrorFromRPCNilTimestamps(t *testing.T) {
	rpcM := &Mirror{
		ID:  3,
	}

	_, err := MirrorFromRPC(rpcM)
	if err == nil {
		t.Fatal("Expected error for nil timestamps")
	}
}

func TestMirrorToAndFromRPCRoundtrip(t *testing.T) {
	original := &mirrors.Mirror{
		ID:            42,
		Name:          "roundtrip",
		HttpURL:       "http://rt.com",
		RsyncURL:      "rsync://rt.com",
		FtpURL:        "ftp://rt.com",
		SponsorName:   "rt-sponsor",
		SponsorURL:    "http://rt-sponsor.com",
		AdminName:     "rt-admin",
		AdminEmail:    "rt@test.com",
		CustomData:    "custom",
		ContinentOnly: true,
		CountryOnly:   false,
		ASOnly:        true,
		Score:         5,
		Latitude:      48.85,
		Longitude:     2.35,
		ContinentCode: "EU",
		CountryCodes:  "FR",
		Asnum:         789,
		Comment:       "rt-comment",
		Enabled:       true,
		Up:            true,
		Country:       "France",
	}

	rpcM, err := MirrorToRPC(original)
	if err != nil {
		t.Fatalf("MirrorToRPC error: %s", err)
	}

	roundtrip, err := MirrorFromRPC(rpcM)
	if err != nil {
		t.Fatalf("MirrorFromRPC error: %s", err)
	}

	if roundtrip.ID != original.ID {
		t.Fatalf("ID mismatch: %d vs %d", roundtrip.ID, original.ID)
	}
	if roundtrip.Name != original.Name {
		t.Fatalf("Name mismatch: %s vs %s", roundtrip.Name, original.Name)
	}
	if roundtrip.HttpURL != original.HttpURL {
		t.Fatalf("HttpURL mismatch: %s vs %s", roundtrip.HttpURL, original.HttpURL)
	}
	if roundtrip.Asnum != original.Asnum {
		t.Fatalf("Asnum mismatch: %d vs %d", roundtrip.Asnum, original.Asnum)
	}
}
