package main

import (
	"example/internal/cmd"
	_ "example/internal/logic"
)

func main() {
	cmd.Main.Run()
}
