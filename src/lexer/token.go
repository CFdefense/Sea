package lexer

import "regexp"

// Regex patterns for token matching
var (
	// Patterns for each token type
	IdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)

	// Operators including all arithmetic, logical, bitwise, and assignment operators
	OperatorPattern = regexp.MustCompile(`^(\+\+|--|==|!=|<=|>=|&&|\|\||->|<<|>>|//|[-+*/%<>!&|^~=]=|[-+*/%<>!&|^~])`)

	// Constants (all uppercase)
	ConstantPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*`)

	// All language keywords
	KeywordPattern = regexp.MustCompile(`^(if|else|while|do|for|match|enum|struct|const|void|int|bool|mut|return|switch|case|default|break|continue|goto|sizeof|asm)\b`)

	// Numeric literals including hex and binary
	NumberPattern = regexp.MustCompile(`^(0x[0-9a-fA-F]+|0b[01]+|[0-9]+)`)

	// Boolean literals
	BoolPattern = regexp.MustCompile(`^(true|false)\b`)

	// Punctuators including all delimiters and structural tokens
	PunctuatorPattern = regexp.MustCompile(`^[{}\[\]();,.:?]`)

	// Special tokens including decorators and reference operators
	SpecialPattern = regexp.MustCompile(`^(@|&mut\b|&|#|\$)`)

	// String literals with escape sequences
	StringPattern = regexp.MustCompile(`^"([^"\\]|\\.)*"`)

	// Comments
	SingleLineCommentPattern = regexp.MustCompile(`^//[^\n]*`)
	MultiLineCommentPattern  = regexp.MustCompile(`^/\*[\s\S]*?\*/`)
)

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
