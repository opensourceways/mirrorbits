// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"net/http/httptest"
	"sync"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/network"
)

func TestMirrorStatsHandler_Unreachable(t *testing.T) {
	SetConfiguration(&Configuration{})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()

	h.mirrorStatsHandler(w, req, NewContext(w, req, h.templates))
}

func TestHTTP_Stop_AlreadyStopped(t *testing.T) {
	h := &HTTP{stopped: true}
	h.Stop(0)
}

func TestHTTP_StopChan(t *testing.T) {
	h := &HTTP{}
	ch := h.StopChan()
	_ = ch
}

func TestHTTP_SetListener2(t *testing.T) {
	h := &HTTP{}
	ln, err := httptest.NewServer(nil).Listener, error(nil)
	_ = err
	if ln != nil {
		h.SetListener(ln)
	}
}

func TestHTTP_Reload_GeoIP(t *testing.T) {
	h := &HTTP{
		geoip:     network.NewGeoIP(),
		templates: Templates{RWMutex: new(sync.RWMutex)},
	}
	h.geoip.LoadGeoIP()
}
