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

// token category enum definitions
const (
	T_IDENTIFIER TokenType = iota
	T_OPERATOR
	T_CONSTANT
	T_KEYWORD
	T_LITERAL
	T_PUNCTUATOR
	T_SPECIAL
	T_UNKNOWN
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
