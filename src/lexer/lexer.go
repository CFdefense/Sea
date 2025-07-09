package lexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"regexp"

	debugger "github.com/CFdefense/compiler/src/debug"
)

// lexer context object
type Lexer struct {
	token_stream []Token
	content      map[string]string
	row          int
	col          int
	debug        *debugger.Debug
	tokenNFAs    []*NFA		// store NFA tokens
}

// lexer object constructor
func InitializeLexer(debug bool) *Lexer {
	return &Lexer{
		token_stream: []Token{},
		content:      make(map[string]string),
		row:          1,
		col:          1,
		debug:        debugger.InitializeDebugger("LEX", debug),
		tokenNFAs:    []*NFA{},	// initialize empty NFA slice
	}
}

// NFA state structure
type NFAState struct {
	id                 int
	isAccepting        bool
	tokenType          TokenType
	transitions        map[rune][]*NFAState
	epsilonTransitions []*NFAState
	compiledRegex      *regexp.Regexp
}

// NFA structure
type NFA struct {
	start *NFAState
	end   *NFAState
}

// function responsible for all things lexical analysis
func (l *Lexer) LexicalAnalysis(path string) {

	// begin lexer scan of target file
	// might want to change this to target directory
	// depending on future limits/scope
	l.Scan(path)

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
	// 1. Define allowed tokens using regex DONE?
	// 2. Convert Regex to NFA (Thompson's Algorithm)
	// 3. Compute ε-Closure of States
	// 4. Convert NFA to DFA
	//

	// 2. convert regex to NFAs via thompson's algorithm
	l.buildTokenNFAs()
}

// thompson's algorithm impl:
// 1. convert each token pattern to NFAs
// 2. store NFAs in lexer struct for later use in steps 2, 3 ^^
func (l *Lexer) buildTokenNFAs() {
	l.debug.DebugLog("lexer: building NFAs", false)

	// convert token patterns, order matters for precedence
	keywordNFA := l.thompsonConstruct(KEYWORD_PATTERN.String(), T_KEYWORD)
	constantNFA := l.thompsonConstruct(CONSTANT_PATTERN.String(), T_CONSTANT)
	identifierNFA := l.thompsonConstruct(IDENTIFIER_PATTERN.String(), T_IDENTIFIER)

	// literals and operators
	boolNFA := l.thompsonConstruct(BOOL_PATTERN.String(), T_LITERAL)
	numberNFA := l.thompsonConstruct(NUMBER_PATTERN.String(), T_LITERAL)
	stringNFA := l.thompsonConstruct(STRING_PATTERN.String(), T_STRING_LITERAL)
	charNFA := l.thompsonConstruct(CHAR_PATTERN.String(), T_CHAR_LITERAL)
	escapeNFA := l.thompsonConstruct(ESCAPE_SEQUENCE_PATTERN.String(), T_ESCAPE_SEQUENCE)

	operatorNFA := l.thompsonConstruct(OPERATOR_PATTERN.String(), T_OPERATOR)
	punctuatorNFA := l.thompsonConstruct(PUNCTUATOR_PATTERN.String(), T_PUNCTUATOR)
	specialNFA := l.thompsonConstruct(SPECIAL_PATTERN.String(), T_SPECIAL)

	// assembly-specific NFAs
	asmInstructionNFA := l.thompsonConstruct(ASM_INSTRUCTION.String(), T_ASM_INSTRUCTION)
	asmRegisterNFA := l.thompsonConstruct(ASM_REGISTER.String(), T_ASM_REGISTER)
	asmImmediateNFA := l.thompsonConstruct(ASM_IMMEDIATE.String(), T_ASM_IMMEDIATE)
	asmMemoryRefNFA := l.thompsonConstruct(ASM_MEMORY_REF.String(), T_ASM_MEMORY_REF)
	asmLabelNFA := l.thompsonConstruct(ASM_LABEL.String(), T_ASM_LABEL)

	// comments
	singleCommentNFA := l.thompsonConstruct(SINGLE_LINE_COMMENT_PATTERN.String(), T_SINGLE_LINE_COMMENT)
	multiCommentNFA := l.thompsonConstruct(MULTI_LINE_COMMENT_PATTERN.String(), T_MULTI_LINE_COMMENT)

	// store NFAs in priority order for lexing
	l.tokenNFAs = []*NFA{
		keywordNFA, constantNFA, identifierNFA,
		boolNFA, numberNFA, stringNFA, charNFA, escapeNFA,
		operatorNFA, punctuatorNFA, specialNFA,
		asmInstructionNFA, asmRegisterNFA, asmImmediateNFA,
		asmMemoryRefNFA, asmLabelNFA,
		singleCommentNFA, multiCommentNFA,
	}

	l.debug.DebugLog("lexer: success on NFAs", false)
}

// thompson's construction
// using golang's regex engine for pattern matching, could be done from scratch
// have our own NFA structure but i think writing a regex parser would be too much too soon
// TODO: more expansive NFA structure handling
func (l *Lexer) thompsonConstruct(regexPattern string, tokenType TokenType) *NFA {
	l.debug.DebugLog(fmt.Sprintf("lexer: converting regex pattern to NFA: %s", regexPattern), false)

	compiledRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		l.debug.DebugLog(fmt.Sprintf("lexer: failed to compile regex pattern: %s, error: %v", regexPattern, err), true)
		// return simple NFA
		return l.buildSimpleNFA(tokenType, false)
	}

	// create a functional NFA
	startState := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        false,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
		compiledRegex:      compiledRegex,	// store compiled regex
	}

	endState := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        true,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
		compiledRegex:      nil,
	}

	nfa := &NFA{
		start: startState,
		end:   endState,
	}

	l.debug.DebugLog(fmt.Sprintf("lexer: successfully built NFA for pattern: %s", regexPattern), false)
	return nfa
}

// build a NFA for a token that doesn't have a regex pattern
func (l *Lexer) buildSimpleNFA(tokenType TokenType, isAccepting bool) *NFA {
	start := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        isAccepting,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	end := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        isAccepting,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	// transition from start to end on any character
	start.transitions['\000'] = []*NFAState{end}	// null character

	return &NFA{start: start, end: end}
}

var stateIDCounter int = 0

func (l *Lexer) generateStateID() int {
	stateIDCounter++
	return stateIDCounter
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
	stateIDCounter = 0     // reset state counter for clean IDs
}

// function to get the token stream
func (l *Lexer) GetTokenStream() []Token {
	return l.token_stream
}
