package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Helper function to check if a character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
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
		tokenType == T_BREAK || tokenType == T_CONTINUE ||
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

// createToken creates a new token with the given parameters
func createToken(tokenType TokenType, lexeme string, row, col int) Token {
	return Token{
		token_type: tokenType,
		lexeme:     lexeme,
		row:        row,
		col:        col,
	}
}

// handleWhitespace handles whitespace characters
func (l *Lexer) handleWhitespace(content string, pos int) (bool, int) {
	if isWhitespace(content[pos]) {
		if content[pos] == '\n' {
			l.row++
			l.col = 1
		} else {
			l.col++
		}
		return true, pos + 1
	}
	return false, pos
}

// handleSingleLineComment handles single-line comments
func (l *Lexer) handleSingleLineComment(content string, pos int) (bool, int) {
	if pos < len(content) && content[pos] == '#' {
		// Check if this is followed by another # (then it's not a comment)
		if pos+1 < len(content) && content[pos+1] == '#' {
			// Treat as two separate # tokens
			token1 := createToken(T_HASH, "#", l.row, l.col)
			l.token_stream = append(l.token_stream, token1)
			l.col++

			token2 := createToken(T_HASH, "#", l.row, l.col)
			l.token_stream = append(l.token_stream, token2)
			l.col++
			pos += 2
			return true, pos
		}

		// Check if we're in an ASM context (inside asm { ... })
		inASMContext := false
		// Look backwards to see if we're inside an asm block
		for i := pos - 1; i >= 0; i-- {
			if content[i] == '}' {
				break // Found closing brace, not in ASM context
			} else if content[i] == '{' {
				// Look for 'asm' before the opening brace
				j := i - 1
				for j >= 0 && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n') {
					j--
				}
				if j >= 2 && content[j-2:j+1] == "asm" {
					inASMContext = true
				}
				break
			}
		}

		// Treat as comment if it's at start of line, preceded by whitespace, or in ASM context
		isComment := false
		if pos == 0 {
			isComment = true
		} else if inASMContext {
			isComment = true
		} else {
			// Check if previous character is whitespace
			prevChar := content[pos-1]
			if prevChar == ' ' || prevChar == '\t' || prevChar == '\n' || prevChar == '\r' {
				isComment = true
			}
			// Also treat as comment if preceded by certain ASM-related characters
			if prevChar == '}' || prevChar == ',' || prevChar == ')' {
				isComment = true
			}
		}

		// Only treat as comment if followed by a space or at end of line
		if isComment && (pos+1 >= len(content) || content[pos+1] == ' ') {
			end := pos + 1
			// Continue to end of line, but don't consume the newline character
			for end < len(content) && content[end] != '\n' && content[end] != '\r' {
				end++
			}
			commentText := content[pos:end]
			token := createToken(T_SINGLE_LINE_COMMENT, commentText, l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			// Update col and pos
			l.col += end - pos
			pos = end
			return true, pos
		} else {
			// Not a comment, treat as T_HASH
			token := createToken(T_HASH, "#", l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col++
			pos++
			return true, pos
		}
	}
	return false, pos
}

// handleMultiLineComment handles multi-line comments
func (l *Lexer) handleMultiLineComment(content string, pos int) (bool, int) {
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
		token := createToken(T_MULTI_LINE_COMMENT, commentText, l.row, l.col)
		l.token_stream = append(l.token_stream, token)
		l.col += end - pos
		pos = end
		return true, pos
	}
	return false, pos
}

// handleOperator handles operators and special characters
func (l *Lexer) handleOperator(content string, pos int) (bool, int) {
	// Handle >>= operator (right shift assignment)
	if pos+3 <= len(content) && content[pos:pos+3] == ">>=" {
		token1 := createToken(T_RIGHT_SHIFT, ">>", l.row, l.col)
		l.token_stream = append(l.token_stream, token1)
		l.col += 2
		token2 := createToken(T_ASSIGN, "=", l.row, l.col)
		l.token_stream = append(l.token_stream, token2)
		l.col++
		pos += 3
		return true, pos
	}

	// Handle &mut as separate tokens
	if pos+4 <= len(content) && content[pos:pos+4] == "&mut" {
		token1 := createToken(T_AMPERSAND, "&", l.row, l.col)
		l.token_stream = append(l.token_stream, token1)
		l.col++
		token2 := createToken(T_MUT, "mut", l.row, l.col)
		l.token_stream = append(l.token_stream, token2)
		l.col += 3
		pos += 4
		return true, pos
	}

	// Handle postfix increment/decrement
	if pos+2 <= len(content) && content[pos:pos+2] == "++" {
		// Check if preceded by an identifier
		if pos > 0 && isWordChar(content[pos-1]) {
			// This is postfix increment, let DFA handle it
			return false, pos
		} else {
			// This is prefix increment
			token1 := createToken(T_PLUS, "+", l.row, l.col)
			l.token_stream = append(l.token_stream, token1)
			l.col++
			token2 := createToken(T_PLUS, "+", l.row, l.col)
			l.token_stream = append(l.token_stream, token2)
			l.col++
			pos += 2
			return true, pos
		}
	}

	if pos+2 <= len(content) && content[pos:pos+2] == "--" {
		// Check if preceded by an identifier
		if pos > 0 && isWordChar(content[pos-1]) {
			// This is postfix decrement, let DFA handle it
			return false, pos
		} else {
			// This is prefix decrement
			token1 := createToken(T_MINUS, "-", l.row, l.col)
			l.token_stream = append(l.token_stream, token1)
			l.col++
			token2 := createToken(T_MINUS, "-", l.row, l.col)
			l.token_stream = append(l.token_stream, token2)
			l.col++
			pos += 2
			return true, pos
		}
	}

	// Handle standalone ++ and -- operators (split them)
	if pos+2 <= len(content) && content[pos:pos+2] == "++" {
		token1 := createToken(T_PLUS, "+", l.row, l.col)
		l.token_stream = append(l.token_stream, token1)
		l.col++
		token2 := createToken(T_PLUS, "+", l.row, l.col)
		l.token_stream = append(l.token_stream, token2)
		l.col++
		pos += 2
		return true, pos
	}

	if pos+2 <= len(content) && content[pos:pos+2] == "--" {
		token1 := createToken(T_MINUS, "-", l.row, l.col)
		l.token_stream = append(l.token_stream, token1)
		l.col++
		token2 := createToken(T_MINUS, "-", l.row, l.col)
		l.token_stream = append(l.token_stream, token2)
		l.col++
		pos += 2
		return true, pos
	}

	if pos < len(content) && content[pos] == '%' {
		// Check if it's followed by a letter (then it's a register)
		if pos+1 < len(content) && ((content[pos+1] >= 'a' && content[pos+1] <= 'z') ||
			(content[pos+1] >= 'A' && content[pos+1] <= 'Z')) {
			// Check if we're in an ASM context (inside asm { ... })
			inASMContext := false
			// Look backwards to see if we're inside an asm block
			for i := pos - 1; i >= 0; i-- {
				if content[i] == '}' {
					break // Found closing brace, not in ASM context
				} else if content[i] == '{' {
					// Look for 'asm' before the opening brace
					j := i - 1
					for j >= 0 && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n') {
						j--
					}
					if j >= 2 && content[j-2:j+1] == "asm" {
						inASMContext = true
					}
					break
				}
			}

			if inASMContext {
				// Let the ASM register handling take care of it
			} else {
				// Not in ASM context, treat as modulo operator
				token := createToken(T_MODULO, "%", l.row, l.col)
				l.token_stream = append(l.token_stream, token)
				l.col++
				pos++
				return true, pos
			}
		} else {
			// Standalone % - treat as T_MODULO
			token := createToken(T_MODULO, "%", l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col++
			pos++
			return true, pos
		}
	}
	if pos+2 <= len(content) && content[pos:pos+2] == ":=" {
		token := createToken(T_DECLARE_ASSIGN, ":=", l.row, l.col)
		l.token_stream = append(l.token_stream, token)
		l.col += 2
		pos += 2
		return true, pos
	}
	if pos < len(content) && content[pos] == '"' {
		end := pos + 1
		for end < len(content) && content[end] != '"' {
			if content[end] == '\\' && end+1 < len(content) {
				// Skip escaped character
				end += 2
			} else {
				end++
			}
		}
		if end < len(content) {
			end++ // include closing quote
			stringText := content[pos:end]
			token := createToken(T_STRING_LITERAL, stringText, l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col += end - pos
			pos = end
			return true, pos
		}
	}
	if pos < len(content) && content[pos] == '\'' {
		end := pos + 1
		for end < len(content) && content[end] != '\'' {
			if content[end] == '\\' && end+1 < len(content) {
				// Skip escaped character
				end += 2
			} else {
				end++
			}
		}
		if end < len(content) {
			end++ // include closing quote
			charText := content[pos:end]
			token := createToken(T_CHAR_LITERAL, charText, l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col += end - pos
			pos = end
			return true, pos
		}
	}
	if pos < len(content) && content[pos] == '$' {
		// Check if it's followed by a digit (then it's an ASM immediate value)
		if pos+1 < len(content) && content[pos+1] >= '0' && content[pos+1] <= '9' {
			// Let the ASM immediate value handling take care of it
		} else {
			// Standalone $ - treat as T_DOLLAR
			token := createToken(T_DOLLAR, "$", l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col++
			pos++
			return true, pos
		}
	}
	if pos < len(content) && content[pos] == '$' {
		// Find the end of the immediate value
		end := pos + 1
		for end < len(content) && ((content[end] >= '0' && content[end] <= '9') ||
			(content[end] >= 'a' && content[end] <= 'f') ||
			(content[end] >= 'A' && content[end] <= 'F') ||
			content[end] == 'x' || content[end] == 'X') &&
			content[end] != ',' && content[end] != ' ' &&
			content[end] != '\t' && content[end] != '\n' && content[end] != '#' {
			end++
		}
		immediateText := content[pos:end]
		token := createToken(T_IDENTIFIER, immediateText, l.row, l.col) // Treat as identifier for ASM
		l.token_stream = append(l.token_stream, token)
		l.col += end - pos
		pos = end
		return true, pos
	}
	if pos < len(content) && content[pos] == '%' {
		// Find the end of the register name
		end := pos + 1
		parenCount := 0
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
			content[end] == '0') {

			// Handle parentheses for register names like %st(1)
			if content[end] == '(' {
				parenCount++
			} else if content[end] == ')' {
				parenCount--
				if parenCount < 0 {
					break // Unmatched closing parenthesis
				}
			}

			// Stop if we hit whitespace or other delimiters (but allow parentheses)
			if content[end] == ',' && parenCount == 0 {
				break
			} else if content[end] == ' ' || content[end] == '\t' || content[end] == '\n' || content[end] == '#' {
				break
			}

			end++
		}

		// Only create the token if we have a valid register name
		if end > pos+1 {
			registerText := content[pos:end]
			token := createToken(T_IDENTIFIER, registerText, l.row, l.col) // Treat as identifier for now
			l.token_stream = append(l.token_stream, token)
			l.col += end - pos
			pos = end
			return true, pos
		}
	}
	if pos < len(content) && content[pos] == '_' {
		// Check if it's followed by a word character (then it's part of an identifier)
		if pos+1 < len(content) && isWordChar(content[pos+1]) {
			// Let the normal identifier handling take care of it
		} else {
			// Standalone underscore - treat as T_UNDERSCORE
			token := createToken(T_UNDERSCORE, "_", l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col++
			pos++
			return true, pos
		}
	}
	if pos+3 <= len(content) && content[pos:pos+3] == ">>=" {
		token1 := createToken(T_GREATER_EQUAL, ">=", l.row, l.col)
		l.token_stream = append(l.token_stream, token1)
		l.col += 2
		token2 := createToken(T_GREATER_THAN, ">", l.row, l.col)
		l.token_stream = append(l.token_stream, token2)
		l.col++
		pos += 3
		return true, pos
	}
	if pos+2 <= len(content) && content[pos:pos+2] == "=>" && pos+3 <= len(content) && content[pos+2] == '>' {
		// Skip this position and let the next iteration handle the >>= pattern
		pos++
		return true, pos
	}
	if pos < len(content) && content[pos] == ',' {
		// Count consecutive commas
		consecutiveCommas := 0
		for i := pos; i < len(content) && content[i] == ','; i++ {
			consecutiveCommas++
		}

		if consecutiveCommas >= 2 {
			// Split consecutive commas: first comma as T_COMMA, rest as T_ERROR
			firstCommaToken := createToken(T_COMMA, ",", l.row, l.col)
			l.token_stream = append(l.token_stream, firstCommaToken)
			l.col++

			// Create T_ERROR token for remaining commas
			remainingCommas := strings.Repeat(",", consecutiveCommas-1)
			errorToken := createToken(T_ERROR, remainingCommas, l.row, l.col)
			l.token_stream = append(l.token_stream, errorToken)
			l.debug.DebugLog(fmt.Sprintf("Consecutive commas: %s + %s at %d:%d", firstCommaToken.lexeme, errorToken.lexeme, l.row, l.col), true)

			// Update position
			for i := 1; i < consecutiveCommas; i++ {
				l.col++
			}
			pos += consecutiveCommas
			return true, pos
		}
	}
	if pos < len(content) && content[pos] == '-' {
		// Check if we're in an ASM context (inside asm { ... })
		inASMContext := false
		// Look backwards to see if we're inside an asm block
		for i := pos - 1; i >= 0; i-- {
			if content[i] == '}' {
				break // Found closing brace, not in ASM context
			} else if content[i] == '{' {
				// Look for 'asm' before the opening brace
				j := i - 1
				for j >= 0 && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n') {
					j--
				}
				if j >= 2 && content[j-2:j+1] == "asm" {
					inASMContext = true
				}
				break
			}
		}

		if inASMContext {
			// Check if this negative number is followed by a parenthesis (memory addressing)
			if pos+1 < len(content) && content[pos+1] >= '0' && content[pos+1] <= '9' {
				numberEnd := pos + 1
				for numberEnd < len(content) && content[numberEnd] >= '0' && content[numberEnd] <= '9' {
					numberEnd++
				}

				if numberEnd < len(content) && content[numberEnd] == '(' {
					// This is ASM memory addressing with negative number, treat as identifier
					numberText := content[pos:numberEnd]
					token := createToken(T_IDENTIFIER, numberText, l.row, l.col)
					l.token_stream = append(l.token_stream, token)
					l.col += numberEnd - pos
					pos = numberEnd
					return true, pos
				}
			}
		}
	}
	if pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		// Check if we're in an ASM context (inside asm { ... })
		inASMContext := false
		// Look backwards to see if we're inside an asm block
		for i := pos - 1; i >= 0; i-- {
			if content[i] == '}' {
				break // Found closing brace, not in ASM context
			} else if content[i] == '{' {
				// Look for 'asm' before the opening brace
				j := i - 1
				for j >= 0 && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n') {
					j--
				}
				if j >= 2 && content[j-2:j+1] == "asm" {
					inASMContext = true
				}
				break
			}
		}

		// Check if this number is followed by a parenthesis (memory addressing)
		numberEnd := pos
		for numberEnd < len(content) && content[numberEnd] >= '0' && content[numberEnd] <= '9' {
			numberEnd++
		}

		if numberEnd < len(content) && content[numberEnd] == '(' {
			// This is ASM memory addressing, treat the number as identifier
			numberText := content[pos:numberEnd]
			token := createToken(T_IDENTIFIER, numberText, l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col += numberEnd - pos
			pos = numberEnd
			return true, pos
		} else if inASMContext {
			// In ASM context, treat standalone numbers as identifiers
			numberEnd := pos
			for numberEnd < len(content) && content[numberEnd] >= '0' && content[numberEnd] <= '9' {
				numberEnd++
			}
			numberText := content[pos:numberEnd]
			token := createToken(T_IDENTIFIER, numberText, l.row, l.col)
			l.token_stream = append(l.token_stream, token)
			l.col += numberEnd - pos
			pos = numberEnd
			return true, pos
		}
	}
	if pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		// Find the end of the number part
		numberEnd := pos
		for numberEnd < len(content) && content[numberEnd] >= '0' && content[numberEnd] <= '9' {
			numberEnd++
		}

		// Check if there's an identifier part after the number
		if numberEnd < len(content) && isWordChar(content[numberEnd]) {
			// Find the end of the identifier part
			identifierEnd := numberEnd
			for identifierEnd < len(content) && isWordChar(content[identifierEnd]) {
				identifierEnd++
			}

			// Create number token
			numberText := content[pos:numberEnd]
			numberToken := createToken(T_INT_LITERAL, numberText, l.row, l.col)
			l.token_stream = append(l.token_stream, numberToken)

			// Create identifier token
			identifierText := content[numberEnd:identifierEnd]
			identifierToken := createToken(T_IDENTIFIER, identifierText, l.row, l.col+len(numberText))
			l.token_stream = append(l.token_stream, identifierToken)

			// Update position
			for _, c := range content[pos:identifierEnd] {
				if c == '\n' {
					l.row++
					l.col = 1
				} else {
					l.col++
				}
			}
			pos = identifierEnd
			return true, pos
		}
	}
	return false, pos
}

// handleUnknownToken handles unknown or invalid tokens
func (l *Lexer) handleUnknownToken(content string, pos int) (bool, int) {
	// Handle Unicode characters properly
	rune, size := utf8.DecodeRuneInString(content[pos:])
	currentChar := string(rune)

	// Check if this is a valid Unicode character
	if rune == utf8.RuneError {
		// Invalid UTF-8 sequence, treat as single byte
		token := createToken(T_UNKNOWN, string(content[pos]), l.row, l.col)
		l.token_stream = append(l.token_stream, token)
		l.debug.DebugLog(fmt.Sprintf("Invalid UTF-8 token: %c at %d:%d", content[pos], l.row, l.col), true)
		l.col++
		return true, pos + 1
	} else if rune > 127 || (rune >= 0 && rune <= 31) || rune == 127 {
		// This is a Unicode character or control character, create a single T_ERROR token
		token := createToken(T_ERROR, currentChar, l.row, l.col)
		l.token_stream = append(l.token_stream, token)
		l.debug.DebugLog(fmt.Sprintf("Unicode/control error token: %s at %d:%d", currentChar, l.row, l.col), true)
		l.col++
		return true, pos + size
	} else {
		// Single byte ASCII character (32-126)
		token := createToken(T_UNKNOWN, currentChar, l.row, l.col)
		l.token_stream = append(l.token_stream, token)
		l.debug.DebugLog(fmt.Sprintf("Unknown token: %c at %d:%d", rune, l.row, l.col), true)
		l.col++
		return true, pos + size
	}
}
