package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger *slog.Logger
}


// Parsing the runtime configuration settings for the application;
// Establishing the dependencies for the handlers; and
// Running the HTTP server.
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

	logger.Debug("starting server", "addr", *addr)

	err := http.ListenAndServe(*addr, app.routes())

	logger.Error("server error", "err", err.Error())
	os.Exit(1)
}
