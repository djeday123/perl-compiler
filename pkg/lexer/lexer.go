// Package lexer implements a lexical analyzer for Perl source code.
package lexer

import (
	"github.com/djeday123/perl-compiler/pkg/token"
)

// Lexer represents a lexical analyzer for Perl.
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

// readChar reads the next character from input and advances position.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing position.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharN returns the character at position n ahead without advancing.
func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = token.Token{Type: token.MATCH, Literal: "=~", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.FATCOMMA, Literal: "=>", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.ASSIGN, l.ch, tok.Line, tok.Column)
		}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = token.Token{Type: token.INC, Literal: "++", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.PLUSEQ, Literal: "+=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.PLUS, l.ch, tok.Line, tok.Column)
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.DEC, Literal: "--", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MINUSEQ, Literal: "-=", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: "->", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.MINUS, l.ch, tok.Line, tok.Column)
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NE, Literal: "!=", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = token.Token{Type: token.NOTMATCH, Literal: "!~", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.NOT, l.ch, tok.Line, tok.Column)
		}
	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.DIVEQ, Literal: "/=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.SLASH, l.ch, tok.Line, tok.Column)
		}
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = token.Token{Type: token.POWEQ, Literal: "**=", Line: tok.Line, Column: tok.Column}
			} else {
				tok = token.Token{Type: token.POWER, Literal: "**", Line: tok.Line, Column: tok.Column}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MULEQ, Literal: "*=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.ASTERISK, l.ch, tok.Line, tok.Column)
		}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MODEQ, Literal: "%=", Line: tok.Line, Column: tok.Column}
		} else if isLetter(l.peekChar()) {
			// Hash sigil
			tok.Type = token.HASH
			l.readChar()
			tok.Literal = "%" + l.readIdentifier()
			return tok
		} else {
			tok = newToken(token.MODULO, l.ch, tok.Line, tok.Column)
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			if l.peekChar() == '>' {
				l.readChar()
				tok = token.Token{Type: token.NUMEQ, Literal: "<=>", Line: tok.Line, Column: tok.Column}
			} else {
				tok = token.Token{Type: token.LE, Literal: "<=", Line: tok.Line, Column: tok.Column}
			}
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = token.Token{Type: token.LSHIFT, Literal: "<<", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.LT, l.ch, tok.Line, tok.Column)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GE, Literal: ">=", Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.RSHIFT, Literal: ">>", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.GT, l.ch, tok.Line, tok.Column)
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: "&&", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.BITAND, l.ch, tok.Line, tok.Column)
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: "||", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.BITOR, l.ch, tok.Line, tok.Column)
		}
	case '^':
		tok = newToken(token.BITXOR, l.ch, tok.Line, tok.Column)
	case '~':
		tok = newToken(token.BITNOT, l.ch, tok.Line, tok.Column)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = token.Token{Type: token.ELLIPSIS, Literal: "...", Line: tok.Line, Column: tok.Column}
			} else {
				tok = token.Token{Type: token.RANGE, Literal: "..", Line: tok.Line, Column: tok.Column}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.CONCATEQ, Literal: ".=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(token.CONCAT, l.ch, tok.Line, tok.Column)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, tok.Line, tok.Column)
	case ':':
		tok = newToken(token.COLON, l.ch, tok.Line, tok.Column)
	case ',':
		tok = newToken(token.COMMA, l.ch, tok.Line, tok.Column)
	case '(':
		tok = newToken(token.LPAREN, l.ch, tok.Line, tok.Column)
	case ')':
		tok = newToken(token.RPAREN, l.ch, tok.Line, tok.Column)
	case '{':
		tok = newToken(token.LBRACE, l.ch, tok.Line, tok.Column)
	case '}':
		tok = newToken(token.RBRACE, l.ch, tok.Line, tok.Column)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, tok.Line, tok.Column)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, tok.Line, tok.Column)
	case '?':
		tok = newToken(token.TERNARY, l.ch, tok.Line, tok.Column)
	case '#':
		tok.Type = token.COMMENT
		tok.Literal = l.readComment()
		return tok
	case '$':
		tok.Type = token.SCALAR
		l.readChar()
		tok.Literal = "$" + l.readIdentifier()
		return tok
	case '@':
		tok.Type = token.ARRAY
		l.readChar()
		tok.Literal = "@" + l.readIdentifier()
		return tok
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString('"')
		return tok
	case '\'':
		tok.Type = token.STRING
		tok.Literal = l.readString('\'')
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber(tok.Line, tok.Column)
		} else {
			tok = newToken(token.ILLEGAL, l.ch, tok.Line, tok.Column)
		}
	}

	l.readChar()
	return tok
}

// readIdentifier reads an identifier from the input.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float) from the input.
func (l *Lexer) readNumber(line, column int) token.Token {
	position := l.position
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for scientific notation
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[position:l.position]
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal, Line: line, Column: column}
	}
	return token.Token{Type: token.INT, Literal: literal, Line: line, Column: column}
}

// readString reads a string literal from the input.
func (l *Lexer) readString(quote byte) string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar()
		}
	}
	str := l.input[position:l.position]
	l.readChar()
	return str
}

// readComment reads a comment from the input.
func (l *Lexer) readComment() string {
	position := l.position + 1
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// newToken creates a new token with the given type and character.
func newToken(tokenType token.TokenType, ch byte, line, column int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: line, Column: column}
}

// isLetter returns true if the character is a letter or underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit returns true if the character is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
