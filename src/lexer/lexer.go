package lexer

// Lexer context object
type Lexer struct {
	token_stream []Token
	content      string
	row          int
	col          int
}

// Lexer constructor
func CreateLexer(content string) *Lexer {
	// create the new lexer object
	lexer := &Lexer{
		token_stream: []Token{},
		content:      content,
		row:          1,
		col:          1,
	}

	// begin lexing
	lexer.start_lexing(content)

	// return the new lexer object
	return lexer
}

// function to start lexing
func (l *Lexer) start_lexing(content string) {
	l.content = content
	l.row = 1
	l.col = 1
	// TODO: implement lexing
}

// function to get the token stream
func (l *Lexer) GetTokenStream() []Token {
	return l.token_stream
}
