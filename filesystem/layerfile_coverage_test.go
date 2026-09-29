// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package filesystem

import (
	"testing"
	"time"

	"github.com/opensourceways/mirrorbits/config"
)

func saveFileTree() *FileStore {
	return &FileStore{
		Mapping:     fileTree.Mapping,
		SelectorMap: fileTree.SelectorMap,
		Root:        fileTree.Root,
	}
}

func restoreFileTree(ft *FileStore) {
	fileTree.Mapping = ft.Mapping
	fileTree.SelectorMap = ft.SelectorMap
	fileTree.Root = ft.Root
}

func TestCheckRepoScenario_Match(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Mapping["v1/ISO"] = &LayerFile{Dir: "v1/ISO", Name: "x86_64"}
	fileTree.Mapping["v1/RPM"] = &LayerFile{Dir: "v1/RPM", Name: "aarch64"}

	result := checkRepoScenario("v1", []string{"ISO", "RPM", "OTHER"})
	if len(result) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(result))
	}
}

func TestCheckRepoScenario_NoMatch(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	result := checkRepoScenario("v1", []string{"ISO"})
	if len(result) != 0 {
		t.Fatalf("Expected 0 matches")
	}
}

func TestCheckRepoArch_Match(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Mapping["v1/ISO/x86_64"] = &LayerFile{Dir: "v1/ISO/x86_64"}
	fileTree.Mapping["v1/ISO/aarch64"] = &LayerFile{Dir: "v1/ISO/aarch64"}

	result := checkRepoArch("v1", []string{"ISO"}, []string{"x86_64", "aarch64", "ppc64le"})
	if len(result) != 2 {
		t.Fatalf("Expected 2 arch matches, got %d", len(result))
	}
}

func TestCheckRepoArch_NoMatch(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	result := checkRepoArch("v1", []string{"ISO"}, []string{"x86_64"})
	if len(result) != 0 {
		t.Fatalf("Expected 0 matches")
	}
}

func TestSelectEveryScenarioArchDir(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	lf1 := &LayerFile{Dir: "v1/ISO/x86_64", Name: "file.iso"}
	lf2 := &LayerFile{Dir: "v1/RPM/aarch64", Name: "file.rpm"}
	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Mapping["v1/ISO/x86_64"] = lf1
	fileTree.Mapping["v1/RPM/aarch64"] = lf2

	result := selectEveryScenarioArchDir("v1", []string{"ISO", "RPM"}, []string{"x86_64", "aarch64"})
	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}
}

func TestSelectEveryScenarioArchDir_Empty(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	result := selectEveryScenarioArchDir("v1", []string{"ISO"}, []string{"x86_64"})
	if len(result) != 0 {
		t.Fatalf("Expected 0 results")
	}
}

func TestLayerFile_Flattening_WithSubs(t *testing.T) {
	scenarioNode := &LayerFile{Dir: "v1", Name: "ISO", Sub: []*LayerFile{
		{Dir: "v1/ISO", Name: "x86_64", Sub: []*LayerFile{
			{Dir: "v1/ISO/x86_64", Name: "file.iso", Size: 100},
		}},
	}}
	root := &LayerFile{Dir: "", Name: "v1", Sub: []*LayerFile{scenarioNode}}

	result := root.flattening()
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
}

func TestLayerFile_CollectFileInfo_Leaf(t *testing.T) {
	lf := &LayerFile{Dir: "v1/ISO", Name: "x86_64"}
	result := lf.collectFileInfo(nil, "ISO")
	if result != nil {
		t.Fatalf("Expected nil for leaf without Sub")
	}
}

func TestLayerFile_CollectFileInfo_WithSubs(t *testing.T) {
	lf := &LayerFile{
		Dir:  "v1/ISO",
		Name: "x86_64",
		Sub: []*LayerFile{
			{Dir: "v1/ISO/x86_64", Name: "file.iso", Size: 100},
		},
	}
	result := lf.collectFileInfo(nil, "ISO")
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
}

func TestLayerFile_CollectFileInfo_EmbeddedImg(t *testing.T) {
	lf := &LayerFile{
		Dir:  "v1/embedded_img",
		Name: "x86_64",
		Sub: []*LayerFile{
			{Dir: "v1/embedded_img/x86_64", Name: "inner", Sub: []*LayerFile{
				{Dir: "v1/embedded_img/x86_64/inner", Name: "file.img", Size: 500},
			}},
		},
	}
	result := lf.collectFileInfo(nil, "embedded_img")
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	if len(result[0].Tree) != 1 {
		t.Fatalf("Expected 1 tree item")
	}
}

func TestLayerFile_DfsEveryFile_WithSubs(t *testing.T) {
	leaf1 := &LayerFile{Dir: "v1/ISO/x86_64", Name: "file1.iso", ModTime: time.Now()}
	leaf2 := &LayerFile{Dir: "v1/ISO/aarch64", Name: "file2.iso", ModTime: time.Now().Add(1 * time.Hour)}
	parent := &LayerFile{Dir: "v1/ISO", Name: "ISO", Sub: []*LayerFile{leaf1, leaf2}}

	parent.dfsEveryFile(nil)
}

func TestLayerFile_DfsEveryFile_SingleChild(t *testing.T) {
	leaf := &LayerFile{Dir: "v1/ISO/x86_64", Name: "file.iso", ModTime: time.Now()}
	parent := &LayerFile{Dir: "v1/ISO", Name: "ISO", Sub: []*LayerFile{leaf}}

	parent.dfsEveryFile(nil)
}

func TestLayerFile_DfsEveryFile_FindLatest(t *testing.T) {
	older := &LayerFile{Dir: "d", Name: "old.iso", ModTime: time.Now().Add(-2 * time.Hour)}
	newer := &LayerFile{Dir: "d", Name: "new.iso", ModTime: time.Now()}
	parent := &LayerFile{Dir: "", Name: "root", Sub: []*LayerFile{older, newer}}

	target := &LayerFile{ModTime: time.Now().Add(-1 * time.Hour)}
	parent.dfsEveryFile(target)
}

func TestLayerFile_DfsFirstFile_SingleChild(t *testing.T) {
	leaf := &LayerFile{Dir: "d", Name: "file.iso"}
	parent := &LayerFile{Dir: "", Name: "root", Sub: []*LayerFile{leaf}}
	result := parent.dfsFirstFile()
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
}

func TestLayerFile_DfsFirstFile_MultipleChildren(t *testing.T) {
	leaf1 := &LayerFile{Dir: "d", Name: "file1.sha256sum"}
	leaf2 := &LayerFile{Dir: "d", Name: "file2.iso"}
	parent := &LayerFile{Dir: "", Name: "root", Sub: []*LayerFile{leaf1, leaf2}}
	result := parent.dfsFirstFile()
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}
	if result.Name != "file2.iso" {
		t.Fatalf("Expected file2.iso, got %s", result.Name)
	}
}

func TestCollectRepoVersionList_WithRoot(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Mapping["v1/ISO"] = &LayerFile{Dir: "v1/ISO"}
	fileTree.Mapping["v1/ISO/x86_64"] = &LayerFile{Dir: "v1/ISO/x86_64"}
	fileTree.Mapping["v1/ISO/aarch64"] = &LayerFile{Dir: "v1/ISO/aarch64"}
	fileTree.Root = LayerFile{
		Sub: []*LayerFile{
			{Dir: "", Name: "v1"},
		},
	}

	repoVersionList = repoVersionList[0:0]
	collectRepoVersionList(config.DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64", "aarch64"},
	})

	if len(repoVersionList) == 0 {
		t.Fatalf("Expected at least 1 version")
	}
}

func TestCollectRepoVersionList_EmptyRoot(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Root = LayerFile{}

	repoVersionList = repoVersionList[0:0]
	collectRepoVersionList(config.DirFilter{
		SecondDir: []string{"ISO"},
		ThirdDir:  []string{"x86_64"},
	})
	if len(repoVersionList) != 0 {
		t.Fatalf("Expected 0 versions for empty root")
	}
}

func TestAppendParticularFile_WithSHA(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Mapping["v1/ISO/file.iso"] = &LayerFile{Dir: "v1/ISO", Name: "file.iso", Size: 1024}

	d := &DisplayFileList{}
	p := config.ParticularFileMapping{
		SourcePath:  []string{"v1/ISO/file.iso"},
		SHA256List:  []string{"abc123"},
	}
	d.appendParticularFile(p, "/repo", "/fallback")
	if len(d.Tree) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(d.Tree))
	}
	if d.Tree[0].ShaCode != "abc123" {
		t.Fatalf("Expected sha 'abc123'")
	}
}

func TestAppendParticularFile_NoSHA(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)

	d := &DisplayFileList{}
	p := config.ParticularFileMapping{
		SourcePath:  []string{"v1/ISO/file.iso"},
		SHA256List:  []string{""},
	}
	d.appendParticularFile(p, "/nonexistent", "/nonexistent")
	if len(d.Tree) != 1 {
		t.Fatalf("Expected 1 file entry even without sha")
	}
}

func TestGetRepoFileList_WithMapping(t *testing.T) {
	orig := saveFileTree()
	defer restoreFileTree(orig)

	fileTree.Mapping = make(map[string]*LayerFile)
	fileTree.Root = LayerFile{
		Sub: []*LayerFile{
			{Dir: "", Name: "v1"},
		},
	}
	fileTree.Mapping["v1/ISO"] = &LayerFile{Dir: "v1/ISO"}
	fileTree.Mapping["v1/ISO/x86_64"] = &LayerFile{Dir: "v1/ISO/x86_64"}

	cnf := &config.Configuration{
		Repository: "/tmp",
		Fallbacks:  []config.Fallback{{URL: "http://fallback.com"}},
		RepositoryFilter: config.DirFilter{
			SecondDir: []string{"ISO"},
			ThirdDir:  []string{"x86_64"},
		},
	}
	result := GetRepoFileList("v1", cnf)
	_ = result
}
