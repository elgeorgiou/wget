package main

import (
	"fmt"
	"os"
	"wgetclone/internal/app"
)

const exitFailure = 1

func main() {
	err := app.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}
}
