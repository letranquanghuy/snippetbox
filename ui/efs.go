package ui

import (
	"embed"
)

// This looks like a comment (below), but it is actually a special comment directive. When our
// application is compiled (as part of either go build or go run ), this comment directive
// instructs Go to store the files from our ui/static folder in an embedded filesystem
// referenced by the global variable Files.
// |
// v

//go:embed "html" "static"
var Files embed.FS
