package lexer

// Based on my research I believe the best way to go about the lexer is as follows
//
// 1. Define allowed tokens using regex
// 2. Convert Regex to NFA (Thompson's Algorithm)
// 3. Compute ε-Closure of States
// 4. Convert NFA to DFA 
//
//

// 'Enum' for token types
// using iota assigns successive integer values
// ie: first defined token is 0, second is 1 etc
type TokenType int

// token enum definitions
const (
	T_OPENING_BRACE TokenType = iota	// '{'
	T_CLOSING_BRACE						// '}'
	T_OPENING_PAREN						// '('
	T_CLOSING_PAREN						// ')'
	T_OPENING_BRACK						// '['
	T_CLOSING_BRACK						// ']'
	T_OP_PLUS							// '+'
	T_OP_DASH							// '
	T_OP_SLASH
	T_OP_
	
)

// Individual token object
type Token struct {
	token_type    TokenType
	token_content string
	row           int
	col           int
}

func (t *Token) GetTokenContent() string {
	return t.token_content
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
