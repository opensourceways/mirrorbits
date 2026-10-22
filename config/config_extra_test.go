// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package config

import (
	"os"
	"testing"
)

func TestLoadConfigAlreadyLoadedExtra(t *testing.T) {
	oldConfig := config
	defer func() { config = oldConfig }()
	config = &Configuration{Repository: "/tmp"}

	LoadConfig()
}

func TestDefaultConfigExtra(t *testing.T) {
	c := defaultConfig()
	if c.WeightDistributionRange <= 0 {
		t.Fatal("Expected positive WeightDistributionRange")
	}
	if c.OutputMode == "" {
		t.Fatal("Expected non-empty OutputMode")
	}
}

func TestSetConfigurationExtra(t *testing.T) {
	c := &Configuration{Repository: "/test"}
	SetConfiguration(c)
	if GetConfig().Repository != "/test" {
		t.Fatal("Configuration not set")
	}
}

func TestSubscribeConfigExtra(t *testing.T) {
	ch := make(chan bool, 1)
	SubscribeConfig(ch)
}

func TestFileExistsExtra(t *testing.T) {
	if fileExists("/nonexistent/file") {
		t.Fatal("Expected false for non-existent file")
	}

	dir, _ := os.MkdirTemp("", "mirrorbits-config-test-*")
	defer os.RemoveAll(dir)
	f, _ := os.CreateTemp(dir, "test.conf")
	f.Close()

	if !fileExists(f.Name()) {
		t.Fatal("Expected true for existing file")
	}
}

func TestIsInSliceExtra(t *testing.T) {
	if !isInSlice("json", []string{"auto", "json", "redirect"}) {
		t.Fatal("Expected true")
	}
	if isInSlice("invalid", []string{"auto", "json", "redirect"}) {
		t.Fatal("Expected false")
	}
}

func TestGetRedisAddressExtra(t *testing.T) {
	SetConfiguration(&Configuration{RedisAddress: "localhost:6379"})
	_ = GetRedisAddress()
}

func TestGetRedisPwdExtra(t *testing.T) {
	SetConfiguration(&Configuration{RedisPassword: "secret"})
	_ = GetRedisPwd()
}

func TestTestSentinelsEqExtra(t *testing.T) {
	s1 := sentinels{Host: "host1"}
	s2 := sentinels{Host: "host2"}
	a := []sentinels{s1}
	b := []sentinels{s1}
	if !testSentinelsEq(a, b) {
		t.Fatal("Expected true for equal sentinels")
	}
	c := []sentinels{s2}
	if testSentinelsEq(a, c) {
		t.Fatal("Expected false for different sentinels")
	}
}
