package main

import (
	"fmt"

	"github.com/Acketuk/Gator/internal/config"
)

func main() {

	config, err := config.Read()
	if err != nil {
		fmt.Printf("\tError: %s\n", err)
	}

	fmt.Println(*config)
}
