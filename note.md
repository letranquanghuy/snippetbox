# Notes

## `go.mod`

`go.mod` tells the Go toolchain exactly which package versions should be used when you run commands like:

- `go run`
- `go test`
- `go build`

from your project directory.

The `// indirect` annotation means a package is required by your dependencies, but does not appear directly in any `import` statement in your own codebase.

## `go.sum`

`go.sum` contains cryptographic checksums representing the contents of required packages.

This file is not intended to be edited manually, and in most cases you do not need to open it. It provides two useful guarantees:

- Running `go mod verify` checks whether the checksums of downloaded packages on your machine match the entries in `go.sum`, helping ensure they have not been altered.
```
$ go mod verify
all modules verified
```

- If someone else downloads dependencies using `go mod download`, Go will report an error when there is any mismatch between downloaded package contents and the checksums listed in `go.sum`.

## Upgrading packages
Once a package has been downloaded and added to your go.mod file the package and
version are ‘fixed’. But there are many reasons why you might want to upgrade to use a
newer version of a package in the future.
To upgrade to latest available minor or patch release of a package, you can simply run
go get with the -u flag like so:
```
$ go get -u github.com/foo/bar
```

## Removing unused packages
Sometimes you might go get a package only to realize later that you don’t need it
anymore. When this happens you’ve got two choices.

You could either run go get and postfix the package path with @none, like so:
```
$ go get github.com/foo/bar@none
```

Or if you’ve removed all references to the package in your code, you can run go mod tidy, which will automatically remove any unused packages from your go.mod and go.sum files.
```
$ go mod tidy
```