// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthorize_NoMetadata(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	ctx := context.Background()

	err := authorize(ctx)
	if err == nil {
		t.Fatalf("Expected error for no metadata")
	}
}

func TestAuthorize_WrongPassword(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "wrong")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := authorize(ctx)
	if err == nil {
		t.Fatalf("Expected error for wrong password")
	}
}

func TestAuthorize_CorrectPassword(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "secret")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := authorize(ctx)
	if err != nil {
		t.Fatalf("Expected no error for correct password, got %v", err)
	}
}

func TestAuthorize_EmptyPassword(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: ""})

	md := metadata.Pairs("password", "")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := authorize(ctx)
	if err != nil {
		t.Fatalf("Expected no error for matching empty passwords, got %v", err)
	}
}

func TestUnaryInterceptor_Unauthorized(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	ctx := context.Background()

	_, err := UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatalf("Handler should not be called when unauthorized")
		return nil, nil
	})
	if err == nil {
		t.Fatalf("Expected error for unauthorized request")
	}
}

func TestUnaryInterceptor_Authorized(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "secret")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	called := false
	resp, err := UnaryInterceptor(ctx, "request", &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !called {
		t.Fatalf("Handler should have been called")
	}
	if resp != "response" {
		t.Fatalf("Expected 'response', got %v", resp)
	}
}

func TestStreamInterceptor_Unauthorized(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	ctx := context.Background()
	stream := &mockServerStream{ctx: ctx}

	err := StreamInterceptor(nil, stream, &grpc.StreamServerInfo{}, func(srv interface{}, stream grpc.ServerStream) error {
		t.Fatalf("Handler should not be called when unauthorized")
		return nil
	})
	if err == nil {
		t.Fatalf("Expected error for unauthorized request")
	}
}

func TestStreamInterceptor_Authorized(t *testing.T) {
	SetConfiguration(&Configuration{RPCPassword: "secret"})

	md := metadata.Pairs("password", "secret")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	stream := &mockServerStream{ctx: ctx}

	called := false
	err := StreamInterceptor(nil, stream, &grpc.StreamServerInfo{}, func(srv interface{}, stream grpc.ServerStream) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !called {
		t.Fatalf("Handler should have been called")
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
