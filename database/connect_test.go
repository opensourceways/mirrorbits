// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package database

import (
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"gopkg.in/yaml.v2"
)

func sentinelsFromYAML(yamlStr string) *Configuration {
	cnf := &Configuration{}
	yaml.Unmarshal([]byte(yamlStr), cnf)
	return cnf
}

func TestRedis_Connect_NoSentinels_EmptyAddress(t *testing.T) {
	SetConfiguration(&Configuration{
		RedisAddress:           "",
		RedisSentinels:         nil,
		RedisSentinelMasterName: "",
	})
	r := &Redis{pool: nil}
	_, err := r.Connect()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestRedis_Connect_SentinelsEmptyMasterName(t *testing.T) {
	cnf := sentinelsFromYAML("RedisSentinels:\n  - Host: sentinel1:26379\nRedisSentinelMasterName: \"\"\nRedisAddress: \"\"")
	SetConfiguration(cnf)
	r := &Redis{pool: nil}
	_, err := r.Connect()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestRedis_Connect_SentinelsWithMasterName(t *testing.T) {
	cnf := sentinelsFromYAML("RedisSentinels:\n  - Host: nonexistent.invalid:26379\nRedisSentinelMasterName: mymaster\nRedisAddress: \"\"")
	SetConfiguration(cnf)
	r := &Redis{pool: nil}
	_, err := r.Connect()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestRedis_Connect_SentinelsMultipleFail(t *testing.T) {
	cnf := sentinelsFromYAML("RedisSentinels:\n  - Host: nonexistent1.invalid:26379\n  - Host: nonexistent2.invalid:26379\nRedisSentinelMasterName: mymaster\nRedisAddress: \"\"")
	SetConfiguration(cnf)
	r := &Redis{pool: nil}
	_, err := r.Connect()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestRedis_FailureState(t *testing.T) {
	r := &Redis{}

	r.setFailureState(true)
	if !r.Failure() {
		t.Fatalf("Expected failure=true")
	}

	r.setFailureState(false)
	if r.Failure() {
		t.Fatalf("Expected failure=false")
	}
}

func TestRedis_ConnectPubsub_AlreadySet(t *testing.T) {
	_, conn := prepareTestRedis()
	conn.ConnectPubsub()
	conn.ConnectPubsub()
}
