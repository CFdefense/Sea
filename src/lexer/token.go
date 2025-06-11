package lexer

// 'Enum' for token types
// using iota assigns successive integer values
// ie: first defined token is 0, second is 1 etc
type TokenType int

// token enum definitions
const (
	T_OPENING_BRACE TokenType = iota
	T_CLOSING_BRACE
	T_OPENING_PAREN
	T_CLOSING_PAREN
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
