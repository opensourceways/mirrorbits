// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
)

type rpcPoolMock struct {
	Conn *redigomock.Conn
}

func (r *rpcPoolMock) Get() redis.Conn { return r.Conn }
func (r *rpcPoolMock) Close() error     { return nil }

func newMockRedis() (*redigomock.Conn, *database.Redis) {
	mock := redigomock.NewConn()
	pool := &rpcPoolMock{Conn: mock}
	return mock, database.NewRedisCustomPool(pool)
}

func TestMain(m *testing.M) {
	SetConfiguration(&Configuration{RedisAddress: ""})
	os.Exit(m.Run())
}

func TestCLI_Ping(t *testing.T) {
	c := &CLI{}
	_, err := c.Ping(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLI_GetVersion(t *testing.T) {
	c := &CLI{}
	reply, err := c.GetVersion(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if reply == nil {
		t.Fatalf("Expected non-nil reply")
	}
}

func TestCLI_Upgrade_Ready(t *testing.T) {
	sig := make(chan os.Signal, 1)
	c := &CLI{sig: sig}

	_, err := c.Upgrade(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	select {
	case <-sig:
	default:
		t.Fatalf("Expected signal to be sent")
	}
}

func TestCLI_Upgrade_NotReady(t *testing.T) {
	c := &CLI{sig: nil}

	_, err := c.Upgrade(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatalf("Expected error when sig is nil")
	}
}

func TestCLI_Reload_Ready(t *testing.T) {
	sig := make(chan os.Signal, 1)
	c := &CLI{sig: sig}

	_, err := c.Reload(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	select {
	case <-sig:
	default:
		t.Fatalf("Expected signal to be sent")
	}
}

func TestCLI_Reload_NotReady(t *testing.T) {
	c := &CLI{sig: nil}

	_, err := c.Reload(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatalf("Expected error when sig is nil")
	}
}

func TestCLI_MatchMirror_NoRedis(t *testing.T) {
	c := &CLI{redis: nil}
	_, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "test"})
	if err == nil {
		t.Fatalf("Expected error when redis is nil")
	}
}

func TestCLI_MatchMirror_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}
	_, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "test"})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_ChangeStatus_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 0, Enabled: true})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_ChangeStatus_Enable(t *testing.T) {
	mock, r := newMockRedis()
	c := &CLI{redis: r}

	mock.Command("HMSET", "MIRROR_1", "enabled", true).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")
	mock.Command("RPUSH", "MIRRORLOGS_1", redigomock.NewAnyData()).Expect(int64(1))

	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 1, Enabled: true})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLI_ChangeStatus_Disable(t *testing.T) {
	mock, r := newMockRedis()
	c := &CLI{redis: r}

	mock.Command("HMSET", "MIRROR_1", "enabled", false).Expect("OK")
	mock.Command("PUBLISH", "_mirrorbits_mirror_update", "1").Expect("OK")
	mock.Command("RPUSH", "MIRRORLOGS_1", redigomock.NewAnyData()).Expect(int64(1))

	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 1, Enabled: false})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLI_List_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}
	_, err := c.List(context.Background(), &empty.Empty{})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_MirrorInfo_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_MirrorInfo_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}
	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 1})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_RemoveMirror_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.RemoveMirror(context.Background(), &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_RemoveMirror_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}
	_, err := c.RemoveMirror(context.Background(), &MirrorIDRequest{ID: 1})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_ScanMirror_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.ScanMirror(context.Background(), &ScanMirrorRequest{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_ScanMirror_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}
	_, err := c.ScanMirror(context.Background(), &ScanMirrorRequest{ID: 1})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_StatsFile_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}

	now := timestampNow()
	_, err := c.StatsFile(context.Background(), &StatsFileRequest{
		Pattern:   ".*",
		DateStart: now,
		DateEnd:   now,
	})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_StatsMirror_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.StatsMirror(context.Background(), &StatsMirrorRequest{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_StatsMirror_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}

	now := timestampNow()
	_, err := c.StatsMirror(context.Background(), &StatsMirrorRequest{
		ID:        1,
		DateStart: now,
		DateEnd:   now,
	})
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCLI_GetMirrorLogs_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_GetMirrorLogs_Success(t *testing.T) {
	mock, r := newMockRedis()
	c := &CLI{redis: r}

	mock.Command("LRANGE", "MIRRORLOGS_1", redigomock.NewAnyInt(), redigomock.NewAnyInt()).Expect([]interface{}{})

	_, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{ID: 1, MaxResults: 500})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCLI_SetDatabase(t *testing.T) {
	c := &CLI{}
	_, r := newMockRedis()
	c.SetDatabase(r)
	if c.redis != r {
		t.Fatalf("Expected redis set")
	}
}

func TestCLI_SetCache(t *testing.T) {
	c := &CLI{}
	c.SetCache(nil)
}

func TestCLI_SetSignals(t *testing.T) {
	c := &CLI{}
	sig := make(chan os.Signal, 1)
	c.SetSignals(sig)
	if c.sig != sig {
		t.Fatalf("Expected sig set")
	}
}

func TestCLI_UpdateMirror_InvalidID(t *testing.T) {
	c := &CLI{}
	_, err := c.UpdateMirror(context.Background(), &Mirror{ID: 0})
	if err == nil {
		t.Fatalf("Expected error for invalid ID")
	}
}

func TestCLI_UpdateMirror_Unreachable(t *testing.T) {
	_, r := newMockRedis()
	c := &CLI{redis: r}

	now := timestampNow()
	m := &Mirror{ID: 1, Name: "test", StateSince: now, LastSync: now, LastSuccessfulSync: now, LastModTime: now}

	_, err := c.UpdateMirror(context.Background(), m)
	if err == nil {
		t.Fatalf("Expected error for unreachable redis")
	}
}

func TestCreateDiff(t *testing.T) {
	m1 := &mirrors.Mirror{Name: "m1", HttpURL: "http://old.com"}
	m2 := &mirrors.Mirror{Name: "m1", HttpURL: "http://new.com"}

	diff := createDiff(m1, m2)
	if diff == "" {
		t.Fatalf("Expected non-empty diff")
	}
}

func TestCreateDiff_Identical(t *testing.T) {
	m1 := &mirrors.Mirror{Name: "m1", HttpURL: "http://same.com"}
	m2 := &mirrors.Mirror{Name: "m1", HttpURL: "http://same.com"}

	diff := createDiff(m1, m2)
	if diff != "" {
		t.Fatalf("Expected empty diff for identical mirrors, got %s", diff)
	}
}

func TestErrNameAlreadyTaken(t *testing.T) {
	if ErrNameAlreadyTaken == nil {
		t.Fatalf("Expected non-nil error")
	}
	if ErrNameAlreadyTaken.Error() != "name already taken" {
		t.Fatalf("Unexpected message: %s", ErrNameAlreadyTaken.Error())
	}
}
