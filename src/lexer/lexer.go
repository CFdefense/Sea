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
}

// lexer object constructor
func InitializeLexer(debug bool) *Lexer {
	return &Lexer{
		token_stream: []Token{},
		content:      make(map[string]string),
		row:          1,
		col:          1,
		debug:        debugger.InitializeDebugger("LEX", debug),
	}
}

// NFA state structure
type NFAState struct {
	id			int
	isAccepting		bool
	tokenType		TokenType
	transitions		map[rune][]*NFAState
	epsilonTransitions	[]*NFAState
}

// NFA structure
type NFA struct {
	start	*NFAState
	end	*NFAState
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
// 2. immediately discard the NFAs since this is just a skeleton and we aren't
//    to step 3 yet
func (l *Lexer) buildTokenNFAs() {
	l.debug.DebugLog("lexer: building NFAs", false)

	// convert token patterns
	identifierNFA := l.thompsonConstruct(IDENTIFIER_PATTERN.String(), T_IDENTIFIER)
	operatorNFA := l.thompsonConstruct(OPERATOR_PATTERN.String(), T_OPERATOR)
	constantNFA := l.thompsonConstruct(CONSTANT_PATTERN.String(), T_CONSTANT)
	keywordNFA := l.thompsonConstruct(KEYWORD_PATTERN.String(), T_KEYWORD)
	numberNFA := l.thompsonConstruct(NUMBER_PATTERN.String(), T_LITERAL)
	boolNFA := l.thompsonConstruct(BOOL_PATTERN.String(), T_LITERAL)
	punctuatorNFA := l.thompsonConstruct(PUNCTUATOR_PATTERN.String(), T_PUNCTUATOR)
	specialNFA := l.thompsonConstruct(SPECIAL_PATTERN.String(), T_SPECIAL)
	stringNFA := l.thompsonConstruct(STRING_PATTERN.String(), T_STRING_LITERAL)
	charNFA := l.thompsonConstruct(CHAR_PATTERN.String(), T_CHAR_LITERAL)
	escapeNFA := l.thompsonConstruct(ESCAPE_SEQUENCE_PATTERN.String(), T_ESCAPE_SEQUENCE)

	// assembly-specific NFAs
	asmInstructionNFA := l.thompsonConstruct(ASM_INSTRUCTION.String(), T_ASM_INSTRUCTION)
	asmRegisterNFA := l.thompsonConstruct(ASM_REGISTER.String(), T_ASM_REGISTER)
	asmImmediateNFA := l.thompsonConstruct(ASM_IMMEDIATE.String(), T_ASM_IMMEDIATE)
	asmMemoryRefNFA := l.thompsonConstruct(ASM_MEMORY_REF.String(), T_ASM_MEMORY_REF)
	asmLabelNFA := l.thompsonConstruct(ASM_LABEL.String(), T_ASM_LABEL)

	// comments
	singleCommentNFA := l.thompsonConstruct(SINGLE_LINE_COMMENT_PATTERN.String(), T_SINGLE_LINE_COMMENT)
	multiCommentNFA := l.thompsonConstruct(MULTI_LINE_COMMENT_PATTERN.String(), T_MULTI_LINE_COMMENT)

	// here we discard them so that it actually compiles
	// we need to either:
	// a: store in an array in the lexer struct
	// b: make a master nfa for later conversion
	// i'm not sure what the best route is right now
	_ = identifierNFA
	_ = operatorNFA
	_ = constantNFA
	_ = keywordNFA
	_ = numberNFA
	_ = boolNFA
	_ = punctuatorNFA
	_ = specialNFA
	_ = stringNFA
	_ = charNFA
	_ = escapeNFA
	_ = asmInstructionNFA
	_ = asmRegisterNFA
	_ = asmImmediateNFA
	_ = asmMemoryRefNFA
	_ = asmLabelNFA
	_ = singleCommentNFA
	_ = multiCommentNFA

	l.debug.DebugLog("lexer: success on NFAs", false)
}

// thompson's construction
// TODO: implement full thing
func (l *Lexer) thompsonConstruct(regexPattern string, tokenType TokenType) *NFA {
	l.debug.DebugLog(fmt.Sprintf("Converting regex pattern to NFA: %s", regexPattern), false)

	// Create start and end states
	startState := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        false,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	endState := &NFAState{
		id:                 l.generateStateID(),
		isAccepting:        true,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	// TODO: implement thompson's construction

	return &NFA{
		start: startState,
		end:   endState,
	}
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
}

// function to get the token stream
func (l *Lexer) GetTokenStream() []Token {
	return l.token_stream
}
