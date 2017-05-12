package main

import (
	"log"
	"os"

	"github.com/bamchoh/pollydent"
)

func main() {
	f, _ := os.Create("pollydent.log")
	logger := log.New(f, "pollydent:", 0)
	p, _ := pollydent.NewPolly(logger, "pollydent.yml")
	p.ReadAloud("こんにちは世界")
}
