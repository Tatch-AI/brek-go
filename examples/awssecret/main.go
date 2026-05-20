package main

import (
	"fmt"

	"github.com/sushantvema-harper/brek-go"
)

var getConfig = brek.GetConfig
var printf = fmt.Printf

func main() {
	brek.SetLoaders(brek.DefaultLoaders())

	conf, err := getConfig()
	if err != nil {
		panic(err)
	}

	printf("service=%v\n", conf["service"])
}
