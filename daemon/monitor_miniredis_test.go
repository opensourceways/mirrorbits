package daemon

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
)

func setupMiniredisMonitor(t *testing.T) (*miniredis.Miniredis, *database.Redis, *mirrors.Cache, func()) {
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

	cache := mirrors.NewCache(r)
	cleanup := func() {
		r.Close()
		mr.Close()
	}
	return mr, r, cache, cleanup
}

func seedMonitor(t *testing.T, mr *miniredis.Miniredis, id int, name, httpURL string) {
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
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "stateSince", strconv.FormatInt(time.Now().Unix(), 10))
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "lastSync", "0")
}

func TestNewMonitor_Miniredis(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	if m == nil {
		t.Fatal("Expected non-nil monitor")
	}
	if m.redis == nil {
		t.Fatal("Expected non-nil redis")
	}
	if m.cache == nil {
		t.Fatal("Expected non-nil cache")
	}
	if m.mirrors == nil {
		t.Fatal("Expected non-nil mirrors map")
	}
	if m.cluster == nil {
		t.Fatal("Expected non-nil cluster")
	}
	if m.trace == nil {
		t.Fatal("Expected non-nil trace")
	}
	m.Stop()
}

func TestMonitor_mirrorsID_Miniredis(t *testing.T) {
	mr, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	mr.HSet("MIRRORS", "1", "mirror1")
	mr.HSet("MIRRORS", "2", "mirror2")

	m := NewMonitor(r, cache)
	defer m.Stop()

	ids, err := m.mirrorsID()
	if err != nil {
		t.Fatalf("mirrorsID failed: %s", err)
	}
	if len(ids) != 2 {
		t.Fatalf("Expected 2 IDs, got %d", len(ids))
	}
}

func TestMonitor_mirrorsID_Empty(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	defer m.Stop()

	ids, err := m.mirrorsID()
	if err != nil {
		t.Fatalf("mirrorsID failed: %s", err)
	}
	if len(ids) != 0 {
		t.Fatalf("Expected 0 IDs, got %d", len(ids))
	}
}

func TestMonitor_syncMirrorList_Miniredis(t *testing.T) {
	mr, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	seedMonitor(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	m := NewMonitor(r, cache)
	defer m.Stop()

	err := m.syncMirrorList(1)
	if err != nil {
		t.Fatalf("syncMirrorList failed: %s", err)
	}

	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	if m, ok := m.mirrors[1]; !ok || m.Name != "mirror1" {
		t.Fatalf("Expected mirror1 in monitor map")
	}
}

func TestMonitor_syncMirrorList_Multiple(t *testing.T) {
	mr, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	seedMonitor(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	seedMonitor(t, mr, 2, "mirror2", "http://mirror2.example.com/")

	m := NewMonitor(r, cache)
	defer m.Stop()

	err := m.syncMirrorList(1, 2)
	if err != nil {
		t.Fatalf("syncMirrorList failed: %s", err)
	}

	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	if len(m.mirrors) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(m.mirrors))
	}
}

func TestMonitor_syncMirrorList_DeletedMirror(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	defer m.Stop()

	// Add a mirror to the monitor's map
	m.mapLock.Lock()
	m.mirrors[1] = &mirror{
		Mirror: mirrors.Mirror{ID: 1, Name: "mirror1"},
	}
	m.mapLock.Unlock()

	// Don't seed the mirror in Redis - it will be deleted
	err := m.syncMirrorList(1)
	if err != nil {
		t.Fatalf("syncMirrorList failed: %s", err)
	}

	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	if _, ok := m.mirrors[1]; ok {
		t.Fatal("Expected mirror1 to be deleted from map")
	}
}

func TestMonitor_Stop_Double_Miniredis(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	m.Stop()
	// Stopping again should not panic
	m.Stop()
}

func TestMonitor_retry_Miniredis(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	defer m.Stop()

	called := 0
	m.retry(func(i uint) error {
		called++
		if i < 2 {
			return fmt.Errorf("retry %d", i)
		}
		return nil
	}, 10*time.Millisecond)

	if called != 3 {
		t.Fatalf("Expected 3 calls, got %d", called)
	}
}

func TestMonitor_retry_ImmediateSuccess_Miniredis(t *testing.T) {
	_, r, cache, cleanup := setupMiniredisMonitor(t)
	defer cleanup()

	m := NewMonitor(r, cache)
	defer m.Stop()

	called := 0
	m.retry(func(i uint) error {
		called++
		return nil
	}, 10*time.Millisecond)

	if called != 1 {
		t.Fatalf("Expected 1 call, got %d", called)
	}
}

func TestMirror_NeedHealthCheck_Miniredis(t *testing.T) {
	m := &mirror{
		Mirror: mirrors.Mirror{},
	}
	m.lastCheck = time.Now().Add(-10 * time.Minute)
	if !m.NeedHealthCheck(5) {
		t.Fatal("Expected health check needed (10 > 5 minutes)")
	}
	if m.NeedHealthCheck(15) {
		t.Fatal("Expected no health check needed (10 < 15 minutes)")
	}
}

func TestMirror_NeedSync_Miniredis(t *testing.T) {
	m := &mirror{
		Mirror: mirrors.Mirror{},
	}
	m.LastSync.Time = time.Now().Add(-10 * time.Minute)
	if !m.NeedSync(5) {
		t.Fatal("Expected sync needed (10 > 5 minutes)")
	}
	if m.NeedSync(15) {
		t.Fatal("Expected no sync needed (10 < 15 minutes)")
	}
}
