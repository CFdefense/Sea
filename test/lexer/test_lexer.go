package test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/CFdefense/compiler/src/lexer"
)

const LEXER_TEST_DIR = "./test/lexer/tests/"

type TestCase struct {
	TestName        string   `json:"test_name"`
	TestDescription string   `json:"description"`
	TestContent     string   `json:"code"`
	ExpectedResult  []string `json:"result"`
}

type TestResult struct {
	TestCase TestCase
	Result   bool
	Expected []string
	Actual   []lexer.Token
}

// function to iterate over all lexer test cases
// will compare actual token stream results to expected
func RunLexerTests(debug bool) []TestResult {
	var test_results []TestResult
	l := lexer.InitializeLexer(debug)

	// get all lexer json test files
	files, err := os.ReadDir(LEXER_TEST_DIR)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	// iterate over all json files and extract
	for _, file := range files {
		// if file is a json file, process it
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			fullPath := filepath.Join(LEXER_TEST_DIR, file.Name())
			tests, err := process_json_file(fullPath)
			if err != nil {
				log.Printf("Error processing %s: %v", fullPath, err)
			}

			// execute tests and add results to test results
			for _, test := range tests {

				// reset lexer in between uses
				l.ResetLexer()

				// get results and compare to expected
				token_stream_result := l.GetTokenStream()

				result := compareTokenContent(token_stream_result, test.ExpectedResult)

				// Create a TestResult instance
				test_result := TestResult{
					TestCase: test,
					Result:   result,
					Expected: test.ExpectedResult,
					Actual:   token_stream_result,
				}

				// add test result to test results
				test_results = append(test_results, test_result)
			}
		}
	}

	return test_results
}

// compareTokenContent compares a slice of tokens with a slice of expected token content strings
func compareTokenContent(actual []lexer.Token, expected []string) bool {
	// compare token streamlengths
	if len(actual) != len(expected) {
		return false
	}

	// Compare token content after length check
	allMatch := true
	for i, token := range actual {
		actualContent := token.GetTokenContent()
		expectedContent := expected[i]
		match := actualContent == expectedContent
		if !match {
			allMatch = false
		}
	}

	return allMatch
}

// function to unmarshal json file into a slice of test cases
func process_json_file(fullPath string) ([]TestCase, error) {
	var tests []TestCase

	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file %s: %w", fullPath, err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&tests); err != nil {
		return nil, fmt.Errorf("failed to decode JSON in %s: %w", fullPath, err)
	}

	if len(tests) == 0 {
		log.Printf("Warning: no test cases found in %s. Possible format mismatch?", fullPath)
	}

	return tests, nil
}
