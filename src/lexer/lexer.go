package lexer

// lexer context object
type Lexer struct {
	token_stream []Token
	content      string
	row          int
	col          int
}

// lexer object constructor
func InitializeLexer() *Lexer {
	return &Lexer{
		token_stream: []Token{},
		content:      "",
		row:          1,
		col:          1,
	}
}

// function responsible for all things lexical analysis
func (l *Lexer) LexicalAnalysis() {

	// begin lexer scan of target file
	// might want to change this to target directory
	// depending on future limits/scope
	l.Scan()

	// begin lexer analyze
	// this will turn our scanned content into
	// a nice stream of tokens
	l.Analyze()
}

// function to allow the lexer to scan
// current implementation is file based
// might want to change to directory based
func (l *Lexer) Scan() {

}

// function to begin analysis
// will result in scan -> token stream
func (l *Lexer) Analyze() {

}

// function to reset a lexer
// mostly used in repeated test executions
// can also be used in between compiling multiple files
func (l *Lexer) ResetLexer() {
	l.token_stream = []Token{}
	l.content = ""
	l.row = 1
	l.col = 1
}

// function to get the token stream
func (l *Lexer) GetTokenStream() []Token {
	return l.token_stream
}
