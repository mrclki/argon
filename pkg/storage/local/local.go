package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peertechde/argon/pkg/storage"
)

const (
	defaultPermissions = os.FileMode(0600)
)

func New(dir string) storage.Storage {
	return &Local{
		dir: dir,
	}
}

type Local struct {
	dir string
}

func (l *Local) path(p string) string {
	return filepath.Join(l.dir, p)
}

func (l *Local) Read(_ context.Context, name string) ([]byte, error) {
	b, err := ReadFile(l.path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &storage.NotFoundError{Name: name}
		}
		return nil, storage.ErrInternal
	}
	return b, nil
}

func (l *Local) Write(_ context.Context, name string, data []byte) error {
	if _, err := os.Stat(l.path(name)); err == nil {
		return &storage.AlreadyExistsError{Name: name}
	}
	if dir := filepath.Dir(name); dir != "" && dir != "." {
		fmt.Println(dir)
		return storage.ErrInvalidName
	}
	return WriteFile(l.path(name), data, defaultPermissions)
}

func (l *Local) List(_ context.Context) ([]string, error) {
	files, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		result = append(result, file.Name())
	}
	return result, nil
}

func (l *Local) Stat(_ context.Context, name string) (*storage.FileInfo, error) {
	fi, err := Stat(l.path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &storage.NotFoundError{Name: name}
		}
		return nil, storage.ErrInternal
	}
	return fi, nil
}

func (l *Local) Rename(_ context.Context, old, new string) error {
	if _, err := Stat(l.path(old)); os.IsNotExist(err) {
		return &storage.NotFoundError{Name: old}
	}
	if _, err := Stat(l.path(new)); err == nil {
		return &storage.AlreadyExistsError{Name: new}
	}
	return Rename(l.path(old), l.path(new))

}

func (l *Local) Remove(_ context.Context, name string) error {
	return Remove(l.path(name))
}

func (l *Local) Close() error {
	return nil
}
