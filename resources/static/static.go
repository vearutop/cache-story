// Package static provides embedded static assets.
package static

import (
	"embed"
)

// Assets provides embedded static assets for web application.
//
//go:embed *
var Assets embed.FS
