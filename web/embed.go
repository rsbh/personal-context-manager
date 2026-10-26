// Package web holds the built UI. npm run build writes it to dist/app.
package web

import (
	"embed"
	"io/fs"
)

// dist/.gitkeep is in git, so the embed works without a UI build.
//
//go:embed all:dist
var dist embed.FS

// FS returns the built app. Without a build it has no index.html, and the server shows a placeholder.
func FS() fs.FS {
	sub, err := fs.Sub(dist, "dist/app")
	if err != nil {
		return dist
	}
	return sub
}
