package compiler

import (
	debugger "github.com/CFdefense/compiler/src/debug"
	"github.com/CFdefense/compiler/src/lexer"
)

// main compiler struct to hold all compiler components
// TODO add the rest of the components
type Compiler struct {
	lexer *lexer.Lexer
	debug *debugger.Debug
}

// compiler constructor
// can have a debug mode for verbose outputs
func InitializeCompiler(debug bool) *Compiler {
	return &Compiler{
		lexer: lexer.InitializeLexer(debug),
		debug: debugger.InitializeDebugger("CMP", debug),
	}
}

// function to initiate lexical
func (c *Compiler) BeginLexicalAnalysis(path string) {
	c.lexer.LexicalAnalysis(path)
}
