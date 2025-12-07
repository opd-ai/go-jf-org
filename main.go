package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	fmt.Printf("go-jf-org v%s\n", version)
	fmt.Println("A tool to organize media files for Jellyfin server")
	fmt.Println()
	fmt.Println("This is a development version.")
	fmt.Println("See IMPLEMENTATION_PLAN.md for the full development roadmap.")
	os.Exit(0)
}
