package main

import (
	"net/http"
	"strings"

	"github.com/justinas/alice"
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

	// dynamic application routes. For now, this chain will only contain the
	// LoadAndSave session middleware but we'll add more to it later.
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Update these routes to use the new dynamic middleware chain followed by
	// the appropriate handler function. Note that because the alice ThenFunc()
	// method returns a http.Handler (rather than a http.HandlerFunc) we also
	// need to switch to registering the route using the mux.Handle() method.
	mux.Handle("GET /", dynamic.ThenFunc(app.home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	
	// Create a new middleware chain containing the middleware specific to our
	// protected routes.
	protected := dynamic.Append(app.requireAuthentication)
	mux.Handle("GET /snippet/create", protected.ThenFunc(app.snippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(app.snippetCreatePost))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	// Create a new middleware chain containing the middleware specific to our
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	// Return the 'standard' middleware chain followed by the servemux
	return standard.Then(mux)
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
