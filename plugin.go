package clasy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

const loggerContextKey = "logger"

type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
}

type prefixedLogger interface {
	Prefix() string
	SetPrefix(string)
}

func WithLogger(parent context.Context, v Logger) context.Context {
	return context.WithValue(parent, loggerContextKey, v)
}

func LoggerFromContext(c context.Context) Logger {
	if v, ok := c.Value(loggerContextKey).(Logger); ok {
		return v
	} else {
		return log.New(ioutil.Discard, "", log.LstdFlags)
	}
}

type Plugin interface {
	Name() string
	TakeMetaInfo(c context.Context, path string, info os.FileInfo) (string, []string, error)
}

type Plugins []Plugin

func (self Plugins) Name() string {
	names := make([]string, 0)
	for _, plugin := range self {
		names = append(names, plugin.Name())
	}
	return strings.Join(names, ", ")
}

func (self Plugins) TakeMetaInfo(c context.Context, path string, info os.FileInfo) (string, []string, error) {
	log := LoggerFromContext(c)
	var logPrefix string
	if log, ok := log.(prefixedLogger); ok {
		logPrefix = log.Prefix()
	}
	defer func() {
		if log, ok := log.(prefixedLogger); ok {
			log.SetPrefix(logPrefix)
		}
	}()

	for _, plug := range self {
		if log, ok := log.(prefixedLogger); ok {
			log.SetPrefix(fmt.Sprintf("[%v] ", plug.Name()))
		}

		if displayName, tags, err := plug.TakeMetaInfo(WithLogger(c, log), path, info); err != nil {
			log.Printf("plugin error: %s", err)
		} else {
			return displayName, tags, nil
		}
	}
	return info.Name(), []string{}, nil
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
		sym, err := plg.Lookup("CreatePlugin")
		if err != nil {
			log.Printf("symbol lookup failed: %s", err)
			continue
		}
		if factory, ok := sym.(func() interface{}); ok {
			plugs = append(plugs, factory().(Plugin))
		} else {
			log.Print("type assertion failed")
		}
	}
	return plugs, nil
}
