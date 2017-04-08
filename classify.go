package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func classify(r io.Reader) error {
	var err error
	meta, err := LoadMetaFromReader(r)
	if err != nil {
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

			if info, err := os.Lstat(destFilename); err != nil {
				// not exist, just create it
				if err = os.Symlink(relSrcFilename, destFilename); err != nil {
					fmt.Fprintf(os.Stderr, "Can't create symlink %v => %v: %s\n", srcFilename, destFilename, err)
				}
			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				// already exists, check conditions and report it
				destination, err := os.Readlink(destFilename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Illegal state detected, file already exists and can't get its attributes: %s\n", destFilename)
				} else if destination != relSrcFilename {
					fmt.Fprintf(os.Stderr, "Illegal state detected, file already exists and it point to another one: %s => %s\n", destFilename, destination)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Can't create symlink, file already exits and it has unknown state: %s\n", destFilename)
			}
		}
	}
	return nil
}
