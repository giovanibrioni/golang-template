# Example report

```markdown
# Review

## Blockers
| Severity | Location | Finding |
| --- | --- | --- |
| high | migrations/002_drop_owner.sql:1 | migration runs destructive SQL (DROP COLUMN) with no expand/contract |
| high | internal/items/service.go:22 | create reads by name then inserts, with no unique constraint |
| high | internal/items/service.go:30 | writes the row and publishes an event with no outbox or compensation |
| high | internal/items/service.go:18 | retries create, and the insert has no idempotency key |

## Warnings
| Severity | Location | Finding |
| --- | --- | --- |
| medium | internal/items/repository.go:40 | new filter on owner_id has no index in the migration |
| medium | internal/core/config/config.go:18 | REDIS_URL is not listed in .env.example |

## Script summary
- lint: pass
- architecture: pass
- orphans: pass
- env: fail
- merge-risk: pass (1 warning)

## Not in scope
- Events: no domain event contract
```
