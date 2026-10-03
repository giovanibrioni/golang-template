---
name: golang-template-review
description: >-
  Reviews this Go template against AGENTS.md: architecture checks, orphan
  packages, hidden bugs, security, and merge risk for handlers, migrations,
  indexes, and env vars. Use when the user asks for a code review, merge
  review, or golang-template-review on this repository.
disable-model-invocation: true
---

# Go template review

Review only. Do not edit code unless the user asks for fixes.

Read [reference.md](reference.md) before judging findings. Scripts are stdlib-only Go; execute them, do not reimplement them.

## Workflow

```
- [ ] 1. Architecture and lint
- [ ] 2. Bugs, orphans, ghost code
- [ ] 3. Security
- [ ] 4. Merge risk
- [ ] 5. Report
```

Diff scope: `branch changes` unless the user asks to review only uncommitted work. There is no commit yet when `git rev-parse HEAD` fails; the merge-risk script then uses untracked files.

### 1. Architecture and lint

Run, and keep going if a command fails:

```bash
make lint
make review
```

`make review` runs `architecture_check.go`, `orphan_check.go`, `env_example_check.go`, and `merge_risk_check.go` from `.cursor/skills/golang-template-review/scripts/`. Treat each `FAIL` line as a blocker. Save `WARN` and `INFO` lines for step 4.

### 2. Bugs, orphans, ghost code

Launch exactly one `bugbot` subagent (`run_in_background: false`, description `Bugbot`). It computes the diff. Prompt:

```text
Full Repository Path: <absolute repository path>
Diff: <branch changes | uncommitted changes>
Custom Instructions: Hidden bugs, unused code, and ghost code in this Go domain template. Commit and rollback belong only in httpx.Wrap, plus db migrations and pgtest cleanup. Services return domain errors with Kind() and PublicDetail(); handlers return httpx.Response and never write status codes themselves. Repository methods use db.Tx from the context. Flag logic bugs and dead code the deterministic scripts would miss. Flag race conditions: check-then-act and lost updates. Flag a dual write with no fallback, outbox, or compensation. Flag a retry around a non-idempotent write.
```

Add `Base Branch:` only when the review must compare against a branch other than the repository default.

If the subagent cannot compute the diff, retry once with `Diff: natural language` and a per-file change description. Do not retry a second time.

The orphan script already fails on internal packages never imported and on domains missing from the router. Do not repeat those hits unless Bugbot adds impact.

### 3. Security

Launch exactly one `security-review` subagent (`run_in_background: false`, description `Security Review`). Prompt:

```text
Full Repository Path: <absolute repository path>
Diff: <branch changes | uncommitted changes>
Custom Instructions: This service has no authentication in the template. Review injection, secret logging, SQL built by string concatenation instead of pgx parameters, and config that accepts URLs. slog only; no fmt.Print; do not log secrets.
```

Use the same diff scope as step 2. Retry once on invocation or subagent failure. Do not invent a natural-language fallback; this subagent only accepts `branch changes` or `uncommitted changes`.

### 4. Merge risk

Use the `merge_risk_check.go` output plus the checklist in [reference.md](reference.md). Read changed handlers, models, repositories, migrations, and `internal/core/config/config.go`.

This repository has no domain event bus. If the script prints `INFO no domain event contract in this diff`, say that events are not in scope.

`WARN` lines are findings, not script failures.

### 5. Report

Use this template. Sort findings by severity. Omit empty sections.

```markdown
# Review

## Blockers
| Severity | Location | Finding |
| --- | --- | --- |
| high | file:line | ... |

## Warnings
| Severity | Location | Finding |
| --- | --- | --- |
| medium | file:line | ... |

## Script summary
- lint: pass/fail
- architecture: pass/fail
- orphans: pass/fail
- env: pass/fail
- merge-risk: pass/fail (N warnings)

## Not in scope
- Events: no domain event contract / reviewed <paths>
```

Severity: `high` for bugs, security issues, breaking API changes, and destructive migration SQL. `medium` for missing indexes, missing env documentation, and architecture drift. `low` for ghost code with no caller.

Do not fix findings or rerun the review unless the user asks.
