package main

import (
	"fmt"
	"os"

	"github.com/janmz/masseneinladung/cmd/masseneinladung/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
