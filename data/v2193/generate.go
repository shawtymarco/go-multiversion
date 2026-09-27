//go:build ignore

// Run from the repository root: go run data/v2193/generate.go -source <dragonfly-checkout>
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	source := flag.String("source", "", "Dragonfly Git checkout")
	out := flag.String("out", "data/v2193/block_states.nbt", "snapshot output")
	flag.Parse()
	if *source == "" {
		panic("-source is required")
	}
	data, err := exec.Command("git", "-C", *source, "show", "4c7b5074be94fa83a1cd98e9c752083ad04a6e21:server/world/block_states.nbt").Output()
	if err != nil {
		panic(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "f0784a6284d6ca7d98cc3472f4ce84241a11e11b18ed16f6591dfd5e6da6fbd6" {
		panic("source checksum mismatch")
	}
	if err := os.WriteFile(*out, data, 0644); err != nil {
		panic(err)
	}
}
