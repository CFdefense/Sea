package compiler

import "github.com/CFdefense/compiler/src/lexer"

// main compiler struct to hold all compiler components
type Compiler struct {
	lexer *lexer.Lexer
}

// compiler constructor
func InitializeCompiler() *Compiler {
	return &Compiler{
		lexer.InitializeLexer(),
	}
}

// function to initiate lexical analysis
func (c *Compiler) BeginLexicalAnalysis() {
	c.lexer.LexicalAnalysis()
}
