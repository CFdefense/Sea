package test

import (
	"fmt"

	lexer_test "github.com/CFdefense/compiler/test/lexer"
)

// function to run all tests
func RunTests() {
	lexer_tests := lexer_test.RunLexerTests()

	fmt.Println("Lexer Tests:")
	fmt.Println("--------------------------------")
	for _, test := range lexer_tests {
		if test.Result {
			fmt.Println(test.TestCase.TestName, "PASSED")
		} else {
			fmt.Println(test.TestCase.TestName, "FAILED")
			fmt.Println("Expected:", test.Expected)
			fmt.Println("Actual:", test.Actual)
		}
		fmt.Print("--------------------------------\n")
	}
}
