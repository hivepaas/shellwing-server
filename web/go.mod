// Not Go: a module of its own only so that the go tool (go test ./...,
// golangci-lint, govulncheck) leaves the web app out of the server's module,
// with the Go files some npm packages carry in node_modules.
module github.com/hivepaas/shellwing-server/web
