package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

const pipelinesFileName string = "/pipelines.yaml"

// Parse decodes YAML describing an environment manifest.
func Parse(in io.Reader) (*Manifest, error) {
	m := &Manifest{}
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseFile is a wrapper around Parse that accepts a filename, it opens and
// parses the file, and closes it.
func ParseFile(fs afero.Fs, filepath string) (*Manifest, error) {
	info, err := fs.Stat(filepath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("The path %s is a file path(required directory path)", filepath)
	}
	f, err := fs.Open(filepath + pipelinesFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
