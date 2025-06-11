package test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/username/go-compiler/src/lexer"
)

const LEXER_TEST_DIR = "./lexer/tests/"

type TestCase struct {
	TestName        string        `json:"test_name"`
	TestDescription string        `json:"test_description"`
	TestContent     string        `json:"test_content"`
	ExpectedResult  []lexer.Token `json:"expected_result"`
}

type TestResult struct {
	TestCase TestCase
	Result   bool
	Expected []lexer.Token
	Actual   []lexer.Token
}

// function to iterate over all lexer test cases
// will compare actual token stream results to expected
func RunLexerTests() {
	var test_results []TestResult

	// get all lexer json test files
	files, err := ioutil.ReadDir(LEXER_TEST_DIR)
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
				lexer_ctx := lexer.CreateLexer(test.TestContent)

				// get results and compare to expected
				token_stream_result := lexer_ctx.GetTokenStream()
				result := slices.Equal(token_stream_result, test.ExpectedResult)

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
}

// function to unmarshal json file into a slice of test cases
func process_json_file(fullPath string) ([]TestCase, error) {
	var tests []TestCase

	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	// decode json file into slice of test cases
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(&tests)

	return tests, err
}
