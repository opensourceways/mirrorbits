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
		data, err1 := os.ReadFile(relPath)
		if err1 == nil {
			ft.Sha256 = string(data)
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
			if currPath == path && strings.HasSuffix(currPath, ".sha256sum") {
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
	Size    string
	ShaCode string
	Type    string
}

type DisplayFileArray struct {
	Scenario string
	Arch     string
	Tree     []DisplayFile
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

	return append(t, DisplayFile{
		Name:    ft.Name,
		Size:    ft.Size,
		ShaCode: FileTree.M[ft.Dir+Sep+ft.Name+".sha256sum"].Sha256,
		Type:    "file",
	})
}

func (ft *LayerFile) AppendDir(flag bool, t []DisplayFile) []DisplayFile {
	if !flag {
		return t
	}

	return append(t, DisplayFile{
		Name: ft.Name,
		Type: "dir",
	})
}

func (ft *LayerFile) Print() string {
	b, _ := json.Marshal(*ft)
	return string(b)
}

// type PathFilter [...]string
var PathFilter = [...]string{
	"ISO" + Sep + "x86_64",
	"ISO" + Sep + "aarch64",
	"ISO" + Sep + "arm32",
	"ISO" + Sep + "loongarch64",
	"ISO" + Sep + "riscv64",
	"ISO" + Sep + "power",
	"ISO" + Sep + "sw64",
	"edge_img" + Sep + "x86_64",
	"edge_img" + Sep + "aarch64",
	"edge_img" + Sep + "arm32",
	"edge_img" + Sep + "loongarch64",
	"edge_img" + Sep + "riscv64",
	"edge_img" + Sep + "power",
	"edge_img" + Sep + "sw64",
	"virtual_machine_img" + Sep + "x86_64",
	"virtual_machine_img" + Sep + "aarch64",
	"virtual_machine_img" + Sep + "arm32",
	"virtual_machine_img" + Sep + "loongarch64",
	"virtual_machine_img" + Sep + "riscv64",
	"virtual_machine_img" + Sep + "power",
	"virtual_machine_img" + Sep + "sw64",
	"embedded_img" + Sep + "x86_64",
	"embedded_img" + Sep + "aarch64",
	"embedded_img" + Sep + "arm32",
	"embedded_img" + Sep + "loongarch64",
	"embedded_img" + Sep + "riscv64",
	"embedded_img" + Sep + "power",
	"embedded_img" + Sep + "sw64",
}

func Filter(path string) bool {
	for _, v := range PathFilter {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}
