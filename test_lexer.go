package main

import (
	"fmt"

	"github.com/CFdefense/compiler/src/lexer"
)

func main() {
	l := lexer.InitializeLexer(true)
	l.SetContent(map[string]string{"test.txt": "asm { mov %rax, %rbx }"})

	l.Analyze()

	tokens := l.GetTokenStream()
	fmt.Println("Tokens:")
	for i, token := range tokens {
		fmt.Printf("  %d: {type: %s, content: '%s'}\n", i, token.GetTokenType().String(), token.GetTokenContent())
	}
}
