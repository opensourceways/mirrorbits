// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package scan

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     300 * time.Second,
	},
	Timeout: 30 * time.Second,
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

	req, err := http.NewRequest("HEAD", uri.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("mirror http url: " + resp.Status + " " + httpUrl + " request failed")
	}

	fileList := filesystem.GetSelectorList()
	fd := filesystem.FileData{}
	for _, fl := range fileList {
		fileUrl := strings.ReplaceAll(fl.Dir, filesystem.Sep, "/") + "/" + fl.Name
		req1, _ := http.NewRequest("HEAD", uri.String()+"/"+fileUrl, nil)
		req1.Header.Set("User-Agent", userAgent)
		resp1, err1 := client.Do(req)
		if err1 != nil {
			return 0, err
		}
		if resp1.StatusCode != http.StatusOK {
			return 0, errors.New("mirror file http url: " + uri.String() + "/" + fileUrl + " request failed")
		}

		sizeStr := resp1.Header.Get("Content-Length")
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		fd.Path = fileUrl
		fd.Size = size

		modTimeStr := resp1.Header.Get("Last-Modified")
		modTime, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modTimeStr)
		fd.ModTime = modTime
		r.scan.ScannerAddFile(fd)
	}

	return 0, nil
}
