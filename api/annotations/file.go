package annotations

import (
	"os"
	"path/filepath"
)

// Save 保存
func (f *File) Save(path string) (string, error) {
	ext := filepath.Ext(path)
	if len(ext) == 0 {
		path = filepath.Join(path, f.Name)
	}
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write(f.Content)
	return path, err
}
