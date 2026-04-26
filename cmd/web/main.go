package main

import (
	"database/sql"
	"flag"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/letranquanghuy/snippetbox/internal/models"
)

type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

// Parsing the runtime configuration settings for the application;
// Establishing the dependencies for the handlers; and
// Running the HTTP server.
func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Define a new command-line flag for the MySQL Data Source Name string.
	// Format: username:password@tcp(host:port)/database?params
	dsn := flag.String("dsn", "web:1@tcp(localhost:3306)/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// log in JSON format with source information and debug level
	// logger := slog.New(slog.NewJSONHandler(os.Stdout,
	// 	&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))

	// log in text format with debug level but without source information
	// logger := slog.New(slog.NewTextHandler(os.Stdout,
	// 	&slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	// log into a file and also to the console
	// 0644 là permission chuẩn cho file log vì:
	// 6 (owner) — đọc + ghi
	// 4 (group) — chỉ đọc
	// 4 (others) — chỉ đọc
	// File log không cần ai ngoài owner ghi vào, nên 0644 là hợp lý và an toàn hơn.
	// 0666 thì group và others cũng ghi được — không cần thiết và kém an toàn hơn. Dùng cho file log thì quá rộng quyền.
	os.MkdirAll("log", 0755)
	logFile, err := os.OpenFile("log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	multi := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(multi, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))

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

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		logger:        logger,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
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
