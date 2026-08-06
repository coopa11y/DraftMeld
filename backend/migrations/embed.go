package migrations

import "embed"

// Files contains the versioned DraftMeld database migrations.
//
//go:embed *.sql
var Files embed.FS
