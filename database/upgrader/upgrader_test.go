// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package upgrader

import (
	"testing"
)

func TestGetUpgraderV1(t *testing.T) {
	u := GetUpgrader(nil, 1)
	if u == nil {
		t.Fatal("Expected non-nil upgrader for version 1")
	}
}

func TestGetUpgraderUnknown(t *testing.T) {
	u := GetUpgrader(nil, 99)
	if u != nil {
		t.Fatal("Expected nil for unknown version")
	}
}
