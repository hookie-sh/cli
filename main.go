package main

import (
	"fmt"
	"os"

	"github.com/hookie-sh/cli/cmd"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			fmt.Println(cmd.VersionString())
			os.Exit(0)
		}
	}
	cmd.Execute()
}

