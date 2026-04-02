package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type application struct {
	logger *slog.Logger
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	// logger := slog.New(slog.NewJSONHandler(os.Stdout,
	// 	&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	logger := slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	
	app := &application{
		logger: logger,
	}
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", neuter(fileServer)))
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	logger.Debug("starting server", "addr", *addr)

	err := http.ListenAndServe(*addr, mux)

	logger.Error("server error", "err", err.Error())
	os.Exit(1)
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
