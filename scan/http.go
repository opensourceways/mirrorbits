// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package scan

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var mirrorCheckClient = resty.New().RemoveProxy().SetHeader(userAgentName, userAgent)

// HttpScanner is the implementation of an http scanner
type HttpScanner struct {
	scan *scan
}

// Scan starts an http head scan of the given mirror
func (r *HttpScanner) Scan(httpUrl, identifier string, repoVersion []*filesystem.LayerFile, stop <-chan struct{}) (core.Precision, string, error) {

	fileList := repoVersion
	recentFile := fileList[0]
	filePath := recentFile.Dir + filesystem.Sep + recentFile.Name

	if !strings.HasPrefix(httpUrl, "https://") {
		return 0, filePath, fmt.Errorf("[%s] %s does not start with https://", identifier, httpUrl)
	}

	if utils.IsStopped(stop) {
		return 0, filePath, ErrScanAborted
	}

	uri, err := url.Parse(httpUrl)
	if err != nil {
		return 0, filePath, err
	}
	client := mirrorCheckClient
	head, err := client.R().Head(uri.String())
	if err != nil {
		return 0, filePath, err
	}
	if err != nil {
		return 0, filePath, err
	}
	if head.StatusCode() != http.StatusOK {
		return 0, filePath, errors.New(identifier + " mirror http url: " + head.Status() + " " + httpUrl + " request failed")
	}

	fd := filesystem.FileData{}
	for i, fl := range fileList {
		fileUrl := fl.Dir + filesystem.Sep + fl.Name

	retry:
		headFileUrl := utils.ConcatURL(uri.String(), fileUrl)
		head1, err1 := client.R().Head(headFileUrl)
		if err1 != nil {
			return 0, filePath, err
		}
		if head1.StatusCode() == http.StatusTooManyRequests {
			time.Sleep(time.Second)
			goto retry
		}
		if head1.StatusCode() != http.StatusOK {
			return 0, filePath, fmt.Errorf("file no.%d, http url: %s request failed, response status: %s", i, headFileUrl, head1.Status())
		}

		sizeStr := head1.Header().Get("Content-Length")
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		fd.Path = filePath
		fd.Size = size

		sourceFile := filesystem.GetRepoFileData(fileUrl)
		if size == 0 || sourceFile.Size != size {
			return 0, filePath, fmt.Errorf("file no.%d, http url: %s, size mismatch: %d[dest] != %d[src]", i, headFileUrl, size, sourceFile.Size)
		}

		modTimeStr := head1.Header().Get("Last-Modified")
		modTime, _ := time.Parse(time.RFC1123, modTimeStr)
		fd.ModTime = modTime
		r.scan.ScannerAddFile(fd)

	}

	return 0, "", nil
}
