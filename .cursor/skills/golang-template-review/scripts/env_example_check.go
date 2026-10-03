// Config env keys must appear in .env.example. Stdlib only.
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
	"strconv"
	"strings"
)

func main() {
	root := rootDir()
	config := filepath.Join(root, "internal", "core", "config", "config.go")
	example := filepath.Join(root, ".env.example")
	if !fileExists(config) {
		fmt.Println("FAIL internal/core/config/config.go missing")
		os.Exit(1)
	}
	if !fileExists(example) {
		fmt.Println("FAIL .env.example missing")
		os.Exit(1)
	}

	expected := getenvKeys(config)
	present := exampleKeys(example)
	var missing []string
	for key := range expected {
		if !present[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		for _, key := range missing {
			fmt.Printf("FAIL .env.example missing %s\n", key)
		}
		os.Exit(1)
	}
	fmt.Printf("OK env (%d settings documented in .env.example)\n", len(expected))
}

func getenvKeys(path string) map[string]bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config:", err)
		os.Exit(2)
	}
	keys := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "getenv" {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if key, err := strconv.Unquote(lit.Value); err == nil {
			keys[key] = true
		}
		return true
	})
	return keys
}

func exampleKeys(path string) map[string]bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read .env.example:", err)
		os.Exit(2)
	}
	keys := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}
		key, _, found := strings.Cut(stripped, "=")
		if !found {
			continue
		}
		keys[strings.TrimSpace(key)] = true
	}
	return keys
}

func rootDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot locate source file")
		os.Exit(2)
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
