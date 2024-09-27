// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package filesystem

import (
	"github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/utils"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// path separator
	Sep = "/"
	// file extension named sha256sum
	FileExtensionSha256 = ".sha256sum"
	// estimate indexed file sum
	EstimateFileNum = 10000
	// openEuler version initial total
	RepoVersionNum = 64
	// openEuler version directory prefix value
	RepoVersionDirectoryPrefix = "openEuler-"
	// Scenario is ISO or edge_img, standard iso file extension
	StandardISOFileExtension = ".iso"
)

var (
	fileTree = &FileStore{
		Mapping:     make(map[string]*LayerFile, EstimateFileNum),
		SelectorMap: make(map[string]*LayerFile, RepoVersionNum),
		Root:        LayerFile{},
	}
	pathFilter      []string
	repoVersionList []DisplayRepoVersion
	repoVersionMap  = make(map[string][]DisplayFileList, RepoVersionNum)
	selectorList    []*LayerFile

	lock            sync.RWMutex
	fileTreeReplica = &FileStore{
		Mapping:     make(map[string]*LayerFile, EstimateFileNum),
		SelectorMap: make(map[string]*LayerFile, RepoVersionNum),
		Root:        LayerFile{},
	}
)

// website display file information structure
type DisplayFile struct {
	Name    string
	Path    string
	Size    string
	ShaCode string
	Type    string
}

// website display file menu structure
type DisplayFileList struct {
	Scenario string
	Arch     string
	Tree     []DisplayFile
}

// website display repo version menu structure
type DisplayRepoVersion struct {
	Version  string
	Scenario []string
	Arch     []string
	LTS      bool
}

// store the files structure
// Root contains tree-structured file information
// Mapping is a map contains per file information
// SelectorMap is a map contains some files that supports to use for mirror check
type FileStore struct {
	Mapping     map[string]*LayerFile
	SelectorMap map[string]*LayerFile
	Root        LayerFile
}

// file information structure in project
type FileData struct {
	Path    string
	Sha1    string
	Sha256  string
	Md5     string
	Size    int64
	ModTime time.Time
}

// tree-structured file information structure in project
type LayerFile struct {
	Dir     string
	Name    string
	Size    int64
	Sha256  string
	Type    string
	ModTime time.Time
	Sub     []*LayerFile
}

// update the tree-structured files
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

// a file append to the tree-structured files, and return the file information
func BuildFileTree(path string, size, modTime []byte, cnf *config.Configuration) *FileData {
	return fileTreeReplica.Root.layeringPath(path, size, modTime, cnf)
}

// get the website displayed repo version list
func GetRepoVersionList() []DisplayRepoVersion {
	var ans []DisplayRepoVersion
	lock.RLock()
	ans = repoVersionList
	lock.RUnlock()
	return ans
}

// get a file list that supports to use for mirror check
func GetSelectorList() []*LayerFile {
	var ans []*LayerFile
	lock.RLock()
	ans = selectorList
	lock.RUnlock()
	return ans
}

// get a repo version file list to display in website
func GetRepoFileList(version string, cnf *config.Configuration) []DisplayFileList {
	lock.RLock()
	defer lock.RUnlock()
	if v, ok := repoVersionMap[version]; ok {
		return v
	}

	if p, ok := fileTree.Mapping[version]; ok {
		repoVersionMap[version] = p.flattening()
	}

	// x86-64 convert to x86_64
	for i, v := range repoVersionMap[version] {
		if v.Arch == "x86-64" {
			repoVersionMap[version][i].Arch = "x86_64"
		}
	}

	// handler some particular file that information loading by the config file
	for _, v := range cnf.RepositoryFilter.ParticularFile {
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
				ans.appendParticularFile(v, cnf.Repository, cnf.Fallbacks[0].URL)
				repoVersionMap[version] = append(repoVersionMap[version], *ans)
			} else {
				ans := &repoVersionMap[version][idx]
				ans.appendParticularFile(v, cnf.Repository, cnf.Fallbacks[0].URL)
				repoVersionMap[version][idx] = *ans
			}
		}
	}
	return repoVersionMap[version]
}

// build a rule list that use for filter the repo source files
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

// filtering by the file path
func Filter(path string) bool {
	if !strings.HasPrefix(path, RepoVersionDirectoryPrefix) {
		return false
	}
	if strings.HasSuffix(path, FileExtensionSha256) {
		return false
	}
	for _, v := range pathFilter {
		if strings.Contains(path, v) {
			arr := strings.Split(path, Sep)
			if len(arr) > 1 && (arr[1] == "ISO" || arr[1] == "edge_img") && !strings.HasSuffix(path, StandardISOFileExtension) {
				return false
			}
			return true
		}
	}
	return false
}

// add the configured particular file to the website display file menu
func (d *DisplayFileList) appendParticularFile(p config.ParticularFileMapping, repoPath, fallbackPath string) {
	for i, v := range p.SourcePath {
		path := v
		pathArr := strings.Split(path, Sep)
		viewSize := ""
		size, ok := fileTree.Mapping[path]
		if ok {
			viewSize = utils.ReadableSize(size.Size)
		}
		shaCode := ""
		if i < len(p.SHA256List) {
			if p.SHA256List[i] != "" {
				shaCode = p.SHA256List[i]
			} else {
				sha, err := os.ReadFile(repoPath + Sep + path + FileExtensionSha256)
				if err == nil {
					shaCode = strings.Split(string(sha), " ")[0]
				}
			}
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

// per repo version, do record the recent file information
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

// file size: string [rrwxrr-x  23,545,123] => int64 [23545123]
func covertFileSize(arr []byte) int64 {
	n := len(arr)
	size := make([]byte, n, n)
	idx := n - 1
	for i := idx; i >= 0; i-- {
		if arr[i] == ' ' {
			idx++
			break
		}
		if arr[i] != ',' {
			size[idx] = arr[i]
			idx--
		}
	}
	sizeInt, _ := strconv.ParseInt(string(size[idx:]), 10, 64)
	return sizeInt
}

// set the file information
func (ft *LayerFile) setFileData(path string, size, modTime []byte, cnf *config.Configuration) *FileData {

	fd := new(FileData)
	ft.Type = "file"
	fd.Path = path
	ft.Size = covertFileSize(size)
	fd.Size = ft.Size
	modTime[4] = '-'
	modTime[7] = '-'
	lastModTime, err := time.Parse(time.DateTime, string(modTime))
	if err == nil {
		ft.ModTime = lastModTime
		fd.ModTime = lastModTime
		ft.setRecentFile()
	}
	sha256FilePath := strings.ReplaceAll(utils.ConcatURL(cnf.Repository, path), Sep, string(os.PathSeparator)) + FileExtensionSha256
	data, err1 := os.ReadFile(sha256FilePath)
	if err1 == nil {
		ft.Sha256 = strings.Split(string(data), " ")[0]
		fd.Sha256 = ft.Sha256
	}
	return fd
}

// a file append to the tree-structured files, and return the file information
func (ft *LayerFile) layeringPath(path string, size, modTime []byte, cnf *config.Configuration) *FileData {
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
				fd = node.setFileData(currPath, size, modTime, cnf)
			}
			if _, ok1 := fm[currPath]; !ok1 {
				fm[currPath] = node
				dp.Sub = append(dp.Sub, node)
			}
		}
	}
	return fd
}

// tree-structured files structure covert into flat files list to display on the website page
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

// collect every file from a tree-structured files node
func (ft *LayerFile) collectFileInfo(ans []DisplayFileList, scenario string) []DisplayFileList {
	if len(ft.Sub) == 0 {
		return ans
	}

	var t []DisplayFile
	arr := strings.Split(ft.Dir, Sep)
	if len(arr) == 2 && arr[1] == "embedded_img" && len(ft.Sub) == 1 {
		for _, p := range ft.Sub[0].Sub {
			t = p.appendFile(len(p.Sub) == 0, t)
			t = p.appendDir(len(p.Sub) != 0, t)
		}
	} else {
		for _, p := range ft.Sub {
			t = p.appendFile(len(p.Sub) == 0, t)
			t = p.appendDir(len(p.Sub) != 0, t)
		}
	}
	return append(ans, DisplayFileList{
		Scenario: scenario,
		Arch:     ft.Name,
		Tree:     t,
	})
}

func (ft *LayerFile) appendFile(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}
	path := ft.Dir + Sep + ft.Name
	return append(t, DisplayFile{
		Name:    ft.Name,
		Path:    path,
		Size:    utils.ReadableSize(ft.Size),
		ShaCode: ft.Sha256,
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

func (ft *LayerFile) dfsEveryFile() {
	if ft == nil {
		return
	}

	subLen := len(ft.Sub)
	if subLen == 0 {
		selectorList = append(selectorList, ft)
		return
	}

	if subLen == 1 {
		ft.Sub[0].dfsEveryFile()
		return
	}

	for _, p := range ft.Sub {
		p.dfsEveryFile()
	}
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
			var editArch []string
			// x86-64 merge to x86_64
			editArch = append(editArch, arch...)
			x, y := -1, -1
			for i, s := range editArch {
				if s == "x86_64" {
					x = i
				}
				if s == "x86-64" {
					y = i
				}
			}
			if x == -1 && y != -1 {
				editArch[y] = "x86_64"
			}
			if x != -1 && y != -1 {
				editArch[y] = editArch[0]
				editArch = editArch[1:]
			}
			sort.Strings(editArch)

			// collect all repo version menu list
			repoVersionList = append(repoVersionList, DisplayRepoVersion{
				Version:  v.Name,
				Scenario: scenario,
				Arch:     editArch,
				LTS:      strings.Contains(v.Name, "LTS"),
			})

			// select some files to do check mirror
			p := fileTree.SelectorMap[v.Name]
			if p == nil {
				continue
			}
			selectorList = append(selectorList, p)
			selectDir := selectEveryScenarioArchDir(v.Name, scenario, arch)
			if p.ModTime.After(time.Now().AddDate(0, -7, 0)) {
				// the long-term maintenance repo version, select every file to check exist or not in the mirror website
				for _, p1 := range selectDir {
					p1.dfsEveryFile()
				}
			} else {
				// the stopped maintenance repo version, select some file to check exist or not in the mirror website
				for _, p2 := range selectDir {
					if selectFile := p2.dfsDictionaryOrderLastFile(); selectFile != nil {
						selectorList = append(selectorList, selectFile)
					}
				}
			}
		}
	}

	for i, v := range repoVersionList {
		// additional file information
		for _, v1 := range filter.ParticularFile {
			if v1.VersionName == v.Version {
				repoVersionList[i].Scenario = appendParticularScenarioArch(repoVersionList[i].Scenario, v1.ScenarioName)
				repoVersionList[i].Arch = appendParticularScenarioArch(repoVersionList[i].Arch, v1.ArchName)
			}
		}
	}
	sort.SliceStable(repoVersionList, func(i, j int) bool {
		return repoVersionList[i].Version < repoVersionList[j].Version
	})
}

func appendParticularScenarioArch(list []string, value string) []string {
	flag := true
	for _, v := range list {
		if v == value {
			flag = false
			break
		}
	}
	if flag {
		list = append(list, value)
	}
	return list
}

func selectEveryScenarioArchDir(version string, scenario, arch []string) (ans []*LayerFile) {
	for _, v := range scenario {
		for j := len(arch) - 1; j >= 0; j-- {
			if p, ok := fileTree.Mapping[version+Sep+v+Sep+arch[j]]; ok {
				ans = append(ans, p)
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
