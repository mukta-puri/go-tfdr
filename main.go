package main

import "github.com/tyler-technologies/go-terraform-state-copy/cmd"

var version = "devbuild"

func main() {
	cmd.Execute(version)
}
