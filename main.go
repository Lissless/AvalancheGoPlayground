package main

import (
	"fmt"

	"main.go/container"
)

func main() {
	blockID, success := container.GetBlockIDFromIndex(1368067)
	if success {
		fmt.Println("Block ID: ", blockID)
	} else {
		fmt.Println("Error: ", blockID)
	}
}
