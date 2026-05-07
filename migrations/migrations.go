// Package migrations embeds the versioned SQL migration files so they are
// compiled into the binary and work regardless of the working directory.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
