// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package config

import (
	"os"
	"testing"

	"github.com/opensourceways/mirrorbits/core"
)

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()

	if c.OutputMode != "auto" {
		t.Fatalf("Expected OutputMode 'auto', got %s", c.OutputMode)
	}
	if c.ListenAddress != ":8080" {
		t.Fatalf("Expected ListenAddress ':8080', got %s", c.ListenAddress)
	}
	if c.RedisAddress != "127.0.0.1:6379" {
		t.Fatalf("Expected RedisAddress '127.0.0.1:6379', got %s", c.RedisAddress)
	}
	if c.ConcurrentSync != 50 {
		t.Fatalf("Expected ConcurrentSync 50, got %d", c.ConcurrentSync)
	}
	if c.ScanInterval != 60 {
		t.Fatalf("Expected ScanInterval 60, got %d", c.ScanInterval)
	}
	if c.CheckInterval != 30 {
		t.Fatalf("Expected CheckInterval 30, got %d", c.CheckInterval)
	}
	if c.RepositoryScanInterval != 50 {
		t.Fatalf("Expected RepositoryScanInterval 50, got %d", c.RepositoryScanInterval)
	}
	if c.MaxLinkHeaders != 10 {
		t.Fatalf("Expected MaxLinkHeaders 10, got %d", c.MaxLinkHeaders)
	}
	if c.Gzip != false {
		t.Fatalf("Expected Gzip false")
	}
	if c.RedisDB != 0 {
		t.Fatalf("Expected RedisDB 0, got %d", c.RedisDB)
	}
	if c.GeoipDatabasePath != "/usr/share/GeoIP/" {
		t.Fatalf("Expected GeoipDatabasePath '/usr/share/GeoIP/', got %s", c.GeoipDatabasePath)
	}
	if c.Hashes.SHA256 != true {
		t.Fatalf("Expected SHA256 true")
	}
	if c.Hashes.SHA1 != false {
		t.Fatalf("Expected SHA1 false")
	}
	if c.Hashes.MD5 != false {
		t.Fatalf("Expected MD5 false")
	}
	if c.WeightDistributionRange != 1.5 {
		t.Fatalf("Expected WeightDistributionRange 1.5, got %f", c.WeightDistributionRange)
	}
	if c.RPCListenAddress != "localhost:3390" {
		t.Fatalf("Expected RPCListenAddress 'localhost:3390', got %s", c.RPCListenAddress)
	}
	if c.SchemaStrictMatch != true {
		t.Fatalf("Expected SchemaStrictMatch true")
	}
	if c.DisallowRedirects != false {
		t.Fatalf("Expected DisallowRedirects false")
	}
	if c.DisableOnMissingFile != false {
		t.Fatalf("Expected DisableOnMissingFile false")
	}
}

func TestGetConfig_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic when config not loaded")
		}
	}()
	old := config
	config = nil
	defer func() { config = old }()
	GetConfig()
}

func TestGetConfig_Valid(t *testing.T) {
	old := config
	defer func() { config = old }()

	c := &Configuration{Repository: "/tmp/test"}
	SetConfiguration(c)

	got := GetConfig()
	if got.Repository != "/tmp/test" {
		t.Fatalf("Expected Repository '/tmp/test', got %s", got.Repository)
	}
}

func TestSetConfiguration(t *testing.T) {
	c := &Configuration{Repository: "/tmp/test2"}
	SetConfiguration(c)

	if GetConfig() != c {
		t.Fatalf("SetConfiguration did not set the config correctly")
	}
}

func TestSubscribeConfig(t *testing.T) {
	old := subscribers
	defer func() { subscribers = old }()
	subscribers = nil

	ch := make(chan bool, 1)
	SubscribeConfig(ch)

	if len(subscribers) != 1 {
		t.Fatalf("Expected 1 subscriber, got %d", len(subscribers))
	}
}

func TestNotifySubscribers(t *testing.T) {
	old := subscribers
	defer func() { subscribers = old }()
	subscribers = nil

	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)
	SubscribeConfig(ch1)
	SubscribeConfig(ch2)

	notifySubscribers()

	select {
	case <-ch1:
	default:
		t.Fatalf("Subscriber 1 should have been notified")
	}

	select {
	case <-ch2:
	default:
		t.Fatalf("Subscriber 2 should have been notified")
	}
}

func TestNotifySubscribers_NotBlocking(t *testing.T) {
	old := subscribers
	defer func() { subscribers = old }()
	subscribers = nil

	ch := make(chan bool, 1)
	SubscribeConfig(ch)

	notifySubscribers()
	notifySubscribers()

	select {
	case <-ch:
	default:
		t.Fatalf("Subscriber should have been notified")
	}
}

func TestFileExists(t *testing.T) {
	if fileExists("/nonexistent/path/to/file") {
		t.Fatalf("Expected false for non-existent file")
	}

	tmpFile, err := os.CreateTemp("", "config-test")
	if err != nil {
		t.Fatalf("Unable to create temp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	if !fileExists(tmpFile.Name()) {
		t.Fatalf("Expected true for existing file")
	}
}

func TestIsInSlice(t *testing.T) {
	list := []string{"a", "b", "c"}

	if !isInSlice("a", list) {
		t.Fatalf("Expected true for 'a'")
	}
	if isInSlice("d", list) {
		t.Fatalf("Expected false for 'd'")
	}
	if isInSlice("", list) {
		t.Fatalf("Expected false for empty string")
	}
}

func TestTestSentinelsEq(t *testing.T) {
	a := []sentinels{{Host: "h1"}, {Host: "h2"}}
	b := []sentinels{{Host: "h1"}, {Host: "h2"}}

	if !testSentinelsEq(a, b) {
		t.Fatalf("Expected true for equal sentinels")
	}

	b = []sentinels{{Host: "h1"}, {Host: "h3"}}
	if testSentinelsEq(a, b) {
		t.Fatalf("Expected false for different sentinels")
	}

	if testSentinelsEq(a, []sentinels{{Host: "h1"}}) {
		t.Fatalf("Expected false for different length")
	}

	if !testSentinelsEq([]sentinels{}, []sentinels{}) {
		t.Fatalf("Expected true for empty sentinels")
	}
}

func TestGetRedisAddress(t *testing.T) {
	os.Setenv("REDIS_ADDRESS_PORT", "localhost:6379")
	defer os.Unsetenv("REDIS_ADDRESS_PORT")

	if r := GetRedisAddress(); r != "localhost:6379" {
		t.Fatalf("Expected 'localhost:6379', got %s", r)
	}
}

func TestGetRedisPwd(t *testing.T) {
	os.Setenv("REDIS_PWD", "secret")
	defer os.Unsetenv("REDIS_PWD")

	if r := GetRedisPwd(); r != "secret" {
		t.Fatalf("Expected 'secret', got %s", r)
	}
}

func TestLoadConfig_AlreadyLoaded(t *testing.T) {
	old := config
	defer func() { config = old }()

	config = &Configuration{Repository: "/tmp"}
	LoadConfig()

	if config.Repository != "/tmp" {
		t.Fatalf("LoadConfig should not reload if already loaded")
	}
}

func TestReloadConfig_ValidFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	configContent := []byte("Repository: " + dir + "\n")
	configFile := dir + "/mirrorbits.conf"
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Unable to write config file: %s", err)
	}

	old := config
	defer func() { config = old }()
	oldConfigFile := core.ConfigFile
	defer func() { core.ConfigFile = oldConfigFile }()

	core.ConfigFile = configFile
	config = nil

	err = ReloadConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	c := GetConfig()
	if c.Repository != dir {
		t.Fatalf("Expected Repository '%s', got '%s'", dir, c.Repository)
	}
}

func TestReloadConfig_InvalidOutputMode(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	configContent := []byte("Repository: " + dir + "\nOutputMode: invalid\n")
	configFile := dir + "/mirrorbits.conf"
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Unable to write config file: %s", err)
	}

	old := config
	defer func() { config = old }()
	oldConfigFile := core.ConfigFile
	defer func() { core.ConfigFile = oldConfigFile }()

	core.ConfigFile = configFile
	config = nil

	err = ReloadConfig()
	if err == nil {
		t.Fatalf("Expected error for invalid OutputMode")
	}
}

func TestReloadConfig_WeightDistributionRangeZero(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	configContent := []byte("Repository: " + dir + "\nWeightDistributionRange: 0\n")
	configFile := dir + "/mirrorbits.conf"
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Unable to write config file: %s", err)
	}

	old := config
	defer func() { config = old }()
	oldConfigFile := core.ConfigFile
	defer func() { core.ConfigFile = oldConfigFile }()

	core.ConfigFile = configFile
	config = nil

	err = ReloadConfig()
	if err == nil {
		t.Fatalf("Expected error for WeightDistributionRange <= 0")
	}
}

func TestReloadConfig_EmptyRepository(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	configContent := []byte("Repository: \"\"\n")
	configFile := dir + "/mirrorbits.conf"
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Unable to write config file: %s", err)
	}

	old := config
	defer func() { config = old }()
	oldConfigFile := core.ConfigFile
	defer func() { core.ConfigFile = oldConfigFile }()

	core.ConfigFile = configFile
	config = nil

	err = ReloadConfig()
	if err == nil {
		t.Fatalf("Expected error for empty Repository")
	}
}
