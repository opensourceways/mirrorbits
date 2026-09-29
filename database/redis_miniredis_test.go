package database

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/gomodule/redigo/redis"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *Redis, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %s", err)
	}

	mr.Server().Register("ROLE", func(c *server.Peer, cmd string, args []string) {
		c.WriteLen(3)
		c.WriteBulk("master")
		c.WriteInt(0)
		c.WriteLen(0)
	})
	SetConfiguration(&Configuration{
		RedisAddress: mr.Addr(),
		RedisDB:      0,
	})

	r := NewRedis()
	r.ConnectPubsub()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn := r.Get()
		if _, ok := conn.(*NotReadyError); !ok {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cleanup := func() {
		r.Close()
		mr.Close()
	}
	return mr, r, cleanup
}

func TestNewRedis_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	if r == nil {
		t.Fatal("Expected non-nil Redis")
	}
	if r.Failure() {
		t.Fatal("Expected failure=false")
	}

	conn, err := r.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %s", err)
	}
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		t.Fatalf("PING failed: %s", err)
	}
}

func TestConnect_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	conn, err := r.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %s", err)
	}
	defer conn.Close()

	roleReply, err := redis.Values(conn.Do("ROLE"))
	if err != nil {
		t.Fatalf("ROLE failed: %s", err)
	}
	role, err := redis.String(roleReply[0], nil)
	if err != nil {
		t.Fatalf("Convert role failed: %s", err)
	}
	if role != "master" {
		t.Fatalf("Expected master role, got %s", role)
	}
}

func TestConnect_NoRedisAddress(t *testing.T) {
	SetConfiguration(&Configuration{
		RedisAddress: "",
		RedisDB:      0,
	})
	r := NewRedisCustomPool(nil)
	defer func() {
		select {
		case <-r.stop:
			return
		default:
			close(r.stop)
		}
	}()

	_, err := r.Connect()
	if err != ErrUnreachable {
		t.Fatalf("Expected ErrUnreachable, got %v", err)
	}
}

func TestConnect_UnreachableAddress(t *testing.T) {
	SetConfiguration(&Configuration{
		RedisAddress: "127.0.0.1:1",
		RedisDB:      0,
	})
	r := NewRedisCustomPool(nil)
	defer func() {
		select {
		case <-r.stop:
			return
		default:
			close(r.stop)
		}
	}()

	_, err := r.Connect()
	if err == nil {
		t.Fatal("Expected error for unreachable address")
	}
}

func TestGetListOfMirrors_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredis(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRRORS", "2", "mirror2")

	mirrors, err := r.GetListOfMirrors()
	if err != nil {
		t.Fatalf("GetListOfMirrors failed: %s", err)
	}

	if len(mirrors) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(mirrors))
	}
	if mirrors[1] != "mirror1" {
		t.Fatalf("Expected mirror1, got %s", mirrors[1])
	}
	if mirrors[2] != "mirror2" {
		t.Fatalf("Expected mirror2, got %s", mirrors[2])
	}
}

func TestGetListOfMirrors_Empty(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	mirrors, err := r.GetListOfMirrors()
	if err != nil {
		t.Fatalf("GetListOfMirrors failed: %s", err)
	}
	if len(mirrors) != 0 {
		t.Fatalf("Expected 0 mirrors, got %d", len(mirrors))
	}
}

func TestCheckVersion_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	err := r.CheckVersion()
	if err == ErrRedisUpgradeRequired {
		t.Fatalf("Should not get ErrRedisUpgradeRequired with miniredis")
	}
}

func TestGetDBFormatVersion_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredis(t)
	defer cleanup()

	version, err := r.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("GetDBFormatVersion failed: %s", err)
	}
	if version != core.DBVersion {
		t.Fatalf("Expected version %d, got %d", core.DBVersion, version)
	}

	mr.Set(core.DBVersionKey, "1")
	version, err = r.GetDBFormatVersion()
	if err != nil {
		t.Fatalf("GetDBFormatVersion failed: %s", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

func TestUpgradeNeeded_Miniredis(t *testing.T) {
	mr, r, cleanup := setupMiniredis(t)
	defer cleanup()

	needed, err := r.UpgradeNeeded()
	if err != nil {
		t.Fatalf("UpgradeNeeded failed: %s", err)
	}
	if needed {
		t.Fatal("Expected no upgrade needed")
	}

	mr.Set(core.DBVersionKey, "0")
	needed, err = r.UpgradeNeeded()
	if err != nil {
		t.Fatalf("UpgradeNeeded failed: %s", err)
	}
	if !needed {
		t.Fatal("Expected upgrade needed")
	}
}

func TestAcquireLock_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	lock, err := r.AcquireLock("testlock")
	if err != nil {
		t.Fatalf("AcquireLock failed: %s", err)
	}
	if lock == nil {
		t.Fatal("Expected non-nil lock")
	}

	_, err = r.AcquireLock("testlock")
	if err != ErrAlreadyLocked {
		t.Fatalf("Expected ErrAlreadyLocked, got %v", err)
	}

	lock.Release()

	lock2, err := r.AcquireLock("testlock")
	if err != nil {
		t.Fatalf("AcquireLock after release failed: %s", err)
	}
	lock2.Release()
}

func TestAcquireLock_EmptyName(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	_, err := r.AcquireLock("")
	if err != ErrInvalidLockName {
		t.Fatalf("Expected ErrInvalidLockName, got %v", err)
	}
}

func TestClose_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	r.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked: %v", r)
		}
	}()

	r.Close()
}

func TestNewRedis_FailureState(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	if r.Failure() {
		t.Fatal("Expected failure=false initially")
	}

	r.setFailureState(true)
	if !r.Failure() {
		t.Fatal("Expected failure=true after setFailureState")
	}

	r.setFailureState(false)
	if r.Failure() {
		t.Fatal("Expected failure=false after reset")
	}
}

func TestGet_Ready_Miniredis(t *testing.T) {
	_, r, cleanup := setupMiniredis(t)
	defer cleanup()

	conn := r.Get()
	defer conn.Close()

	if conn.Err() != nil {
		t.Fatalf("Expected nil error from Get, got %s", conn.Err())
	}

	_, err := conn.Do("PING")
	if err != nil {
		t.Fatalf("PING failed: %s", err)
	}
}
