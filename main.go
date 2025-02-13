package main

import (
	"flag"
	"fmt"
	"os"
)

type RenameOptions struct {
    dryRun      bool
    recursive   bool
    ignoreCase  bool
    showChanges bool
}

func main() {
	opts := parseFlags()

    args := flag.Args()
    if err := run(args, opts); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func parseFlags() *RenameOptions {
    opts := &RenameOptions{}

    flag.BoolVar(&opts.dryRun, "n", false, "dry run - show what would be renamed")
    flag.BoolVar(&opts.recursive, "r", false, "recursive - include subdirectories")
    flag.BoolVar(&opts.ignoreCase, "i", false, "ignore case in pattern matching")
    flag.BoolVar(&opts.showChanges, "v", false, "verbose - show renamed files")

    flag.Parse()
    return opts
}

func run(args []string, opts *RenameOptions) error {
    pattern, files, err := validateArgs(args)
    if err != nil {
        return err
    }
    fmt.Println("Pattern:", pattern)
    fmt.Println("Files:", files)
    return nil
}

func validateArgs(args []string) (pattern string, files []string, err error) {
    if len(args) < 1 {
        return "", nil, fmt.Errorf("pattern argument is required")
    }
    if len(args) < 2 {
        return "", nil, fmt.Errorf("at least one file argument is required")
    }
    return args[0], args[1:], nil
}