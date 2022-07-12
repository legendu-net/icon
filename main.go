/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"legendu.net/icon/cmd"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cmd.Execute()
}
