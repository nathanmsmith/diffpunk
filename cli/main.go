package main

import (
	"os"
)

var version = "dev"

// Release builds disable Vite integration through -X main.development=false.
var development = "true"

func main() { os.Exit(run()) }
