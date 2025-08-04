package test

import (
	"fmt"

	lexer_test "github.com/CFdefense/compiler/test/lexer"
)

// function to run all tests
func RunTests(debug bool) {
	lexer_tests := lexer_test.RunLexerTests(debug)

	fmt.Println("Lexer Tests:")
	fmt.Println("==========================================")

	passed := 0
	failed := 0

	for _, test := range lexer_tests {
		if test.Result {
			fmt.Printf("%s PASSED\n", test.TestCase.TestName)
			passed++
		} else {
			fmt.Printf("%s FAILED\n", test.TestCase.TestName)
			fmt.Printf("   Description: %s\n", test.TestCase.TestDescription)
			fmt.Printf("   Input: %s\n", test.TestCase.TestContent)
			fmt.Printf("   Error: %s\n", test.Error)
			fmt.Printf("   Expected: %d tokens\n", len(test.Expected))
			fmt.Printf("   Actual: %d tokens\n", len(test.Actual))

			// Show first few tokens for debugging
			if len(test.Actual) > 0 {
				fmt.Printf("   First few actual tokens:\n")
				for i, token := range test.Actual {
					if i >= 5 { // Limit to first 5 tokens
						fmt.Printf("     ... and %d more\n", len(test.Actual)-5)
						break
					}
					fmt.Printf("     %d: {type: %s, content: %s}\n",
						i, token.GetTokenType().String(), token.GetTokenContent())
				}
			}
			failed++
		}
		fmt.Print("------------------------------------------\n")
	}

	fmt.Printf("\nTest Summary: %d passed, %d failed\n", passed, failed)
	if failed > 0 {
		fmt.Println("Some tests failed - check the lexer implementation")
	} else {
		fmt.Println("All tests passed!")
	}
}
