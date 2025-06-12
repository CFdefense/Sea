package main

import (
	"flag"

	"github.com/CFdefense/compiler/src/compiler"
	"github.com/CFdefense/compiler/test"
)

// main entry point for the compiler
func main() {
	// parse command line argumentsd
	run_tests := flag.Bool("test", false, "Run all compiler tests")
	target_path := flag.String("path", "", "Directory to compile")
	flag.Parse()

	// optionally run tests
	if *run_tests {
		test.RunTests()
	}

	// create the compiler ctx
	compiler_ctx := compiler.InitializeCompiler()

	// start lexical analysis
	compiler_ctx.BeginLexicalAnalysis(*target_path)

	// TODO add next steps

}
