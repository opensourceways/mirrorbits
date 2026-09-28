// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package network

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
)

type redisPoolMockNet struct {
	Conn *redigomock.Conn
}

func (r *redisPoolMockNet) Get() redis.Conn { return r.Conn }
func (r *redisPoolMockNet) Close() error    { return nil }

func prepareNetRedis() (*redigomock.Conn, *database.Redis) {
	mock := redigomock.NewConn()
	pool := &redisPoolMockNet{Conn: mock}
	return mock, database.NewRedisCustomPool(pool)
}

func TestClusterLock_Get_Success(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := prepareNetRedis()

	mock.Command("SET", "SCANNING_1", 1, "NX", "EX", 10).Expect("OK")

	lock := NewClusterLock(conn, "SCANNING_1", "mirror1")
	done, err := lock.Get()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if done == nil {
		t.Fatalf("Expected non-nil done channel")
	}

	lock.Release()
}

func TestClusterLock_Get_AlreadyLocked(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := prepareNetRedis()

	mock.Command("SET", "SCANNING_2", 1, "NX", "EX", 10).ExpectError(redis.ErrNil)

	lock := NewClusterLock(conn, "SCANNING_2", "mirror2")
	done, err := lock.Get()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if done != nil {
		t.Fatalf("Expected nil done channel for already locked")
	}
}

func TestClusterLock_Get_Error(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := prepareNetRedis()

	mock.Command("SET", "SCANNING_3", 1, "NX", "EX", 10).ExpectError(redis.Error("ERR fail"))

	lock := NewClusterLock(conn, "SCANNING_3", "mirror3")
	_, err := lock.Get()
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestClusterLock_Get_AlreadyInUse(t *testing.T) {
	SetConfiguration(&Configuration{})
	mock, conn := prepareNetRedis()

	mock.Command("SET", "SCANNING_4", 1, "NX", "EX", 10).Expect("OK")

	lock := NewClusterLock(conn, "SCANNING_4", "mirror4")
	done1, err := lock.Get()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	_ = done1

	_, err = lock.Get()
	if err == nil {
		t.Fatalf("Expected error for lock already in use")
	}

	lock.Release()
}

func TestClusterLock_NewClusterLock(t *testing.T) {
	SetConfiguration(&Configuration{})
	_, conn := prepareNetRedis()

	lock := NewClusterLock(conn, "key", "identifier")
	if lock == nil {
		t.Fatalf("Expected non-nil lock")
	}
	if lock.key != "key" {
		t.Fatalf("Expected key 'key'")
	}
	if lock.identifier != "identifier" {
		t.Fatalf("Expected identifier 'identifier'")
	}
}

func TestGeoIP_OpenDatabase_NotFound(t *testing.T) {
	SetConfiguration(&Configuration{GeoipDatabasePath: "/nonexistent/path/"})

	g := NewGeoIP()
	_, err := g.openDatabase("test.mmdb")
	if err == nil {
		t.Fatalf("Expected error for non-existent database")
	}
}

func TestGeoIP_LoadGeoIP_Error(t *testing.T) {
	SetConfiguration(&Configuration{GeoipDatabasePath: "/nonexistent/path/"})

	g := NewGeoIP()
	err := g.LoadGeoIP()
	if err == nil {
		t.Fatalf("Expected error for non-existent databases")
	}
}

func TestGeoIP_LoadGeoIP_EmptyPath(t *testing.T) {
	SetConfiguration(&Configuration{GeoipDatabasePath: ""})

	g := NewGeoIP()
	err := g.LoadGeoIP()
	if err == nil {
		t.Fatalf("Expected error for empty path")
	}
}
