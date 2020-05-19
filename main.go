package main

import (
	"github.com/fairwindsops/nova/cmd"
)

var (
	// VERSION is set during build
	VERSION = "0.0.1"
)

func main() {
	cmd.Execute(VERSION)
}
