package api

import (
	"io/fs"
	"net/http"
	"strings"
)

func SPA(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")

		if p == "" {
			p = "index.html"
		}

		if _, err := fs.Stat(fsys, p); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
