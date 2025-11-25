// Package lexer provides tokenization for Perl source code.
package lexer

// Lexer performs lexical analysis on Perl source code.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: EQ, Literal: "==", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = Token{Type: MATCH, Literal: "=~", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = Token{Type: FATARROW, Literal: "=>", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: ASSIGN, Literal: "=", Line: tok.Line, Column: tok.Column}
		}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = Token{Type: PLUSPLUS, Literal: "++", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: PLUS, Literal: "+", Line: tok.Line, Column: tok.Column}
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = Token{Type: MINUSMINUS, Literal: "--", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = Token{Type: ARROW, Literal: "->", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: MINUS, Literal: "-", Line: tok.Line, Column: tok.Column}
		}
	case '*':
		tok = Token{Type: ASTERISK, Literal: "*", Line: tok.Line, Column: tok.Column}
	case '/':
		tok = Token{Type: SLASH, Literal: "/", Line: tok.Line, Column: tok.Column}
	case '%':
		tok = Token{Type: PERCENT, Literal: "%", Line: tok.Line, Column: tok.Column}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: "!=", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = Token{Type: NOTMATCH, Literal: "!~", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: BANG, Literal: "!", Line: tok.Line, Column: tok.Column}
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: LT_EQ, Literal: "<=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: LT, Literal: "<", Line: tok.Line, Column: tok.Column}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: GT_EQ, Literal: ">=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: GT, Literal: ">", Line: tok.Line, Column: tok.Column}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = Token{Type: AND, Literal: "&&", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = Token{Type: OR, Literal: "||", Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '.':
		tok = Token{Type: CONCAT, Literal: ".", Line: tok.Line, Column: tok.Column}
	case ',':
		tok = Token{Type: COMMA, Literal: ",", Line: tok.Line, Column: tok.Column}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: ";", Line: tok.Line, Column: tok.Column}
	case ':':
		tok = Token{Type: COLON, Literal: ":", Line: tok.Line, Column: tok.Column}
	case '(':
		tok = Token{Type: LPAREN, Literal: "(", Line: tok.Line, Column: tok.Column}
	case ')':
		tok = Token{Type: RPAREN, Literal: ")", Line: tok.Line, Column: tok.Column}
	case '{':
		tok = Token{Type: LBRACE, Literal: "{", Line: tok.Line, Column: tok.Column}
	case '}':
		tok = Token{Type: RBRACE, Literal: "}", Line: tok.Line, Column: tok.Column}
	case '[':
		tok = Token{Type: LBRACKET, Literal: "[", Line: tok.Line, Column: tok.Column}
	case ']':
		tok = Token{Type: RBRACKET, Literal: "]", Line: tok.Line, Column: tok.Column}
	case '$':
		tok = Token{Type: SCALAR, Literal: "$", Line: tok.Line, Column: tok.Column}
	case '@':
		tok = Token{Type: ARRAY, Literal: "@", Line: tok.Line, Column: tok.Column}
	case '#':
		tok.Type = COMMENT
		tok.Literal = l.readComment()
		return tok
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString('"')
		return tok
	case '\'':
		tok.Type = STRING
		tok.Literal = l.readString('\'')
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (string, TokenType) {
	position := l.position
	tokenType := INT

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = FLOAT
		l.readChar() // consume the '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], tokenType
}

func (l *Lexer) readString(quote byte) string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar() // skip escaped character
		}
	}
	result := l.input[position:l.position]
	l.readChar()
	return result
}

func (l *Lexer) readComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
