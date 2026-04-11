package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/letranquanghuy/snippetbox/internal/models"
)

type application struct {
	logger   *slog.Logger
	snippets *models.SnippetModel
}

// Parsing the runtime configuration settings for the application;
// Establishing the dependencies for the handlers; and
// Running the HTTP server.
func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:1@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// logger := slog.New(slog.NewJSONHandler(os.Stdout,
	// 	&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	logger := slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	app := &application{
		logger: logger,
		snippets: &models.SnippetModel{DB: db},
	}

	logger.Debug("starting server", "addr", *addr)

	err = http.ListenAndServe(*addr, app.routes())

	logger.Error("server error", "err", err.Error())
	os.Exit(1)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
