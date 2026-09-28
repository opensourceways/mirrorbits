// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package filesystem

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/config"
)

func TestConvertFileSize(t *testing.T) {
	tests := []struct {
		input    []byte
		expected int64
	}{
		{[]byte("  23,545,123"), 23545123},
		{[]byte("  1,000"), 1000},
		{[]byte("  42"), 42},
	}
	for _, tt := range tests {
		result := covertFileSize(tt.input)
		if result != tt.expected {
			t.Fatalf("For input %q, expected %d, got %d", string(tt.input), tt.expected, result)
		}
	}
}

func TestSetRecentFile(t *testing.T) {
	fileTreeReplica.Mapping = make(map[string]*LayerFile)
	fileTreeReplica.SelectorMap = make(map[string]*LayerFile)

	lf1 := &LayerFile{Dir: "v1/ISO", Name: "file.iso", ModTime: time.Now().Add(-1 * time.Hour)}
	lf1.setRecentFile()

	if v, ok := fileTreeReplica.SelectorMap["v1"]; !ok || v != lf1 {
		t.Fatalf("Expected selector map to have v1")
	}

	lf2 := &LayerFile{Dir: "v1/ISO", Name: "newer.iso", ModTime: time.Now()}
	lf2.setRecentFile()

	if v, ok := fileTreeReplica.SelectorMap["v1"]; !ok || v != lf2 {
		t.Fatalf("Expected selector map to have newer file for v1")
	}
}

func TestLayerFile_flattening_Nil(t *testing.T) {
	var lf *LayerFile
	result := lf.flattening()
	if result != nil {
		t.Fatalf("Expected nil for nil LayerFile")
	}
}

func TestLayerFile_flattening_Empty(t *testing.T) {
	lf := &LayerFile{}
	result := lf.flattening()
	if result != nil {
		t.Fatalf("Expected nil for empty LayerFile")
	}
}

func TestLayerFile_toDisplayFile_File(t *testing.T) {
	lf := &LayerFile{
		Dir:    "/test",
		Name:   "file.txt",
		Size:   1024,
		Sha256: "abc123",
	}
	df := lf.toDisplayFile()
	if df.Name != "file.txt" {
		t.Fatalf("Expected name 'file.txt'")
	}
	if df.Type != "file" {
		t.Fatalf("Expected type 'file'")
	}
	if df.Size == "" {
		t.Fatalf("Expected non-empty size")
	}
}

func TestLayerFile_toDisplayFile_Dir(t *testing.T) {
	sub := &LayerFile{Dir: "/test/sub", Name: "child.txt", Size: 100}
	lf := &LayerFile{
		Dir:  "/test",
		Name: "dir",
		Sub:  []*LayerFile{sub},
	}
	df := lf.toDisplayFile()
	if df.Type != "dir" {
		t.Fatalf("Expected type 'dir'")
	}
	if len(df.Sub) != 1 {
		t.Fatalf("Expected 1 sub item")
	}
}

func TestLayerFile_dfsEveryFile_Nil(t *testing.T) {
	var lf *LayerFile
	lf.dfsEveryFile(nil)
}

func TestLayerFile_dfsFirstFile_Nil(t *testing.T) {
	var lf *LayerFile
	result := lf.dfsFirstFile()
	if result != nil {
		t.Fatalf("Expected nil for nil LayerFile")
	}
}

func TestLayerFile_dfsFirstFile_Leaf(t *testing.T) {
	lf := &LayerFile{Name: "leaf.txt"}
	result := lf.dfsFirstFile()
	if result == nil {
		t.Fatalf("Expected non-nil for leaf")
	}
}

func TestGetRepoFileList_NotInMap(t *testing.T) {
	cnf := &config.Configuration{
		Repository:   "/tmp",
		Fallbacks:    []config.Fallback{{URL: "http://fallback.com"}},
	}
	result := GetRepoFileList("nonexistent-version", cnf)
	_ = result
}

func TestUpdateFileTree(t *testing.T) {
	fileTreeReplica.Mapping = make(map[string]*LayerFile, 10)
	fileTreeReplica.SelectorMap = make(map[string]*LayerFile, 10)
	fileTreeReplica.Root = LayerFile{}

	UpdateFileTree(config.DirFilter{})
}

func TestBuildFileTree(t *testing.T) {
	fileTreeReplica.Mapping = make(map[string]*LayerFile, 10)
	fileTreeReplica.SelectorMap = make(map[string]*LayerFile, 10)
	fileTreeReplica.Root = LayerFile{}

	dir, _ := os.MkdirTemp("", "buildfile-test")
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.iso")
	os.WriteFile(testFile, []byte("hello"), 0644)

	cnf := &config.Configuration{Repository: dir}

	fd := BuildFileTree("openEuler-22.03/ISO/file.iso", []byte("  1,000"), []byte("2024 08 08 11:01:29"), cnf)
	_ = fd
}

func TestGetRepoVersionList_ReturnsValue(t *testing.T) {
	result := GetRepoVersionList()
	_ = result
}

func TestLayerFile_setFileData(t *testing.T) {
	dir, _ := os.MkdirTemp("", "setfile-test")
	defer os.RemoveAll(dir)

	cnf := &config.Configuration{Repository: dir}

	lf := &LayerFile{Dir: "v1/ISO", Name: "file.iso"}
	fd := lf.setFileData("v1/ISO/file.iso", []byte("  1,000"), []byte("2024 08 08 11:01:29"), cnf)

	if fd == nil {
		t.Fatalf("Expected non-nil FileData")
	}
	if fd.Size != 1000 {
		t.Fatalf("Expected size 1000, got %d", fd.Size)
	}
	if fd.Path != "v1/ISO/file.iso" {
		t.Fatalf("Expected path 'v1/ISO/file.iso'")
	}
}

func TestAppendParticularScenarioArch(t *testing.T) {
	list := []string{"a", "b"}
	result := appendParticularScenarioArch(list, "c")
	if len(result) != 3 || result[2] != "c" {
		t.Fatalf("Expected 3 items with 'c' appended")
	}

	result = appendParticularScenarioArch(list, "a")
	if len(result) != 2 {
		t.Fatalf("Expected no duplicate")
	}
}
