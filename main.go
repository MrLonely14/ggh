package main

import "github.com/MrLonely14/ggh/cmd"

// version is set via ldflags during build by GoReleaser
var version = "dev"

func main() {
	cmd.Main(version)
}
