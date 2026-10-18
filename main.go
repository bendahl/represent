package main

import (
	"fmt"
	"os"
)

var version string
var commit string

var commands = map[string]string{
	"render":  "Render HTML slides from Markdown",
	"present": "Run presentation web server",
	"version": "Print version information",
}

func main() {
	args := os.Args[1:]
	numArgs := len(args)

	if numArgs < 1 {
		fmt.Printf("Expected at least two arguments, got %v\n", numArgs)
		fmt.Println()
		usage()
		os.Exit(1)
	}

	command := args[0]
	if !isValidCommand(command) {
		fmt.Printf("Unknown command \"%s\"\n", command)
		fmt.Println()
		usage()
		os.Exit(1)
	}

	switch command {
	case "render":
		panic("ERROR: Not implemented")
	case "present":

		panic("ERROR: Not implemented")
	case "list":

		panic("ERROR: Not implemented")
	case "version":
		printVersion()
	default:
		panic("ERROR: Not implemented - This should be unreachable")
	}
}

func isValidCommand(cmd string) bool {
	_, ok := commands[cmd]
	return ok
}

func usage() {
	fmt.Println("Usage: represent <command> <args>")
	fmt.Println()
	fmt.Println("Available commands:")
	for c, description := range commands {
		fmt.Printf("	%s - %s\n", c, description)
	}
}

func printVersion() {
	if version == commit {
		version = "dev"
	}
	fmt.Printf("Version: %s\nCommit Hash: %s\n\n", version, commit)
}
