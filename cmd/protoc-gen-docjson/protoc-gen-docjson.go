package main

import (
	// Built-in/core modules.

	"flag"
	"os"

	// "path"

	// Third-party modules.
	log "github.com/sirupsen/logrus"

	// Generated code.
	// First-party modules.
	plugin "github.com/cuberat/protoc-gen-docjson/internal/plugin"
)

func main() {
	var (
		infile  string
		outfile string
	)

	flag.StringVar(&infile, "infile", "", "Input file")
	flag.StringVar(&outfile, "outfile", "", "Output file")

	flag.Parse()

	debug := true
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	err := plugin.ProcessCodeGenRequest(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("couldn't process code generation request: %s", err)
	}

}
