package main

import (
	"fmt"
	"os"

	"github.com/sushantvema-harper/brek-go"
)

var exitFunc = os.Exit

func main() {
	if err := brek.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitFunc(1)
	}
}
