// Package db embeds SQL migrations so the binary can apply them without
// shipping the .sql files separately.
package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
