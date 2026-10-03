// Merge-risk signals on the diff against the base branch. Stdlib only.
//
// FAIL: destructive SQL in a changed migration, or model changes without a migration.
// WARN: handler, router, repository, model, or config files changed. Warnings do not fail the script.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var destructive = []string{
	"DROP TABLE",
	"DROP COLUMN",
	"DROP INDEX",
	"DROP CONSTRAINT",
	"DROP DEFAULT",
	"ALTER COLUMN",
	"RENAME TO",
	"RENAME COLUMN",
	"TRUNCATE",
}

func main() {
	root := rootDir()
	changed, scope := changedFiles(root)
	fmt.Printf("INFO scope: %s (%d file(s))\n", scope, len(changed))
	if len(changed) == 0 {
		fmt.Println("OK merge-risk (empty diff)")
		return
	}

	var failures, warnings []string
	var migrations, models, api []string
	for _, path := range changed {
		switch {
		case strings.HasPrefix(path, "migrations/") && strings.HasSuffix(path, ".sql"):
			migrations = append(migrations, path)
		case strings.HasPrefix(path, "internal/") &&
			(filepath.Base(path) == "model.go" || filepath.Base(path) == "repository.go"):
			models = append(models, path)
		}
		if (strings.HasPrefix(path, "internal/") && filepath.Base(path) == "handler.go") || path == "internal/api/router.go" {
			api = append(api, path)
		}
		if path == "internal/core/config/config.go" {
			warnings = append(warnings, "Config changed; confirm .env.example and deploy env: "+path)
		}
	}

	for _, path := range migrations {
		failures = append(failures, destructiveLines(root, path)...)
	}
	if len(models) > 0 && len(migrations) == 0 {
		failures = append(failures, "model files changed without a migration: "+strings.Join(models, ", "))
	}
	if len(api) > 0 {
		warnings = append(warnings, "API surface changed; check backward compatibility: "+strings.Join(api, ", "))
	}
	if len(models) > 0 {
		warnings = append(warnings, "query or model shape changed; confirm indexes for new filters and foreign keys: "+strings.Join(models, ", "))
	}

	fmt.Println("INFO no domain event contract in this diff")

	for _, item := range failures {
		fmt.Printf("FAIL %s\n", item)
	}
	for _, item := range warnings {
		fmt.Printf("WARN %s\n", item)
	}
	if len(failures) > 0 {
		os.Exit(1)
	}
	fmt.Println("OK merge-risk")
}

func destructiveLines(root, path string) []string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return nil
	}
	var failures []string
	for i, line := range strings.Split(string(data), "\n") {
		upper := strings.ToUpper(line)
		for _, token := range destructive {
			if strings.Contains(upper, token) {
				failures = append(failures, fmt.Sprintf("%s:%d migration runs destructive SQL (%s)", path, i+1, token))
				break
			}
		}
	}
	return failures
}

func changedFiles(root string) ([]string, string) {
	if !gitOK(root, "rev-parse", "--verify", "HEAD") {
		return uniqueSorted(gitOut(root, "ls-files", "--others", "--exclude-standard")), "untracked files (no commits yet)"
	}
	for _, branch := range []string{"main", "master"} {
		if !gitOK(root, "rev-parse", "--verify", branch) {
			continue
		}
		out, ok := gitOutOk(root, "merge-base", "HEAD", branch)
		if !ok || len(out) == 0 {
			continue
		}
		base := out[0]
		changed := append(
			gitOut(root, "diff", "--name-only", "--diff-filter=ACMRD", base),
			gitOut(root, "ls-files", "--others", "--exclude-standard")...,
		)
		return uniqueSorted(changed), fmt.Sprintf("diff against %s (%s)", branch, base[:12])
	}
	changed := append(
		gitOut(root, "diff", "--name-only", "--diff-filter=ACMRD", "HEAD"),
		gitOut(root, "ls-files", "--others", "--exclude-standard")...,
	)
	return uniqueSorted(changed), "working tree (no main/master base)"
}

func gitOut(root string, args ...string) []string {
	out, _ := gitOutOk(root, args...)
	return out
}

func gitOutOk(root string, args ...string) ([]string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, false
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, true
}

func gitOK(root string, args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	return cmd.Run() == nil
}

func uniqueSorted(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out
}

func rootDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot locate source file")
		os.Exit(2)
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
}
