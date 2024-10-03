package annotations

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Save 保存
func (f *File) Save(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if len(ext) == 0 {
		filename = filepath.Join(filename, f.Name)
	}
	newFilename := uuid.New().String() + ext
	path := filepath.Dir(filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return "", err
		}
	}
	filename = filepath.Join(path, newFilename)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write(f.Content)
	return path, err
}
