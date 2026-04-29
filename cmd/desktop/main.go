package main

import (
	"fmt"
	"os"
	"github.com/localdev/minux-mesh/core"
)

func main() {
	node, err := core.NewNode()
	if err != nil {
		fmt.Println("Failed to start node:", err)
		os.Exit(1)
	}
	defer node.Close()

	node.Run()
}
