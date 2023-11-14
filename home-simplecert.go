package main

import (
	"fmt"
	"os"

	"github.com/jodydadescott/home-simplecert/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
