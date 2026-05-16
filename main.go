package main

import "github.com/aholbreich/taskledger/cmd"

// version is injected at build time via -ldflags "-X main.version=<value>".
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
