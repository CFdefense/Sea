package lexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	debugger "github.com/CFdefense/compiler/src/debug"
)

// lexer context object
type Lexer struct {
	token_stream []Token
	content      map[string]string
	row          int
	col          int
	debug        *debugger.Debug
	tokenNFAs    []*NFA // store NFA tokens
}

// lexer object constructor
func InitializeLexer(debug bool) *Lexer {
	return &Lexer{
		token_stream: []Token{},
		content:      make(map[string]string),
		row:          1,
		col:          1,
		debug:        debugger.InitializeDebugger("LEX", debug),
		tokenNFAs:    []*NFA{}, // initialize empty NFA slice
	}
}

// function responsible for all things lexical analysis
func (l *Lexer) LexicalAnalysis(path string) {

	// begin lexer scan of target file
	// might want to change this to target directory
	// depending on future limits/scope
	// if were testing, we can skip this as content is already set
	if path != "" {
		l.Scan(path)
	}

	// begin lexer analyze
	// this will turn our scanned content into
	// a nice stream of tokens
	l.Analyze()
}

// function to allow the lexer to scan
// going to implement directory based
// this will allow future scaling to
// 'project-compilation' vs 'file-compilation'
func (l *Lexer) Scan(path string) {
	// open target directory
	l.debug.DebugLog(fmt.Sprintf("Beginning Lexer Scan on Directory: %v", path), false)
	files, err := os.ReadDir(path)

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			fileNames = append(fileNames, file.Name())
		}
	}
	l.debug.DebugLog(fmt.Sprintf("Found Files: %v", fileNames), false)

	if err != nil {
		l.debug.DebugLog(fmt.Sprintf("Failed to read directory: %v", err), true)
	}

	for _, file := range files {
		// open the file and read entire file contents
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			// Use filepath.Join to create the full path
			fullPath := filepath.Join(path, file.Name())

			data, err := os.ReadFile(fullPath)
			if err != nil {
				l.debug.DebugLog(fmt.Sprintf("Error reading file: %v", err), true)
				return
			}

			// add file name and contents to the content map
			// convert bytes slice to string
			// in format of key: [file_name, value: file_content]
			l.content[file.Name()] = string(data)

			l.debug.DebugLog(fmt.Sprintf("Opened File: %v", file.Name()), false)
			l.debug.DebugLog("Found Data:", false)
			for line := range strings.SplitSeq(string(data), "\n") {
				l.debug.DebugLog(fmt.Sprintf("\t%s", line), false)
			}
		}
	}
}

// function to begin analysis
// will result in scan -> token stream
func (l *Lexer) Analyze() {
	// TODO Analyze and create token stream
	// Based on my research I believe the best way to go about the lexer is as follows
	//
	// 1. Define allowed tokens using regex
	// 2. Convert Regex to NFA (Thompson's Algorithm)
	// 3. Compute ε-Closure of States
	// 4. Convert NFA to DFA
	//

	// 1. define regex patterns (DONE)

	// 2. convert regex to NFAs via thompson's algorithm
	l.buildTokenNFAs()
}

// thompson's algorithm impl:
// 1. convert each token regex pattern to NFAs
// 2. store NFAs in lexer struct for later use in steps 2, 3 ^^
func (l *Lexer) buildTokenNFAs() {
	l.debug.DebugLog("lexer: building NFAs", false)

	// Define all our regex token definitions
	regexDefs := []TokenRegexDef{
		{"KEYWORD", KEYWORD_PATTERN_STR, "", T_KEYWORD},
		{"CONSTANT", CONSTANT_PATTERN_STR, "", T_CONSTANT},
		{"IDENTIFIER", IDENTIFIER_PATTERN_STR, "", T_IDENTIFIER},
		{"BOOL", BOOL_PATTERN_STR, "", T_LITERAL},
		{"NUMBER", NUMBER_PATTERN_STR, "", T_LITERAL},
		{"STRING", STRING_PATTERN_STR, "", T_STRING_LITERAL},
		{"CHAR", CHAR_PATTERN_STR, "", T_CHAR_LITERAL},
		{"ESCAPE", ESCAPE_SEQUENCE_PATTERN_STR, "", T_ESCAPE_SEQUENCE},
		{"OPERATOR", OPERATOR_PATTERN_STR, "", T_OPERATOR},
		{"PUNCTUATOR", PUNCTUATOR_PATTERN_STR, "", T_PUNCTUATOR},
		{"SPECIAL", SPECIAL_PATTERN_STR, "", T_SPECIAL},
		{"ASM_INSTRUCTION", ASM_INSTRUCTION_STR, "", T_ASM_INSTRUCTION},
		{"ASM_REGISTER", ASM_REGISTER_STR, "", T_ASM_REGISTER},
		{"ASM_IMMEDIATE", ASM_IMMEDIATE_STR, "", T_ASM_IMMEDIATE},
		{"ASM_MEMORY_REF", ASM_MEMORY_REF_STR, "", T_ASM_MEMORY_REF},
		{"ASM_LABEL", ASM_LABEL_STR, "", T_ASM_LABEL},
		{"SINGLE_COMMENT", SINGLE_LINE_COMMENT_PATTERN_STR, "", T_SINGLE_LINE_COMMENT},
		{"MULTI_COMMENT", MULTI_LINE_COMMENT_PATTERN_STR, "", T_MULTI_LINE_COMMENT},
	}

	// Build NFAs for each token
	l.tokenNFAs = []*NFA{} // reset NFAs
	for i := range regexDefs {
		regexDefs[i].Postfix = postfix(regexDefs[i].Pattern, regexDefs[i].Name, l.debug)
		l.debug.DebugLog(fmt.Sprintf("Pattern: %s -> Postfix: %s", regexDefs[i].Pattern, regexDefs[i].Postfix), false)
		if regexDefs[i].Name == "STRING" {
			tokens := tokenizeRegex(regexDefs[i].Pattern)
			l.debug.DebugLog(fmt.Sprintf("STRING pattern tokens: %v", tokens), false)
			for j, t := range tokens {
				l.debug.DebugLog(fmt.Sprintf("  Token %d: %s (%s)", j, t.Value, t.Type), false)
			}
		}
		nfa := thompsonConstruct(regexDefs[i].Postfix, regexDefs[i].TokenType)
		nfa.Print(l.debug)
		l.testNFA(nfa, regexDefs[i].Name, regexDefs[i].Pattern)
		l.tokenNFAs = append(l.tokenNFAs, nfa)
	}

	l.debug.DebugLog("lexer: success on NFAs", false)
}

// function to reset a lexer
// mostly used in repeated test executions
// can also be used in between compiling multiple files
func (l *Lexer) ResetLexer() {
	l.token_stream = []Token{}
	l.content = make(map[string]string)
	l.row = 1
	l.col = 1
	l.tokenNFAs = []*NFA{} // reset NFAs
}

// function to get the token stream
func (l *Lexer) GetTokenStream() []Token {
	return l.token_stream
}

// function to set the content of the lexer
func (l *Lexer) SetContent(content map[string]string) {
	l.content = content
}

func (l *Lexer) testNFA(nfa *NFA, name, pattern string) {
	// TODO: Not sure if we need this testing functionality if converting to DFA but wrote logic for it anyways
}
