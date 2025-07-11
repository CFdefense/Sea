package test

import (
	"fmt"

	lexer_test "github.com/CFdefense/compiler/test/lexer"
)

// function to run all tests
func RunTests(debug bool) {
	lexer_tests := lexer_test.RunLexerTests(debug)

	fmt.Println("Lexer Tests:")
	fmt.Println("--------------------------------")
	for _, test := range lexer_tests {
		if test.Result {
			fmt.Println(test.TestCase.TestName, "PASSED")
		} else {
			fmt.Println(test.TestCase.TestName, "FAILED")
			fmt.Println("Expected:", test.Expected)
			fmt.Println("Actual:")
			for _, token := range test.Actual {
				fmt.Printf("  {type: %s, content: %s}\n", token.GetTokenType().String(), token.GetTokenContent())
			}
		}
		fmt.Print("--------------------------------\n")
	}
}
