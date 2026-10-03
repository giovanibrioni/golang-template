// Unimported internal packages. Stdlib only.
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type pkgInfo struct {
	importPath string
	dir        string
	rel        string
}

func main() {
	rootDir, module := locate()
	packages := internalPackages(rootDir, module)
	importerDirs := importersByPackage(rootDir, packages)

	var failures, rels []string
	for importPath, info := range packages {
		rels = append(rels, info.rel)
		if len(importerDirs[importPath]) == 0 {
			failures = append(failures, fmt.Sprintf("%s is never imported outside its own package", info.rel))
		}
	}
	sort.Strings(failures)
	sort.Strings(rels)

	if len(failures) > 0 {
		for _, item := range failures {
			fmt.Printf("FAIL %s\n", item)
		}
		os.Exit(1)
	}
	fmt.Printf("OK orphans (%d internal packages imported)\n", len(packages))
}

func internalPackages(rootDir, module string) map[string]pkgInfo {
	packages := map[string]pkgInfo{}
	_ = filepath.WalkDir(filepath.Join(rootDir, "internal"), func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry == nil || !entry.IsDir() {
			return nil
		}
		goFiles, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		for _, file := range goFiles {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
				rel, err := filepath.Rel(rootDir, path)
				if err != nil {
					return nil
				}
				importPath := module + "/" + filepath.ToSlash(rel)
				packages[importPath] = pkgInfo{importPath: importPath, dir: path, rel: filepath.ToSlash(rel)}
				break
			}
		}
		return nil
	})
	return packages
}

func importersByPackage(rootDir string, packages map[string]pkgInfo) map[string]map[string]bool {
	importerDirs := map[string]map[string]bool{}
	for _, root := range []string{"cmd", "internal", "migrations"} {
		base := filepath.Join(rootDir, root)
		if !dirExists(base) {
			continue
		}
		_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry == nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return nil
			}
			dir := filepath.Dir(path)
			for _, spec := range file.Imports {
				imported := strings.Trim(spec.Path.Value, `"`)
				info, ok := packages[imported]
				if !ok || info.dir == dir {
					continue
				}
				if importerDirs[imported] == nil {
					importerDirs[imported] = map[string]bool{}
				}
				importerDirs[imported][dir] = true
			}
			return nil
		})
	}
	return importerDirs
}

func locate() (string, string) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot locate source file")
		os.Exit(2)
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "read go.mod:", err)
		os.Exit(2)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return root, strings.TrimSpace(after)
		}
	}
	fmt.Fprintln(os.Stderr, "no module directive in go.mod")
	os.Exit(2)
	return "", ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
