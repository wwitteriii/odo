package ioutils

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// IsExisting returns bool whether path exists
func IsExisting(path string) (bool, error) {
	appFS := afero.NewOsFs()
	fileInfo, err := appFS.Stat(path)

	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return true, fmt.Errorf("%q: Dir already exists at %s", filepath.Base(path), path)
	}
	return true, fmt.Errorf("%q: File already exists at %s", filepath.Base(path), path)
}
