package main

import (
	"os"

	"github.com/GoCodeAlone/workflow-plugin-crypto/internal/validatorreward"
)

func main() {
	os.Exit(validatorreward.Main(os.Args[1:], os.Stdout, os.Stderr))
}
