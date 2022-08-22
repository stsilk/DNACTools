package main

import (
	"fmt"

	dnaclib "github.com/stsilk/DNACTools/dnaclib"
)

func main() {
	client, err := dnaclib.LoginToDNAC()
	if err != nil {
		panic(err)
	}
	result, err := dnaclib.RenameDevice(client)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
