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
	"time"
)

const Sep = string(os.PathSeparator)

type fileStore struct {
	Mapping map[string]*LayerFile
	Root    LayerFile
}

var FileTree = fileStore{
	Root:    LayerFile{},
	Mapping: make(map[string]*LayerFile, 1024*10),
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
	Dir    string
	Name   string
	Size   string
	Sha256 string
	Type   string
	Sub    []*LayerFile
}

func (ft *LayerFile) ViewFileSize(size int64) {
	if size <= 1024 {
		ft.Size = strconv.FormatInt(size, 10) + " Bytes"
	} else if size <= 1024*1024 {
		ft.Size = strconv.FormatInt(size>>10, 10) + " K"
	} else if size <= 1024*1024*1024 {
		ft.Size = strconv.FormatInt(size>>20, 10) + " M"
	} else {
		ft.Size = strconv.FormatInt(size>>30, 10) + " GiB"
	}

}

func (ft *LayerFile) SetFileData(fullPath string, cnf *config.Configuration) *FileData {

	relPath, err := filepath.Abs(cnf.Repository + Sep + fullPath)
	if err != nil {
		return nil
	} else {

		fd := new(FileData)
		stat, err1 := os.Stat(relPath)
		if err1 == nil {
			ft.ViewFileSize(stat.Size())
			ft.Type = "file"
			fd.Path = fullPath
			fd.Size = stat.Size()
			fd.ModTime = stat.ModTime()
		}
		for _, ext := range cnf.DBExcludeFileExtension {
			if strings.HasSuffix(relPath, ext) {
				data, err2 := os.ReadFile(relPath)
				if err2 == nil {
					ft.Sha256 = strings.Split(string(data), " ")[0]
				}
				fd = nil
				break
			}
		}
		return fd
	}
}

func (ft *LayerFile) LayeringPath(path string, cnf *config.Configuration) *FileData {
	fm := FileTree.Mapping
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
				fd = node.SetFileData(currPath, cnf)
			}
			if _, ok1 := fm[currPath]; !ok1 {
				fm[currPath] = node
				dp.Sub = append(dp.Sub, node)
			}
		}
	}
	return fd
}

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

func (ft *LayerFile) Flattening() []DisplayFileList {
	if ft == nil || len(ft.Sub) == 0 {
		return nil
	}

	var ans []DisplayFileList

	for _, p1 := range ft.Sub {
		if len(p1.Sub) != 0 {
			for _, p2 := range p1.Sub {
				ans = p2.CollectFileInfo(ans, p1.Name)
			}
		}
	}
	return ans
}

func (ft *LayerFile) CollectFileInfo(ans []DisplayFileList, scenario string) []DisplayFileList {
	if len(ft.Sub) == 0 {
		return ans
	}

	var t []DisplayFile
	arr := strings.Split(ft.Dir, Sep)
	if len(arr) == 2 && arr[1] == "embedded_img" && len(ft.Sub) == 1 {
		for _, p := range ft.Sub[0].Sub {
			t = p.AppendFile(!strings.HasSuffix(p.Name, ".sha256sum"), t)
		}
	} else {
		for _, p := range ft.Sub {
			t = p.AppendFile(len(p.Sub) == 0 && !strings.HasSuffix(p.Name, ".sha256sum"), t)
			t = p.AppendDir(len(p.Sub) != 0, t)
		}
	}
	return append(ans, DisplayFileList{
		Scenario: scenario,
		Arch:     ft.Name,
		Tree:     t,
	})
}

func (ft *LayerFile) AppendFile(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}
	path := ft.Dir + Sep + ft.Name
	shaCode := ""
	sha, ok := FileTree.Mapping[path+".sha256sum"]
	if ok {
		shaCode = strings.Split(sha.Sha256, " ")[0]
	}
	return append(t, DisplayFile{
		Name:    ft.Name,
		Path:    path,
		Size:    ft.Size,
		ShaCode: shaCode,
		Type:    "file",
	})
}

func (ft *LayerFile) AppendDir(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}

	return append(t, DisplayFile{
		Name: ft.Name,
		Path: ft.Dir + Sep + ft.Name,
		Type: "dir",
	})
}

var RepoVersionList []DisplayRepoVersion
var RepoSourceFileNum int

func CollectRepoVersionList() {
	RepoSourceFileNum = len(FileTree.Mapping)
	if len(RepoVersionList) > 0 {
		RepoVersionList = RepoVersionList[0:0]
	}
	for k := range FileTree.Mapping {
		pathArr := strings.Split(k, Sep)
		if len(pathArr) == 1 {
			scenario := checkRepoScenario(pathArr[0])
			arch := checkRepoArch(scenario, pathArr[0])
			if len(arch) > 0 {
				RepoVersionList = append(RepoVersionList, DisplayRepoVersion{
					Version:  pathArr[0],
					Scenario: scenario,
					Arch:     arch,
					LTS:      strings.Contains(pathArr[0], "LTS"),
				})
			}
		}
	}
	sort.SliceStable(RepoVersionList, func(i, j int) bool {
		return RepoVersionList[i].Version < RepoVersionList[j].Version
	})
}

func checkRepoScenario(versionName string) []string {
	var scenario []string
	for _, v := range Scenario {
		if _, ok := FileTree.Mapping[versionName+Sep+v]; ok {
			scenario = append(scenario, v)
		}
	}
	return scenario
}

func checkRepoArch(scenario []string, versionName string) []string {
	var arch []string
	for _, v := range Arch {
		for _, v1 := range scenario {
			if _, ok := FileTree.Mapping[versionName+Sep+v1+Sep+v]; ok {
				arch = append(arch, v)
				break
			}
		}
	}
	return arch
}

var PathFilter []string
var Scenario []string
var Arch []string

func InitPathFilter(filter config.DirFilter) {
	if len(filter.SecondDir) == 0 && len(filter.ThirdDir) == 0 {
		return
	}

	if len(filter.SecondDir) != 0 && len(filter.ThirdDir) == 0 {
		PathFilter = filter.SecondDir
		Scenario = filter.SecondDir
	} else if len(filter.SecondDir) == 0 && len(filter.ThirdDir) != 0 {
		PathFilter = filter.ThirdDir
		Arch = filter.ThirdDir
	} else {
		Scenario = filter.SecondDir
		Arch = filter.ThirdDir
		for _, v1 := range filter.SecondDir {
			for _, v2 := range filter.ThirdDir {
				PathFilter = append(PathFilter, v1+Sep+v2)
			}
		}
	}
}

func Filter(path string) bool {
	if path[:10] != "openEuler-" {
		return false
	}
	for _, v := range PathFilter {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}
