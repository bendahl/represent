package main

import (
	"flag"
	"fmt"
	"log"
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
		handleRenderCmd(args[1:])
	case "present":
		handlePresent(args[1:])
	case "version":
		printVersion()
	default:
		panic("ERROR: Not implemented - This should be unreachable")
	}
}

func handlePresent(args []string) {
	fl := flag.NewFlagSet("present", flag.ExitOnError)
	var inputMd string
	fl.StringVar(&inputMd, "in", "slides.md", "input markdown filename")
	var host string
	fl.StringVar(&host, "host", "localhost", "hostname to listen on")
	var port int
	fl.IntVar(&port, "port", 8080, "port to listen on")
	err := fl.Parse(args)
	if err != nil {
		log.Fatalln("Failed to process program arguments.")
	}
	err = startServer(inputMd, host, port)
	if err != nil {
		log.Fatalf("Error running server: %v\n", err)
	}
}

func handleRenderCmd(args []string) {
	fl := flag.NewFlagSet("render", flag.ExitOnError)
	var outDir string
	fl.StringVar(&outDir, "out", "slides", "output directory name")
	var inputMd string
	fl.StringVar(&inputMd, "in", "slides.md", "input markdown filename")
	err := fl.Parse(args)
	if err != nil {
		log.Fatalln("Failed to process program arguments.")
	}
	err = generateSlides(outDir, inputMd)
	if err != nil {
		log.Fatalf("Failed to generate slides: %v\n", err)
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
