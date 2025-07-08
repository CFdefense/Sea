package lexer

// 'Enum' for token types
// using iota assigns successive integer values
// ie: first defined token is 0, second is 1 etc
type TokenType int

// token category enum definitions
const (
	T_IDENTIFIER TokenType = iota // Names for functions, variables, types (e.g., main, x, myVar)
	T_OPERATOR                    // Mathematical and logical operators (+, -, *, /, %, ==, !=, &&, ||)
	T_CONSTANT                    // Named constants and enum values (e.g., RED, GREEN, BLUE in enum)
	T_KEYWORD                     // Language keywords (if, while, match, enum, struct, void, int, bool)
	T_LITERAL                     // Direct values (numbers, booleans: 42, true, false)
	T_PUNCTUATOR                  // Structural characters ({, }, (, ), [, ], ;, ,)
	T_SPECIAL                     // Special tokens (@, #, $) and decorators
	T_UNKNOWN                     // Invalid or unrecognized tokens
)

// Individual token object
type Token struct {
	token_type TokenType
	lexeme     string
	row        int
	col        int
}

func (t *Token) GetTokenContent() string {
	return t.lexeme
}

func (t *Token) GetTokenType() TokenType {
	return t.token_type
}

func (t *Token) GetRow() int {
	return t.row
}

func (t *Token) GetCol() int {
	return t.col
}
