package main

import (
	"os"

	"nova/internal/cli"
)

func main() {
	app := cli.NewApp()
	code := app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(code)
}
