package main

import (
	selfBuild "github.com/tokenize-x/tx-callisto/build"
	"github.com/tokenize-x/tx-crust/build"
)

func main() {
	build.Main(selfBuild.Commands)
}
