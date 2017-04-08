package main

import (
	"io"

	"gopkg.in/yaml.v2"
)

func generateTemplate(w io.Writer) error {
	b, err := yaml.Marshal(&MetaFile{
		Schema:  "0.0",
		SrcDir:  "Original file stored directory",
		DestDir: "Directory will be created symlinks",
		Files: []*FileData{
			&FileData{
				Name:        "filename relative in the directory",
				DisplayName: "display name",
				Tags:        []string{"tag"},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
