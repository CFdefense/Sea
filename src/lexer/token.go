package lexer

import "regexp"

// Original regex pattern strings (for postfix conversion)
const (
	// Comments first (highest priority)
	SINGLE_LINE_COMMENT_PATTERN_STR = `^#[^\n]*`
	MULTI_LINE_COMMENT_PATTERN_STR  = `^/\*[\s\S]*?\*/`

	// String literals (must start with quote)
	STRING_PATTERN_STR          = `^"([^"\\]|\\.)*"`
	CHAR_PATTERN_STR            = `^'([^'\\]|\\.)'`
	ESCAPE_SEQUENCE_PATTERN_STR = `^\\[ntr\\"']`

	// Boolean literals
	BOOL_PATTERN_STR = `^(true|false)\b`

	// Numbers
	NUMBER_PATTERN_STR = `^(0x[0-9a-fA-F]+|0b[01]+|[0-9]+)`

	// Keywords (must be before identifiers)
	KEYWORD_PATTERN_STR = `^(if|else|while|do|for|match|enum|struct|const|void|int|bool|mut|return|default|break|continue|sizeof|asm)`

	// Constants (all uppercase)
	CONSTANT_PATTERN_STR = `^[A-Z][A-Z0-9_]*`

	// Identifiers (after keywords)
	IDENTIFIER_PATTERN_STR = `^[a-zA-Z_][a-zA-Z0-9_]*`

	// Operators (specific patterns first)
	OPERATOR_PATTERN_STR = `^(\+\+|--|==|!=|<=|>=|&&|\|\||->|<<|>>|//|[-+*/%<>!&|^~=]=|[-+*/%<>!&|^~])`

	// Punctuators
	PUNCTUATOR_PATTERN_STR = `^[{}\[\]();,.:?]`

	// Special tokens
	SPECIAL_PATTERN_STR = `^(@|&mut\b|&|\$|~|` + "`" + `)`

	// ASM patterns (lowest priority)
	ASM_INSTRUCTION_STR = `^[a-z][a-z0-9]*\b`
	ASM_REGISTER_STR    = `^%([re][a-z]{2}|[re]?[0-9]+|[re]?[ax|bx|cx|dx|si|di|sp|bp]|[re]?ip)\b`
	ASM_IMMEDIATE_STR   = `^\$?(0x[0-9a-fA-F]+|[0-9]+|[a-zA-Z_][a-zA-Z0-9_]*)\b`
	ASM_MEMORY_REF_STR  = `^\(\s*%[a-zA-Z0-9]+\s*\)`
	ASM_LABEL_STR       = `^[a-zA-Z_][a-zA-Z0-9_]*:`
)

// Regex patterns for token matching
var (
	// Patterns for each token type
	IDENTIFIER_PATTERN = regexp.MustCompile(IDENTIFIER_PATTERN_STR)

	// Operators including all arithmetic, logical, bitwise, and assignment operators
	OPERATOR_PATTERN = regexp.MustCompile(OPERATOR_PATTERN_STR)

	// Constants (all uppercase)
	CONSTANT_PATTERN = regexp.MustCompile(CONSTANT_PATTERN_STR)

	// All language keywords
	KEYWORD_PATTERN = regexp.MustCompile(KEYWORD_PATTERN_STR)

	// Numeric literals including hex and binary
	NUMBER_PATTERN = regexp.MustCompile(NUMBER_PATTERN_STR)

	// Boolean literals
	BOOL_PATTERN = regexp.MustCompile(BOOL_PATTERN_STR)

	// Punctuators including all delimiters and structural tokens
	PUNCTUATOR_PATTERN = regexp.MustCompile(PUNCTUATOR_PATTERN_STR)

	// Special tokens including decorators and reference operators
	SPECIAL_PATTERN = regexp.MustCompile(SPECIAL_PATTERN_STR)

	// String literals with escape sequences
	STRING_PATTERN = regexp.MustCompile(STRING_PATTERN_STR)

	// Character literals
	CHAR_PATTERN = regexp.MustCompile(CHAR_PATTERN_STR)

	// Escape sequences
	ESCAPE_SEQUENCE_PATTERN = regexp.MustCompile(ESCAPE_SEQUENCE_PATTERN_STR)

	// Comments
	SINGLE_LINE_COMMENT_PATTERN = regexp.MustCompile(SINGLE_LINE_COMMENT_PATTERN_STR)
	MULTI_LINE_COMMENT_PATTERN  = regexp.MustCompile(MULTI_LINE_COMMENT_PATTERN_STR)

	// ASM-specific patterns
	ASM_BLOCK_START = regexp.MustCompile(`^asm\s*{`)
	ASM_BLOCK_END   = regexp.MustCompile(`^}`)

	// Assembly instruction pattern (e.g., "mov", "push", "pop", "call")
	ASM_INSTRUCTION = regexp.MustCompile(ASM_INSTRUCTION_STR)

	// Register pattern (e.g., %rax, %rbx, %eax, %r8)
	ASM_REGISTER = regexp.MustCompile(ASM_REGISTER_STR)

	// Immediate value pattern (e.g., $123, $0xFF, label_name)
	ASM_IMMEDIATE = regexp.MustCompile(ASM_IMMEDIATE_STR)

	// Memory reference pattern (e.g., (%rax), 8(%rbp), -16(%rbp))
	ASM_MEMORY_REF = regexp.MustCompile(ASM_MEMORY_REF_STR)

	// Assembly label pattern (e.g., loop_start:, main:)
	ASM_LABEL = regexp.MustCompile(ASM_LABEL_STR)

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
	T_TERNARY_OPERATOR // Ternary operator (? :)
	T_UNARY_OPERATOR   // Unary operators (++, --, !, ~)
	T_BINARY_OPERATOR  // Binary operators (+, -, *, /, etc.)
	T_MEMBER_OPERATOR  // Member access operators (., ->)

	// Control flow
	// Assembly-specific tokens
	T_ASM_BLOCK       // asm { ... } block markers
	T_ASM_INSTRUCTION // Assembly mnemonics (mov, push, pop, etc.)
	T_ASM_REGISTER    // Register references (%rax, %rbp, etc.)
	T_ASM_IMMEDIATE   // Immediate values ($42, $0xFF)
	T_ASM_MEMORY_REF  // Memory references ((%rax), 8(%rbp))
	T_ASM_SEPARATOR   // Operand separators (comma)
	T_ASM_LABEL       // Labels for jumps and functions

	// Specific operator types
	T_PLUS           // Addition operator (+)
	T_MINUS          // Subtraction operator (-)
	T_MULTIPLY       // Multiplication operator (*)
	T_DIVIDE         // Division operator (/)
	T_MODULO         // Modulo operator (%)
	T_INT_DIVIDE     // Integer division operator (//)
	T_LEFT_SHIFT     // Left shift operator (<<)
	T_RIGHT_SHIFT    // Right shift operator (>>)
	T_EQUALS         // Equality operator (==)
	T_NOT_EQUALS     // Inequality operator (!=)
	T_LESS_THAN      // Less than operator (<)
	T_GREATER_THAN   // Greater than operator (>)
	T_LESS_EQUAL     // Less than or equal operator (<=)
	T_GREATER_EQUAL  // Greater than or equal operator (>=)
	T_AND            // Logical AND operator (&&)
	T_OR             // Logical OR operator (||)
	T_NOT            // Logical NOT operator (!)
	T_XOR            // XOR operator (^)
	T_ASSIGN         // Assignment operator (=)
	T_DECLARE_ASSIGN // Declaration assignment operator (:=)
	T_AT             // At symbol (@)
	T_MATCH_ARROW    // Match arrow operator (=>)
	T_ARROW          // Arrow operator (=>)

	// Specific punctuator types
	T_OPENING_BRACE   // Opening brace ({)
	T_CLOSING_BRACE   // Closing brace (})
	T_OPENING_PAREN   // Opening parenthesis (()
	T_CLOSING_PAREN   // Closing parenthesis ())
	T_OPENING_BRACKET // Opening bracket ([)
	T_CLOSING_BRACKET // Closing bracket (])
	T_COMMA           // Comma (,)
	T_SEMICOLON       // Semicolon (;)
	T_DOT             // Period (.)
	T_COLON           // Colon (:)
	T_QUESTION        // Question mark (?)

	// Specific special character types
	T_HASH      // Hash symbol (#)
	T_DOLLAR    // Dollar symbol ($)
	T_AMPERSAND // Ampersand (&)
	T_TILDE     // Tilde (~)
	T_BACKTICK  // Backtick (`)

	// Specific keyword types
	T_IF           // if keyword
	T_ELSE         // else keyword
	T_WHILE        // while keyword
	T_DO           // do keyword
	T_FOR          // for keyword
	T_MATCH        // match keyword
	T_ENUM         // enum keyword
	T_STRUCT       // struct keyword
	T_CONST        // const keyword
	T_VOID         // void keyword
	T_INT          // int keyword
	T_BOOL_KEYWORD // bool keyword
	T_MUT          // mut keyword
	T_RETURN       // return keyword
	T_DEFAULT      // default keyword
	T_BREAK        // break keyword
	T_CONTINUE     // continue keyword
	T_SIZEOF       // sizeof keyword
	T_ASM          // asm keyword

	// Error and additional type tokens
	T_ERROR     // Error token for invalid input
	T_INT_TYPE  // int type keyword (alias for T_INT)
	T_BOOL_TYPE // bool type keyword (alias for T_BOOL_KEYWORD)
	T_VOID_TYPE // void type keyword (alias for T_VOID)
	T_FUNCTION  // function keyword

	// Specific literal types
	T_INT_LITERAL        // Integer literal
	T_BOOL_LITERAL       // Boolean literal
	T_STRING_LITERAL_RAW // Raw string literal
	T_CHAR_LITERAL_RAW   // Raw character literal
	T_UNDERSCORE         // Underscore pattern for match arms
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
	case T_MEMBER_OPERATOR:
		return "T_MEMBER_OPERATOR"
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
	case T_PLUS:
		return "T_PLUS"
	case T_MINUS:
		return "T_MINUS"
	case T_MULTIPLY:
		return "T_MULTIPLY"
	case T_DIVIDE:
		return "T_DIVIDE"
	case T_MODULO:
		return "T_MODULO"
	case T_INT_DIVIDE:
		return "T_INT_DIVIDE"
	case T_LEFT_SHIFT:
		return "T_LEFT_SHIFT"
	case T_RIGHT_SHIFT:
		return "T_RIGHT_SHIFT"
	case T_EQUALS:
		return "T_EQUALS"
	case T_NOT_EQUALS:
		return "T_NOT_EQUALS"
	case T_LESS_THAN:
		return "T_LESS_THAN"
	case T_GREATER_THAN:
		return "T_GREATER_THAN"
	case T_LESS_EQUAL:
		return "T_LESS_EQUAL"
	case T_GREATER_EQUAL:
		return "T_GREATER_EQUAL"
	case T_AND:
		return "T_AND"
	case T_OR:
		return "T_OR"
	case T_NOT:
		return "T_NOT"
	case T_XOR:
		return "T_XOR"
	case T_ASSIGN:
		return "T_ASSIGN"
	case T_DECLARE_ASSIGN:
		return "T_DECLARE_ASSIGN"
	case T_AT:
		return "T_AT"
	case T_MATCH_ARROW:
		return "T_MATCH_ARROW"
	case T_ARROW:
		return "T_ARROW"
	case T_OPENING_BRACE:
		return "T_OPENING_BRACE"
	case T_CLOSING_BRACE:
		return "T_CLOSING_BRACE"
	case T_OPENING_PAREN:
		return "T_OPENING_PAREN"
	case T_CLOSING_PAREN:
		return "T_CLOSING_PAREN"
	case T_OPENING_BRACKET:
		return "T_OPENING_BRACKET"
	case T_CLOSING_BRACKET:
		return "T_CLOSING_BRACKET"
	case T_COMMA:
		return "T_COMMA"
	case T_SEMICOLON:
		return "T_SEMICOLON"
	case T_DOT:
		return "T_DOT"
	case T_COLON:
		return "T_COLON"
	case T_QUESTION:
		return "T_QUESTION"
	case T_HASH:
		return "T_HASH"
	case T_DOLLAR:
		return "T_DOLLAR"
	case T_AMPERSAND:
		return "T_AMPERSAND"
	case T_TILDE:
		return "T_TILDE"
	case T_BACKTICK:
		return "T_BACKTICK"
	case T_IF:
		return "T_IF"
	case T_ELSE:
		return "T_ELSE"
	case T_WHILE:
		return "T_WHILE"
	case T_DO:
		return "T_DO"
	case T_FOR:
		return "T_FOR"
	case T_MATCH:
		return "T_MATCH"
	case T_ENUM:
		return "T_ENUM"
	case T_STRUCT:
		return "T_STRUCT"
	case T_CONST:
		return "T_CONST"
	case T_VOID:
		return "T_VOID"
	case T_INT:
		return "T_INT"
	case T_BOOL_KEYWORD:
		return "T_BOOL_KEYWORD"
	case T_MUT:
		return "T_MUT"
	case T_RETURN:
		return "T_RETURN"
	case T_DEFAULT:
		return "T_DEFAULT"
	case T_BREAK:
		return "T_BREAK"
	case T_CONTINUE:
		return "T_CONTINUE"
	case T_SIZEOF:
		return "T_SIZEOF"
	case T_ASM:
		return "T_ASM"
	case T_ERROR:
		return "T_ERROR"
	case T_INT_TYPE:
		return "T_INT_TYPE"
	case T_BOOL_TYPE:
		return "T_BOOL_TYPE"
	case T_VOID_TYPE:
		return "T_VOID_TYPE"
	case T_FUNCTION:
		return "T_FUNCTION"
	case T_INT_LITERAL:
		return "T_INT_LITERAL"
	case T_BOOL_LITERAL:
		return "T_BOOL_LITERAL"
	case T_STRING_LITERAL_RAW:
		return "T_STRING_LITERAL_RAW"
	case T_CHAR_LITERAL_RAW:
		return "T_CHAR_LITERAL_RAW"
	case T_UNDERSCORE:
		return "T_UNDERSCORE"
	default:
		return "T_UNKNOWN"
	}
}
