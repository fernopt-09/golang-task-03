package migrations

import "embed"

// FS contains all SQL migrations, they are built into the binary.
//
//go:embed *.sql
var FS embed.FS
