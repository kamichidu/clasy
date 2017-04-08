package main

import (
	"flag"
	"fmt"
	"os"
)

func run() int {
	var metaFilename string
	var mode string
	flag.StringVar(&metaFilename, "f", "meta.yml", "Meta `FILE`")
	flag.StringVar(&mode, "m", "classify", "Execution mode: classify, generate-template")
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
	case "generate-template":
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
