package main

//go:generate go run install_tools.go

import "github.com/ppowo/zzk/cmd"

func main() {
	cmd.Execute()
}