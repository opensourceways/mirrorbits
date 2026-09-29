package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/golang/protobuf/ptypes/empty"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/rpc"
	"google.golang.org/grpc"
)

func setupCLIIntegration(t *testing.T) (*miniredis.Miniredis, *rpc.CLI, func()) {
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
		GeoipDatabasePath: "/nonexistent/",
		RPCListenAddress:  "localhost:0",
		RPCPassword:       "",
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

	grpcServer := &rpc.CLI{}
	grpcServer.SetDatabase(r)

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %s", err)
	}

	server := grpc.NewServer()
	rpc.RegisterCLIServer(server, grpcServer)
	go server.Serve(ln)

	core.RPCHost = "localhost"
	core.RPCPort = uint(ln.Addr().(*net.TCPAddr).Port)
	core.RPCPassword = ""

	cleanup := func() {
		server.Stop()
		r.Close()
		mr.Close()
	}
	return mr, grpcServer, cleanup
}

func seedCLI(t *testing.T, mr *miniredis.Miniredis, id int, name, httpURL string) {
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
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "continentCode", "AS")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "countryCodes", "CN")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "country", "China")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "excludedCountryCodes", "")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "networkBandwidth", "1000")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "allowredirects", "0")
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "stateSince", strconv.FormatInt(time.Now().Unix(), 10))
	mr.HSet(fmt.Sprintf("MIRROR_%d", id), "lastSync", "0")
}

func TestCLIIntegration_List(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	seedCLI(t, mr, 2, "mirror2", "http://mirror2.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := c.CmdList()

	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatalf("CmdList failed: %s", err)
	}
}

func TestCLIIntegration_Version(t *testing.T) {
	_, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := c.CmdVersion()

	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatalf("CmdVersion failed: %s", err)
	}
}

func TestCLIIntegration_Reload(t *testing.T) {
	_, srv, cleanup := setupCLIIntegration(t)
	defer cleanup()

	sig := make(chan os.Signal, 1)
	srv.SetSignals(sig)

	c := newTestCli()
	c.rpcconn = nil

	err := c.CmdReload()
	_ = err
}

func TestCLIIntegration_Upgrade(t *testing.T) {
	_, srv, cleanup := setupCLIIntegration(t)
	defer cleanup()

	sig := make(chan os.Signal, 1)
	srv.SetSignals(sig)

	c := newTestCli()
	c.rpcconn = nil

	err := c.CmdUpgrade()
	_ = err
}

func TestCLIIntegration_Show_InvalidArgs(t *testing.T) {
	_, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	c := newTestCli()
	c.rpcconn = nil

	err := c.CmdShow()
	if err != nil {
		t.Fatalf("CmdShow with no args should return nil, got %v", err)
	}
}

func TestCLIIntegration_Enable_Disabled(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdEnable("mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Disable(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdDisable("mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Remove(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdRemove("-f", "mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Export(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdExport()

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Stats(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	now := time.Now().UTC()
	key := now.Format("2006_01_02")
	mr.HSet("STATS_MIRROR_"+key, "1", "100")
	mr.HSet("STATS_MIRROR_BYTES_"+key, "1", "1024")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdStats("mirror", "mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Logs(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	logEntry := `{"Type":2,"MirrorID":1,"Engine":"test","Date":1000}`
	mr.Push("MIRRORLOGS_1", logEntry)

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdLogs("mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Scan_NoArgs(t *testing.T) {
	_, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	c := newTestCli()
	c.rpcconn = nil

	err := c.CmdScan()
	if err != nil {
		t.Fatalf("CmdScan with no args should return nil, got %v", err)
	}
}

func TestGetRPC_Integration(t *testing.T) {
	_, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	c := newTestCli()
	c.rpcconn = nil

	client := c.GetRPC()
	if client == nil {
		t.Fatal("Expected non-nil RPC client")
	}

	_, err := client.Ping(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Ping failed: %s", err)
	}

	if c.rpcconn == nil {
		t.Fatal("Expected rpcconn to be set after GetRPC")
	}
	c.rpcconn.Close()
	c.rpcconn = nil
}

func TestCLIIntegration_Show(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdShow("mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_StatsFile(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")

	now := time.Now().UTC()
	key := now.Format("2006_01_02")
	mr.HSet("STATS_FILE_"+key, "/file1.txt", "10")
	mr.HSet("STATS_FILE_"+key, "/file2.txt", "20")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdStats("file", "file1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCLIIntegration_Scan_WithMirror(t *testing.T) {
	mr, _, cleanup := setupCLIIntegration(t)
	defer cleanup()

	seedCLI(t, mr, 1, "mirror1", "http://mirror1.example.com/")
	mr.Set("FILES", "file1.txt")

	c := newTestCli()
	c.rpcconn = nil

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = c.CmdScan("-all", "mirror1")

	w.Close()
	os.Stdout = oldStdout
	_ = r
}

func TestCompareAndUpdate_Changed(t *testing.T) {
	original := &mirrors.Mirror{
		HttpURL:  "http://old.example.com/",
		Score:    10,
		Country:  "US",
		Enabled:  true,
	}
	updated := &mirrors.Mirror{
		HttpURL:  "http://new.example.com/",
		Score:    20,
		Country:  "US",
		Enabled:  true,
	}

	changed := CompareAndUpdate(original, updated)
	if !changed {
		t.Fatal("Expected changed=true")
	}
	if original.HttpURL != "http://new.example.com/" {
		t.Fatalf("Expected updated HttpURL, got %s", original.HttpURL)
	}
	if original.Score != 20 {
		t.Fatalf("Expected updated score 20, got %d", original.Score)
	}
}
