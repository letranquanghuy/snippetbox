package main

import (
	// New import
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time" // New import

	"github.com/alexedwards/scs/mysqlstore" // New import
	"github.com/alexedwards/scs/v2"         // New import
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/letranquanghuy/snippetbox/internal/models"
)

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	users          *models.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Make sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web
	// browser when a HTTPS connection is being used (and won't be sent over an
	// unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we
	// want the server to use. In this case the only thing that we're changing
	// is the curve preferences value, so that only elliptic curves with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:      *addr,
		Handler:   app.routes(),
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server.
		IdleTimeout: time.Minute,
		// Maximum time allowed server to read the entire request,
		// including the headers and body.
		ReadTimeout: 5 * time.Second,
		// Maximum time allowed server to write a response to the client.
		WriteTimeout: 10 * time.Second,
	}
	logger.Info(fmt.Sprintf("Starting server: https://localhost%s/", *addr), "addr", srv.Addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	// err = http.ListenAndServe(*addr, app.routes())

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
