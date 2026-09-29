// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package http

import (
	"net/http/httptest"
	"sync"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/network"
)

func TestMirrorHandler_EmptyPath(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp/repo"})
	h := &HTTP{
		geoip:     network.NewGeoIP(),
		engine:    DefaultEngine{},
		templates: Templates{RWMutex: new(sync.RWMutex)},
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req, h.templates)

	h.mirrorHandler(w, req, c)
}

func TestRequestDispatcher_MirrorStats(t *testing.T) {
	SetConfiguration(&Configuration{Repository: "/tmp/repo"})
	_, h := newHTTPWithMockRedis()

	req := httptest.NewRequest("GET", "/?mirrorstats", nil)
	w := httptest.NewRecorder()

	h.requestDispatcher(w, req)
}
