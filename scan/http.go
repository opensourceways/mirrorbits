// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package scan

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     300 * time.Second,
	},
	Timeout: 10 * time.Second,
}

// HttpScanner is the implementation of an rsync scanner
type HttpScanner struct {
	scan *scan
}

// Scan starts an rsync scan of the given mirror
func (r *HttpScanner) Scan(httpUrl, identifier string, conn redis.Conn, stop <-chan struct{}) (core.Precision, error) {

	if !strings.HasPrefix(httpUrl, "https://") {
		return 0, fmt.Errorf("%s does not start with https://", httpUrl)
	}

	if utils.IsStopped(stop) {
		return 0, ErrScanAborted
	}

	uri, err := url.Parse(httpUrl)
	if err != nil {
		return 0, err
	}

	resp, err := client.Head(uri.String())
	if err != nil {
		return 0, err
	}
	if resp.StatusCode == http.StatusOK {
		return core.Precision(time.Second), nil
	}

	return 0, nil
}
