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
	tokenDFAs    []*DFA // store DFA tokens
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
		tokenDFAs:    []*DFA{}, // initialize empty DFA slice
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

	// 3 & 4. Convert NFAs to DFAs
	l.convertNFAsToDFAs()

	// 5. Use DFAs for tokenization
	l.tokenize()
}

// thompson's algorithm impl:
// 1. convert each token regex pattern to NFAs
// 2. store NFAs in lexer struct for later use in steps 2, 3 ^^
func (l *Lexer) buildTokenNFAs() {
	l.debug.DebugLog("lexer: building NFAs", false)

	// Define all our regex token definitions in priority order
	regexDefs := []TokenRegexDef{
		// Single-character operators first (highest priority)
		{"PLUS", "^\\+", "", T_PLUS},
		{"MINUS", "^-", "", T_MINUS},
		{"MULTIPLY", "^\\*", "", T_MULTIPLY},
		{"DIVIDE", "^/", "", T_DIVIDE},
		{"MODULO", "^%", "", T_MODULO},
		{"ASSIGN", "^=", "", T_ASSIGN},
		{"LESS_THAN", "^<", "", T_LESS_THAN},
		{"GREATER_THAN", "^>", "", T_GREATER_THAN},
		{"NOT", "^!", "", T_NOT},
		{"XOR", "^\\^", "", T_XOR},
		{"AMPERSAND", "^&", "", T_AMPERSAND},

		// Multi-character operators (longer patterns first for greedy matching)
		{"EQUALS", "^==", "", T_EQUALS},
		{"NOT_EQUALS", "^!=", "", T_NOT_EQUALS},
		{"LESS_EQUAL", "^<=", "", T_LESS_EQUAL},
		{"GREATER_EQUAL", "^>=", "", T_GREATER_EQUAL},
		{"AND", "^&&", "", T_AND},
		{"OR", "^\\|\\|", "", T_OR},
		{"LEFT_SHIFT", "^<<", "", T_LEFT_SHIFT},
		{"RIGHT_SHIFT", "^>>", "", T_RIGHT_SHIFT},
		{"INT_DIVIDE", "^//", "", T_INT_DIVIDE},
		{"ARROW", "^->", "", T_MEMBER_OPERATOR},
		{"MATCH_ARROW", "^=>", "", T_ARROW},

		// Boolean literals
		{"BOOL", BOOL_PATTERN_STR, "", T_LITERAL},

		// Numbers
		{"NUMBER", NUMBER_PATTERN_STR, "", T_LITERAL},

		// Underscore pattern for match arms
		{"UNDERSCORE", "^_", "", T_UNDERSCORE},
		// Identifiers (before keywords to ensure full identifier matching)
		{"IDENTIFIER", IDENTIFIER_PATTERN_STR, "", T_IDENTIFIER},

		// Keywords (individual patterns to avoid alternation issues) - AFTER identifiers
		{"IF", "if", "", T_IF},
		{"ELSE", "else", "", T_ELSE},
		{"WHILE", "while", "", T_WHILE},
		{"DO", "do", "", T_DO},
		{"FOR", "for", "", T_FOR},
		{"MATCH", "match", "", T_MATCH},
		{"ENUM", "enum", "", T_ENUM},
		{"STRUCT", "struct", "", T_STRUCT},
		{"CONST", "const", "", T_CONST},
		{"VOID", "void", "", T_VOID},
		{"INT", "int", "", T_INT},
		{"BOOL_KEYWORD", "bool", "", T_BOOL_KEYWORD},
		{"MUT", "mut", "", T_MUT},
		{"RETURN", "return", "", T_RETURN},
		{"DEFAULT", "default", "", T_DEFAULT},
		{"BREAK", "break", "", T_BREAK},
		{"CONTINUE", "continue", "", T_CONTINUE},
		{"SIZEOF", "sizeof", "", T_SIZEOF},
		{"ASM", "asm", "", T_ASM},

		// Additional keywords
		{"FUNCTION", "function", "", T_FUNCTION},

		// Constants
		{"CONSTANT", CONSTANT_PATTERN_STR, "", T_LITERAL},

		// String literals (after identifiers)
		{"STRING", STRING_PATTERN_STR, "", T_STRING_LITERAL},
		{"CHAR", CHAR_PATTERN_STR, "", T_CHAR_LITERAL},
		{"ESCAPE", ESCAPE_SEQUENCE_PATTERN_STR, "", T_ESCAPE_SEQUENCE},

		// Punctuators
		{"PUNCTUATOR", PUNCTUATOR_PATTERN_STR, "", T_PUNCTUATOR},

		// Special tokens
		{"SPECIAL", SPECIAL_PATTERN_STR, "", T_SPECIAL},

		// ASM patterns (temporarily disabled)
		// {"ASM_INSTRUCTION", ASM_INSTRUCTION_STR, "", T_ASM_INSTRUCTION},
		// {"ASM_REGISTER", ASM_REGISTER_STR, "", T_ASM_REGISTER},
		// {"ASM_IMMEDIATE", ASM_IMMEDIATE_STR, "", T_ASM_IMMEDIATE},
		// {"ASM_MEMORY_REF", ASM_MEMORY_REF_STR, "", T_ASM_MEMORY_REF},
		// {"ASM_LABEL", ASM_LABEL_STR, "", T_ASM_LABEL},
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
		l.debug.DebugLog(fmt.Sprintf("Added NFA %d for pattern %s with token type %s", len(l.tokenNFAs)-1, regexDefs[i].Name, regexDefs[i].TokenType.String()), false)
	}

	l.debug.DebugLog("lexer: success on NFAs", false)
}

// Convert NFAs to DFAs using the subset construction algorithm
func (l *Lexer) convertNFAsToDFAs() {
	l.debug.DebugLog("lexer: converting NFAs to DFAs", false)

	l.tokenDFAs = []*DFA{} // reset DFAs

	// Convert each NFA to a DFA
	for i, nfa := range l.tokenNFAs {
		tokenTypeName := "unknown"
		if nfa.end != nil && nfa.end.tokenType != 0 {
			tokenTypeName = nfa.end.tokenType.String()
		}

		l.debug.DebugLog(fmt.Sprintf("Converting NFA %d to DFA for %s", i, tokenTypeName), false)

		// Add timeout and error handling
		dfa := ConvertNFAtoDFA(nfa)
		if dfa == nil {
			l.debug.DebugLog(fmt.Sprintf("Failed to convert NFA %d to DFA, using fallback", i), true)
			// Create a simple fallback DFA
			dfa = &DFA{
				start:    &DFAState{id: generateDFAStateID(), isAccepting: true, tokenType: nfa.end.tokenType},
				states:   []*DFAState{},
				alphabet: make(map[string]bool),
			}
		}

		l.tokenDFAs = append(l.tokenDFAs, dfa)

		// Debug output - only print small DFAs to avoid log flooding
		if len(dfa.states) < 50 {
			l.debug.DebugLog(dfa.PrintDFA(), false)
		} else {
			l.debug.DebugLog(fmt.Sprintf("DFA %d has %d states (too large to print)", i, len(dfa.states)), false)
		}
	}

	l.debug.DebugLog(fmt.Sprintf("lexer: successfully converted %d NFAs to DFAs", len(l.tokenDFAs)), false)
}

// Tokenize the input using the DFAs
func (l *Lexer) tokenize() {
	l.debug.DebugLog("lexer: starting tokenization", false)

	for filename, content := range l.content {
		l.debug.DebugLog(fmt.Sprintf("Tokenizing file: %s", filename), false)
		l.row = 1
		l.col = 1
		pos := 0
		for pos < len(content) {
			if handled, newPos := l.handleWhitespace(content, pos); handled {
				pos = newPos
				continue
			}
			if handled, newPos := l.handleSingleLineComment(content, pos); handled {
				pos = newPos
				continue
			}
			if handled, newPos := l.handleMultiLineComment(content, pos); handled {
				pos = newPos
				continue
			}
			if handled, newPos := l.handleOperator(content, pos); handled {
				pos = newPos
				continue
			}
			// Fallback: DFA-based matching and unknowns
			if handled, newPos := l.handleDfaOrUnknown(content, pos); handled {
				pos = newPos
				continue
			}
			// If nothing handled, move forward to avoid infinite loop
			pos++
		}
	}
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
	l.tokenDFAs = []*DFA{} // reset DFAs
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

// TestLexerWithInput tests the lexer with a simple input string
func (l *Lexer) TestLexerWithInput(input string) {
	l.debug.DebugLog("Testing lexer with input: "+input, false)

	// Set the content
	l.content = map[string]string{"test.txt": input}

	// Run the full lexer pipeline
	l.buildTokenNFAs()
	l.convertNFAsToDFAs()
	l.tokenize()

	// Print results
	l.debug.DebugLog(fmt.Sprintf("Generated %d tokens:", len(l.token_stream)), false)
	for i, token := range l.token_stream {
		l.debug.DebugLog(fmt.Sprintf("  %d: %s (%s) at %d:%d",
			i, token.lexeme, token.token_type.String(), token.row, token.col), false)
	}
}

func (l *Lexer) handleDfaOrUnknown(content string, pos int) (bool, int) {
	// Try to match a token starting at the current position
	matched := false
	maxMatchLength := 0
	var matchedTokenType TokenType

	// Try each DFA and find the longest match
	for dfaIndex, dfa := range l.tokenDFAs {
		// Try to match as much as possible from the current position
		i := pos
		currentState := dfa.start
		lastAcceptingPos := -1
		var lastAcceptingType TokenType

		for i <= len(content) {
			// Check if current state is accepting
			if currentState.isAccepting {
				lastAcceptingPos = i
				lastAcceptingType = currentState.tokenType
			}

			if i >= len(content) {
				break
			}

			// Try to find a transition
			found := false
			var nextState *DFAState

			// First, try multi-character patterns (for operators and keywords)
			for transitionSymbol, transition := range currentState.transitions {
				if transitionSymbol != "any" && len(transitionSymbol) > 1 {
					// Check if the remaining content starts with this transition symbol
					if i+len(transitionSymbol) <= len(content) &&
						content[i:i+len(transitionSymbol)] == transitionSymbol {
						nextState = transition
						found = true
						i += len(transitionSymbol)
						break
					}
				}
			}

			if !found {
				// Then try exact single character match
				symbol := string(content[i])
				if transition, exists := currentState.transitions[symbol]; exists {
					nextState = transition
					found = true
					i++
				}
			}

			if !found {
				// Try "any" transition as fallback
				if transition, exists := currentState.transitions["any"]; exists {
					nextState = transition
					i++
				} else {
					break // No valid transition
				}
			}

			// Safety check to prevent infinite loops
			if nextState == nil {
				break
			}

			currentState = nextState
		}

		// Check final state
		if currentState.isAccepting && (lastAcceptingPos == -1 || i > lastAcceptingPos) {
			lastAcceptingPos = i
			lastAcceptingType = currentState.tokenType
		}

		if lastAcceptingPos > pos {
			// Found a match, check if it's longer than previous matches
			matchLength := lastAcceptingPos - pos
			if matchLength > maxMatchLength {
				maxMatchLength = matchLength
				matchedTokenType = lastAcceptingType
				matched = true
				l.debug.DebugLog(fmt.Sprintf("New longest match: DFA %d matched token type %s (length %d)", dfaIndex, matchedTokenType.String(), matchLength), false)
			}
		}
	}

	if matched {
		// Create a token for the matched text
		tokenText := content[pos : pos+maxMatchLength]

		// Special handling for invalid identifiers (starting with digits)
		if matchedTokenType == T_STRING_LITERAL && len(tokenText) > 0 && tokenText[0] >= '0' && tokenText[0] <= '9' {
			// Check if this looks like a number followed by letters
			hasLetters := false
			numberEnd := 0
			for i, c := range tokenText {
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
					hasLetters = true
					numberEnd = i
					break
				}
			}

			if hasLetters && numberEnd > 0 {
				// Split into number + identifier
				numberText := tokenText[:numberEnd]
				identifierText := tokenText[numberEnd:]

				// Create number token
				numberToken := Token{
					token_type: T_INT_LITERAL,
					lexeme:     numberText,
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, numberToken)

				// Create identifier token
				identifierToken := Token{
					token_type: T_IDENTIFIER,
					lexeme:     identifierText,
					row:        l.row,
					col:        l.col + len(numberText),
				}
				l.token_stream = append(l.token_stream, identifierToken)

				// Update position
				for _, c := range tokenText {
					if c == '\n' {
						l.row++
						l.col = 1
					} else {
						l.col++
					}
				}
				pos += maxMatchLength
				return true, pos
			}
		}

		// Special handling for comment-like operators (split consecutive operators)
		if matchedTokenType == T_INT_DIVIDE && len(tokenText) > 2 {
			// Split consecutive // operators
			slashCount := 0
			for _, c := range tokenText {
				if c == '/' {
					slashCount++
				}
			}

			if slashCount > 2 {
				// Split into individual // and / tokens
				remaining := tokenText
				for len(remaining) >= 2 {
					if len(remaining) >= 2 && remaining[0] == '/' && remaining[1] == '/' {
						// Add // token
						intDivideToken := Token{
							token_type: T_INT_DIVIDE,
							lexeme:     "//",
							row:        l.row,
							col:        l.col,
						}
						l.token_stream = append(l.token_stream, intDivideToken)
						l.col += 2
						remaining = remaining[2:]
					} else if len(remaining) >= 1 && remaining[0] == '/' {
						// Add / token
						divideToken := Token{
							token_type: T_DIVIDE,
							lexeme:     "/",
							row:        l.row,
							col:        l.col,
						}
						l.token_stream = append(l.token_stream, divideToken)
						l.col++
						remaining = remaining[1:]
					} else {
						break
					}
				}

				// Update position
				for _, c := range tokenText {
					if c == '\n' {
						l.row++
						l.col = 1
					}
				}
				pos += maxMatchLength
				return true, pos
			}
		}

		// For keyword tokens, check word boundaries (keep this simple)
		if isKeywordTokenType(matchedTokenType) {
			// Check if this is a complete word (not part of a larger identifier)
			nextPos := pos + maxMatchLength
			if nextPos < len(content) && isWordChar(content[nextPos]) {
				// This keyword is part of a larger identifier, treat it as identifier instead
				matchedTokenType = T_IDENTIFIER
			}
		}

		// Handle postfix increment/decrement operators
		if matchedTokenType == T_IDENTIFIER && maxMatchLength >= 3 {
			// Check if the identifier ends with ++ or --
			if len(tokenText) >= 3 {
				suffix := tokenText[len(tokenText)-2:]
				if suffix == "++" || suffix == "--" {
					l.debug.DebugLog(fmt.Sprintf("Splitting postfix operator: %s into %s + %s", tokenText, tokenText[:len(tokenText)-2], suffix), false)
					// Split into identifier + operator
					identifierText := tokenText[:len(tokenText)-2]
					operatorText := suffix

					// Create identifier token
					identifierToken := Token{
						token_type: T_IDENTIFIER,
						lexeme:     identifierText,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, identifierToken)

					// Create operator tokens
					if operatorText == "++" {
						op1 := createToken(T_PLUS, "+", l.row, l.col+len(identifierText))
						l.token_stream = append(l.token_stream, op1)
						op2 := createToken(T_PLUS, "+", l.row, l.col+len(identifierText)+1)
						l.token_stream = append(l.token_stream, op2)
					} else { // "--"
						op1 := createToken(T_MINUS, "-", l.row, l.col+len(identifierText))
						l.token_stream = append(l.token_stream, op1)
						op2 := createToken(T_MINUS, "-", l.row, l.col+len(identifierText)+1)
						l.token_stream = append(l.token_stream, op2)
					}

					// Update position
					for _, c := range tokenText {
						if c == '\n' {
							l.row++
							l.col = 1
						} else {
							l.col++
						}
					}
					pos += maxMatchLength
					return true, pos
				}
			}
		}

		// Map tokens to specific types based on their category
		tokenType := matchedTokenType
		switch matchedTokenType {
		case T_IDENTIFIER:
			// Check if this identifier is actually a keyword
			if keywordType := mapIdentifierToKeywordIfMatch(tokenText); keywordType != T_IDENTIFIER {
				tokenType = keywordType
			}
			// Check if identifier starts with a digit (invalid)
			if len(tokenText) > 0 && tokenText[0] >= '0' && tokenText[0] <= '9' {
				// Split into number + identifier
				numberEnd := 0
				for i, c := range tokenText {
					if c < '0' || c > '9' {
						numberEnd = i
						break
					}
					numberEnd = i + 1
				}

				if numberEnd < len(tokenText) {
					// Create number token
					numberText := tokenText[:numberEnd]
					identifierText := tokenText[numberEnd:]

					// Create number token
					numberToken := Token{
						token_type: T_INT_LITERAL,
						lexeme:     numberText,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, numberToken)

					// Create identifier token
					identifierToken := Token{
						token_type: T_IDENTIFIER,
						lexeme:     identifierText,
						row:        l.row,
						col:        l.col + len(numberText),
					}
					l.token_stream = append(l.token_stream, identifierToken)

					// Update position
					for _, c := range tokenText {
						if c == '\n' {
							l.row++
							l.col = 1
						} else {
							l.col++
						}
					}
					pos += maxMatchLength
					return true, pos
				} else {
					// All digits, treat as number
					tokenType = T_INT_LITERAL
				}
			}
		case T_LITERAL:
			// Map generic literals to specific types
			if tokenText == "true" || tokenText == "false" {
				tokenType = T_BOOL_LITERAL
			} else if isNumeric(tokenText) {
				// Check if this is an oversized decimal integer
				if len(tokenText) > 19 && tokenText[0] >= '1' && tokenText[0] <= '9' {
					// Check if it's all digits (decimal number)
					allDigits := true
					for _, c := range tokenText {
						if c < '0' || c > '9' {
							allDigits = false
							break
						}
					}
					if allDigits {
						tokenType = T_ERROR
					} else {
						tokenType = T_INT_LITERAL
					}
				} else {
					tokenType = T_INT_LITERAL
				}
			} else {
				tokenType = T_LITERAL // Keep as generic literal for other cases
			}
		case T_PUNCTUATOR:
			tokenType = mapPunctuatorToTokenType(tokenText)
		case T_SPECIAL:
			tokenType = mapSpecialToTokenType(tokenText)
		}

		// Map T_LITERAL to T_IDENTIFIER unless it's a number or boolean
		if tokenType == T_LITERAL {
			if isNumeric(tokenText) {
				tokenType = T_INT_LITERAL
			} else if tokenText == "true" || tokenText == "false" {
				tokenType = T_BOOL_LITERAL
			} else {
				tokenType = T_IDENTIFIER
				// Check if it's a keyword
				if keywordType := mapIdentifierToKeywordIfMatch(tokenText); keywordType != T_IDENTIFIER {
					tokenType = keywordType
				}
			}
		}

		token := Token{
			token_type: tokenType,
			lexeme:     tokenText,
			row:        l.row,
			col:        l.col,
		}

		l.token_stream = append(l.token_stream, token)
		l.debug.DebugLog(fmt.Sprintf("Token: %s (%s) at %d:%d", tokenText, tokenType.String(), l.row, l.col), false)

		// Update position and column
		for _, c := range tokenText {
			if c == '\n' {
				l.row++
				l.col = 1
			} else {
				l.col++
			}
		}
		pos += maxMatchLength
		return true, pos
	} else {
		// No token matched, create an unknown token for this character
		return l.handleUnknownToken(content, pos)
	}
}
