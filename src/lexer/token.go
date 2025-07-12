package lexer

import "regexp"

// Regex patterns for token matching
var (
	// Patterns for each token type
	IDENTIFIER_PATTERN = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)

	// Operators including all arithmetic, logical, bitwise, and assignment operators
	OPERATOR_PATTERN = regexp.MustCompile(`^(\+\+|--|==|!=|<=|>=|&&|\|\||->|<<|>>|//|[-+*/%<>!&|^~=]=|[-+*/%<>!&|^~])`)

	// Constants (all uppercase)
	CONSTANT_PATTERN = regexp.MustCompile(`^[A-Z][A-Z0-9_]*`)

	// All language keywords
	KEYWORD_PATTERN = regexp.MustCompile(`^(if|else|while|do|for|match|enum|struct|const|void|int|bool|mut|return|switch|case|default|break|continue|goto|sizeof|asm)\b`)

	// Numeric literals including hex and binary
	NUMBER_PATTERN = regexp.MustCompile(`^(0x[0-9a-fA-F]+|0b[01]+|[0-9]+)`)

	// Boolean literals
	BOOL_PATTERN = regexp.MustCompile(`^(true|false)\b`)

	// Punctuators including all delimiters and structural tokens
	PUNCTUATOR_PATTERN = regexp.MustCompile(`^[{}\[\]();,.:?]`)

	// Special tokens including decorators and reference operators
	SPECIAL_PATTERN = regexp.MustCompile(`^(@|&mut\b|&|#|\$)`)

	// String literals with escape sequences
	STRING_PATTERN = regexp.MustCompile(`^"([^"\\]|\\.)*"`)

	// Character literals
	CHAR_PATTERN = regexp.MustCompile(`^'([^'\\]|\\.)'`)

	// Escape sequences
	ESCAPE_SEQUENCE_PATTERN = regexp.MustCompile(`^\\[ntr\\"']`)

	// Comments
	SINGLE_LINE_COMMENT_PATTERN = regexp.MustCompile(`^//[^\n]*`)
	MULTI_LINE_COMMENT_PATTERN  = regexp.MustCompile(`^/\*[\s\S]*?\*/`)

	// ASM-specific patterns
	ASM_BLOCK_START = regexp.MustCompile(`^asm\s*{`)
	ASM_BLOCK_END   = regexp.MustCompile(`^}`)

	// Assembly instruction pattern (e.g., "mov", "push", "pop", "call")
	ASM_INSTRUCTION = regexp.MustCompile(`^[a-z][a-z0-9]*\b`)

	// Register pattern (e.g., %rax, %rbx, %eax, %r8)
	ASM_REGISTER = regexp.MustCompile(`^%([re][a-z]{2}|[re]?[0-9]+|[re]?[ax|bx|cx|dx|si|di|sp|bp]|[re]?ip)\b`)

	// Immediate value pattern (e.g., $123, $0xFF, label_name)
	ASM_IMMEDIATE = regexp.MustCompile(`^\$?(0x[0-9a-fA-F]+|[0-9]+|[a-zA-Z_][a-zA-Z0-9_]*)\b`)

	// Memory reference pattern (e.g., (%rax), 8(%rbp), -16(%rbp))
	ASM_MEMORY_REF = regexp.MustCompile(`^(-?\d+)?\s*\(\s*%[a-zA-Z0-9]+\s*\)`)

	// Assembly label pattern (e.g., loop_start:, main:)
	ASM_LABEL = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*:`)

	// Operand separator
	ASM_SEPARATOR = regexp.MustCompile(`^,\s*`)
)

// 'Enum' for token types
// using iota assigns successive integer values
// ie: first defined token is 0, second is 1 etc
type TokenType int

// token category enum definitions
const (
	// Basic tokens
	T_IDENTIFIER TokenType = iota // Names for functions, variables, types (e.g., main, x, myVar)
	T_OPERATOR                    // Mathematical and logical operators (+, -, *, /, %, ==, !=, &&, ||)
	T_CONSTANT                    // Named constants and enum values (e.g., RED, GREEN, BLUE in enum)
	T_KEYWORD                     // Language keywords (if, while, match, enum, struct, void, int, bool)
	T_LITERAL                     // Direct values (numbers, booleans: 42, true, false)
	T_PUNCTUATOR                  // Structural characters ({, }, (, ), [, ], ;, ,)
	T_SPECIAL                     // Special tokens (@, #, $) and decorators
	T_UNKNOWN                     // Invalid or unrecognized tokens

	// Comments and documentation
	T_SINGLE_LINE_COMMENT // Single line comments (//)
	T_MULTI_LINE_COMMENT  // Multi-line comments (/* */)

	// String-related
	T_STRING_LITERAL      // String literals with escape sequences
	T_CHAR_LITERAL        // Character literals ('a', '\n')
	T_ESCAPE_SEQUENCE     // Escape sequences in strings (\n, \t, \", etc.)
	T_RAW_STRING_LITERAL  // Raw string literals (r"...") - for regex
	T_BYTE_STRING_LITERAL // Byte string literals (b"...") - for regex

	// Type-related
	T_TYPE_QUALIFIER  // Type qualifiers (mut)
	T_TYPE_IDENTIFIER // Built-in type names (int, bool, void)
	T_ARRAY_TYPE      // Array type declarations
	T_POINTER_TYPE    // Pointer type declarations
	T_FUNCTION_TYPE   // Function type declarations

	// Expression-related
	T_TERNARY_OPERATOR    // Ternary operator (? :)
	T_UNARY_OPERATOR      // Unary operators (++, --, !, ~)
	T_BINARY_OPERATOR     // Binary operators (+, -, *, /, etc.)
	T_ASSIGNMENT_OPERATOR // Assignment operators (=, +=, -=, etc.)
	T_MEMBER_OPERATOR     // Member access operators (., ->)

	// Control flow
	T_LABEL         // Labels for goto statements
	T_CASE_LABEL    // Case labels in switch statements
	T_DEFAULT_LABEL // Default label in switch statements

	// Assembly-specific tokens
	T_ASM_BLOCK       // asm { ... } block markers
	T_ASM_INSTRUCTION // Assembly mnemonics (mov, push, pop, etc.)
	T_ASM_REGISTER    // Register references (%rax, %rbp, etc.)
	T_ASM_IMMEDIATE   // Immediate values ($42, $0xFF)
	T_ASM_MEMORY_REF  // Memory references ((%rax), 8(%rbp))
	T_ASM_SEPARATOR   // Operand separators (comma)
	T_ASM_LABEL       // Labels for jumps and functions
)

// TokenRegexDef represents a regex definition for a token
type TokenRegexDef struct {
	Name      string
	Pattern   string
	Postfix   string
	TokenType TokenType
}

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

// String converts TokenType to its string representation
func (tt TokenType) String() string {
	switch tt {
	case T_IDENTIFIER:
		return "T_IDENTIFIER"
	case T_OPERATOR:
		return "T_OPERATOR"
	case T_CONSTANT:
		return "T_CONSTANT"
	case T_KEYWORD:
		return "T_KEYWORD"
	case T_LITERAL:
		return "T_LITERAL"
	case T_PUNCTUATOR:
		return "T_PUNCTUATOR"
	case T_SPECIAL:
		return "T_SPECIAL"
	case T_UNKNOWN:
		return "T_UNKNOWN"
	case T_SINGLE_LINE_COMMENT:
		return "T_SINGLE_LINE_COMMENT"
	case T_MULTI_LINE_COMMENT:
		return "T_MULTI_LINE_COMMENT"
	case T_STRING_LITERAL:
		return "T_STRING_LITERAL"
	case T_CHAR_LITERAL:
		return "T_CHAR_LITERAL"
	case T_ESCAPE_SEQUENCE:
		return "T_ESCAPE_SEQUENCE"
	case T_RAW_STRING_LITERAL:
		return "T_RAW_STRING_LITERAL"
	case T_BYTE_STRING_LITERAL:
		return "T_BYTE_STRING_LITERAL"
	case T_TYPE_QUALIFIER:
		return "T_TYPE_QUALIFIER"
	case T_TYPE_IDENTIFIER:
		return "T_TYPE_IDENTIFIER"
	case T_ARRAY_TYPE:
		return "T_ARRAY_TYPE"
	case T_POINTER_TYPE:
		return "T_POINTER_TYPE"
	case T_FUNCTION_TYPE:
		return "T_FUNCTION_TYPE"
	case T_TERNARY_OPERATOR:
		return "T_TERNARY_OPERATOR"
	case T_UNARY_OPERATOR:
		return "T_UNARY_OPERATOR"
	case T_BINARY_OPERATOR:
		return "T_BINARY_OPERATOR"
	case T_ASSIGNMENT_OPERATOR:
		return "T_ASSIGNMENT_OPERATOR"
	case T_MEMBER_OPERATOR:
		return "T_MEMBER_OPERATOR"
	case T_LABEL:
		return "T_LABEL"
	case T_CASE_LABEL:
		return "T_CASE_LABEL"
	case T_DEFAULT_LABEL:
		return "T_DEFAULT_LABEL"
	case T_ASM_BLOCK:
		return "T_ASM_BLOCK"
	case T_ASM_INSTRUCTION:
		return "T_ASM_INSTRUCTION"
	case T_ASM_REGISTER:
		return "T_ASM_REGISTER"
	case T_ASM_IMMEDIATE:
		return "T_ASM_IMMEDIATE"
	case T_ASM_MEMORY_REF:
		return "T_ASM_MEMORY_REF"
	case T_ASM_SEPARATOR:
		return "T_ASM_SEPARATOR"
	case T_ASM_LABEL:
		return "T_ASM_LABEL"
	default:
		return "T_UNKNOWN"
	}
}
