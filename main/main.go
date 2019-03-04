// Package main contains the main function.
package main

import (
	"fmt"
	"os"

	"github.com/boreq/guinea"
	"github.com/boreq/plum/main/commands"
)

func main() {
	if err := guinea.Run(&commands.MainCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
