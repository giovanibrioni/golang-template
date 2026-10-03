// Package migrations holds SQL files applied in lexical order.
package migrations

import "embed"

// Files contains the versioned SQL migrations.
//
//go:embed *.sql
var Files embed.FS
