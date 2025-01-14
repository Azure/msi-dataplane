package dataplane

import (
	"errors"
	"os"
	"strings"

	"github.com/gofrs/flock"
)

type File struct {
	Path     string // Path to the file including the file name
	fileLock *flock.Flock
}

// initializeFileLock creates a new file lock if it doesn't already exist
func (f *File) initializeFileLock() {
	if f.fileLock != nil {
		return
	}
	f.fileLock = flock.New(f.Path)
}

// CheckFileExists checks if the file exists. It does not concern itself with any other errors such as permissions.
func (f *File) checkFileExists() error {
	if strings.TrimSpace(f.Path) == "" {
		return os.ErrNotExist
	}

	_, err := os.Stat(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}

	return nil
}
