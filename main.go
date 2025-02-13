package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

    oldPattern, newPattern, err := parsePattern(pattern)
    if err != nil {
        return err
    }

    re, err := compilePattern(oldPattern, opts.ignoreCase)
    if err != nil {
        return err
    }

    return processFiles(files, re, newPattern, opts)
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

func parsePattern(pattern string) (oldPattern, newPattern string, err error) {
    if !strings.HasPrefix(pattern, "s/") {
        return "", "", fmt.Errorf("pattern must start with 's/'")
    }

    parts := strings.Split(strings.TrimPrefix(pattern, "s/"), "/")
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid pattern format. Expected: 's/old/new/'")
    }

	newPattern = strings.ReplaceAll(parts[1], "\\.", ".")
    return parts[0], newPattern, nil
}

func compilePattern(oldPattern string, ignoreCase bool) (*regexp.Regexp, error) {
    if ignoreCase {
        oldPattern = "(?i)" + oldPattern
    }
    return regexp.Compile(oldPattern)
}

func processFiles(files []string, re *regexp.Regexp, newPattern string, opts *RenameOptions) error {
    for _, glob := range files {
        matches, err := filepath.Glob(glob)
        if err != nil {
            return fmt.Errorf("error processing glob pattern %s: %v", glob, err)
        }

        for _, path := range matches {
            if err := processFile(path, re, newPattern, opts); err != nil {
                fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
                continue
            }
        }
    }
    return nil
}

func processFile(path string, re *regexp.Regexp, newPattern string, opts *RenameOptions) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }

    if info.IsDir() {
        if !opts.recursive {
            return nil
        }
        return processDirectory(path, re, newPattern, opts)
    }

    return renameFile(path, re, newPattern, opts)
}

func processDirectory(path string, re *regexp.Regexp, newPattern string, opts *RenameOptions) error {
    entries, err := os.ReadDir(path)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        subPath := filepath.Join(path, entry.Name())
        if err := processFile(subPath, re, newPattern, opts); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", subPath, err)
        }
    }
    return nil
}

func renameFile(path string, re *regexp.Regexp, newPattern string, opts *RenameOptions) error {
    dir := filepath.Dir(path)
    oldName := filepath.Base(path)
    newName := re.ReplaceAllString(oldName, newPattern)

    if oldName == newName {
        return nil
    }

    newPath := filepath.Join(dir, newName)

    if _, err := os.Stat(newPath); err == nil {
        return fmt.Errorf("target file already exists: %s", newPath)
    }

    if opts.showChanges || opts.dryRun {
        fmt.Printf("%s -> %s\n", path, newPath)
    }

    if !opts.dryRun {
        if err := os.Rename(path, newPath); err != nil {
            return fmt.Errorf("failed to rename: %v", err)
        }
    }

    return nil
}