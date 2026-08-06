//go:build !production

package webui

import (
	"embed"
	"io/fs"
)

//go:embed fallback/*
var developmentAssets embed.FS

func assets() (fs.FS, error) {
	return fs.Sub(developmentAssets, "fallback")
}
