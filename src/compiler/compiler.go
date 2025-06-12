package compiler

import "github.com/CFdefense/compiler/src/lexer"

// main compiler struct to hold all compiler components
// TODO add the rest of the components
type Compiler struct {
	lexer *lexer.Lexer
}

// compiler constructor
func InitializeCompiler() *Compiler {
	return &Compiler{
		lexer.InitializeLexer(),
	}
}

// function to initiate lexical
// TODO: pass necessary path information
func (c *Compiler) BeginLexicalAnalysis(path string) {
	c.lexer.LexicalAnalysis(path)
}
