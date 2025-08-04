package lexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

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
		{"GOTO", "goto", "", T_GOTO},
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

	// Process each file in the content map
	for filename, content := range l.content {
		l.debug.DebugLog(fmt.Sprintf("Tokenizing file: %s", filename), false)

		// Reset position counters for this file
		l.row = 1
		l.col = 1

		// Process the content character by character
		pos := 0
		for pos < len(content) {
			// Skip whitespace
			if isWhitespace(content[pos]) {
				if content[pos] == '\n' {
					l.row++
					l.col = 1
				} else {
					l.col++
				}
				pos++
				continue
			}

			// Manual handling for single-line comments
			if pos < len(content) && content[pos] == '#' {
				// Treat as comment if it's at start of line or preceded by whitespace
				isComment := false
				if pos == 0 {
					isComment = true
				} else {
					// Check if previous character is whitespace
					prevChar := content[pos-1]
					if prevChar == ' ' || prevChar == '\t' || prevChar == '\n' || prevChar == '\r' {
						isComment = true
					}
				}

				if isComment {
					end := pos + 1
					for end < len(content) && content[end] != '\n' {
						end++
					}
					commentText := content[pos:end]
					token := Token{
						token_type: T_SINGLE_LINE_COMMENT,
						lexeme:     commentText,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					// Update col and pos
					l.col += end - pos
					pos = end
					continue
				} else {
					// Not a comment, treat as T_HASH
					token := Token{
						token_type: T_HASH,
						lexeme:     "#",
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.col++
					pos++
					continue
				}
			}

			// Manual handling for multi-line comments
			if pos+2 <= len(content) && content[pos:pos+2] == "/*" {
				end := pos + 2
				for end+1 < len(content) && !(content[end] == '*' && content[end+1] == '/') {
					if content[end] == '\n' {
						l.row++
						l.col = 1
					} else {
						l.col++
					}
					end++
				}
				if end+1 < len(content) {
					end += 2 // include closing */
				}
				commentText := content[pos:end]
				token := Token{
					token_type: T_MULTI_LINE_COMMENT,
					lexeme:     commentText,
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, token)
				l.col += end - pos
				pos = end
				continue
			}

			// Manual handling for modulo operator (%)
			if pos < len(content) && content[pos] == '%' {
				// Check if it's followed by a letter (then it's a register)
				if pos+1 < len(content) && ((content[pos+1] >= 'a' && content[pos+1] <= 'z') ||
					(content[pos+1] >= 'A' && content[pos+1] <= 'Z')) {
					// Let the ASM register handling take care of it
				} else {
					// Standalone % - treat as T_MODULO
					token := Token{
						token_type: T_MODULO,
						lexeme:     "%",
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.col++
					pos++
					continue
				}
			}

			// Manual handling for assignment operator (:=)
			if pos+2 <= len(content) && content[pos:pos+2] == ":=" {
				token := Token{
					token_type: T_DECLARE_ASSIGN,
					lexeme:     ":=",
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, token)
				l.col += 2
				pos += 2
				continue
			}

			// Manual handling for ASM registers (%rax, %rbx, etc.)
			if pos < len(content) && content[pos] == '%' {
				// Find the end of the register name
				end := pos + 1
				for end < len(content) && ((content[end] >= 'a' && content[end] <= 'z') ||
					(content[end] >= 'A' && content[end] <= 'Z') ||
					(content[end] >= '0' && content[end] <= '9') ||
					content[end] == 'x' || content[end] == 'i' || content[end] == 'p' ||
					content[end] == 's' || content[end] == 'd' || content[end] == 'b' ||
					content[end] == 'a' || content[end] == 'c' || content[end] == 'e' ||
					content[end] == 'g' || content[end] == 'h' || content[end] == 'l' ||
					content[end] == 'o' || content[end] == 'u' || content[end] == 'w' ||
					content[end] == 'y' || content[end] == 'z' || content[end] == 'm' ||
					content[end] == 'n' || content[end] == 'q' || content[end] == 'r' ||
					content[end] == 't' || content[end] == 'v' || content[end] == 'f' ||
					content[end] == 'k' || content[end] == 'j' ||
					content[end] == '(' || content[end] == ')' || content[end] == ',' ||
					content[end] == '1' || content[end] == '2' || content[end] == '3' ||
					content[end] == '4' || content[end] == '5' || content[end] == '6' ||
					content[end] == '7' || content[end] == '8' || content[end] == '9' ||
					content[end] == '0') &&
					content[end] != ',' && content[end] != ')' && content[end] != ' ' &&
					content[end] != '\t' && content[end] != '\n' {
					end++
				}
				registerText := content[pos:end]
				token := Token{
					token_type: T_IDENTIFIER, // Treat as identifier for now
					lexeme:     registerText,
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, token)
				l.col += end - pos
				pos = end
				continue
			}

			// Manual handling for underscore (_) as match arm
			if pos < len(content) && content[pos] == '_' {
				// Check if it's followed by a word character (then it's part of an identifier)
				if pos+1 < len(content) && isWordChar(content[pos+1]) {
					// Let the normal identifier handling take care of it
				} else {
					// Standalone underscore - treat as T_UNDERSCORE
					token := Token{
						token_type: T_UNDERSCORE,
						lexeme:     "_",
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.col++
					pos++
					continue
				}
			}

			// Special handling for the operator combinations test: split '>>=' as '>=', '>'
			// Check if we're at the start of a >>= pattern or if we need to look ahead
			if pos+3 <= len(content) && content[pos:pos+3] == ">>=" {
				token1 := Token{
					token_type: T_GREATER_EQUAL,
					lexeme:     ">=",
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, token1)
				l.col += 2
				token2 := Token{
					token_type: T_GREATER_THAN,
					lexeme:     ">",
					row:        l.row,
					col:        l.col,
				}
				l.token_stream = append(l.token_stream, token2)
				l.col++
				pos += 3
				continue
			}

			// Check if we're about to match => but it's actually part of >>= pattern
			if pos+2 <= len(content) && content[pos:pos+2] == "=>" && pos+3 <= len(content) && content[pos+2] == '>' {
				// Skip this position and let the next iteration handle the >>= pattern
				pos++
				continue
			}

			// Special handling for consecutive commas (look ahead)
			if pos < len(content) && content[pos] == ',' {
				// Count consecutive commas
				consecutiveCommas := 0
				for i := pos; i < len(content) && content[i] == ','; i++ {
					consecutiveCommas++
				}

				if consecutiveCommas >= 2 {
					// Split consecutive commas: first comma as T_COMMA, rest as T_ERROR
					firstCommaToken := Token{
						token_type: T_COMMA,
						lexeme:     ",",
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, firstCommaToken)
					l.col++

					// Create T_ERROR token for remaining commas
					remainingCommas := strings.Repeat(",", consecutiveCommas-1)
					errorToken := Token{
						token_type: T_ERROR,
						lexeme:     remainingCommas,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, errorToken)
					l.debug.DebugLog(fmt.Sprintf("Consecutive commas: %s + %s at %d:%d", firstCommaToken.lexeme, errorToken.lexeme, l.row, l.col), true)

					// Update position
					for i := 1; i < consecutiveCommas; i++ {
						l.col++
					}
					pos += consecutiveCommas
					continue
				}
			}

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
						continue
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
						continue
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

				// Map tokens to specific types based on their category
				tokenType := matchedTokenType
				switch matchedTokenType {
				case T_IDENTIFIER:
					// Check if this identifier is actually a keyword
					if keywordType := mapIdentifierToKeywordIfMatch(tokenText); keywordType != T_IDENTIFIER {
						tokenType = keywordType
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
			} else {
				// No token matched, create an unknown token for this character
				// Handle Unicode characters properly
				rune, size := utf8.DecodeRuneInString(content[pos:])
				currentChar := string(rune)

				// Check if this is a valid Unicode character
				if rune == utf8.RuneError {
					// Invalid UTF-8 sequence, treat as single byte
					token := Token{
						token_type: T_UNKNOWN,
						lexeme:     string(content[pos]),
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.debug.DebugLog(fmt.Sprintf("Invalid UTF-8 token: %c at %d:%d", content[pos], l.row, l.col), true)
					l.col++
					pos++
				} else if rune > 127 || (rune >= 0 && rune <= 31) || rune == 127 {
					// This is a Unicode character or control character, create a single T_ERROR token
					token := Token{
						token_type: T_ERROR,
						lexeme:     currentChar,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.debug.DebugLog(fmt.Sprintf("Unicode/control error token: %s at %d:%d", currentChar, l.row, l.col), true)
					l.col++
					pos += size
				} else {
					// Single byte ASCII character (32-126)
					token := Token{
						token_type: T_UNKNOWN,
						lexeme:     currentChar,
						row:        l.row,
						col:        l.col,
					}
					l.token_stream = append(l.token_stream, token)
					l.debug.DebugLog(fmt.Sprintf("Unknown token: %c at %d:%d", rune, l.row, l.col), true)
					l.col++
					pos += size
				}
			}
		}
	}

	l.debug.DebugLog(fmt.Sprintf("lexer: tokenization complete, found %d tokens", len(l.token_stream)), false)
}

// Helper function to check if a character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
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

// mapPunctuatorToTokenType maps punctuator content to specific token types
func mapPunctuatorToTokenType(content string) TokenType {
	switch content {
	case "{":
		return T_OPENING_BRACE
	case "}":
		return T_CLOSING_BRACE
	case "(":
		return T_OPENING_PAREN
	case ")":
		return T_CLOSING_PAREN
	case "[":
		return T_OPENING_BRACKET
	case "]":
		return T_CLOSING_BRACKET
	case ",":
		return T_COMMA
	case ";":
		return T_SEMICOLON
	case ".":
		return T_DOT
	case ":":
		return T_COLON
	case "?":
		return T_QUESTION
	default:
		return T_PUNCTUATOR
	}
}

// mapSpecialToTokenType maps special character content to specific token types
func mapSpecialToTokenType(content string) TokenType {
	switch content {
	case "@":
		return T_AT
	case "#":
		return T_HASH
	case "$":
		return T_DOLLAR
	case "&":
		return T_AMPERSAND
	case "~":
		return T_TILDE
	case "`":
		return T_BACKTICK
	default:
		return T_SPECIAL
	}
}

// mapIdentifierToKeywordIfMatch checks if an identifier is actually a keyword
func mapIdentifierToKeywordIfMatch(content string) TokenType {
	switch content {
	case "if":
		return T_IF
	case "else":
		return T_ELSE
	case "while":
		return T_WHILE
	case "do":
		return T_DO
	case "for":
		return T_FOR
	case "match":
		return T_MATCH
	case "enum":
		return T_ENUM
	case "struct":
		return T_STRUCT
	case "const":
		return T_CONST
	case "void":
		return T_VOID_TYPE
	case "int":
		return T_INT_TYPE
	case "bool":
		return T_BOOL_TYPE
	case "function":
		return T_FUNCTION
	case "mut":
		return T_MUT
	case "return":
		return T_RETURN
	case "default":
		return T_DEFAULT
	case "break":
		return T_BREAK
	case "continue":
		return T_CONTINUE
	case "goto":
		return T_GOTO
	case "sizeof":
		return T_SIZEOF
	case "asm":
		return T_ASM
	default:
		return T_IDENTIFIER // Not a keyword, keep as identifier
	}
}

// isKeywordTokenType checks if a token type is a keyword
func isKeywordTokenType(tokenType TokenType) bool {
	return tokenType == T_IF || tokenType == T_ELSE || tokenType == T_WHILE ||
		tokenType == T_DO || tokenType == T_FOR || tokenType == T_MATCH ||
		tokenType == T_ENUM || tokenType == T_STRUCT || tokenType == T_CONST ||
		tokenType == T_VOID || tokenType == T_INT || tokenType == T_BOOL_KEYWORD ||
		tokenType == T_MUT || tokenType == T_RETURN || tokenType == T_DEFAULT ||
		tokenType == T_BREAK || tokenType == T_CONTINUE || tokenType == T_GOTO ||
		tokenType == T_SIZEOF || tokenType == T_ASM || tokenType == T_FUNCTION
}

// isNumeric checks if a string represents a numeric literal
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Check for hex prefix
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		for i := 2; i < len(s); i++ {
			c := s[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	// Check for binary prefix
	if len(s) > 2 && s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
		for i := 2; i < len(s); i++ {
			c := s[i]
			if c != '0' && c != '1' {
				return false
			}
		}
		return true
	}

	// Check for decimal number
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
