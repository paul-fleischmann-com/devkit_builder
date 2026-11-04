// Package templates embeds the reference devkit that `devkit init` copies.
package templates

import "embed"

// FS holds the tree below reference-devkit/.
//
//go:embed all:reference-devkit
var FS embed.FS
