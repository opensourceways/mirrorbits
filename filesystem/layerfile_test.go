// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package filesystem

import (
	"testing"

	"github.com/opensourceways/mirrorbits/config"
)

func TestGetRepoFileData_EmptyPath(t *testing.T) {
	result := GetRepoFileData("")
	if result.Dir != "" || result.Name != "" {
		t.Fatalf("Expected empty LayerFile for empty path")
	}
}

func TestGetRepoFileData_NotFound(t *testing.T) {
	result := GetRepoFileData("/nonexistent/path/file.txt")
	if result.Dir != "" || result.Name != "" {
		t.Fatalf("Expected empty LayerFile for not found path")
	}
}

func TestGetRepoVersionList(t *testing.T) {
	result := GetRepoVersionList()
	_ = result // Can be nil or empty initially
}

func TestGetSelectorList(t *testing.T) {
	result := GetSelectorList()
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
}

func TestInitPathFilter_Empty(t *testing.T) {
	filter := config.DirFilter{}
	InitPathFilter(filter)
}

func TestInitPathFilter_WithDirs(t *testing.T) {
	filter := config.DirFilter{
		SecondDir: []string{"dir1", "dir2"},
		ThirdDir:  []string{"sub1", "sub2"},
	}
	InitPathFilter(filter)
}

func TestInitPathFilter_WithParticularFile(t *testing.T) {
	filter := config.DirFilter{
		SecondDir: []string{"dir1"},
		ThirdDir:  []string{"sub1"},
		ParticularFile: []config.ParticularFileMapping{
			{
				VersionName:  "v1",
				ScenarioName: "ISO",
				ArchName:     "x86_64",
				SourcePath:   []string{"path/to/file1.iso", "path/to/file2.iso"},
				SHA256List:   []string{"sha1", "sha2"},
			},
		},
	}
	InitPathFilter(filter)
}

func TestFilter_NotRepoVersion(t *testing.T) {
	result := Filter("some/random/path.txt")
	if result {
		t.Fatalf("Expected false for non-repo-version path")
	}
}

func TestFilter_Sha256Suffix(t *testing.T) {
	result := Filter("openEuler-22.03/file.sha256sum")
	if result {
		t.Fatalf("Expected false for sha256sum suffix")
	}
}

func TestFilter_NoPathFilter(t *testing.T) {
	// Reset pathFilter
	pathFilter = nil
	result := Filter("openEuler-22.03/ISO/file.iso")
	if result {
		t.Fatalf("Expected false when no pathFilter set")
	}
}

func TestFilter_WithPathFilter(t *testing.T) {
	// Set up path filter for a non-ISO directory
	pathFilter = []string{"Everything/Package"}
	result := Filter("openEuler-22.03/Everything/Package/file.rpm")
	if !result {
		t.Fatalf("Expected true for matching path filter")
	}
}

func TestFilter_ISOWithoutISOExtension(t *testing.T) {
	pathFilter = []string{"ISO"}
	result := Filter("openEuler-22.03/ISO/Package/file.rpm")
	if result {
		t.Fatalf("Expected false for ISO dir with non-iso file")
	}
}

func TestFilter_ISOWithISOExtension(t *testing.T) {
	pathFilter = []string{"ISO"}
	result := Filter("openEuler-22.03/ISO/file.iso")
	if !result {
		t.Fatalf("Expected true for ISO dir with .iso file")
	}
}

func TestFilter_EdgeImgWithoutISOExtension(t *testing.T) {
	pathFilter = []string{"edge_img"}
	result := Filter("openEuler-22.03/edge_img/file.txt")
	if result {
		t.Fatalf("Expected false for edge_img dir with non-iso file")
	}
}

func TestFilter_EdgeImgWithISOExtension(t *testing.T) {
	pathFilter = []string{"edge_img"}
	result := Filter("openEuler-22.03/edge_img/file.iso")
	if !result {
		t.Fatalf("Expected true for edge_img dir with .iso file")
	}
}

func TestConstants(t *testing.T) {
	if Sep != "/" {
		t.Fatalf("Expected Sep '/', got %s", Sep)
	}
	if FileExtensionSha256 != ".sha256sum" {
		t.Fatalf("Expected '.sha256sum', got %s", FileExtensionSha256)
	}
	if StandardISOFileExtension != ".iso" {
		t.Fatalf("Expected '.iso', got %s", StandardISOFileExtension)
	}
	if RepoVersionDirectoryPrefix != "openEuler-" {
		t.Fatalf("Expected 'openEuler-', got %s", RepoVersionDirectoryPrefix)
	}
}

func TestDisplayFileStruct(t *testing.T) {
	df := DisplayFile{
		Name:    "test.iso",
		Path:    "/path/to/test.iso",
		Size:    "1.0 GiB",
		ShaCode: "abc123",
		Type:    "file",
	}
	if df.Name != "test.iso" {
		t.Fatalf("Name mismatch")
	}
	if df.Path != "/path/to/test.iso" {
		t.Fatalf("Path mismatch")
	}
}

func TestDisplayFileListStruct(t *testing.T) {
	dfl := DisplayFileList{
		Scenario: "ISO",
		Arch:     "x86_64",
		Tree:     []DisplayFile{{Name: "file.iso"}},
	}
	if dfl.Scenario != "ISO" {
		t.Fatalf("Scenario mismatch")
	}
	if len(dfl.Tree) != 1 {
		t.Fatalf("Tree length mismatch")
	}
}

func TestDisplayRepoVersionStruct(t *testing.T) {
	drv := DisplayRepoVersion{
		Version:  "22.03",
		Scenario: []string{"ISO", "Everything"},
		Arch:     []string{"x86_64", "aarch64"},
		LTS:      true,
	}
	if drv.Version != "22.03" {
		t.Fatalf("Version mismatch")
	}
	if len(drv.Scenario) != 2 {
		t.Fatalf("Scenario length mismatch")
	}
	if !drv.LTS {
		t.Fatalf("LTS mismatch")
	}
}

func TestFileDataStruct(t *testing.T) {
	fd := FileData{
		Path:    "/test",
		Sha1:    "sha1",
		Sha256:  "sha256",
		Md5:     "md5",
		Size:    100,
	}
	if fd.Path != "/test" {
		t.Fatalf("Path mismatch")
	}
	if fd.Sha256 != "sha256" {
		t.Fatalf("Sha256 mismatch")
	}
}

func TestLayerFileStruct(t *testing.T) {
	lf := LayerFile{
		Dir:     "/dir",
		Name:    "file.txt",
		Size:    200,
		Sha256:  "hash",
		Type:    "file",
	}
	if lf.Dir != "/dir" {
		t.Fatalf("Dir mismatch")
	}
	if lf.Name != "file.txt" {
		t.Fatalf("Name mismatch")
	}
}

func TestFileStoreStruct(t *testing.T) {
	fs := FileStore{
		Mapping:     make(map[string]*LayerFile),
		SelectorMap: make(map[string]*LayerFile),
	}
	if len(fs.Mapping) != 0 {
		t.Fatalf("Mapping should be empty")
	}
}
