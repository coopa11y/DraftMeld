//go:build production

package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var productionAssets embed.FS

func assets() (fs.FS, error) {
	return fs.Sub(productionAssets, "dist")
}
