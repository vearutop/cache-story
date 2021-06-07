package storage

import (
	"embed"
)

// Migrations provide database migrations.
//go:embed migrations
var Migrations embed.FS
