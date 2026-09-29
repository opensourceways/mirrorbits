// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"os"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/golang/protobuf/ptypes"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type rpcTestPool struct {
	Conn *redigomock.Conn
}

func (r *rpcTestPool) Get() redis.Conn {
	return r.Conn
}

func (r *rpcTestPool) Close() error {
	return nil
}

func newRPCRedis(mock *redigomock.Conn) *database.Redis {
	pool := &rpcTestPool{Conn: mock}
	return database.NewRedisCustomPool(pool)
}

func TestCLIMatchMirrorWithRedis(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	mock.Command("HGETALL", "MIRRORS").Expect([]interface{}{})

	r := newRPCRedis(mock)
	c := &CLI{redis: r}

	ctx := context.Background()
	_, err := c.MatchMirror(ctx, &MatchRequest{Pattern: "test"})
	_ = err
}

func TestCLIChangeStatusEnable(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	mock.Command("HMSET", "MIRROR_1", "enabled", true).Expect("OK")
	mock.GenericCommand("PUBLISH").Expect("OK")

	r := newRPCRedis(mock)
	c := &CLI{redis: r}

	ctx := context.Background()
	_, err := c.ChangeStatus(ctx, &ChangeStatusRequest{ID: 1, Enabled: true})
	_ = err
}

func TestCLIChangeStatusDisable(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	mock.Command("HMSET", "MIRROR_1", "enabled", false).Expect("OK")
	mock.GenericCommand("PUBLISH").Expect("OK")

	r := newRPCRedis(mock)
	c := &CLI{redis: r}

	ctx := context.Background()
	_, err := c.ChangeStatus(ctx, &ChangeStatusRequest{ID: 1, Enabled: false})
	_ = err
}

func TestCLIGetMirrorLogsWithRedis(t *testing.T) {
	SetConfiguration(&Configuration{})

	mock := redigomock.NewConn()
	mock.Command("LRANGE", "MIRRORLOGS_1", 0, 99).Expect([]interface{}{})

	r := newRPCRedis(mock)
	c := &CLI{redis: r}

	ctx := context.Background()
	_, err := c.GetMirrorLogs(ctx, &GetMirrorLogsRequest{ID: 1, MaxResults: 100})
	_ = err
}

func TestCLIUpgradeNoSignal(t *testing.T) {
	c := &CLI{sig: make(chan<- os.Signal, 1)}
	ctx := context.Background()
	_, err := c.Upgrade(ctx, nil)
	_ = err
}

func TestCLIReloadNoSignal(t *testing.T) {
	c := &CLI{}
	ctx := context.Background()
	_, err := c.Reload(ctx, nil)
	_ = err
}

func TestCreateDiffMultiple(t *testing.T) {
	m1 := &mirrors.Mirror{
		Name:          "test",
		HttpURL:       "http://old.com",
		Enabled:       true,
		Score:         5,
		ContinentCode: "EU",
	}
	m2 := &mirrors.Mirror{
		Name:          "test",
		HttpURL:       "http://new.com",
		Enabled:       false,
		Score:         10,
		ContinentCode: "AS",
	}
	diff := createDiff(m1, m2)
	if diff == "" {
		t.Fatal("Expected non-empty diff")
	}
}

func TestCLISetMirrorNilRedis(t *testing.T) {
	c := &CLI{redis: nil}
	err := c.setMirror(&mirrors.Mirror{Name: "test", HttpURL: "http://test.com"})
	if err == nil {
		t.Fatal("Expected error for nil redis")
	}
}

func TestMirrorFromRPCNilTimestamp(t *testing.T) {
	rpcM := &Mirror{
		ID:        1,
		Name:      "test",
		StateSince: nil,
	}
	_, err := MirrorFromRPC(rpcM)
	if err == nil {
		t.Fatal("Expected error for nil timestamp")
	}
}

func TestMirrorToRPCWithTimestamps(t *testing.T) {
	now := ptypes.TimestampNow()
	m := &mirrors.Mirror{
		ID:            1,
		Name:          "test",
		HttpURL:       "http://test.com",
		StateSince:    mirrors.Time{},
		LastSync:      mirrors.Time{},
	}
	m.StateSince = m.StateSince.FromTime(now.AsTime())
	rpcM, err := MirrorToRPC(m)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rpcM.ID != 1 {
		t.Fatal("ID mismatch")
	}
}
