// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package daemon

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/opensourceways/mirrorbits/config"
	. "github.com/opensourceways/mirrorbits/testing"
)

func TestMonitor_GetRandomFile_Success(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	m := &monitor{
		redis: conn,
		mirrors: make(map[int]*mirror),
	}

	mock.Command("SRANDMEMBER", "HANDLEDFILES_1").Expect("file.iso")
	mock.Command("HGET", "FILE_file.iso", "size").Expect(int64(1000))

	file, size, err := m.getRandomFile(1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if file != "file.iso" {
		t.Fatalf("Expected 'file.iso', got %s", file)
	}
	if size != 1000 {
		t.Fatalf("Expected 1000, got %d", size)
	}
}

func TestMonitor_GetRandomFile_NoFiles(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	m := &monitor{
		redis: conn,
		mirrors: make(map[int]*mirror),
	}

	mock.Command("SRANDMEMBER", "HANDLEDFILES_2").ExpectError(redis.ErrNil)

	_, _, err := m.getRandomFile(2)
	if err == nil {
		t.Fatalf("Expected error for no files")
	}
}

func TestMonitor_MirrorsID_Unreachable(t *testing.T) {
	SetConfiguration(&Configuration{RedisAddress: ""})
	_, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	m := &monitor{
		redis: conn,
		mirrors: make(map[int]*mirror),
	}

	_, err := m.mirrorsID()
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestMonitor_Stop_AlreadyStopped(t *testing.T) {
	_, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	m := &monitor{
		redis: conn,
		cluster: NewCluster(conn),
		mirrors: make(map[int]*mirror),
		healthCheckChan: make(chan int, 1),
		syncChan: make(chan int),
		stop: make(chan struct{}),
	}

	m.Stop()
	m.Stop()
}

func TestMonitor_Wait(t *testing.T) {
	m := &monitor{}
	m.Wait()
}
