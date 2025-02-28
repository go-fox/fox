package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Module 模块的配置信息
type Module struct {
	Path      string
	Main      bool
	Dir       string
	GoMod     string
	GoVersion string
}

// isMod 判断是否是mod
func isMod(workDir string) (bool, error) {
	if len(workDir) == 0 {
		return false, errors.New("the work directory is not found")
	}
	if _, err := os.Stat(workDir); err != nil {
		return false, err
	}

	data, err := Exec("go list -m -f '{{.GoMod}}'", workDir)
	if err != nil || len(data) == 0 {
		return false, nil
	}

	return true, nil
}

// getRealModule 获取指定工作目录下的模块信息
func getRealModule(workDir string) (*Module, error) {
	data, err := Exec("go list -json -m", workDir)
	if err != nil {
		return nil, err
	}
	modules, err := decodePackages(strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	for _, m := range modules {
		if strings.HasPrefix(workDir, m.Dir) {
			return &m, nil
		}
	}
	return nil, errors.New("no matched module")
}

// decodePackages 解析包
func decodePackages(rc io.Reader) ([]Module, error) {
	var modules []Module
	decoder := json.NewDecoder(rc)
	for decoder.More() {
		var m Module
		if err := decoder.Decode(&m); err != nil {
			return nil, fmt.Errorf("invalid module: %v", err)
		}
		modules = append(modules, m)
	}
	return modules, nil
}
