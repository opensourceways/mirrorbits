// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// Licensed under the MIT license
package filesystem

import (
	"encoding/json"
	"github.com/opensourceways/mirrorbits/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const Sep = string(os.PathSeparator)

type fileStore struct {
	M    map[string]*LayerFile
	Root LayerFile
}

var FileTree = fileStore{
	Root: LayerFile{},
	M:    make(map[string]*LayerFile, 1024*10),
}

type LayerFile struct {
	Dir    string
	Name   string
	Size   string
	Sha256 string
	Type   string
	Sub    []*LayerFile
}

func (ft *LayerFile) FileSizeView(size int64) {
	if size <= 1024 {
		ft.Size = strconv.FormatInt(size, 10) + " Bytes"
		return
	}

	if size <= 1024*1024 {
		ft.Size = strconv.FormatInt(size>>10, 10) + " K"
		return
	}

	if size <= 1024*1024*1024 {
		ft.Size = strconv.FormatInt(size>>20, 10) + " M"
	}

}

func (ft *LayerFile) SetFileData(fullPath string) {

	relPath, err := filepath.Abs(config.GetConfig().Repository + Sep + fullPath)
	if err != nil {
		println("error")
	} else {
		stat, err1 := os.Stat(relPath)
		if err1 == nil {
			ft.FileSizeView(stat.Size())
			ft.Type = "file"
		}
		if strings.HasSuffix(relPath, ".sha256sum") {
			data, err2 := os.ReadFile(relPath)
			if err2 == nil {
				ft.Sha256 = string(data)
			}
		}
	}
}

func (ft *LayerFile) LayeringPath(fm map[string]*LayerFile, path string) {
	fileLayer := strings.Split(path, Sep)
	layerLength := len(fileLayer)
	if _, ok := fm[fileLayer[0]]; !ok {
		fm[fileLayer[0]] = &LayerFile{
			Name: fileLayer[0],
		}
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
				node.SetFileData(currPath)
			}
			if _, ok1 := fm[currPath]; !ok1 {
				fm[currPath] = node
				dp.Sub = append(dp.Sub, node)
			}
		}
	}
}

type DisplayFile struct {
	Name    string
	Path    string
	Size    string
	ShaCode string
	Type    string
}

type DisplayFileArray struct {
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

func (ft *LayerFile) Flattening() []DisplayFileArray {
	if ft == nil || len(ft.Sub) == 0 {
		return nil
	}

	var ans []DisplayFileArray

	for _, p1 := range ft.Sub {
		if len(p1.Sub) != 0 {
			for _, p2 := range p1.Sub {
				ans = p2.DepthSearch(ans, p1.Name)
			}
		}
	}
	return ans
}

func (ft *LayerFile) DepthSearch(ans []DisplayFileArray, scenario string) []DisplayFileArray {
	if len(ft.Sub) == 0 {
		return ans
	}

	var t []DisplayFile
	for _, p := range ft.Sub {
		t = p.AppendFile(len(p.Sub) == 0 && !strings.HasSuffix(p.Name, ".sha256sum"), t)
		t = p.AppendDir(len(p.Sub) != 0, t)
	}
	return append(ans, DisplayFileArray{
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
	sha, ok := FileTree.M[path+".sha256sum"]
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

func (ft *LayerFile) Print() string {
	b, _ := json.Marshal(*ft)
	return string(b)
}

// type PathFilter [...]string
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
	for _, v := range PathFilter {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}
