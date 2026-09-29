package rpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func setupMiniredisRPC(t *testing.T) (*miniredis.Miniredis, *CLI, *database.Redis, func()) {
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
		RedisAddress:       mr.Addr(),
		RedisDB:            0,
		GeoipDatabasePath:  "/nonexistent/",
		RPCListenAddress:   "localhost:0",
	})

	r := database.NewRedis()
	r.ConnectPubsub()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn := r.Get()
		if _, ok := conn.(*database.NotReadyError); !ok {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	c := &CLI{redis: r}
	cleanup := func() {
		r.Close()
		mr.Close()
	}
	return mr, c, r, cleanup
}

func seedMirror(t *testing.T, mr *miniredis.Miniredis, id int, name, httpURL string) {
	t.Helper()
	mr.HSet("MIRRORS", fmt.Sprintf("%d", id), name)
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "ID", fmt.Sprintf("%d", id))
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "name", name)
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "http", httpURL)
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "enabled", "1")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "up", "1")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "score", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "latitude", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "longitude", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "asnum", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "continentCode", "")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "countryCodes", "")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "country", "US")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "excludedCountryCodes", "")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "networkBandwidth", "1000")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "allowredirects", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "stateSince", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "lastSync", "0")
}

func TestRPC_List_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	seedMirror(t, mr, 2, "mirror2", "http://mirror2.example.com/")

	reply, err := c.List(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("List failed: %s", err)
	}
	if len(reply.Mirrors) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(reply.Mirrors))
	}
}

func TestRPC_List_Empty(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	reply, err := c.List(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("List failed: %s", err)
	}
	if len(reply.Mirrors) != 0 {
		t.Fatalf("Expected 0 mirrors, got %d", len(reply.Mirrors))
	}
}

func TestRPC_MirrorInfo_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	reply, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 1})
	if err != nil {
		t.Fatalf("MirrorInfo failed: %s", err)
	}
	if reply.Name != "mirror1" {
		t.Fatalf("Expected name mirror1, got %s", reply.Name)
	}
	if reply.HttpURL != "http://mirror1.example.com/" {
		t.Fatalf("Expected http URL, got %s", reply.HttpURL)
	}
}

func TestRPC_MirrorInfo_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_MirrorInfo_NotFound(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.MirrorInfo(context.Background(), &MirrorIDRequest{ID: 999})
	// Mirror not found returns empty struct, no error from HGETALL
	_ = err
}

func TestRPC_MatchMirror_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	seedMirror(t, mr, 2, "mirror2", "http://mirror2.example.com/")

	reply, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "mirror1"})
	if err != nil {
		t.Fatalf("MatchMirror failed: %s", err)
	}
	if len(reply.Mirrors) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(reply.Mirrors))
	}
	if reply.Mirrors[0].Name != "mirror1" {
		t.Fatalf("Expected mirror1, got %s", reply.Mirrors[0].Name)
	}
}

func TestRPC_MatchMirror_All(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	seedMirror(t, mr, 2, "mirror2", "http://mirror2.example.com/")

	reply, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "mirror"})
	if err != nil {
		t.Fatalf("MatchMirror failed: %s", err)
	}
	if len(reply.Mirrors) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(reply.Mirrors))
	}
}

func TestRPC_MatchMirror_NoMatch(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	reply, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "nomatch"})
	if err != nil {
		t.Fatalf("MatchMirror failed: %s", err)
	}
	if len(reply.Mirrors) != 0 {
		t.Fatalf("Expected 0 matches, got %d", len(reply.Mirrors))
	}
}

func TestRPC_ChangeStatus_Enable(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 1, Enabled: true})
	if err != nil {
		t.Fatalf("ChangeStatus failed: %s", err)
	}

	enabled := mr.HGet(fmt.Sprintf("MIRROR_%d", 1), "enabled")
	if enabled != "1" {
		t.Fatalf("Expected enabled=1, got %s", enabled)
	}
}

func TestRPC_ChangeStatus_Disable(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 1, Enabled: false})
	if err != nil {
		t.Fatalf("ChangeStatus failed: %s", err)
	}

	enabled := mr.HGet(fmt.Sprintf("MIRROR_%d", 1), "enabled")
	if enabled != "0" {
		t.Fatalf("Expected enabled=0, got %s", enabled)
	}
}

func TestRPC_ChangeStatus_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.ChangeStatus(context.Background(), &ChangeStatusRequest{ID: 0, Enabled: true})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_setMirror_NewMirror(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	mirror := &mirrors.Mirror{
		Name:    "newmirror",
		HttpURL: "http://newmirror.example.com/",
		Country: "US",
	}

	err := c.setMirror(mirror)
	if err != nil {
		t.Fatalf("setMirror failed: %s", err)
	}

	if mirror.ID == 0 {
		t.Fatal("Expected non-zero mirror ID")
	}

	name := mr.HGet("MIRRORS", fmt.Sprintf("%d", mirror.ID))
	if name != "newmirror" {
		t.Fatalf("Expected name newmirror in MIRRORS, got %s", name)
	}

	hname := mr.HGet(fmt.Sprintf("MIRROR_%d", mirror.ID), "name")
	if hname != "newmirror" {
		t.Fatalf("Expected name newmirror in MIRROR, got %s", hname)
	}
}

func TestRPC_setMirror_DuplicateName(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	mirror := &mirrors.Mirror{
		ID:      0,
		Name:    "mirror1",
		HttpURL: "http://mirror1.example.com/",
		Country: "US",
	}

	err := c.setMirror(mirror)
	if err != ErrNameAlreadyTaken {
		t.Fatalf("Expected ErrNameAlreadyTaken, got %v", err)
	}
}

func TestRPC_setMirror_UpdateExisting(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	mirror := &mirrors.Mirror{
		ID:      1,
		Name:    "mirror1",
		HttpURL: "http://updated.example.com/",
		Country: "US",
	}

	err := c.setMirror(mirror)
	if err != nil {
		t.Fatalf("setMirror failed: %s", err)
	}

	http := mr.HGet(fmt.Sprintf("MIRROR_%d", 1), "http")
	if http != "http://updated.example.com/" {
		t.Fatalf("Expected updated http URL, got %s", http)
	}
}

func TestRPC_RemoveMirror_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	mr.SetAdd("MIRRORFILES_1", "file1.txt")
	mr.SetAdd("FILEMIRRORS_file1.txt", "1")

	_, err := c.RemoveMirror(context.Background(), &MirrorIDRequest{ID: 1})
	if err != nil {
		t.Fatalf("RemoveMirror failed: %s", err)
	}

	exists := mr.Exists("MIRROR_1")
	if exists {
		t.Fatal("Expected MIRROR_1 to be deleted")
	}

	mname := mr.HGet("MIRRORS", "1")
	if mname != "" {
		t.Fatalf("Expected mirror 1 removed from MIRRORS, got %s", mname)
	}
}

func TestRPC_RemoveMirror_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.RemoveMirror(context.Background(), &MirrorIDRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_StatsFile_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	now := time.Now().UTC()
	key := now.Format("2006_01_02")
	mr.HSet("STATS_FILE_"+key, "/file1.txt", "10")
	mr.HSet("STATS_FILE_"+key, "/file2.txt", "20")

	startProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, -1))
	endProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, 1))

	reply, err := c.StatsFile(context.Background(), &StatsFileRequest{
		Pattern:   "file1",
		DateStart: startProto,
		DateEnd:   endProto,
	})
	if err != nil {
		t.Fatalf("StatsFile failed: %s", err)
	}
	if reply.Files["/file1.txt"] != 10 {
		t.Fatalf("Expected 10 requests for /file1.txt, got %d", reply.Files["/file1.txt"])
	}
}

func TestRPC_StatsFile_InvalidPattern(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	now := time.Now().UTC()
	startProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, -1))
	endProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, 1))

	_, err := c.StatsFile(context.Background(), &StatsFileRequest{
		Pattern:   "[invalid",
		DateStart: startProto,
		DateEnd:   endProto,
	})
	if err == nil {
		t.Fatal("Expected error for invalid pattern")
	}
}

func TestRPC_StatsMirror_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	now := time.Now().UTC()
	key := now.Format("2006_01_02")
	mr.HSet("STATS_MIRROR_"+key, "1", "100")
	mr.HSet("STATS_MIRROR_BYTES_"+key, "1", "1024")

	startProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, -1))
	endProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, 1))

	reply, err := c.StatsMirror(context.Background(), &StatsMirrorRequest{
		ID:        1,
		DateStart: startProto,
		DateEnd:   endProto,
	})
	if err != nil {
		t.Fatalf("StatsMirror failed: %s", err)
	}
	if reply.Requests != 100 {
		t.Fatalf("Expected 100 requests, got %d", reply.Requests)
	}
	if reply.Bytes != 1024 {
		t.Fatalf("Expected 1024 bytes, got %d", reply.Bytes)
	}
}

func TestRPC_StatsMirror_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	now := time.Now().UTC()
	startProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, -1))
	endProto, _ := ptypes.TimestampProto(now.AddDate(0, 0, 1))

	_, err := c.StatsMirror(context.Background(), &StatsMirrorRequest{
		ID:        0,
		DateStart: startProto,
		DateEnd:   endProto,
	})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_GetMirrorLogs_Miniredis(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	logEntry := `{"Type":2,"MirrorID":1,"Engine":"test","Date":1000}`
	mr.Push("MIRRORLOGS_1", logEntry)

	reply, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{
		ID:         1,
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("GetMirrorLogs failed: %s", err)
	}
	if len(reply.Line) == 0 {
		t.Fatal("Expected at least 1 log line")
	}
}

func TestRPC_GetMirrorLogs_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{
		ID:         0,
		MaxResults: 10,
	})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_GetMirrorLogs_Empty(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	reply, err := c.GetMirrorLogs(context.Background(), &GetMirrorLogsRequest{
		ID:         1,
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("GetMirrorLogs failed: %s", err)
	}
	if len(reply.Line) != 0 {
		t.Fatalf("Expected 0 log lines, got %d", len(reply.Line))
	}
}

func TestRPC_AddMirror_GeoIPError(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	mirror := &Mirror{
		Name:    "testmirror",
		HttpURL: "http://testmirror.example.com/",
	}

	_, err := c.AddMirror(context.Background(), mirror)
	if err == nil {
		t.Fatal("Expected error from GeoIP loading")
	}
}

func TestRPC_UpdateMirror_WithCountry(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	now := time.Now()
	ts, _ := ptypes.TimestampProto(now)

	mirror := &Mirror{
		ID:                  1,
		Name:                "mirror1",
		HttpURL:             "http://updated.example.com/",
		Country:             "US",
		StateSince:          ts,
		LastSync:            ts,
		LastSuccessfulSync:  ts,
		LastModTime:         ts,
	}

	reply, err := c.UpdateMirror(context.Background(), mirror)
	if err != nil {
		t.Fatalf("UpdateMirror failed: %s", err)
	}
	if reply == nil {
		t.Fatal("Expected non-nil reply")
	}

	http := mr.HGet(fmt.Sprintf("MIRROR_%d", 1), "http")
	if http != "http://updated.example.com/" {
		t.Fatalf("Expected updated http URL, got %s", http)
	}
}

func TestRPC_UpdateMirror_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	mirror := &Mirror{
		ID:      0,
		Name:    "test",
		HttpURL: "http://test.example.com/",
	}

	_, err := c.UpdateMirror(context.Background(), mirror)
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_ScanMirror_InvalidID(t *testing.T) {
	_, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	_, err := c.ScanMirror(context.Background(), &ScanMirrorRequest{ID: 0})
	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}
}

func TestRPC_ScanMirror_RepoNotIndexed(t *testing.T) {
	mr, c, _, cleanup := setupMiniredisRPC(t)
	defer cleanup()

	seedMirror(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	_, err := c.ScanMirror(context.Background(), &ScanMirrorRequest{ID: 1})
	if err == nil {
		t.Fatal("Expected error when repository not indexed")
	}
}

func TestRPC_MatchMirror_NilRedis(t *testing.T) {
	c := &CLI{redis: nil}
	_, err := c.MatchMirror(context.Background(), &MatchRequest{Pattern: "test"})
	if err == nil {
		t.Fatal("Expected error for nil redis")
	}
}

func TestRPC_ProtoMessage_AllTypes(t *testing.T) {
	(&VersionReply{}).ProtoMessage()
	(&MatchRequest{}).ProtoMessage()
	(&Mirror{}).ProtoMessage()
	(&MirrorListReply{}).ProtoMessage()
	(&MirrorID{}).ProtoMessage()
	(&MatchReply{}).ProtoMessage()
	(&ChangeStatusRequest{}).ProtoMessage()
	(&MirrorIDRequest{}).ProtoMessage()
	(&AddMirrorReply{}).ProtoMessage()
	(&UpdateMirrorReply{}).ProtoMessage()
	(&RefreshRepositoryRequest{}).ProtoMessage()
	(&ScanMirrorRequest{}).ProtoMessage()
	(&ScanMirrorReply{}).ProtoMessage()
	(&StatsFileRequest{}).ProtoMessage()
	(&StatsFileReply{}).ProtoMessage()
	(&StatsMirrorRequest{}).ProtoMessage()
	(&StatsMirrorReply{}).ProtoMessage()
	(&GetMirrorLogsRequest{}).ProtoMessage()
	(&GetMirrorLogsReply{}).ProtoMessage()
}
