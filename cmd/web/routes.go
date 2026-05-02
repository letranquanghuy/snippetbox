package main

import (
	"net/http"
	"strings"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// Route:       GET /staticall/css/main.css
	// Sau StripPrefix("/static", ...):  /all/css/main.css  ← sai!
	// FileServer tìm file:  ./ui/static/all/css/main.css   ← không tồn tại → 404
	mux.Handle("GET /static/", http.StripPrefix("/static", neuter(fileServer)))

	// Don't use neuter for this route, because we want to allow directory listings
	// for the /static-all/ path.
	// Note: if a directory contains an index.html file, FileServer will serve that
	// file instead of showing the directory listing.
	mux.Handle("GET /static-all/", http.StripPrefix("/static-all", fileServer))
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)
	return app.logRequest(commonHeaders(mux))
}

// neuter is a middleware that prevents directory listings by returning a 404
// Not Found response for any request that ends with a slash.
func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
