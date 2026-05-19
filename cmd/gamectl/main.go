package main

import (
	"os"

	"go-demo/internal/gamectl"
)

func main() {
	os.Exit(gamectl.Run(os.Args[1:], os.Stdout, os.Stderr))
}
