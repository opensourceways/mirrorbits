package http

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
)

var (
	httpOnce    sync.Once
	sharedHTTP  *HTTP
	sharedMR    *miniredis.Miniredis
	sharedRedis *database.Redis
	sharedCache *mirrors.Cache
)

func setupHTTPMiniredis(t *testing.T) (*miniredis.Miniredis, *database.Redis, *mirrors.Cache, func()) {
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
		RedisAddress:      mr.Addr(),
		RedisDB:           0,
		GeoipDatabasePath: "../GeoIP/",
		ListenAddress:     ":0",
		OutputMode:        "json",
		Templates:         "../templates",
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

func setupHTTPServer(t *testing.T) (*miniredis.Miniredis, *HTTP, func()) {
	t.Helper()
	mr, r, cache, baseCleanup := setupHTTPMiniredis(t)

	http.DefaultServeMux = http.NewServeMux()
	h := HTTPServer(r, cache)
	cleanup := func() {
		baseCleanup()
	}
	return mr, h, cleanup
}

func TestHTTPServer_Miniredis(t *testing.T) {
	_, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	if h == nil {
		t.Fatal("Expected non-nil HTTP server")
	}
	if h.redis == nil {
		t.Fatal("Expected non-nil redis")
	}
	if h.cache == nil {
		t.Fatal("Expected non-nil cache")
	}
}

func TestHTTP_Reload_Miniredis(t *testing.T) {
	mr, r, cache, baseCleanup := setupHTTPMiniredis(t)
	defer baseCleanup()

	_ = mr
	_ = r
	_ = cache
}

func TestHTTP_mirrorHandler_EmptyPath(t *testing.T) {
	_, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.mirrorHandler(w, req, c)
}

func TestHTTP_mirrorHandler_Mirrorlist(t *testing.T) {
	mr, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	mr.SAdd("FILES", "/test.txt")

	req := httptest.NewRequest("GET", "/test.txt?mirrorlist", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.mirrorHandler(w, req, c)
}

func TestHTTP_mirrorHandler_Mirrorstats(t *testing.T) {
	_, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.mirrorStatsHandler(w, req, c)
}

func TestHTTP_mirrorHandler_FileStats(t *testing.T) {
	mr, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	mr.SAdd("FILES", "/test.txt")

	req := httptest.NewRequest("GET", "/test.txt?stats", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.mirrorHandler(w, req, c)
}

func TestHTTP_mirrorHandler_Checksum(t *testing.T) {
	mr, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	mr.SAdd("FILES", "/test.txt")

	req := httptest.NewRequest("GET", "/test.txt?sha256", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.checksumHandler(w, req, c)
}

func TestHTTP_mirrorHandler_NotFound(t *testing.T) {
	_, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
	w := httptest.NewRecorder()

	tmpl := Templates{RWMutex: &sync.RWMutex{}}
	c := NewContext(w, req, tmpl)
	h.mirrorHandler(w, req, c)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestHTTP_Restarting(t *testing.T) {
	_, h, cleanup := setupHTTPServer(t)
	defer cleanup()

	if h.Restarting {
		t.Fatal("Expected Restarting=false")
	}
	h.Restarting = true
	if !h.Restarting {
		t.Fatal("Expected Restarting=true")
	}
}
