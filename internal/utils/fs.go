package utils

import (
	"errors"
	"io/fs"
)

func FileExists(fsys fs.FS, path string) (bool, error) {
	_, err := fs.Stat(fsys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil // File doesn't exist, but no error
		}
		return false, err // Other error occurred
	}
	return true, nil // File exists
}
