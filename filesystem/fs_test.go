// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestNewFileInfo(t *testing.T) {
	fi := NewFileInfo("/test/path.txt")
	if fi.Path != "/test/path.txt" {
		t.Fatalf("Expected /test/path.txt, got %s", fi.Path)
	}
}

func TestIsInRepository(t *testing.T) {
	tests := []struct {
		repo string
		path string
		want bool
	}{
		{"/repo", "/repo", true},
		{"/repo", "/repo/file.txt", true},
		{"/repo", "/repo/sub/file.txt", true},
		{"/repo", "/other/file.txt", false},
		{"/repo", "/repofile.txt", false},
		{"/repo/", "/repo", false},
	}
	for _, tt := range tests {
		if got := IsInRepository(tt.repo, tt.path); got != tt.want {
			t.Fatalf("IsInRepository(%q, %q): expected %v, got %v", tt.repo, tt.path, tt.want, got)
		}
	}
}

func TestEvaluateFilePath(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-fs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0755)

	f, err := os.Create(filepath.Join(subdir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello")
	f.Close()

	result, err := EvaluateFilePath(dir, "/sub/test.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result != "/sub/test.txt" {
		t.Fatalf("Expected /sub/test.txt, got %s", result)
	}
}

func TestEvaluateFilePathOutsideRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-fs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	_, err = EvaluateFilePath(dir, "/../../../etc/passwd")
	if err != ErrOutsideRepo {
		t.Fatalf("Expected ErrOutsideRepo, got %v", err)
	}
}

func TestHashFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-hash-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "test.txt")
	f, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello world")
	f.Close()

	cnf := &Configuration{}
	cnf.Hashes.SHA256 = true
	cnf.Hashes.SHA1 = true
	cnf.Hashes.MD5 = true

	fi, err := HashFile(fpath, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if fi.Sha256 == "" {
		t.Fatal("Expected non-empty Sha256")
	}
	if fi.Sha1 == "" {
		t.Fatal("Expected non-empty Sha1")
	}
	if fi.Md5 == "" {
		t.Fatal("Expected non-empty Md5")
	}
}

func TestHashFileNoHashes(t *testing.T) {
	dir, _ := os.MkdirTemp("", "mirrorbits-hash-test-*")
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "test.txt")
	f, _ := os.Create(fpath)
	f.WriteString("hello world")
	f.Close()

	cnf := &Configuration{}

	fi, err := HashFile(fpath, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if fi.Sha256 != "" {
		t.Fatal("Expected empty Sha256")
	}
}

func TestHashFileNonExistent(t *testing.T) {
	cnf := &Configuration{}
	cnf.Hashes.SHA256 = true

	_, err := HashFile("/nonexistent/file.txt", cnf)
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestSha256sum(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-sha-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "test.txt")
	f, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello world")
	f.Close()

	hash, err := Sha256sum(fpath)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(hash) == 0 {
		t.Fatal("Expected non-empty hash")
	}
}

func TestSha256sumNonExistent(t *testing.T) {
	_, err := Sha256sum("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestCovertFileSize(t *testing.T) {
	tests := []struct {
		input    []byte
		expected int64
	}{
		{[]byte("  23,545,123"), 23545123},
		{[]byte("  1,024"), 1024},
		{[]byte("  42"), 42},
		{[]byte(" 0"), 0},
	}
	for _, tt := range tests {
		if got := covertFileSize(tt.input); got != tt.expected {
			t.Fatalf("covertFileSize(%q): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestFilter(t *testing.T) {
	InitPathFilter(DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64"},
	})

	tests := []struct {
		path     string
		expected bool
	}{
		{"openEuler-22.03/ISO/x86_64/test.iso", true},
		{"openEuler-22.03/ISO/x86_64/test.txt", false},
		{"non-prefix/ISO/x86_64/test.iso", false},
		{"openEuler-22.03/ISO/x86_64/test.sha256sum", false},
		{"openEuler-22.03/repo/x86_64/test.iso", false},
	}
	for _, tt := range tests {
		if got := Filter(tt.path); got != tt.expected {
			t.Fatalf("Filter(%q): expected %v, got %v", tt.path, tt.expected, got)
		}
	}
}

func TestGetRepoFileData(t *testing.T) {
	r := GetRepoFileData("")
	if r.Dir != "" && r.Name != "" {
		t.Fatal("Expected empty LayerFile for empty path")
	}
}

func TestGetRepoVersionList(t *testing.T) {
	l := GetRepoVersionList()
	_ = l
}

func TestGetSelectorList(t *testing.T) {
	m := GetSelectorList()
	if m == nil {
		t.Fatal("Expected non-nil map")
	}
}
