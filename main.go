package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

//go:embed VERSION
var VERSION string

// counter implements flag.Value for counting repeated flag occurrences.
type counter int

func (c *counter) String() string   { return strconv.Itoa(int(*c)) }
func (c *counter) Set(string) error { *c++; return nil }
func (c *counter) IsBoolFlag() bool { return true }

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] [version]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Validate or increment a semantic version 2.0.0 string.\n\n")
	fmt.Fprintf(os.Stderr, "If no flags are given, validates the input and prints the recognized\n")
	fmt.Fprintf(os.Stderr, "semantic version. If bump flags are given, increments the specified\n")
	fmt.Fprintf(os.Stderr, "component(s) and prints the result.\n\n")
	fmt.Fprintf(os.Stderr, "Input is read from the first non-flag argument, or from stdin if\n")
	fmt.Fprintf(os.Stderr, "no arguments are provided.\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --help         Print this help message\n\n")
	fmt.Fprintf(os.Stderr, "  --major        Increment the major version (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  --minor        Increment the minor version (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  --patch        Increment the patch version (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  --prerelease   Retain pre-release label when bumping\n")
	fmt.Fprintf(os.Stderr, "  --build        Retain build metadata when bumping\n\n")
	fmt.Fprintf(os.Stderr, "  --json         Encode components using JSON (compact)\n")
	fmt.Fprintf(os.Stderr, "  --jsonpp       Encode components using JSON (pretty-printed)\n")
}

// readFirstLine reads from r until EOF, \n, or \0 and returns the bytes.
func readFirstLine(r io.Reader) ([]byte, error) {
	var buf []byte
	one := make([]byte, 1)
	for {
		n, err := r.Read(one)
		if n > 0 {
			if one[0] == '\n' || one[0] == 0 {
				break
			}
			buf = append(buf, one[0])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return buf, err
		}
	}
	return buf, nil
}

func run() int {
	var major, minor, patch counter
	var prerelease, build bool

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&major, "major", "")
	fs.Var(&minor, "minor", "")
	fs.Var(&patch, "patch", "")
	fs.BoolVar(&prerelease, "prerelease", false, "")
	fs.BoolVar(&build, "build", false, "")

	var json, jsonpp bool
	fs.BoolVar(&json, "json", false, "")
	fs.BoolVar(&jsonpp, "jsonpp", false, "")

	help := false
	fs.BoolVar(&help, "help", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		usage()
		return 1
	}

	if help {
		usage()
		return 0
	}

	var input []byte
	args := fs.Args()
	if len(args) > 0 {
		// Scan non-flag args for a semver, using each arg as input
		// Concatenate with spaces to allow scanning across args
		for i, arg := range args {
			if i > 0 {
				input = append(input, ' ')
			}
			input = append(input, arg...)
		}
		// Truncate at first \n or \0
		for i, b := range input {
			if b == '\n' || b == 0 {
				input = input[:i]
				break
			}
		}
	} else {
		var err error
		input, err = readFirstLine(os.Stdin)
		if err != nil {
			return 1
		}
	}

	v, ok := MakeSemVer(input, WithJSONIndent(`  `))
	if !ok {
		return 1
	}

	bumping := major > 0 || minor > 0 || patch > 0

	if bumping {
		for range int(major) {
			v.BumpMajor()
		}
		for range int(minor) {
			v.BumpMinor()
		}
		for range int(patch) {
			v.BumpPatch()
		}
		if !prerelease {
			v.StripPreRelease()
		}
		if !build {
			v.StripBuild()
		}
	}

	format := v.String
	switch {
	// --jsonpp takes precedence over --json
	case jsonpp:
		format = v.PrettyJSON
	case json:
		format = v.JSON
	}

	fmt.Println(format())
	return 0
}

func main() {
	os.Exit(run())
}
