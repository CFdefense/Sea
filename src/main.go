package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/CFdefense/compiler/src/compiler"
	"github.com/CFdefense/compiler/test"
)

func main() {
	// Parse command-line flags
	runTests := flag.String("test", "", "Run compiler test suite (e.g., lexer)")
	targetPath := flag.String("path", "", "Directory to compile")
	flag.Parse()

	// Run tests if requested
	if *runTests != "" {
		switch *runTests {
		case "lexer":
			test.RunTests()
		default:
			fmt.Fprintf(os.Stderr, "Unknown test target: %s\n", *runTests)
			os.Exit(1)
		}
		fmt.Println("Tests Completed -> Exiting...")
		return
	}

	// Ensure a path is provided if were not running tests
	if *runTests == "" && *targetPath == "" {
		fmt.Fprintln(os.Stderr, "Error: no path provided. Use -path to specify source directory.")
		flag.Usage()
		os.Exit(1)
	}

	// Begin compilation process
	fmt.Printf("Compiling project at: %s\n", *targetPath)

	// create the compiler ctx
	compiler_ctx := compiler.InitializeCompiler()

	// start lexical analysis
	compiler_ctx.BeginLexicalAnalysis(*targetPath)

	// TODO: Next compiler steps
}
