package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: settle <payments.csv>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open %s\n", os.Args[1])
		os.Exit(1)
	}
	defer file.Close() // runs when main returns, Go's answer to RAII

	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()

	if err := newProcessor().run(file, output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
