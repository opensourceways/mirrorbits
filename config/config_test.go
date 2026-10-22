// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	if c.ListenAddress != ":8080" {
		t.Fatalf("Expected :8080, got %s", c.ListenAddress)
	}
	if c.RedisAddress != "127.0.0.1:6379" {
		t.Fatalf("Expected 127.0.0.1:6379, got %s", c.RedisAddress)
	}
	if c.ConcurrentSync != 50 {
		t.Fatalf("Expected 50, got %d", c.ConcurrentSync)
	}
	if c.Hashes.SHA256 != true {
		t.Fatalf("Expected SHA256 to be true")
	}
	if c.Hashes.SHA1 != false {
		t.Fatalf("Expected SHA1 to be false")
	}
	if c.OutputMode != "auto" {
		t.Fatalf("Expected auto, got %s", c.OutputMode)
	}
	if c.WeightDistributionRange != 1.5 {
		t.Fatalf("Expected 1.5, got %f", c.WeightDistributionRange)
	}
	if c.RPCListenAddress != "localhost:3390" {
		t.Fatalf("Expected localhost:3390, got %s", c.RPCListenAddress)
	}
	if c.SchemaStrictMatch != true {
		t.Fatalf("Expected SchemaStrictMatch to be true")
	}
}

func TestSetAndGetConfiguration(t *testing.T) {
	c := &Configuration{
		Repository:    "/tmp/testrepo",
		ListenAddress: ":9090",
	}
	SetConfiguration(c)
	got := GetConfig()
	if got.Repository != "/tmp/testrepo" {
		t.Fatalf("Expected /tmp/testrepo, got %s", got.Repository)
	}
	if got.ListenAddress != ":9090" {
		t.Fatalf("Expected :9090, got %s", got.ListenAddress)
	}
}

func TestGetConfigPanics(t *testing.T) {
	old := config
	config = nil
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic when config is nil")
		}
		config = old
	}()
	GetConfig()
}

func TestFileExists(t *testing.T) {
	if !fileExists(".") {
		t.Fatal("Expected . to exist")
	}
	if fileExists("/nonexistent/path/that/should/not/exist") {
		t.Fatal("Expected nonexistent path to return false")
	}
}

func TestIsInSlice(t *testing.T) {
	list := []string{"a", "b", "c"}
	if !isInSlice("a", list) {
		t.Fatal("Expected true for 'a'")
	}
	if isInSlice("d", list) {
		t.Fatal("Expected false for 'd'")
	}
	if isInSlice("", list) {
		t.Fatal("Expected false for empty string")
	}
}

func TestTestSentinelsEq(t *testing.T) {
	a := []sentinels{{Host: "h1"}, {Host: "h2"}}
	b := []sentinels{{Host: "h1"}, {Host: "h2"}}
	if !testSentinelsEq(a, b) {
		t.Fatal("Expected equal sentinels to be equal")
	}
	c := []sentinels{{Host: "h1"}}
	if testSentinelsEq(a, c) {
		t.Fatal("Expected different length sentinels to be unequal")
	}
	d := []sentinels{{Host: "h1"}, {Host: "h3"}}
	if testSentinelsEq(a, d) {
		t.Fatal("Expected different hosts to be unequal")
	}
}

func TestGetRedisAddress(t *testing.T) {
	old := os.Getenv("REDIS_ADDRESS_PORT")
	defer os.Setenv("REDIS_ADDRESS_PORT", old)

	os.Setenv("REDIS_ADDRESS_PORT", "localhost:6380")
	if r := GetRedisAddress(); r != "localhost:6380" {
		t.Fatalf("Expected localhost:6380, got %s", r)
	}
	os.Setenv("REDIS_ADDRESS_PORT", "")
	if r := GetRedisAddress(); r != "" {
		t.Fatalf("Expected empty, got %s", r)
	}
}

func TestGetRedisPwd(t *testing.T) {
	old := os.Getenv("REDIS_PWD")
	defer os.Setenv("REDIS_PWD", old)

	os.Setenv("REDIS_PWD", "secret")
	if r := GetRedisPwd(); r != "secret" {
		t.Fatalf("Expected secret, got %s", r)
	}
	os.Setenv("REDIS_PWD", "")
	if r := GetRedisPwd(); r != "" {
		t.Fatalf("Expected empty, got %s", r)
	}
}

func TestSubscribeConfigAndNotify(t *testing.T) {
	oldSubscribers := subscribers
	defer func() {
		subscribers = oldSubscribers
	}()
	subscribers = nil

	ch := make(chan bool, 1)
	SubscribeConfig(ch)

	if len(subscribers) != 1 {
		t.Fatalf("Expected 1 subscriber, got %d", len(subscribers))
	}

	notifySubscribers()

	select {
	case <-ch:
	default:
		t.Fatal("Expected to receive notification")
	}
}

func TestParseConfigValid(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-repo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	content := []byte("Repository: \"" + dir + "\"\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err = parseConfig(content, "test.conf")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	got := GetConfig()
	if got.Repository != dir {
		t.Fatalf("Expected %s, got %s", dir, got.Repository)
	}
}

func TestParseConfigEmptyRepo(t *testing.T) {
	content := []byte("Repository: \"\"\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err := parseConfig(content, "test.conf")
	if err == nil {
		t.Fatal("Expected error for empty repository")
	}
}

func TestParseConfigBadOutputMode(t *testing.T) {
	dir, _ := os.MkdirTemp("", "mirrorbits-repo-*")
	defer os.RemoveAll(dir)

	content := []byte("Repository: \"" + dir + "\"\nOutputMode: \"invalid\"\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err := parseConfig(content, "test.conf")
	if err == nil {
		t.Fatal("Expected error for invalid outputMode")
	}
}

func TestParseConfigBadWeightRange(t *testing.T) {
	dir, _ := os.MkdirTemp("", "mirrorbits-repo-*")
	defer os.RemoveAll(dir)

	content := []byte("Repository: \"" + dir + "\"\nWeightDistributionRange: 0\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err := parseConfig(content, "test.conf")
	if err == nil {
		t.Fatal("Expected error for zero WeightDistributionRange")
	}
}

func TestParseConfigNegativeScanInterval(t *testing.T) {
	dir, _ := os.MkdirTemp("", "mirrorbits-repo-*")
	defer os.RemoveAll(dir)

	content := []byte("Repository: \"" + dir + "\"\nRepositoryScanInterval: -5\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err := parseConfig(content, "test.conf")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	got := GetConfig()
	if got.RepositoryScanInterval != 0 {
		t.Fatalf("Expected 0, got %d", got.RepositoryScanInterval)
	}
}

func TestParseConfigInvalidYAML(t *testing.T) {
	content := []byte("Repository: [invalid yaml\n")
	oldConfig := config
	defer func() { config = oldConfig }()
	config = nil

	err := parseConfig(content, "test.conf")
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
}

func TestLoadConfigAlreadyLoaded(t *testing.T) {
	oldConfig := config
	defer func() { config = oldConfig }()

	c := &Configuration{Repository: "/test"}
	config = c
	LoadConfig()
	if config != c {
		t.Fatal("Expected config to remain unchanged when already loaded")
	}
}
