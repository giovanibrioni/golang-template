// Deterministic architecture checks from AGENTS.md. Stdlib only.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var (
	rootDir  string
	module   string
	failures []string
)

var domainFiles = []string{"model.go", "repository.go", "service.go", "handler.go"}

var commitOwners = map[string]bool{
	"internal/core/httpx/httpx.go":   true,
	"internal/core/db/migrate.go":    true,
	"internal/core/pgtest/pgtest.go": true,
}

func main() {
	rootDir, module = locate()

	domains := domainPackages()
	if len(domains) == 0 {
		fail("no domain packages under internal/ besides core, api, and health")
	}

	routerText := readText(filepath.Join(rootDir, "internal", "api", "router.go"))
	for _, domain := range domains {
		dir := filepath.Join(rootDir, "internal", domain)
		for _, name := range domainFiles {
			if !fileExists(filepath.Join(dir, name)) {
				fail(fmt.Sprintf("internal/%s/%s missing", domain, name))
			}
		}
		if !strings.Contains(routerText, module+"/internal/"+domain) {
			fail(fmt.Sprintf("internal/%s is not imported from internal/api/router.go", domain))
		}
		if !fakeRepositoryExists(dir) {
			fail(fmt.Sprintf("internal/%s tests declare no fake repository", domain))
		}
		for _, name := range []string{"repository_test.go", "handler_test.go"} {
			text := readText(filepath.Join(dir, name))
			if !strings.Contains(text, "pgtest.") || !strings.Contains(text, "func Test") {
				fail(fmt.Sprintf("internal/%s/%s does not run pgtest integration tests", domain, name))
			}
		}
	}

	walkGo(filepath.Join(rootDir, "internal"), checkOwnership)
	walkGo(filepath.Join(rootDir, "cmd"), checkOwnership)
	walkGo(filepath.Join(rootDir, "internal", "core"), checkCoreImports)

	if len(failures) > 0 {
		for _, item := range failures {
			fmt.Printf("FAIL %s\n", item)
		}
		os.Exit(1)
	}
	fmt.Printf("OK architecture (%d domain package(s): %s)\n", len(domains), strings.Join(domains, ", "))
}

func domainPackages() []string {
	entries, err := os.ReadDir(filepath.Join(rootDir, "internal"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "read internal/:", err)
		os.Exit(2)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		switch entry.Name() {
		case "core", "api", "health":
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

func fakeRepositoryExists(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name
				fake := strings.HasPrefix(name, "Fake") || strings.HasPrefix(name, "fake")
				if fake && strings.Contains(name, "Repository") {
					return true
				}
			}
		}
	}
	return false
}

func checkOwnership(path string) {
	rel := relRoot(path)
	lines := strings.Split(readText(path), "\n")
	for i, line := range lines {
		if !commitOwners[rel] && (strings.Contains(line, ".Commit(") || strings.Contains(line, ".Rollback(")) {
			fail(fmt.Sprintf("%s:%d commit/rollback outside httpx.Wrap", rel, i+1))
		}
		for _, call := range []string{"fmt.Print(", "fmt.Println(", "fmt.Printf("} {
			if strings.Contains(line, call) {
				fail(fmt.Sprintf("%s:%d %s in application code; use slog", rel, i+1, strings.TrimSuffix(call, "(")))
			}
		}
	}
}

func checkCoreImports(path string) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		fail(fmt.Sprintf("%s cannot parse: %v", relRoot(path), err))
		return
	}
	for _, spec := range file.Imports {
		imported := strings.Trim(spec.Path.Value, `"`)
		if !strings.HasPrefix(imported, module+"/internal/") {
			continue
		}
		if !strings.HasPrefix(imported, module+"/internal/core/") {
			fail(fmt.Sprintf("%s core package imports %s", relRoot(path), imported))
		}
	}
}

func walkGo(dir string, fn func(path string)) {
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry == nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fn(path)
		return nil
	})
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func relRoot(path string) string {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func fail(message string) {
	failures = append(failures, message)
}
