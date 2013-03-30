package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
)

type Package struct {
	Dir          string
	ImportPath   string
	Name         string
	Target       string
	Goroot       bool
	Standard     bool
	Stale        bool
	Root         string
	Imports      []string
	Deps         []string
	TestGoFiles  []string
	TestImports  []string
	XTestGoFiles []string
	XTestImports []string
}

func getPackages(list string, shuffle bool) []Package {
	js, err := stdout(".", "go", "list", "-json", list)
	if err != nil {
		panic(err)
	}
	var packages []Package
	dec := json.NewDecoder(bytes.NewBuffer([]byte(js)))
	for {
		var pkg Package
		err := dec.Decode(&pkg)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if pkg.Stale {
			panic("stale package: " + pkg.ImportPath)
		}

		// skip package without tests
		if len(pkg.TestGoFiles) == 0 && len(pkg.XTestGoFiles) == 0 {
			continue
		}

		packages = append(packages, pkg)
	}

	if !shuffle {
		return packages
	}

	shuffled := make([]Package, 0, len(packages))
	for _, i := range rand.Perm(len(packages)) {
		shuffled = append(shuffled, packages[i])
	}
	return shuffled
}
