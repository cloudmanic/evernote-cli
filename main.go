package main

import "github.com/cloudmanic/evernote-cli/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
