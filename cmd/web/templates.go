package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/letranquanghuy/snippetbox/internal/models"
)

// Define a templateData type to act as the holding structure for
// any dynamic data that we want to pass to our HTML templates.
// At the moment it only contains one field, but we'll add more
// to it as the build progresses.
type templateData struct {
	CurrentYear int
	Snippet     models.Snippet
	Snippets    []models.Snippet
}

// Create a humanDate fucntion which returns a nicely formatted string representation of a time.
// Time object.
// The format string "02 Jan 2006 at 15:04" is a specific layout that Go uses to format time values. 
// It can not be changed to any other value, 
// as it is used as a reference point for Go to understand how to format the time.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}
	// Use the filepath.Glob() function to get a slice of all filepaths that
	// match the pattern "./ui/html/pages/*.tmpl". This will essentially gives
	// us a slice of all the filepaths for our application 'page' templates
	// like: [ui/html/pages/home.tmpl ui/html/pages/view.tmpl]
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.tmpl') from the full filepath
		// and assign it to the name variable.
		name := filepath.Base(page)

		// // option: 1 hardcode the list of files to parse for each page template
		// // Create a slice containing the filepaths for our base template, any
		// // partials and the page.
		// files := []string{
		// 	"./ui/html/base.tmpl",
		// 	"./ui/html/partials/nav.tmpl",
		// 	page,
		// }
		// // Parse the files into a template set.
		// ts, err := template.ParseFiles(files...)

		// option: 2 use filepath.Glob() to get the list of files to parse for each page template
		// Create a slice containing the filepaths for our base template, any
		// partials and the page. Again, we use filepath.Glob() to get the
		// filepaths for the partials, which gives us flexibility to add more
		// partials in the future without needing to change this code.

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.tmpl') as the key.
		cache[name] = ts
	}
	// // Print key and value of the map to verify that the templates have been parsed and added to the cache correctly.
	// for k, v := range cache {
	// 	fmt.Printf("key: %s, value: %v\n", k, v)
	// }

	// Return the map.
	return cache, nil
}
