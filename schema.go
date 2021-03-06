package clasy

import (
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type MetaFile struct {
	Schema  string      `yaml:"schema"`
	SrcDir  string      `yaml:"source_directory"`
	DestDir string      `yaml:"destination_directory"`
	Files   []*FileData `yaml:"files"`
}

type FileData struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Tags        []string `yaml:"tags"`
}

func LoadMetaFromReader(r io.Reader) (*MetaFile, error) {
	in, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	meta := new(MetaFile)
	if err = yaml.Unmarshal(in, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func LoadMetaFromFilename(filename string) (*MetaFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadMetaFromReader(file)
}
