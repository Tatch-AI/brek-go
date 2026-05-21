package brek

import (
	"fmt"
	"os"
)

func debug(args ...any) {
	if os.Getenv("BREK_DEBUG") == "" {
		return
	}

	fmt.Println(append([]any{"[BREK][DEBUG]"}, args...)...)
}
