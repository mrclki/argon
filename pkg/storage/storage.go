package storage

import (
	"context"
	"fmt"
	"time"
)

var (
	ErrInternal     = fmt.Errorf("internal error")
	ErrAccessDenied = fmt.Errorf("access denied")
	ErrInvalidName  = fmt.Errorf("name is invalid")
)

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Unable to find %s file", e.Name)
}

func (e *NotFoundError) Is(target error) bool {
	t, ok := target.(*NotFoundError)
	if !ok {
		return false
	}
	return e.Name == t.Name
}

type AlreadyExistsError struct {
	Name string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("File %s already exists", e.Name)
}

func (e *AlreadyExistsError) Is(target error) bool {
	t, ok := target.(*AlreadyExistsError)
	if !ok {
		return false
	}
	return e.Name == t.Name
}

type Storage interface {
	Read(ctx context.Context, name string) ([]byte, error)
	Write(ctx context.Context, name string, data []byte) error
	List(ctx context.Context) ([]string, error)
	Stat(ctx context.Context, name string) (*FileInfo, error)
	Rename(ctx context.Context, old, new string) error
	Remove(ctx context.Context, name string) error
	Close() error
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    uint32    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	Dir     bool      `json:"dir"`
}
