package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamichidu/clasy"
	"gopkg.in/urfave/cli.v1"
)

func chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		for _, fn := range funcs {
			if err := fn(c); err != nil {
				return err
			}
		}
		return nil
	}
}

func confirm(r io.Reader) (bool, error) {
	s, err := bufio.NewReader(r).ReadString('\n')
	if err != nil {
		return false, err
	}
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		s = "N"
	}
	return strings.HasPrefix(s, "Y"), nil
}

func classifyAction(c *cli.Context) error {
	plug := c.App.Metadata["Plugin"].(clasy.Plugin)
	logger := c.App.Metadata["Logger"].(clasy.Logger)
	in := c.App.Metadata["Reader"].(io.Reader)
	verbose := c.GlobalBool("verbose")

	if c.NArg() != 2 {
		return cli.NewExitError("Must specify position arguments", 128)
	}
	srcDir := filepath.Clean(c.Args().Get(0))
	destDir := filepath.Clean(c.Args().Get(1))
	logger.Printf("Source directory: %v", srcDir)
	logger.Printf("Destination directory: %v", destDir)

	fmt.Fprint(c.App.Writer, "Are you sure? [y/n]")
	proceed, err := confirm(in)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Confirmation error: %s", err), 128)
	} else if !proceed {
		return nil
	}

	if verbose {
		logger.Printf("Gather files from %v", srcDir)
	}
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Printf("Gathering files error: %s", err)
			return nil
		} else if info.IsDir() {
			if verbose {
				logger.Printf("Ignore directory item: %v", info.Name())
			}
			return nil
		}

		if verbose {
			logger.Printf("Found file item: %v", path)
		}

		var (
			displayName = info.Name()
			tags        = make([]string, 0)
		)

		// overwrite file data
		plugCtx := context.Background()
		plugCtx = clasy.WithLogger(plugCtx, logger)
		displayName, tags, err = plug.TakeMetaInfo(plugCtx, path, info)
		if err != nil {
			logger.Printf("Can't plugged-in, proceed any way: %s", err)
		}
		if displayName != "" {
			logger.Printf("Plugin modified display name: %s: %v -> %v", path, info.Name(), displayName)
		}
		if tags != nil {
			logger.Printf("Plugin modified tags: %s: %v", path, tags)
		}

		for _, tag := range tags {
			destFilename := filepath.Join(destDir, tag, displayName)

			destDir := filepath.Dir(destFilename)
			if _, err := os.Stat(destDir); err != nil {
				if verbose {
					logger.Printf("Create directory: %v", destDir)
				}
				if err = os.MkdirAll(destDir, 0755); err != nil {
					logger.Printf("Can't create directory %v: %s", destDir, err)
					continue
				}
			}

			relSrcFilename, err := filepath.Rel(destDir, path)
			if err != nil {
				logger.Printf("Can't get relative path, basedir %v and file %v: %s", destDir, path, err)
				continue
			}

			if info, err := os.Lstat(destFilename); err != nil {
				// not exist, just create it
				if verbose {
					logger.Printf("Create symlink: %v => %v", path, destFilename)
				}
				if err = os.Symlink(relSrcFilename, destFilename); err != nil {
					logger.Printf("Can't create symlink %v => %v: %s", path, destFilename, err)
				}
			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				// already exists, check conditions and report it
				if verbose {
					logger.Printf("Symlink %v already exists, check it", destFilename)
				}
				destination, err := os.Readlink(destFilename)
				if err != nil {
					logger.Printf("Illegal state detected, file already exists and can't get its attributes: %s", destFilename)
				} else if destination != relSrcFilename {
					logger.Printf("Illegal state detected, file already exists and it point to another one: %s => %s", destFilename, destination)
				}
			} else {
				logger.Printf("Can't create symlink, file already exits and it has unknown state: %s", destFilename)
			}
		}
		return nil
	})
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Gathering error: %s", err), 1)
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
