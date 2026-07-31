package main

import (
	"fmt"
	"os"

	"httpFromScratch/server"
)

func main() {
	if err := server.Start(":8080"); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting server:", err)
		os.Exit(1)
	}
}
