// Copyright (c) 2014-2019 Ludovic Fauzet
// Licensed under the MIT license

package cli

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/rpc"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type mockCLIClient struct {
	listFn          func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error)
	mirrorInfoFn    func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*rpc.Mirror, error)
	addMirrorFn     func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error)
	updateMirrorFn  func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.UpdateMirrorReply, error)
	removeMirrorFn  func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	refreshFn       func(ctx context.Context, in *rpc.RefreshRepositoryRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	scanMirrorFn    func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error)
	statsFileFn     func(ctx context.Context, in *rpc.StatsFileRequest, opts ...grpc.CallOption) (*rpc.StatsFileReply, error)
	statsMirrorFn   func(ctx context.Context, in *rpc.StatsMirrorRequest, opts ...grpc.CallOption) (*rpc.StatsMirrorReply, error)
	pingFn          func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	getMirrorLogsFn func(ctx context.Context, in *rpc.GetMirrorLogsRequest, opts ...grpc.CallOption) (*rpc.GetMirrorLogsReply, error)
	matchMirrorFn   func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error)
	changeStatusFn  func(ctx context.Context, in *rpc.ChangeStatusRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	getVersionFn    func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.VersionReply, error)
	upgradeFn       func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	reloadFn        func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
}

func (m *mockCLIClient) GetVersion(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.VersionReply, error) {
	if m.getVersionFn != nil {
		return m.getVersionFn(ctx, in, opts...)
	}
	return &rpc.VersionReply{Version: "1.0.0"}, nil
}

func (m *mockCLIClient) Upgrade(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.upgradeFn != nil {
		return m.upgradeFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) Reload(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.reloadFn != nil {
		return m.reloadFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) ChangeStatus(ctx context.Context, in *rpc.ChangeStatusRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.changeStatusFn != nil {
		return m.changeStatusFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) List(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
	if m.listFn != nil {
		return m.listFn(ctx, in, opts...)
	}
	return &rpc.MirrorListReply{}, nil
}

func (m *mockCLIClient) MirrorInfo(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*rpc.Mirror, error) {
	if m.mirrorInfoFn != nil {
		return m.mirrorInfoFn(ctx, in, opts...)
	}
	return &rpc.Mirror{Name: "test"}, nil
}

func (m *mockCLIClient) AddMirror(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error) {
	if m.addMirrorFn != nil {
		return m.addMirrorFn(ctx, in, opts...)
	}
	return &rpc.AddMirrorReply{}, nil
}

func (m *mockCLIClient) UpdateMirror(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.UpdateMirrorReply, error) {
	if m.updateMirrorFn != nil {
		return m.updateMirrorFn(ctx, in, opts...)
	}
	return &rpc.UpdateMirrorReply{}, nil
}

func (m *mockCLIClient) RemoveMirror(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.removeMirrorFn != nil {
		return m.removeMirrorFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) RefreshRepository(ctx context.Context, in *rpc.RefreshRepositoryRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) ScanMirror(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
	if m.scanMirrorFn != nil {
		return m.scanMirrorFn(ctx, in, opts...)
	}
	return &rpc.ScanMirrorReply{FilesIndexed: 10, KnownIndexed: 5, Removed: 2}, nil
}

func (m *mockCLIClient) StatsFile(ctx context.Context, in *rpc.StatsFileRequest, opts ...grpc.CallOption) (*rpc.StatsFileReply, error) {
	if m.statsFileFn != nil {
		return m.statsFileFn(ctx, in, opts...)
	}
	return &rpc.StatsFileReply{Files: map[string]int64{"file.iso": 100}}, nil
}

func (m *mockCLIClient) StatsMirror(ctx context.Context, in *rpc.StatsMirrorRequest, opts ...grpc.CallOption) (*rpc.StatsMirrorReply, error) {
	if m.statsMirrorFn != nil {
		return m.statsMirrorFn(ctx, in, opts...)
	}
	return &rpc.StatsMirrorReply{Requests: 50, Bytes: 1024}, nil
}

func (m *mockCLIClient) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	if m.pingFn != nil {
		return m.pingFn(ctx, in, opts...)
	}
	return &empty.Empty{}, nil
}

func (m *mockCLIClient) GetMirrorLogs(ctx context.Context, in *rpc.GetMirrorLogsRequest, opts ...grpc.CallOption) (*rpc.GetMirrorLogsReply, error) {
	if m.getMirrorLogsFn != nil {
		return m.getMirrorLogsFn(ctx, in, opts...)
	}
	return &rpc.GetMirrorLogsReply{Line: []string{"log1", "log2"}}, nil
}

func (m *mockCLIClient) MatchMirror(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
	if m.matchMirrorFn != nil {
		return m.matchMirrorFn(ctx, in, opts...)
	}
	return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
}

func newMockCli(mock *mockCLIClient) *cli {
	return &cli{
		creds:      &loginCreds{Password: "test"},
		mockClient: mock,
	}
}

func makeTimestamp(sec int64) *timestamp.Timestamp {
	ts, _ := ptypes.TimestampProto(time.Unix(sec, 0))
	return ts
}

func TestCmdListWithMirrors(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "mirror1", HttpURL: "http://m1.com/", Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
					{Name: "mirror2", HttpURL: "http://m2.com/", Enabled: false, Up: false, StateSince: makeTimestamp(2000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-http", "-state")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListWithLocation(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", CountryCodes: "US CA", ContinentCode: "NA", StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-location")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListWithScore(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", Score: 100, Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-score")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListDisabledOnly(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", Enabled: false, Up: true, StateSince: makeTimestamp(1000)},
					{Name: "m2", Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-disabled")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListEnabledOnly(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", Enabled: false, Up: true, StateSince: makeTimestamp(1000)},
					{Name: "m2", Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-enabled")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListDownOnly(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", Enabled: true, Up: false, StateSince: makeTimestamp(1000)},
					{Name: "m2", Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-down")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListTooManyArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdList("extra-arg")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdAddSuccess(t *testing.T) {
	mock := &mockCLIClient{
		addMirrorFn: func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error) {
			if in.Name != "testmirror" {
				t.Fatalf("Expected name testmirror, got %s", in.Name)
			}
			return &rpc.AddMirrorReply{Country: "US", Latitude: 40.0, Longitude: -74.0}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdAdd("-http", "http://example.com", "testmirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdAddWithHttps(t *testing.T) {
	mock := &mockCLIClient{
		addMirrorFn: func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error) {
			return &rpc.AddMirrorReply{}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdAdd("-http", "https://example.com", "testmirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdAddWithWarnings(t *testing.T) {
	mock := &mockCLIClient{
		addMirrorFn: func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error) {
			return &rpc.AddMirrorReply{Warnings: []string{"warning1", "warning2"}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdAdd("-http", "http://example.com", "testmirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdAddNoIdentifier(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdAdd("-http", "http://example.com")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdRemoveForce(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		removeMirrorFn: func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
			if in.ID != 1 {
				t.Fatalf("Expected ID 1, got %d", in.ID)
			}
			return &empty.Empty{}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdRemove("-f", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdRemoveWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdRemove("-f")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdScanSingleMirror(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			return &rpc.ScanMirrorReply{FilesIndexed: 10, KnownIndexed: 5, Removed: 2, Enabled: true, TZOffsetMs: 1000}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-enable", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanAllMirrors(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{
				{ID: 1, Name: "m1"},
				{ID: 2, Name: "m2"},
			}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			return &rpc.ScanMirrorReply{FilesIndexed: 5}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-all")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanRsyncMethod(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			if in.Protocol != rpc.ScanMirrorRequest_RSYNC {
				t.Fatalf("Expected RSYNC method, got %v", in.Protocol)
			}
			return &rpc.ScanMirrorReply{FilesIndexed: 3}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-rsync", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanFtpMethod(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			if in.Protocol != rpc.ScanMirrorRequest_FTP {
				t.Fatalf("Expected FTP method, got %v", in.Protocol)
			}
			return &rpc.ScanMirrorReply{FilesIndexed: 3}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-ftp", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdScan()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdScanAllWithMirror(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdScan("-all", "test-mirror")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdRefresh(t *testing.T) {
	mock := &mockCLIClient{
		refreshFn: func(ctx context.Context, in *rpc.RefreshRepositoryRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
			if !in.Rehash {
				t.Fatal("Expected rehash false")
			}
			return &empty.Empty{}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdRefresh("-rehash")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdRefreshWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdRefresh("extra")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdEditWithMirrorFile(t *testing.T) {
	tmpDir := t.TempDir()
	mirrorFile := filepath.Join(tmpDir, "mirror.yaml")

	mirror := &mirrors.Mirror{
		Name:     "test-mirror",
		HttpURL:  "http://updated.com",
		Enabled:  true,
		Score:    10,
	}
	data, _ := yamlMarshalMirror(mirror)
	ioutil.WriteFile(mirrorFile, data, 0644)

	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		mirrorInfoFn: func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*rpc.Mirror, error) {
			return &rpc.Mirror{Name: "test-mirror", HttpURL: "http://old.com", Enabled: true, Score: 5,
				StateSince: makeTimestamp(1000), LastSync: makeTimestamp(1000), LastSuccessfulSync: makeTimestamp(1000), LastModTime: makeTimestamp(1000)}, nil
		},
		updateMirrorFn: func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.UpdateMirrorReply, error) {
			return &rpc.UpdateMirrorReply{Diff: "updated"}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdEdit("-mirror-file", mirrorFile, "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdEditMirrorFileNoChange(t *testing.T) {
	tmpDir := t.TempDir()
	mirrorFile := filepath.Join(tmpDir, "mirror.yaml")

	mirror := &mirrors.Mirror{
		Name:     "test-mirror",
		HttpURL:  "http://same.com",
		Enabled:  true,
		Score:    5,
	}
	data, _ := yamlMarshalMirror(mirror)
	ioutil.WriteFile(mirrorFile, data, 0644)

	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		mirrorInfoFn: func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*rpc.Mirror, error) {
			return &rpc.Mirror{Name: "test-mirror", HttpURL: "http://same.com", Enabled: true, Score: 5,
				StateSince: makeTimestamp(1000), LastSync: makeTimestamp(1000), LastSuccessfulSync: makeTimestamp(1000), LastModTime: makeTimestamp(1000)}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdEdit("-mirror-file", mirrorFile, "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdEditWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdEdit()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdShow(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		mirrorInfoFn: func(ctx context.Context, in *rpc.MirrorIDRequest, opts ...grpc.CallOption) (*rpc.Mirror, error) {
			return &rpc.Mirror{Name: "test-mirror", HttpURL: "http://test.com", Comment: "a comment",
				StateSince: makeTimestamp(1000), LastSync: makeTimestamp(1000), LastSuccessfulSync: makeTimestamp(1000), LastModTime: makeTimestamp(1000)}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdShow("test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdShowWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdShow()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdExportMirmon(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", HttpURL: "http://m1.com/", RsyncURL: "rsync://m1.com/", FtpURL: "ftp://m1.com/", CountryCodes: "US", AdminEmail: "admin@m1.com", Enabled: true},
					{Name: "m2", HttpURL: "http://m2.com/", CountryCodes: "FR", Enabled: false},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdExport("mirmon")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdExportMirmonExcludeDisabled(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", HttpURL: "http://m1.com/", CountryCodes: "US", Enabled: true},
					{Name: "m2", HttpURL: "http://m2.com/", CountryCodes: "FR", Enabled: false},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdExport("-disabled=false", "mirmon")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdExportUnsupportedFormat(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdExport("unknown")
	if err != nil {
		t.Fatalf("Expected nil error for unsupported format, got: %v", err)
	}
}

func TestCmdExportWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdExport()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdEnable(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		changeStatusFn: func(ctx context.Context, in *rpc.ChangeStatusRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
			if !in.Enabled {
				t.Fatal("Expected enabled=true")
			}
			return &empty.Empty{}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdEnable("test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdDisable(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		changeStatusFn: func(ctx context.Context, in *rpc.ChangeStatusRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
			if in.Enabled {
				t.Fatal("Expected enabled=false")
			}
			return &empty.Empty{}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdDisable("test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdEnableWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdEnable()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdDisableWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdDisable()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdStatsFile(t *testing.T) {
	mock := &mockCLIClient{
		statsFileFn: func(ctx context.Context, in *rpc.StatsFileRequest, opts ...grpc.CallOption) (*rpc.StatsFileReply, error) {
			return &rpc.StatsFileReply{Files: map[string]int64{"file1.iso": 10, "file2.iso": 20}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdStats("file", "*.iso")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdStatsMirror(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		statsMirrorFn: func(ctx context.Context, in *rpc.StatsMirrorRequest, opts ...grpc.CallOption) (*rpc.StatsMirrorReply, error) {
			return &rpc.StatsMirrorReply{Requests: 100, Bytes: 1048576, Mirror: &rpc.Mirror{Enabled: true, Up: true}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdStats("mirror", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdStatsWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdStats("file")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdStatsInvalidType(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdStats("invalid", "test")
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdStatsWithDateRange(t *testing.T) {
	mock := &mockCLIClient{
		statsFileFn: func(ctx context.Context, in *rpc.StatsFileRequest, opts ...grpc.CallOption) (*rpc.StatsFileReply, error) {
			return &rpc.StatsFileReply{Files: map[string]int64{"f.iso": 5}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdStats("-start-date", "2024-01-01", "-end-date", "2024-12-31", "file", "*.iso")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdStatsMirrorHumanReadable(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		statsMirrorFn: func(ctx context.Context, in *rpc.StatsMirrorRequest, opts ...grpc.CallOption) (*rpc.StatsMirrorReply, error) {
			return &rpc.StatsMirrorReply{Requests: 50, Bytes: 524288, Mirror: &rpc.Mirror{Enabled: true, Up: true}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdStats("-h=false", "mirror", "m1")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdLogs(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		getMirrorLogsFn: func(ctx context.Context, in *rpc.GetMirrorLogsRequest, opts ...grpc.CallOption) (*rpc.GetMirrorLogsReply, error) {
			return &rpc.GetMirrorLogsReply{Line: []string{"log line 1", "log line 2"}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdLogs("test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdLogsEmpty(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "test-mirror"}}}, nil
		},
		getMirrorLogsFn: func(ctx context.Context, in *rpc.GetMirrorLogsRequest, opts ...grpc.CallOption) (*rpc.GetMirrorLogsReply, error) {
			return &rpc.GetMirrorLogsReply{Line: nil}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdLogs("test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdLogsWrongArgs(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdLogs()
	if err != nil {
		t.Fatalf("Expected nil error for usage, got: %v", err)
	}
}

func TestCmdReload(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdReload()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdUpgrade(t *testing.T) {
	c := newMockCli(&mockCLIClient{})
	err := c.CmdUpgrade()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdVersion(t *testing.T) {
	mock := &mockCLIClient{
		getVersionFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.VersionReply, error) {
			return &rpc.VersionReply{Version: "1.0.0", Build: "abc", GoVersion: "go1.20", OS: "linux", Arch: "amd64", GoMaxProcs: 4}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdVersionServerError(t *testing.T) {
	mock := &mockCLIClient{
		getVersionFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.VersionReply, error) {
			return nil, errors.New("server unavailable")
		},
	}
	c := newMockCli(mock)
	err := c.CmdVersion()
	if err == nil {
		t.Fatal("Expected error for server unavailable")
	}
}

func TestCmdVersionEmptyServerVersion(t *testing.T) {
	mock := &mockCLIClient{
		getVersionFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.VersionReply, error) {
			return &rpc.VersionReply{Version: ""}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdVersion()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdListWithRsyncFtp(t *testing.T) {
	mock := &mockCLIClient{
		listFn: func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*rpc.MirrorListReply, error) {
			return &rpc.MirrorListReply{
				Mirrors: []*rpc.Mirror{
					{Name: "m1", RsyncURL: "rsync://m1.com/", FtpURL: "ftp://m1.com/", Enabled: true, Up: true, StateSince: makeTimestamp(1000)},
				},
			}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdList("-rsync", "-ftp")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdStatsMirrorDisabled(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		statsMirrorFn: func(ctx context.Context, in *rpc.StatsMirrorRequest, opts ...grpc.CallOption) (*rpc.StatsMirrorReply, error) {
			return &rpc.StatsMirrorReply{Requests: 10, Bytes: 100, Mirror: &rpc.Mirror{Enabled: false, Up: false}}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdStats("mirror", "m1")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdAddWithAllFlags(t *testing.T) {
	mock := &mockCLIClient{
		addMirrorFn: func(ctx context.Context, in *rpc.Mirror, opts ...grpc.CallOption) (*rpc.AddMirrorReply, error) {
			return &rpc.AddMirrorReply{Country: "FR", Latitude: 48.85, Longitude: 2.35, Continent: "EU", ASN: "AS123", NetworkBandwidth: 1000}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdAdd(
		"-http", "http://example.com",
		"-rsync", "rsync://example.com/",
		"-ftp", "ftp://example.com/",
		"-sponsor-name", "Sponsor",
		"-sponsor-url", "http://sponsor.com",
		"-sponsor-logo", "http://sponsor.com/logo.png",
		"-admin-name", "Admin",
		"-admin-email", "admin@example.com",
		"-custom-data", "{\"key\":\"value\"}",
		"-continent-only",
		"-country-only",
		"-as-only",
		"-score", "100",
		"-comment", "test comment",
		"-net-bandwidth", "500",
		"-latitude", "48.85",
		"-longitude", "2.35",
		"-country", "France",
		"testmirror",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanScanErrorContinue(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{
				{ID: 1, Name: "m1"},
				{ID: 2, Name: "m2"},
			}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			return nil, errors.New("scan failed")
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-all")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestCmdScanWithTimeout(t *testing.T) {
	mock := &mockCLIClient{
		matchMirrorFn: func(ctx context.Context, in *rpc.MatchRequest, opts ...grpc.CallOption) (*rpc.MatchReply, error) {
			return &rpc.MatchReply{Mirrors: []*rpc.MirrorID{{ID: 1, Name: "m1"}}}, nil
		},
		scanMirrorFn: func(ctx context.Context, in *rpc.ScanMirrorRequest, opts ...grpc.CallOption) (*rpc.ScanMirrorReply, error) {
			return &rpc.ScanMirrorReply{FilesIndexed: 5}, nil
		},
	}
	c := newMockCli(mock)
	err := c.CmdScan("-timeout", "30", "test-mirror")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestParseCommandsExitOnRpcClose(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic: %v", r)
		}
	}()
	_ = ParseCommands("help")
}

func yamlMarshalMirror(m *mirrors.Mirror) ([]byte, error) {
	return yaml.Marshal(m)
}
