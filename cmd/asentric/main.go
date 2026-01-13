// Package main is the entry point for the asentric CLI.
package main

import (
	"os"

	"github.com/asentric/asentric/cmd/asentric/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
