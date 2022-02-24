//go:build linux
// +build linux

package local

import (
	"os"

	"github.com/peertechde/argon/pkg/storage"
)

func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func WriteFile(name string, data []byte, perm os.FileMode) error {
	if err := checkName(name); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

func Stat(name string) (*storage.FileInfo, error) {
	if err := checkName(name); err != nil {
		return nil, err
	}
	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	fileInfo := &storage.FileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    uint32(fi.Mode()),
		ModTime: fi.ModTime(),
		Dir:     fi.IsDir(),
	}
	return fileInfo, nil
}

func Rename(old, new string) error {
	if err := checkName(new); err != nil {
		return err
	}
	return os.Rename(old, new)
}

func Remove(name string) error {
	if err := checkName(name); err != nil {
		return err
	}
	return os.Remove(name)
}

func checkName(name string) error {
	if name == "" {
		return storage.ErrInvalidName
	}
	switch name {
	case ".", "..":
		return storage.ErrInvalidName
	case "/":
		return storage.ErrInvalidName
	case "*":
		return storage.ErrInvalidName
	default:
		return nil
	}
	if len(name) > 255 {
		return storage.ErrInvalidName
	}

	return nil
}
