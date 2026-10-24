package main

import (
	"github.com/dpologdvinity/promptpet/internal/cli"
	"os"
)

var version = "dev"

func main() { os.Exit(cli.Run(os.Args[1:], cli.RealDeps(version))) }
