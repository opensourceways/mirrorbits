// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved
// Licensed under the MIT license

package scan

import (
	"testing"

	"github.com/opensourceways/mirrorbits/filesystem"
)

func TestHttpScanner_Scan_InvalidURL(t *testing.T) {
	r := &HttpScanner{scan: &scan{}}
	repoVersion := []*filesystem.LayerFile{
		{Dir: "v1/ISO", Name: "file.iso"},
	}
	_, _, err := r.Scan("https://\x00invalid", "test-mirror", repoVersion, nil)
	if err == nil {
		t.Fatalf("Expected error for invalid URL")
	}
}
