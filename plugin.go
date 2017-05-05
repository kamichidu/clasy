package clasy

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

type Plugin interface {
	Name() string
	TakeMetaInfo(context.Context, os.FileInfo) (*FileData, error)
}

type Plugins []Plugin

func (self Plugins) Name() string {
	names := make([]string, 0)
	for _, plugin := range self {
		names = append(names, plugin.Name())
	}
	return strings.Join(names, ", ")
}

func (self Plugins) TakeMetaInfo(ctx context.Context, info os.FileInfo) (*FileData, error) {
	for _, plug := range self {
		if meta, err := plug.TakeMetaInfo(ctx, info); err != nil {
			log.Printf("plugin error: %s", err)
		} else {
			return meta, nil
		}
	}
	return nil, nil
}

var _ Plugin = Plugins(nil)

func LoadPlugin(pluginDir string) (Plugin, error) {
	globPattern := filepath.Join(pluginDir, "*.so")
	filenames, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, err
	}

	plugs := make(Plugins, 0)
	for _, filename := range filenames {
		plg, err := plugin.Open(filename)
		if err != nil {
			log.Printf("plugin load failed: %s", err)
			continue
		}
		sym, err := plg.Lookup("ClasyEnabled")
		if err != nil {
			log.Printf("symbol lookup failed: %s", err)
			continue
		}
		if v, ok := sym.(Plugin); ok {
			plugs = append(plugs, v)
		} else {
			log.Print("type assertion failed")
		}
	}
	return plugs, nil
}
