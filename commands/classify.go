package commands

import (
	"context"
	"fmt"
	"github.com/kamichidu/clasy"
	"gopkg.in/urfave/cli.v1"
	"io"
	"os"
	"path/filepath"
)

func classifyAction(c *cli.Context) error {
	plugins, err := clasy.LoadPlugins("./plugins-enabled/")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Can't load plugins, disabling it...")
	}
	fmt.Fprintf(os.Stderr, "Loaded plugins: %s\n", plugins.Name())

	var r io.Reader
	meta, err := clasy.LoadMetaFromReader(r)
	if err != nil {
		return err
	}

	for _, fileData := range meta.Files {
		srcFilename := filepath.Join(meta.SrcDir, fileData.Name)
		if fileInfo, err := os.Stat(srcFilename); err != nil {
			fmt.Fprintf(os.Stderr, "Can't stat file: %v: %s\n", srcFilename, err)
			continue
		} else {
			// overwrite file data
			modFileData, err := plugins.TakeMetaInfo(context.Background(), fileInfo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Can't plugged-in, proceed any way: %s", err)
			} else if modFileData != nil {
				fmt.Fprintf(os.Stderr, "Plugin modified file data: %s\n", srcFilename)
				fileData = modFileData
			}
		}

		for _, tag := range fileData.Tags {
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

func init() {
	Commands = append(Commands, cli.Command{
		Name: "classify",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "",
			},
		},
		Action: classifyAction,
	})
}
