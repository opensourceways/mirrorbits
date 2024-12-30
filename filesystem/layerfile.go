// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package filesystem

import (
	"github.com/opensourceways/mirrorbits/config"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Sep                 = string(os.PathSeparator)
	FileExtensionSha256 = ".sha256sum"
)

var (
	fileTree = &FileStore{
		Mapping:     make(map[string]*LayerFile, 1024*10),
		SelectorMap: make(map[string]*LayerFile, 64),
		Root:        LayerFile{},
	}
	pathFilter      []string
	repoVersionList []DisplayRepoVersion
	repoVersionMap  = make(map[string][]DisplayFileList, 64)
	selectorList    []LayerFile

	lock            sync.RWMutex
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, 1024*10),
		SelectorMap: make(map[string]*LayerFile, 64),
		Root:        LayerFile{},
	}
)

type DisplayFile struct {
	Name    string
	Path    string
	Size    string
	ShaCode string
	Type    string
}

type DisplayFileList struct {
	Scenario string
	Arch     string
	Tree     []DisplayFile
}

type DisplayRepoVersion struct {
	Version  string
	Scenario []string
	Arch     []string
	LTS      bool
}

type FileStore struct {
	Mapping     map[string]*LayerFile
	SelectorMap map[string]*LayerFile
	Root        LayerFile
}

type FileData struct {
	Path    string
	Sha1    string
	Sha256  string
	Md5     string
	Size    int64
	ModTime time.Time
}

type LayerFile struct {
	Dir     string
	Name    string
	Size    int64
	Sha256  string
	Type    string
	ModTime time.Time
	Sub     []*LayerFile
}

func UpdateFileTree(filter config.DirFilter) {
	lock.Lock()
	defer lock.Unlock()

	fileTree = fileTreeReplica
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, len(fileTree.Mapping)<<1),
		SelectorMap: make(map[string]*LayerFile, len(fileTree.SelectorMap)<<1),
		Root:        LayerFile{},
	}
	repoVersionList = repoVersionList[0:0]
	selectorList = selectorList[0:0]
	collectRepoVersionList(filter)

}

func BuildFileTree(path string, cnf *config.Configuration) *FileData {
	return fileTreeReplica.Root.layeringPath(path, cnf)
}

func GetRepoVersionList() (ans []DisplayRepoVersion) {
	lock.RLock()
	ans = repoVersionList
	lock.RUnlock()
	return
}

func GetSelectorList() (ans []LayerFile) {
	lock.RLock()
	ans = selectorList
	lock.RUnlock()
	return
}

func GetRepoFileList(version string, filter config.DirFilter) []DisplayFileList {
	lock.RLock()
	defer lock.RUnlock()
	if v, ok := repoVersionMap[version]; ok {
		return v
	}

	if p, ok := fileTree.Mapping[version]; ok {
		repoVersionMap[version] = p.flattening()
	}

	for _, v := range filter.ParticularFile {
		if v.VersionName == version {
			idx := -1
			for i := range repoVersionMap[version] {
				if repoVersionMap[version][i].Scenario == v.ScenarioName && repoVersionMap[version][i].Arch == v.ArchName {
					idx = i
					break
				}
			}
			if idx == -1 {
				ans := &DisplayFileList{
					Scenario: v.ScenarioName,
					Arch:     v.ArchName,
				}
				appendParticularFile(ans, v)
				repoVersionMap[version] = append(repoVersionMap[version], *ans)
			} else {
				ans := &repoVersionMap[version][idx]
				appendParticularFile(ans, v)
				repoVersionMap[version][idx] = *ans
			}
		}
	}
	return repoVersionMap[version]
}

func InitPathFilter(filter config.DirFilter) {
	if len(filter.SecondDir) == 0 || len(filter.ThirdDir) == 0 {
		return
	}

	for _, v1 := range filter.SecondDir {
		for _, v2 := range filter.ThirdDir {
			pathFilter = append(pathFilter, v1+Sep+v2)
		}
	}

	for _, v3 := range filter.ParticularFile {
		for _, v4 := range v3.SourcePath {
			pathFilter = append(pathFilter, v4)
		}
	}
}

func Filter(path string) bool {
	if path[:10] != "openEuler-" {
		return false
	}
	if strings.HasSuffix(path, FileExtensionSha256) {
		return false
	}
	for _, v := range pathFilter {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}

func appendParticularFile(d *DisplayFileList, p config.ParticularFileMapping) {
	for _, v := range p.SourcePath {
		path := v
		pathArr := strings.Split(path, Sep)
		viewSize := ""
		size, ok := fileTree.Mapping[path]
		if ok {
			viewSize = size.viewFileSize()
		}
		shaCode := ""
		sha, ok := fileTree.Mapping[path+FileExtensionSha256]
		if ok {
			shaCode = strings.Split(sha.Sha256, " ")[0]
		}
		d.Tree = append(d.Tree, DisplayFile{
			Name:    pathArr[len(pathArr)-1],
			Path:    path,
			Size:    viewSize,
			ShaCode: shaCode,
			Type:    "file",
		})
	}
}

func (ft *LayerFile) setRecentFile() {
	version := strings.Split(ft.Dir, Sep)[0]
	if v, ok := fileTreeReplica.SelectorMap[version]; ok {
		if !ft.ModTime.Before(v.ModTime) {
			fileTreeReplica.SelectorMap[version] = ft
		}
	} else {
		fileTreeReplica.SelectorMap[version] = ft
	}
}

func (ft *LayerFile) setFileData(fullPath string, cnf *config.Configuration) *FileData {

	relPath, err := filepath.Abs(cnf.Repository + Sep + fullPath)
	if err != nil {
		return nil
	} else {

		fd := new(FileData)
		stat, err1 := os.Stat(relPath)
		if err1 == nil {
			ft.Size = stat.Size()
			ft.Type = "file"
			ft.ModTime = stat.ModTime()
			fd.Path = strings.ReplaceAll(fullPath, Sep, "/")
			fd.Size = stat.Size()
			fd.ModTime = stat.ModTime()
			ft.setRecentFile()

			data, err2 := os.ReadFile(relPath + FileExtensionSha256)
			if err2 == nil {
				ft.Sha256 = strings.Split(string(data), " ")[0]
				fd.Sha256 = ft.Sha256
			}
		}
		return fd
	}
}

func (ft *LayerFile) layeringPath(path string, cnf *config.Configuration) *FileData {
	fm := fileTreeReplica.Mapping
	var fd *FileData
	fileLayer := strings.Split(path, Sep)
	layerLength := len(fileLayer)
	if _, ok := fm[fileLayer[0]]; !ok {
		node := &LayerFile{
			Name: fileLayer[0],
		}
		ft.Sub = append(ft.Sub, node)
		fm[fileLayer[0]] = node
	}
	for i := 1; i < layerLength; i++ {
		dir := strings.Join(fileLayer[:i], Sep)
		if dp, ok := fm[dir]; ok {
			node := &LayerFile{
				Dir:  dir,
				Name: fileLayer[i],
			}
			currPath := dir + Sep + node.Name
			if currPath == path {
				fd = node.setFileData(currPath, cnf)
			}
			if _, ok1 := fm[currPath]; !ok1 {
				fm[currPath] = node
				dp.Sub = append(dp.Sub, node)
			}
		}
	}
	return fd
}

func (ft *LayerFile) flattening() []DisplayFileList {
	if ft == nil || len(ft.Sub) == 0 {
		return nil
	}

	var ans []DisplayFileList

	for _, p1 := range ft.Sub {
		if len(p1.Sub) != 0 {
			for _, p2 := range p1.Sub {
				ans = p2.collectFileInfo(ans, p1.Name)
			}
		}
	}
	return ans
}

func (ft *LayerFile) collectFileInfo(ans []DisplayFileList, scenario string) []DisplayFileList {
	if len(ft.Sub) == 0 {
		return ans
	}

	var t []DisplayFile
	arr := strings.Split(ft.Dir, Sep)
	if len(arr) == 2 && arr[1] == "embedded_img" && len(ft.Sub) == 1 {
		for _, p := range ft.Sub[0].Sub {
			t = p.appendFile(!strings.HasSuffix(p.Name, FileExtensionSha256), t)
		}
	} else {
		for _, p := range ft.Sub {
			t = p.appendFile(len(p.Sub) == 0 && !strings.HasSuffix(p.Name, FileExtensionSha256), t)
			t = p.appendDir(len(p.Sub) != 0, t)
		}
	}
	return append(ans, DisplayFileList{
		Scenario: scenario,
		Arch:     ft.Name,
		Tree:     t,
	})
}

func (ft *LayerFile) viewFileSize() (ans string) {
	if ft.Size <= 1024 {
		ans = strconv.FormatInt(ft.Size, 10) + " B"
	} else if ft.Size <= 1024*1024 {
		ans = strconv.FormatInt(ft.Size>>10, 10) + "." + strconv.FormatInt((ft.Size>>9)%10, 10) + " KiB"
	} else if ft.Size <= 1024*1024*1024 {
		ans = strconv.FormatInt(ft.Size>>20, 10) + "." + strconv.FormatInt((ft.Size>>19)%10, 10) + " MiB"
	} else {
		ans = strconv.FormatInt(ft.Size>>30, 10) + "." + strconv.FormatInt((ft.Size>>29)%10, 10) + " GiB"
	}
	return
}

func (ft *LayerFile) appendFile(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}
	path := ft.Dir + Sep + ft.Name
	shaCode := ""
	sha, ok := fileTree.Mapping[path+FileExtensionSha256]
	if ok {
		shaCode = strings.Split(sha.Sha256, " ")[0]
	}
	return append(t, DisplayFile{
		Name:    ft.Name,
		Path:    path,
		Size:    ft.viewFileSize(),
		ShaCode: shaCode,
		Type:    "file",
	})
}

func (ft *LayerFile) appendDir(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}

	return append(t, DisplayFile{
		Name: ft.Name,
		Path: ft.Dir + Sep + ft.Name,
		Type: "dir",
	})
}

func (ft *LayerFile) dfsDictionaryOrderLastFile() *LayerFile {
	if ft == nil {
		return nil
	}
	subLen := len(ft.Sub)
	if subLen == 0 {
		return ft
	}
	if subLen == 1 {
		return ft.Sub[0].dfsDictionaryOrderLastFile()
	}

	sort.SliceStable(ft.Sub, func(i, j int) bool {
		return ft.Sub[i].Name > ft.Sub[j].Name
	})

	for _, p := range ft.Sub {
		if !strings.HasSuffix(p.Name, FileExtensionSha256) {
			if ans := p.dfsDictionaryOrderLastFile(); ans != nil {
				return ans
			}
		}
	}
	return nil
}

func collectRepoVersionList(filter config.DirFilter) {
	if len(fileTree.Root.Sub) == 0 || len(filter.SecondDir) == 0 || len(filter.ThirdDir) == 0 {
		return
	}
	for _, v := range fileTree.Root.Sub {
		scenario := checkRepoScenario(v.Name, filter.SecondDir)
		arch := checkRepoArch(v.Name, scenario, filter.ThirdDir)
		if len(arch) > 0 {
			sort.Strings(scenario)
			sort.Strings(arch)
			repoVersionList = append(repoVersionList, DisplayRepoVersion{
				Version:  v.Name,
				Scenario: scenario,
				Arch:     arch,
				LTS:      strings.Contains(v.Name, "LTS"),
			})

			var selectDir []*LayerFile
			p := fileTree.SelectorMap[v.Name]
			if p.ModTime.After(time.Now().AddDate(-1, 0, 0)) {
				selectDir = selectEveryScenarioDir(v.Name, scenario, arch)
				selectorList = append(selectorList, *p)
			} else {
				selectDir = selectDictionaryOrderLastDir(v.Name, scenario, arch)
			}
			for _, p1 := range selectDir {
				if selectFile := p1.dfsDictionaryOrderLastFile(); selectFile != nil {

					selectorList = append(selectorList, *p)
				}
			}
		}
	}
	sort.SliceStable(repoVersionList, func(i, j int) bool {
		return repoVersionList[i].Version < repoVersionList[j].Version
	})
}

func selectDictionaryOrderLastDir(version string, scenario, arch []string) []*LayerFile {
	for i := len(scenario) - 1; i >= 0; i-- {
		for j := len(arch) - 1; j >= 0; j-- {
			if p, ok := fileTree.Mapping[version+Sep+scenario[i]+Sep+arch[j]]; ok {
				return []*LayerFile{p}
			}
		}
	}
	return nil
}

func selectEveryScenarioDir(version string, scenario, arch []string) (ans []*LayerFile) {
	for _, v := range scenario {
		for j := len(arch) - 1; j >= 0; j-- {
			if p, ok := fileTree.Mapping[version+Sep+v+Sep+arch[j]]; ok {
				ans = append(ans, p)
				break
			}
		}
	}
	return
}

func checkRepoScenario(versionName string, filter []string) []string {
	var scenario []string
	for _, v := range filter {
		if _, ok := fileTree.Mapping[versionName+Sep+v]; ok {
			scenario = append(scenario, v)
		}
	}
	return scenario
}

func checkRepoArch(versionName string, scenario, filter []string) []string {
	var arch []string
	for _, v := range filter {
		for _, v1 := range scenario {
			if _, ok := fileTree.Mapping[versionName+Sep+v1+Sep+v]; ok {
				arch = append(arch, v)
				break
			}
		}
	}
	return arch
}
