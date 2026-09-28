// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package logs

import (
	"os"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
)

func TestReloadLogs_NonDaemon(t *testing.T) {
	SetConfiguration(&Configuration{})
	core.Daemon = false
	ReloadLogs()
}

func TestReloadLogs_DaemonMode(t *testing.T) {
	SetConfiguration(&Configuration{LogDir: ""})
	core.Daemon = true
	ReloadLogs()
	core.Daemon = false
}

func TestReloadDownloadLogs_WithLogDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-dllogs")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	SetConfiguration(&Configuration{LogDir: dir})
	ReloadDownloadLogs()
	dlogger.Close()
}

func TestReloadDownloadLogs_InvalidLogDir(t *testing.T) {
	SetConfiguration(&Configuration{LogDir: "/nonexistent/path/that/does/not/exist"})
	ReloadDownloadLogs()
}

func TestReloadRuntimeLogs_WithDebug(t *testing.T) {
	rlogger.f = nil
	core.RunLog = ""
	core.Debug = true
	ReloadRuntimeLogs()
	core.Debug = false
	if rlogger.f == nil {
		t.Fatalf("Expected logger to be set up")
	}
}

func TestSetDownloadLogWriter_WithHeader(t *testing.T) {
	var buf = &writeCloserBuf{}
	setDownloadLogWriter(buf, true)
	if dlogger.l == nil {
		t.Fatalf("Expected logger to be created")
	}
	if buf.Len() == 0 {
		t.Fatalf("Expected header to be written")
	}
	dlogger.Close()
}

type writeCloserBuf struct {
	buf []byte
}

func (w *writeCloserBuf) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *writeCloserBuf) Close() error {
	return nil
}

func (w *writeCloserBuf) Len() int {
	return len(w.buf)
}
