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

func TestInitPathFilterEmpty(t *testing.T) {
	old := pathFilter
	defer func() { pathFilter = old }()

	pathFilter = nil
	InitPathFilter(DirFilter{})
	if len(pathFilter) != 0 {
		t.Fatalf("Expected empty pathFilter, got %d", len(pathFilter))
	}

	InitPathFilter(DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64"},
	})
	if len(pathFilter) == 0 {
		t.Fatal("Expected non-empty pathFilter")
	}

	InitPathFilter(DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64"},
		ParticularFile: []ParticularFileMapping{
			{SourcePath: []string{"openEuler-22.03/special/path"}},
		},
	})
	found := false
	for _, p := range pathFilter {
		if p == "openEuler-22.03/special/path" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected particular file path in filter")
	}

	pathFilter = nil
}

func TestGetRepoFileDataNonExistent(t *testing.T) {
	got := GetRepoFileData("/nonexistent/path")
	if got.Dir != "" || got.Name != "" {
		t.Fatalf("Expected empty LayerFile for nonexistent path, got %v", got)
	}
}

func TestGetRepoVersionListNonNil(t *testing.T) {
	list := GetRepoVersionList()
	_ = list
}

func TestGetSelectorListNonNil(t *testing.T) {
	m := GetSelectorList()
	if m == nil {
		t.Fatal("Expected non-nil map")
	}
}

func TestLayerFileToDisplayFileLeaf(t *testing.T) {
	lf := &LayerFile{
		Dir:    "/test",
		Name:   "file.iso",
		Size:   1024,
		Sha256: "abc123",
		Type:   "file",
	}
	df := lf.toDisplayFile()
	if df.Name != "file.iso" {
		t.Fatalf("Expected file.iso, got %s", df.Name)
	}
	if df.Path != "/test/file.iso" {
		t.Fatalf("Expected /test/file.iso, got %s", df.Path)
	}
	if df.Type != "file" {
		t.Fatalf("Expected file, got %s", df.Type)
	}
	if df.Size == "" {
		t.Fatal("Expected non-empty size")
	}
	if df.ShaCode != "abc123" {
		t.Fatalf("Expected abc123, got %s", df.ShaCode)
	}
	if len(df.Sub) != 0 {
		t.Fatalf("Expected no sub, got %d", len(df.Sub))
	}
}

func TestLayerFileToDisplayFileDir(t *testing.T) {
	leaf := &LayerFile{
		Dir:   "/test/sub",
		Name:  "file.iso",
		Size:  1024,
		Type:  "file",
	}
	parent := &LayerFile{
		Dir:  "/test",
		Name: "subdir",
		Type: "dir",
		Sub:  []*LayerFile{leaf},
	}
	df := parent.toDisplayFile()
	if df.Type != "dir" {
		t.Fatalf("Expected dir, got %s", df.Type)
	}
	if len(df.Sub) != 1 {
		t.Fatalf("Expected 1 sub, got %d", len(df.Sub))
	}
	if df.Sub[0].Name != "file.iso" {
		t.Fatalf("Expected file.iso, got %s", df.Sub[0].Name)
	}
}

func TestLayerFileCollectFileInfoEmpty(t *testing.T) {
	lf := &LayerFile{Dir: "/test", Name: "arch"}
	ans := lf.collectFileInfo(nil, "ISO")
	if len(ans) != 0 {
		t.Fatalf("Expected empty, got %d", len(ans))
	}
}

func TestLayerFileCollectFileInfoWithSub(t *testing.T) {
	leaf := &LayerFile{
		Dir:   "/test/ISO/x86_64",
		Name:  "file.iso",
		Size:  1024,
		Type:  "file",
	}
	parent := &LayerFile{
		Dir:  "openEuler-22.03/ISO",
		Name: "x86_64",
		Sub:  []*LayerFile{leaf},
	}
	ans := parent.collectFileInfo(nil, "ISO")
	if len(ans) != 1 {
		t.Fatalf("Expected 1, got %d", len(ans))
	}
	if ans[0].Scenario != "ISO" {
		t.Fatalf("Expected ISO, got %s", ans[0].Scenario)
	}
	if ans[0].Arch != "x86_64" {
		t.Fatalf("Expected x86_64, got %s", ans[0].Arch)
	}
}

func TestLayerFileCollectFileInfoEmbeddedImg(t *testing.T) {
	leaf := &LayerFile{Name: "file.iso", Size: 1024, Type: "file"}
	mid := &LayerFile{Sub: []*LayerFile{leaf}}
	parent := &LayerFile{
		Dir:  "openEuler-22.03/embedded_img",
		Name: "x86_64",
		Sub:  []*LayerFile{mid},
	}
	ans := parent.collectFileInfo(nil, "embedded_img")
	if len(ans) != 1 {
		t.Fatalf("Expected 1, got %d", len(ans))
	}
}

func TestLayerFileFlatteningEmpty(t *testing.T) {
	var lf *LayerFile
	if lf.flattening() != nil {
		t.Fatal("Expected nil for nil")
	}

	empty := &LayerFile{}
	if empty.flattening() != nil {
		t.Fatal("Expected nil for empty")
	}

	empty2 := &LayerFile{Sub: []*LayerFile{}}
	if empty2.flattening() != nil {
		t.Fatal("Expected nil for empty sub")
	}
}

func TestLayerFileDfsEveryFileNil(t *testing.T) {
	var lf *LayerFile
	lf.dfsEveryFile(nil)
}

func TestLayerFileDfsFirstFileNil(t *testing.T) {
	var lf *LayerFile
	if lf.dfsFirstFile() != nil {
		t.Fatal("Expected nil for nil")
	}
}

func TestLayerFileDfsFirstFileLeaf(t *testing.T) {
	leaf := &LayerFile{Name: "file.iso", Type: "file"}
	if got := leaf.dfsFirstFile(); got == nil || got.Name != "file.iso" {
		t.Fatalf("Expected file.iso, got %v", got)
	}
}

func TestLayerFileDfsFirstFileSingleSub(t *testing.T) {
	leaf := &LayerFile{Name: "file.iso", Type: "file"}
	parent := &LayerFile{Name: "dir", Sub: []*LayerFile{leaf}}
	if got := parent.dfsFirstFile(); got == nil || got.Name != "file.iso" {
		t.Fatalf("Expected file.iso, got %v", got)
	}
}

func TestLayerFileDfsFirstFileMultipleSubs(t *testing.T) {
	leaf1 := &LayerFile{Name: "file1.iso", Type: "file"}
	leaf2 := &LayerFile{Name: "file2.iso", Type: "file"}
	parent := &LayerFile{Name: "dir", Sub: []*LayerFile{leaf1, leaf2}}
	got := parent.dfsFirstFile()
	if got == nil {
		t.Fatal("Expected non-nil")
	}
}

func TestLayerFileDfsEveryFileLeaf(t *testing.T) {
	leaf := &LayerFile{Name: "file.iso", Type: "file", ModTime: time.Now()}
	leaf.dfsEveryFile(nil)
}

func TestAppendParticularScenarioArch(t *testing.T) {
	list := []string{"ISO", "debuginfo"}
	result := appendParticularScenarioArch(list, "ISO")
	if len(result) != 2 {
		t.Fatalf("Expected 2 (no duplicate), got %d", len(result))
	}

	result = appendParticularScenarioArch(list, "newarch")
	if len(result) != 3 {
		t.Fatalf("Expected 3, got %d", len(result))
	}
}

func TestCheckRepoScenario(t *testing.T) {
	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"openEuler-22.03/ISO": {},
			"openEuler-22.03/debuginfo": {},
		},
	}
	result := checkRepoScenario("openEuler-22.03", []string{"ISO", "debuginfo", "other"})
	if len(result) != 2 {
		t.Fatalf("Expected 2, got %d", len(result))
	}
}

func TestCheckRepoArch(t *testing.T) {
	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"openEuler-22.03/ISO/x86_64": {},
			"openEuler-22.03/ISO/aarch64": {},
		},
	}
	result := checkRepoArch("openEuler-22.03", []string{"ISO"}, []string{"x86_64", "aarch64", "other"})
	if len(result) != 2 {
		t.Fatalf("Expected 2, got %d", len(result))
	}
}

func TestSelectEveryScenarioArchDir(t *testing.T) {
	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"openEuler-22.03/ISO/x86_64": {Dir: "openEuler-22.03/ISO/x86_64", Name: "x86_64"},
		},
	}
	result := selectEveryScenarioArchDir("openEuler-22.03", []string{"ISO"}, []string{"x86_64"})
	if len(result) != 1 {
		t.Fatalf("Expected 1, got %d", len(result))
	}
}

func TestSetRecentFile(t *testing.T) {
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, 100),
		SelectorMap: make(map[string]*LayerFile, 64),
	}
	ft := &LayerFile{Dir: "openEuler-22.03/ISO/x86_64", Name: "file.iso", ModTime: time.Now()}
	ft.setRecentFile()
	if v, ok := fileTreeReplica.SelectorMap["openEuler-22.03"]; !ok || v == nil {
		t.Fatal("Expected selector map to have entry")
	}

	ft2 := &LayerFile{Dir: "openEuler-22.03/ISO/x86_64", Name: "file2.iso", ModTime: time.Now().Add(time.Hour)}
	ft2.setRecentFile()
	if fileTreeReplica.SelectorMap["openEuler-22.03"].Name != "file2.iso" {
		t.Fatalf("Expected file2.iso to replace file.iso")
	}
}

func TestSetFileData(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-fs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	shaPath := filepath.Join(dir, "openEuler-22.03"+Sep+"test.txt.sha256sum")
	os.MkdirAll(filepath.Dir(shaPath), 0755)
	os.WriteFile(shaPath, []byte("abc123  test.txt"), 0644)

	cnf := &Configuration{
		Repository: dir,
	}

	ft := &LayerFile{Dir: "openEuler-22.03", Name: "test.txt"}
	size := []byte("         1,024")
	modTime := []byte("2024/08/08 11:01:29")
	fd := ft.setFileData("openEuler-22.03/test.txt", size, modTime, cnf)
	if fd == nil {
		t.Fatal("Expected non-nil FileData")
	}
	if fd.Path != "openEuler-22.03/test.txt" {
		t.Fatalf("Expected openEuler-22.03/test.txt, got %s", fd.Path)
	}
	if fd.Size != 1024 {
		t.Fatalf("Expected 1024, got %d", fd.Size)
	}
	if fd.Sha256 != "abc123" {
		t.Fatalf("Expected abc123, got %s", fd.Sha256)
	}
	if ft.Type != "file" {
		t.Fatalf("Expected file, got %s", ft.Type)
	}
}

func TestBuildFileTree(t *testing.T) {
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, 100),
		SelectorMap: make(map[string]*LayerFile, 64),
		Root:        LayerFile{},
	}
	cnf := &Configuration{Repository: "/tmp"}
	size := []byte("         1,024")
	modTime := []byte("2024/08/08 11:01:29")
	fd := BuildFileTree("openEuler-22.03/ISO/x86_64/file.iso", size, modTime, cnf)
	if fd == nil {
		t.Fatal("Expected non-nil FileData")
	}
	if fd.Path != "openEuler-22.03/ISO/x86_64/file.iso" {
		t.Fatalf("Unexpected path: %s", fd.Path)
	}
}

func TestBuildFileTreeNil(t *testing.T) {
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, 100),
		SelectorMap: make(map[string]*LayerFile, 64),
		Root:        LayerFile{},
	}
	cnf := &Configuration{Repository: "/tmp"}
	size := []byte("0")
	modTime := []byte("2024/08/08 11:01:29")
	fd := BuildFileTree("single", size, modTime, cnf)
	if fd != nil {
		t.Fatal("Expected nil for single-segment path")
	}
}

func TestUpdateFileTree(t *testing.T) {
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, 100),
		SelectorMap: make(map[string]*LayerFile, 64),
		Root:        LayerFile{},
	}
	cnf := &Configuration{Repository: "/tmp"}
	size := []byte("         1,024")
	modTime := []byte("2024/08/08 11:01:29")
	BuildFileTree("openEuler-22.03/ISO/x86_64/file.iso", size, modTime, cnf)

	UpdateFileTree(DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64"},
	})

	list := GetRepoVersionList()
	_ = list
}

func TestFileDataStruct(t *testing.T) {
	fd := FileData{
		Path:    "/test",
		Sha1:    "sha1",
		Sha256:  "sha256",
		Md5:     "md5",
		Size:    100,
		ModTime: time.Now(),
	}
	if fd.Path != "/test" || fd.Sha256 != "sha256" || fd.Size != 100 {
		t.Fatal("Fields mismatch")
	}
}

func TestLayerFileStruct(t *testing.T) {
	lf := LayerFile{
		Dir:    "/dir",
		Name:   "name",
		Size:   100,
		Sha256: "sha",
		Type:   "file",
	}
	if lf.Dir != "/dir" || lf.Name != "name" {
		t.Fatal("Fields mismatch")
	}
}

func TestDisplayFileStruct(t *testing.T) {
	df := DisplayFile{
		Name:    "test",
		Path:    "/test",
		Size:    "1 KB",
		ShaCode: "abc",
		Type:    "file",
	}
	if df.Name != "test" || df.Path != "/test" {
		t.Fatal("Fields mismatch")
	}
}

func TestDisplayFileListStruct(t *testing.T) {
	dfl := DisplayFileList{
		Scenario: "ISO",
		Arch:     "x86_64",
		Tree:     []DisplayFile{{Name: "test"}},
	}
	if dfl.Scenario != "ISO" || dfl.Arch != "x86_64" || len(dfl.Tree) != 1 {
		t.Fatal("Fields mismatch")
	}
}

func TestDisplayRepoVersionStruct(t *testing.T) {
	drv := DisplayRepoVersion{
		Version:  "22.03",
		Scenario: []string{"ISO"},
		Arch:     []string{"x86_64"},
		LTS:      true,
	}
	if drv.Version != "22.03" || !drv.LTS {
		t.Fatal("Fields mismatch")
	}
}

func TestFileStoreStruct(t *testing.T) {
	fs := FileStore{
		Mapping:     make(map[string]*LayerFile),
		SelectorMap: make(map[string]*LayerFile),
		Root:        LayerFile{},
	}
	if fs.Mapping == nil || fs.SelectorMap == nil {
		t.Fatal("Fields mismatch")
	}
}

func TestAppendParticularFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pf-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	shaPath := filepath.Join(dir, "test.iso.sha256sum")
	os.WriteFile(shaPath, []byte("abcdef  test.iso"), 0644)

	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"test.iso": {Size: 1024},
		},
	}

	d := &DisplayFileList{}
	p := ParticularFileMapping{
		SourcePath: []string{"test.iso"},
		SHA256List: []string{""},
	}
	d.appendParticularFile(p, dir, "http://fallback.com")
	if len(d.Tree) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(d.Tree))
	}
	if d.Tree[0].ShaCode != "abcdef" {
		t.Fatalf("Expected abcdef, got %s", d.Tree[0].ShaCode)
	}
}

func TestAppendParticularFileWithSHA256List(t *testing.T) {
	dir, err := os.MkdirTemp("", "mirrorbits-pf2-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"test.iso": {Size: 1024},
		},
	}

	d := &DisplayFileList{}
	p := ParticularFileMapping{
		SourcePath: []string{"test.iso"},
		SHA256List: []string{"preset_sha"},
	}
	d.appendParticularFile(p, dir, "http://fallback.com")
	if len(d.Tree) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(d.Tree))
	}
	if d.Tree[0].ShaCode != "preset_sha" {
		t.Fatalf("Expected preset_sha, got %s", d.Tree[0].ShaCode)
	}
}

func TestGetRepoFileListNonExistent(t *testing.T) {
	fileTree = &FileStore{
		Mapping: make(map[string]*LayerFile),
	}
	repoVersionMap = make(map[string][]DisplayFileList, 64)

	cnf := &Configuration{
		Repository:      "/tmp",
		RepositoryFilter: DirFilter{},
		Fallbacks:       []Fallback{{URL: "http://fallback.com"}},
	}

	result := GetRepoFileList("nonexistent", cnf)
	if result != nil {
		t.Fatalf("Expected nil for nonexistent version, got %v", result)
	}
}

func TestGetRepoFileListWithMapping(t *testing.T) {
	fileNode := &LayerFile{
		Dir:   "openEuler-22.03/ISO/x86_64",
		Name:  "file.iso",
		Size:  1024,
		Type:  "file",
	}
	archNode := &LayerFile{
		Dir:  "openEuler-22.03/ISO",
		Name: "x86_64",
		Sub:  []*LayerFile{fileNode},
	}
	scenarioNode := &LayerFile{
		Dir:  "openEuler-22.03",
		Name: "ISO",
		Sub:  []*LayerFile{archNode},
	}
	versionNode := &LayerFile{
		Name: "openEuler-22.03",
		Sub:  []*LayerFile{scenarioNode},
	}
	fileTree = &FileStore{
		Mapping: map[string]*LayerFile{
			"openEuler-22.03": versionNode,
		},
	}
	repoVersionMap = make(map[string][]DisplayFileList, 64)

	cnf := &Configuration{
		Repository:       "/tmp",
		RepositoryFilter: DirFilter{},
		Fallbacks:        []Fallback{{URL: "http://fallback.com"}},
	}

	result := GetRepoFileList("openEuler-22.03", cnf)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}
