// Package migrations bundles all SQL migration files into the binary so they
// can be applied at startup without shipping the source tree separately.
package migrations

import "embed"

// FS embeds every *.sql file located beside this Go file.
//
//go:embed *.sql
var FS embed.FS
