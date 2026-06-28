package main

import (
	"os"

	"github.com/uesugitorachiyo/ao-atlas/internal/atlas"
)

func main() {
	os.Exit(atlas.Run(os.Args[1:], os.Stdout, os.Stderr))
}
