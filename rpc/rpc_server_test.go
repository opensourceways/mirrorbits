// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package rpc

import (
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestCLI_Start_Success(t *testing.T) {
	SetConfiguration(&Configuration{
		RPCListenAddress: "127.0.0.1:0",
	})
	c := &CLI{}
	if err := c.Start(); err != nil {
		t.Fatalf("Start failed: %s", err)
	}
	if c.listener == nil {
		t.Fatalf("Expected listener to be set")
	}
	if c.server == nil {
		t.Fatalf("Expected server to be set")
	}
}

func TestCLI_Start_InvalidAddress(t *testing.T) {
	SetConfiguration(&Configuration{
		RPCListenAddress: "invalid:address:format",
	})
	c := &CLI{}
	err := c.Start()
	if err == nil {
		t.Fatalf("Expected error for invalid address")
	}
}
