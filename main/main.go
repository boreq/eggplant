// Package main contains the main function.
package main

import (
	"fmt"
	"github.com/boreq/plum/main/commands"
	"github.com/boreq/guinea"
	"os"
)

func main() {
	e := guinea.Run(&commands.MainCmd)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
