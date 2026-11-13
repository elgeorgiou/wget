// Package main provides the command-line entry point for the wget clone.
//
// It starts the application with the command-line arguments and connects
// standard output and standard error to the application's output streams.
//
// If the application returns an error, main prints it with an "ERROR:"
// prefix to standard error and exits with a non-zero status code.
package main

import (
	"fmt"
	"os"
	"wgetclone/internal/app"
)

const exitFailure = 1

// main starts the wget clone command-line application.
//
// It passes the command-line arguments, standard output, and standard error
// to app.Run. If an error occurs, it prints the error message to standard
// error and terminates the program with exit code 1.
//
// Parameters:
// - None.
//
// Returns:
// - None. The program exits with code 1 if app.Run returns an error.
func main() {
	err := app.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(exitFailure)
	}
}
