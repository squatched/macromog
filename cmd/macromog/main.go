package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	root, state := newRootCmd()
	root.SetArgs(args[1:])
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "macromog: "+err.Error())
		return 1
	}
	return state.code
}
