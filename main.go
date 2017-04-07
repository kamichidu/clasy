package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func classify(r io.Reader) error {
	var err error
	in, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	meta := new(MetaFile)
	if err = yaml.Unmarshal(in, meta); err != nil {
		return err
	}

	for _, fileData := range meta.Files {
		for _, tag := range fileData.Tags {
			srcFilename := filepath.Join(meta.SrcDir, fileData.Name)
			destFilename := filepath.Join(meta.DestDir, tag, fileData.DisplayName)

			destDir := filepath.Dir(destFilename)
			if _, err = os.Stat(destDir); err != nil {
				if err = os.MkdirAll(destDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Can't create directory %v: %s\n", destDir, err)
					continue
				}
			}

			relSrcFilename, err := filepath.Rel(destDir, srcFilename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Can't get relative path, basedir %v and file %v: %s\n", destDir, srcFilename, err)
				continue
			}

			if err = os.Symlink(relSrcFilename, destFilename); err != nil {
				fmt.Fprintf(os.Stderr, "Can't create symlink %v => %v: %s\n", srcFilename, destFilename, err)
			}
		}
	}
	return nil
}

func run() int {
	var metaFilename string
	var mode string
	flag.StringVar(&metaFilename, "f", "meta.yml", "Meta `FILE`")
	flag.StringVar(&mode, "m", "classify", "Execution mode: classify, template")
	flag.Parse()

	flags := os.O_RDWR | os.O_EXCL
	if _, err := os.Stat(metaFilename); err != nil {
		flags |= os.O_CREATE
	}

	metaFile, err := os.OpenFile(metaFilename, flags, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't open file: %s\n", err)
		return 128
	}
	defer metaFile.Close()

	err = nil
	switch mode {
	case "classify":
		err = classify(metaFile)
	case "template":
		err = generateTemplate(metaFile)
	default:
		err = fmt.Errorf("No such mode: %s", mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	} else {
		return 0
	}
}

func main() {
	os.Exit(run())
}
