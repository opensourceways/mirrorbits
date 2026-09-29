// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"encoding/json"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/network"
	"github.com/rafaeljusto/redigomock"
	. "github.com/opensourceways/mirrorbits/testing"
)

func TestLRUCache_SetIfAbsent(t *testing.T) {
	lru := NewLRUCache(1000)

	tv := &TestValue{value: "a"}
	lru.SetIfAbsent("k1", tv)
	v, ok := lru.Get("k1")
	if !ok || v.(*TestValue).value != "a" {
		t.Fatalf("Expected to find 'a' after SetIfAbsent")
	}

	tv2 := &TestValue{value: "b"}
	lru.SetIfAbsent("k1", tv2)
	v, ok = lru.Get("k1")
	if !ok || v.(*TestValue).value != "a" {
		t.Fatalf("SetIfAbsent should not overwrite existing key")
	}
}

func TestLRUCache_SetCapacity(t *testing.T) {
	lru := NewLRUCache(1000)
	lru.Set("a", &TestValue{value: "1"})
	lru.Set("b", &TestValue{value: "2"})

	lru.SetCapacity(0)
	if _, ok := lru.Get("a"); ok {
		t.Fatalf("Expected 'a' evicted after capacity shrink")
	}
}

func TestLRUCache_StatsJSON(t *testing.T) {
	lru := NewLRUCache(1000)
	s := lru.StatsJSON()
	if s == "" {
		t.Fatalf("Expected non-empty JSON stats")
	}

	var nilLRU *LRUCache
	if nilLRU.StatsJSON() != "{}" {
		t.Fatalf("Expected '{}' for nil LRU")
	}
}

func TestLRUCache_Keys(t *testing.T) {
	lru := NewLRUCache(1000)
	lru.Set("a", &TestValue{value: "1"})
	lru.Set("b", &TestValue{value: "2"})

	keys := lru.Keys()
	if len(keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(keys))
	}
}

func TestLRUCache_Items(t *testing.T) {
	lru := NewLRUCache(1000)
	lru.Set("a", &TestValue{value: "1"})

	items := lru.Items()
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Key != "a" {
		t.Fatalf("Expected key 'a'")
	}
}

func TestLRUCache_Stats(t *testing.T) {
	lru := NewLRUCache(1000)
	lru.Set("a", &TestValue{value: "1"})

	length, size, capacity, _ := lru.Stats()
	if length != 1 {
		t.Fatalf("Expected length 1, got %d", length)
	}
	if capacity != 1000 {
		t.Fatalf("Expected capacity 1000, got %d", capacity)
	}
	if size == 0 {
		t.Fatalf("Expected non-zero size")
	}
}

func TestCache_GetMirrorInvalidationEvent(t *testing.T) {
	_, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	c := NewCache(conn)
	if c == nil {
		t.Fatalf("Expected non-nil cache")
	}
	ch := c.GetMirrorInvalidationEvent()
	if ch == nil {
		t.Fatalf("Expected non-nil channel")
	}
}

func TestCache_GetFileInfoMirror_CacheMiss(t *testing.T) {
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	c := NewCache(conn)

	_, err := c.GetFileInfoMirror(1, "/test/file.tgz")
	if err == nil {
		t.Fatalf("Error expected, mock command not yet registered")
	}

	mock.Command("HMGET", "FILEINFO_1_/test/file.tgz", "size", "modTime", "sha256").Expect([]interface{}{
		[]byte("44000"),
		[]byte(""),
		[]byte(""),
	})

	fi, err := c.GetFileInfoMirror(1, "/test/file.tgz")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if fi.Size != 44000 {
		t.Fatalf("Expected size 44000, got %d", fi.Size)
	}

	fi2, err := c.GetFileInfoMirror(1, "/test/file.tgz")
	if err != nil {
		t.Fatalf("Unexpected error on cache hit: %s", err)
	}
	if fi2.Size != 44000 {
		t.Fatalf("Expected size 44000 from cache, got %d", fi2.Size)
	}
}

func TestPushLog_Success(t *testing.T) {
	mock, conn := PrepareRedisTest()

	cmd := mock.Command("RPUSH", "MIRRORLOGS_1", redigomock.NewAnyData()).Expect(int64(1))

	err := PushLog(conn, NewLogError(1, errTest("boom")))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if mock.Stats(cmd) < 1 {
		t.Fatalf("RPUSH not executed")
	}
}

func TestPushLog_Error(t *testing.T) {
	mock, conn := PrepareRedisTest()

	mock.Command("RPUSH", "MIRRORLOGS_1", redigomock.NewAnyData()).ExpectError(redis.Error("ERR fail"))

	err := PushLog(conn, NewLogError(1, errTest("boom")))
	if err == nil {
		t.Fatalf("Expected error from RPUSH")
	}
}

func TestReadLogs_Success(t *testing.T) {
	mock, conn := PrepareRedisTest()

	logEntry := map[string]interface{}{
		"Type":      float64(LOGTYPE_ADDED),
		"MirrorID":  float64(1),
		"Timestamp": "2020-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(logEntry)

	mock.Command("LRANGE", "MIRRORLOGS_1", redigomock.NewAnyInt(), redigomock.NewAnyInt()).Expect([]interface{}{
		[]byte(data),
	})

	lines, err := ReadLogs(conn, 1, 500)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
}

func TestReadLogs_DefaultMax(t *testing.T) {
	mock, conn := PrepareRedisTest()

	mock.Command("LRANGE", "MIRRORLOGS_2", redigomock.NewAnyInt(), redigomock.NewAnyInt()).Expect([]interface{}{})

	lines, err := ReadLogs(conn, 2, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(lines) != 0 {
		t.Fatalf("Expected 0 lines, got %d", len(lines))
	}
}

func TestReadLogs_UnparseableLine(t *testing.T) {
	mock, conn := PrepareRedisTest()

	mock.Command("LRANGE", "MIRRORLOGS_3", redigomock.NewAnyInt(), redigomock.NewAnyInt()).Expect([]interface{}{
		[]byte("not-json"),
	})

	lines, err := ReadLogs(conn, 3, 500)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(lines) != 0 {
		t.Fatalf("Expected 0 lines for unparseable data, got %d", len(lines))
	}
}

func TestReadLogs_UnknownType(t *testing.T) {
	mock, conn := PrepareRedisTest()

	logEntry := map[string]interface{}{
		"Type":     float64(999),
		"MirrorID": float64(1),
	}
	data, _ := json.Marshal(logEntry)

	mock.Command("LRANGE", "MIRRORLOGS_4", redigomock.NewAnyInt(), redigomock.NewAnyInt()).Expect([]interface{}{
		[]byte(data),
	})

	lines, err := ReadLogs(conn, 4, 500)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(lines) != 0 {
		t.Fatalf("Expected 0 lines for unknown type, got %d", len(lines))
	}
}

func TestNewCache_NilPubsub(t *testing.T) {
	_, conn := PrepareRedisTest()
	c := NewCache(conn)
	if c != nil {
		t.Fatalf("Expected nil cache when Pubsub is nil")
	}
}

func TestCache_fetchMirror_EmptyReply(t *testing.T) {
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	c := NewCache(conn)

	mock.Command("HGETALL", "MIRROR_99").Expect([]interface{}{})

	_, err := c.fetchMirror(99)
	if err != redis.ErrNil {
		t.Fatalf("Expected redis.ErrNil for empty reply, got %v", err)
	}
}

func TestCache_GetMirrors_NoIDs(t *testing.T) {
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	c := NewCache(conn)
	filename := "/test/empty.tgz"

	mock.Command("SMEMBERS", "FILEMIRRORS_"+filename).Expect([]interface{}{})

	mirrors, err := c.GetMirrors(filename, validGeoIPRecord())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(mirrors) != 0 {
		t.Fatalf("Expected 0 mirrors, got %d", len(mirrors))
	}
}

func TestCache_GetMirrors_CachedIDs(t *testing.T) {
	mock, conn := PrepareRedisTest()
	conn.ConnectPubsub()

	c := NewCache(conn)
	filename := "/test/cached.tgz"

	mock.Command("SMEMBERS", "FILEMIRRORS_"+filename).Expect([]interface{}{
		[]byte("1"),
	})
	mock.Command("HGETALL", "MIRROR_1").ExpectMap(map[string]string{
		"ID":        "1",
		"latitude":  "48.856700",
		"longitude": "2.350800",
	})
	mock.Command("HMGET", "FILEINFO_1_"+filename, "size", "modTime", "sha256").Expect([]interface{}{
		[]byte("100"),
		[]byte(""),
		[]byte(""),
	})

	mirrors, err := c.GetMirrors(filename, validGeoIPRecord())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(mirrors) != 1 {
		t.Fatalf("Expected 1 mirror, got %d", len(mirrors))
	}
}

func validGeoIPRecord() network.GeoIPRecord {
	return network.GeoIPRecord{
		CountryCode:   "FR",
		ContinentCode: "EU",
		Latitude:      48.8567,
		Longitude:     2.3508,
	}
}
