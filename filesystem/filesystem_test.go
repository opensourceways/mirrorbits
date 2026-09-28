// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package filesystem

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
)

func TestNewFileInfo(t *testing.T) {
	fi := NewFileInfo("/test/path/file.tgz")
	if fi.Path != "/test/path/file.tgz" {
		t.Fatalf("Expected path '/test/path/file.tgz', got %s", fi.Path)
	}
}

func TestIsInRepository(t *testing.T) {
	repo := "/tmp/repo"

	if !IsInRepository(repo, repo) {
		t.Fatalf("Expected true for repository root")
	}

	if !IsInRepository(repo, "/tmp/repo/file.txt") {
		t.Fatalf("Expected true for file in repository")
	}

	if !IsInRepository(repo, "/tmp/repo/subdir/file.txt") {
		t.Fatalf("Expected true for file in subdirectory")
	}

	if IsInRepository(repo, "/tmp/other/file.txt") {
		t.Fatalf("Expected false for file outside repository")
	}

	if IsInRepository(repo, "/tmp/repofile.txt") {
		t.Fatalf("Expected false for path that looks like prefix but isn't in repo")
	}
}

func TestEvaluateFilePath(t *testing.T) {
	dir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Unable to create subdir: %s", err)
	}

	testFile := filepath.Join(subdir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Unable to write test file: %s", err)
	}

	result, err := EvaluateFilePath(dir, "/subdir/test.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result != "/subdir/test.txt" {
		t.Fatalf("Expected '/subdir/test.txt', got %s", result)
	}
}

func TestEvaluateFilePath_OutsideRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	_, err = EvaluateFilePath(dir, "/../../etc/passwd")
	if err != ErrOutsideRepo {
		t.Fatalf("Expected ErrOutsideRepo, got %v", err)
	}
}

func TestEvaluateFilePath_Symlink(t *testing.T) {
	dir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	targetDir := filepath.Join(dir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Unable to create target dir: %s", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	linkPath := filepath.Join(dir, "link")
	if err := os.Symlink(targetDir, linkPath); err != nil {
		t.Fatalf("Unable to create symlink: %s", err)
	}

	result, err := EvaluateFilePath(dir, "/link/file.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if result != "/target/file.txt" {
		t.Fatalf("Expected '/target/file.txt', got %s", result)
	}
}

func TestHashFile_SHA256(t *testing.T) {
	dir, err := os.MkdirTemp("", "hash-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	cnf := &Configuration{}
	cnf.Hashes.SHA256 = true

	hashes, err := HashFile(testFile, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if hashes.Sha256 == "" {
		t.Fatalf("Expected non-empty SHA256")
	}
}

func TestHashFile_SHA1(t *testing.T) {
	dir, err := os.MkdirTemp("", "hash-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	cnf := &Configuration{}
	cnf.Hashes.SHA1 = true

	hashes, err := HashFile(testFile, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if hashes.Sha1 == "" {
		t.Fatalf("Expected non-empty SHA1")
	}
}

func TestHashFile_MD5(t *testing.T) {
	dir, err := os.MkdirTemp("", "hash-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	cnf := &Configuration{}
	cnf.Hashes.MD5 = true

	hashes, err := HashFile(testFile, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if hashes.Md5 == "" {
		t.Fatalf("Expected non-empty MD5")
	}
}

func TestHashFile_NoHashes(t *testing.T) {
	dir, err := os.MkdirTemp("", "hash-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	cnf := &Configuration{}

	_, err = HashFile(testFile, cnf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestHashFile_FileNotFound(t *testing.T) {
	cnf := &Configuration{}
	cnf.Hashes.SHA256 = true

	_, err := HashFile("/nonexistent/file", cnf)
	if err == nil {
		t.Fatalf("Expected error for non-existent file")
	}
}

func TestSha256sum(t *testing.T) {
	dir, err := os.MkdirTemp("", "sha256-test")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	testFile := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Unable to write file: %s", err)
	}

	hash, err := Sha256sum(testFile)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(hash) == 0 {
		t.Fatalf("Expected non-empty hash")
	}
}

func TestSha256sum_FileNotFound(t *testing.T) {
	_, err := Sha256sum("/nonexistent/file")
	if err == nil {
		t.Fatalf("Expected error for non-existent file")
	}
}

func TestFileInfoStruct(t *testing.T) {
	fi := FileInfo{
		Path:    "/test",
		Size:    100,
		ModTime: time.Now(),
		Sha1:    "sha1hash",
		Sha256:  "sha256hash",
		Md5:     "md5hash",
	}
	if fi.Path != "/test" {
		t.Fatalf("Path mismatch")
	}
	if fi.Size != 100 {
		t.Fatalf("Size mismatch")
	}
	if fi.Sha1 != "sha1hash" {
		t.Fatalf("Sha1 mismatch")
	}
	if fi.Sha256 != "sha256hash" {
		t.Fatalf("Sha256 mismatch")
	}
	if fi.Md5 != "md5hash" {
		t.Fatalf("Md5 mismatch")
	}
}
