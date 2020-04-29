package ioutils

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// IsExisting returns bool whether path exists
func IsExisting(fs afero.Fs, path string) (bool, error) {
	fileInfo, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return true, fmt.Errorf("%q: Dir already exists at %s", filepath.Base(path), path)
	}
	return true, fmt.Errorf("%q: File already exists at %s", filepath.Base(path), path)
}
