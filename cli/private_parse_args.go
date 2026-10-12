package main

import "fmt"

type options struct {
	Pager   bool
	Version bool
	GitArgs []string
}

func parseArgs(args []string) (options, error) {
	opts := options{GitArgs: args}
	if len(args) == 1 && args[0] == "--version" {
		opts.Version = true
		return opts, nil
	}
	if len(args) > 0 && args[0] == "--pager" {
		if len(args) != 1 {
			return opts, fmt.Errorf("--pager reads Git output from stdin and takes no diff arguments")
		}
		opts.Pager = true
		return opts, nil
	}

	return opts, nil
}
