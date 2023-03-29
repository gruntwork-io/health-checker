package main

import (
	"github.com/gruntwork-io/go-commons/entrypoint"
	"github.com/gruntwork-io/health-checker/commands"
)

// This variable is set at build time using -ldflags parameters. For example, we typically set this flag in circle.yml
// to the latest Git tag when building our Go apps:
//
// build-go-binaries --app-name my-app --dest-path bin --ld-flags "-X main.VERSION=$CIRCLE_TAG"
//
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string

// This is the main entry point for the app.
func main() {
	app := commands.CreateCli(VERSION)
	entrypoint.RunApp(app)
}
