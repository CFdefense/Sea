package main

import (
	"flag"
	"log"
	"os"

	"github.com/CFdefense/compiler/src/compiler"
	"github.com/CFdefense/compiler/test"
)

func main() {
	// initialize command-line flags
	runTests := flag.String("test", "", "Run compiler test suite (e.g., lexer)")
	targetPath := flag.String("path", "", "Directory to compile")
	debugMode := flag.Bool("debug", false, "Enable verbose debug mode")

	// parse the inputted command-line flags
	flag.Parse()

	// run tests if requested
	// TODO add other component tests
	if *runTests != "" {
		switch *runTests {
		case "lexer":
			test.RunTests(*debugMode)
		default:
			log.Printf("Unknown test target: %s\n", *runTests)
			os.Exit(1)
		}
		log.Println("Tests Completed -> Exiting...")
		return
	}

	// ensure a path is provided if were not running tests
	if *runTests == "" && *targetPath == "" {
		log.Println("Error: no path provided. Use -path to specify source directory.")
		flag.Usage()
		os.Exit(1)
	}

	// begin compilation process
	log.Printf("Compiling project at: %s\n", *targetPath)

	// create the compiler ctx
	compiler_ctx := compiler.InitializeCompiler(*debugMode)

	// start lexical analysis
	compiler_ctx.BeginLexicalAnalysis(*targetPath)

	// TODO: Next compiler steps
}
