package main

import (
	"flag"

	"github.com/CFdefense/compiler/src/compiler"
	"github.com/CFdefense/compiler/test"
)

// main entry point for the compiler
func main() {
	// parse command line argumentsd
	runTests := flag.Bool("test", false, "Run all compiler tests")
	flag.Parse()

	// optionally run tests
	if *runTests {
		test.RunTests()
	}

	// create the compiler ctx
	compiler_ctx := compiler.InitializeCompiler()

	// start lexical analysis
	compiler_ctx.BeginLexicalAnalysis()

	// TODO add next steps

}
