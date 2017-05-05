package commands

import (
	"github.com/kamichidu/clasy"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func generateAction(c *cli.Context) error {
	b, err := yaml.Marshal(&clasy.MetaFile{
		Schema:  "0.0",
		SrcDir:  "Original file stored directory",
		DestDir: "Directory will be created symlinks",
		Files: []*clasy.FileData{
			&clasy.FileData{
				Name:        "filename relative in the directory",
				DisplayName: "display name",
				Tags:        []string{"tag"},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = c.App.Writer.Write(b)
	return err
}

func init() {
	Commands = append(Commands, cli.Command{
		Name:      "generate",
		Aliases:   []string{},
		Usage:     "Generate skeleton template for meta.yml",
		ArgsUsage: "[FILE]",
		Action:    generateAction,
	})
}
