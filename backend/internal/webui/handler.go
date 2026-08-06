package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func Handler() http.Handler {
	content, err := assets()
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(content))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested != "." {
			if info, statErr := fs.Stat(content, requested); statErr == nil && !info.IsDir() {
				files.ServeHTTP(response, request)
				return
			}
		}
		request.URL.Path = "/"
		files.ServeHTTP(response, request)
	})
}
