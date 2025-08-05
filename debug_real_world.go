package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/CFdefense/compiler/src/lexer"
)

type DebugTestCase struct {
	TestName string `json:"test_name"`
	Code     string `json:"code"`
	Result   []struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"result"`
}

func main() {
	// Read the test case
	data, err := os.ReadFile("test/lexer/tests/real_world_program_tests.json")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	var tests []DebugTestCase
	err = json.Unmarshal(data, &tests)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	if len(tests) == 0 {
		fmt.Println("No tests found")
		return
	}

	test := tests[0]
	fmt.Printf("Test: %s\n", test.TestName)
	fmt.Printf("Code length: %d characters\n", len(test.Code))
	fmt.Printf("Expected tokens: %d\n", len(test.Result))

	// Run the lexer
	l := lexer.InitializeLexer(false)
	l.SetContent(map[string]string{"test.txt": test.Code})
	l.LexicalAnalysis("")
	tokens := l.GetTokenStream()

	fmt.Printf("Actual tokens: %d\n", len(tokens))

	// Compare tokens
	fmt.Println("\n=== TOKEN COMPARISON ===")

	expectedCount := len(test.Result)
	actualCount := len(tokens)

	fmt.Printf("Expected: %d tokens\n", expectedCount)
	fmt.Printf("Actual: %d tokens\n", actualCount)
	fmt.Printf("Difference: %d tokens\n", actualCount-expectedCount)

	// Show first 20 tokens from both
	fmt.Println("\n=== FIRST 20 EXPECTED TOKENS ===")
	for i := 0; i < 20 && i < len(test.Result); i++ {
		fmt.Printf("%2d: {type: %s, content: %s}\n", i, test.Result[i].Type, test.Result[i].Content)
	}

	fmt.Println("\n=== FIRST 20 ACTUAL TOKENS ===")
	for i := 0; i < 20 && i < len(tokens); i++ {
		fmt.Printf("%2d: {type: %s, content: %s}\n", i, tokens[i].GetTokenType(), tokens[i].GetTokenContent())
	}

	// Show last 20 tokens from both
	fmt.Println("\n=== LAST 20 EXPECTED TOKENS ===")
	start := len(test.Result) - 20
	if start < 0 {
		start = 0
	}
	for i := start; i < len(test.Result); i++ {
		fmt.Printf("%2d: {type: %s, content: %s}\n", i, test.Result[i].Type, test.Result[i].Content)
	}

	fmt.Println("\n=== LAST 20 ACTUAL TOKENS ===")
	start = len(tokens) - 20
	if start < 0 {
		start = 0
	}
	for i := start; i < len(tokens); i++ {
		fmt.Printf("%2d: {type: %s, content: %s}\n", i, tokens[i].GetTokenType(), tokens[i].GetTokenContent())
	}

	// Find first mismatch
	fmt.Println("\n=== FINDING FIRST MISMATCH ===")
	minLen := len(test.Result)
	if len(tokens) < minLen {
		minLen = len(tokens)
	}

	for i := 0; i < minLen; i++ {
		expected := test.Result[i]
		actual := tokens[i]

		if expected.Type != actual.GetTokenType().String() || expected.Content != actual.GetTokenContent() {
			fmt.Printf("Mismatch at token %d:\n", i)
			fmt.Printf("  Expected: {type: %s, content: %s}\n", expected.Type, expected.Content)
			fmt.Printf("  Actual:   {type: %s, content: %s}\n", actual.GetTokenType(), actual.GetTokenContent())

			// Show context around the mismatch
			fmt.Printf("\nContext around token %d:\n", i)
			start := i - 5
			if start < 0 {
				start = 0
			}
			end := i + 5
			if end > len(test.Result) {
				end = len(test.Result)
			}

			fmt.Printf("Expected tokens %d-%d:\n", start, end-1)
			for j := start; j < end; j++ {
				fmt.Printf("  %2d: {type: %s, content: %s}\n", j, test.Result[j].Type, test.Result[j].Content)
			}

			if end > len(tokens) {
				end = len(tokens)
			}
			fmt.Printf("Actual tokens %d-%d:\n", start, end-1)
			for j := start; j < end; j++ {
				fmt.Printf("  %2d: {type: %s, content: %s}\n", j, tokens[j].GetTokenType(), tokens[j].GetTokenContent())
			}
			break
		}
	}

	// If actual has more tokens, show the extra ones
	if len(tokens) > len(test.Result) {
		fmt.Println("\n=== EXTRA TOKENS ===")
		for i := len(test.Result); i < len(tokens); i++ {
			fmt.Printf("%2d: {type: %s, content: %s}\n", i, tokens[i].GetTokenType(), tokens[i].GetTokenContent())
		}
	}

	// If expected has more tokens, show the missing ones
	if len(test.Result) > len(tokens) {
		fmt.Println("\n=== MISSING TOKENS ===")
		for i := len(tokens); i < len(test.Result); i++ {
			fmt.Printf("%2d: {type: %s, content: %s}\n", i, test.Result[i].Type, test.Result[i].Content)
		}
	}
}
