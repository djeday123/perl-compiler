package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes Perl source code.
// Lexer, Perl kaynak kodunu tokenize eder.
type Lexer struct {
	input   string // Source code / Kaynak kod
	file    string // Filename / Dosya adı
	pos     int    // Current position in input / input'taki geçerli konum
	readPos int    // Reading position (after current char) / Okuma konumu
	line    int    // Current line number / Geçerli satır numarası
	column  int    // Current column number / Geçerli sütun numarası
	ch      rune   // Current character / Geçerli karakter

	// Context for disambiguation
	// Belirsizlik giderme için bağlam
	lastToken TokenType // Previous token type / Önceki token türü
}

// New creates a new lexer for the given input.
// New, verilen input için yeni bir lexer oluşturur.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		file:   "<input>",
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// NewFile creates a lexer with filename for error reporting.
// NewFile, hata raporlaması için dosya adı ile lexer oluşturur.
func NewFile(input, filename string) *Lexer {
	l := New(input)
	l.file = filename
	return l
}

// readChar advances to the next character.
// readChar, sonraki karaktere ilerler.
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch, _ = utf8.DecodeRuneInString(l.input[l.readPos:])
	}
	l.pos = l.readPos
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.readPos += utf8.RuneLen(l.ch)
}

// peekChar returns next character without advancing.
// peekChar, ilerlemeden sonraki karakteri döndürür.
func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return ch
}

// peekChars returns next n characters without advancing.
// peekChars, ilerlemeden sonraki n karakteri döndürür.
// func (l *Lexer) peekChars(n int) string {
// 	end := l.readPos + n
// 	if end > len(l.input) {
// 		end = len(l.input)
// 	}
// 	return l.input[l.readPos:end]
// }

// skipWhitespace skips spaces and tabs (not newlines).
// skipWhitespace, boşlukları ve tabları atlar (satır sonlarını değil).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips a comment until end of line.
// skipComment, satır sonuna kadar yorumu atlar.
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// NextToken returns the next token.
// NextToken, sonraki tokeni döndürür.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	// Skip comments
	// Yorumları atla
	if l.ch == '#' {
		l.skipComment()
		l.skipWhitespace()
	}

	tok := Token{
		Line:   l.line,
		Column: l.column,
		File:   l.file,
	}

	switch l.ch {
	case 0:
		tok.Type = TokEOF

	case '\n':
		tok.Type = TokNewline
		tok.Value = "\n"
		l.readChar()

	// Single character tokens
	case '(':
		tok.Type = TokLParen
		tok.Value = "("
		l.readChar()
	case ')':
		tok.Type = TokRParen
		tok.Value = ")"
		l.readChar()
	case '[':
		tok.Type = TokLBracket
		tok.Value = "["
		l.readChar()
	case ']':
		tok.Type = TokRBracket
		tok.Value = "]"
		l.readChar()
	case '{':
		tok.Type = TokLBrace
		tok.Value = "{"
		l.readChar()
	case '}':
		tok.Type = TokRBrace
		tok.Value = "}"
		l.readChar()
	case ';':
		tok.Type = TokSemi
		tok.Value = ";"
		l.readChar()
	case ',':
		tok.Type = TokComma
		tok.Value = ","
		l.readChar()
	case '?':
		tok.Type = TokQuestion
		tok.Value = "?"
		l.readChar()
	case '~':
		tok.Type = TokBitNot
		tok.Value = "~"
		l.readChar()
	case '\\':
		tok.Type = TokBackslash
		tok.Value = "\\"
		l.readChar()

	// Multi-character operators
	case '+':
		tok = l.readPlus()
	case '-':
		tok = l.readMinus()
	case '*':
		tok = l.readStar()
	case '/':
		tok = l.readSlash()
	case '%':
		tok = l.readPercent()
	case '.':
		tok = l.readDot()
	case '=':
		tok = l.readEquals()
	case '!':
		tok = l.readBang()
	case '<':
		tok = l.readLess()
	case '>':
		tok = l.readGreater()
	case '&':
		tok = l.readAmpersand()
	case '|':
		tok = l.readPipe()
	case '^':
		tok = l.readCaret()
	case ':':
		tok = l.readColon()

	// Variables
	case '$':
		tok = l.readScalar()
	case '@':
		tok = l.readArray()

	// String literals
	case '"':
		tok = l.readDoubleQuotedString()
	case '\'':
		tok = l.readSingleQuotedString()
	case '`':
		tok = l.readBacktickString()

	default:
		if isDigit(l.ch) {
			tok = l.readNumber()
		} else if l.ch == 's' && l.peekChar() == '/' {
			tok = l.readSubst()
		} else if l.ch == 'm' && l.peekChar() == '/' {
			tok = l.readMatchOp()
		} else if isIdentStart(l.ch) {
			tok = l.readIdentifier()
		} else {
			tok.Type = TokError
			tok.Value = fmt.Sprintf("unexpected character: %c", l.ch)
			l.readChar()
		}
	}

	l.lastToken = tok.Type
	return tok
}

// ============================================================
// Operator readers
// Operatör okuyucuları
// ============================================================

func (l *Lexer) readPlus() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '+':
		tok.Type = TokIncr
		tok.Value = "++"
		l.readChar()
	case '=':
		tok.Type = TokPlusEq
		tok.Value = "+="
		l.readChar()
	default:
		tok.Type = TokPlus
		tok.Value = "+"
	}
	return tok
}

func (l *Lexer) readMinus() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '-':
		tok.Type = TokDecr
		tok.Value = "--"
		l.readChar()
	case '=':
		tok.Type = TokMinusEq
		tok.Value = "-="
		l.readChar()
	case '>':
		tok.Type = TokArrow
		tok.Value = "->"
		l.readChar()
	default:
		tok.Type = TokMinus
		tok.Value = "-"
	}
	return tok
}

func (l *Lexer) readStar() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '*':
		l.readChar()
		if l.ch == '=' {
			tok.Type = TokStarStarEq
			tok.Value = "**="
			l.readChar()
		} else {
			tok.Type = TokStarStar
			tok.Value = "**"
		}
	case '=':
		tok.Type = TokStarEq
		tok.Value = "*="
		l.readChar()
	default:
		// Could be glob or multiplication
		// Glob veya çarpma olabilir
		tok.Type = TokStar
		tok.Value = "*"
	}
	return tok
}

func (l *Lexer) readSlash() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}

	// Check for // or //= first (defined-or) - before regex check
	// Önce // veya //= kontrol et (defined-or) - regex kontrolünden önce
	if l.peekChar() == '/' {
		l.readChar() // consume first /
		l.readChar() // consume second /
		if l.ch == '=' {
			tok.Type = TokDefinedOrEq
			tok.Value = "//="
			l.readChar()
		} else {
			tok.Type = TokDefinedOr
			tok.Value = "//"
		}
		return tok
	}

	// Check if this could be a regex
	// Bu bir regex olabilir mi kontrol et
	if l.expectRegex() {
		return l.readRegex('/')
	}

	l.readChar()
	if l.ch == '=' {
		tok.Type = TokSlashEq
		tok.Value = "/="
		l.readChar()
	} else {
		tok.Type = TokSlash
		tok.Value = "/"
	}
	return tok
}

func (l *Lexer) readPercent() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()

	// Check for hash variable
	// Hash değişkeni kontrol et
	if isIdentStart(l.ch) {
		name := l.readIdentName()
		tok.Type = TokHash
		tok.Value = "%" + name
		return tok
	}

	if l.ch == '=' {
		tok.Type = TokPercentEq
		tok.Value = "%="
		l.readChar()
	} else {
		tok.Type = TokPercent
		tok.Value = "%"
	}
	return tok
}

func (l *Lexer) readDot() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '.':
		l.readChar()
		if l.ch == '.' {
			tok.Type = TokRange3
			tok.Value = "..."
			l.readChar()
		} else {
			tok.Type = TokRange
			tok.Value = ".."
		}
	case '=':
		tok.Type = TokDotEq
		tok.Value = ".="
		l.readChar()
	default:
		tok.Type = TokDot
		tok.Value = "."
	}
	return tok
}

func (l *Lexer) readEquals() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '=':
		tok.Type = TokEq
		tok.Value = "=="
		l.readChar()
	case '~':
		tok.Type = TokMatch
		tok.Value = "=~"
		l.readChar()
	case '>':
		tok.Type = TokFatArrow
		tok.Value = "=>"
		l.readChar()
	default:
		tok.Type = TokAssign
		tok.Value = "="
	}
	return tok
}

func (l *Lexer) readBang() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '=':
		tok.Type = TokNe
		tok.Value = "!="
		l.readChar()
	case '~':
		tok.Type = TokNotMatch
		tok.Value = "!~"
		l.readChar()
	default:
		tok.Type = TokNot
		tok.Value = "!"
	}
	return tok
}

func (l *Lexer) readLess() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '<':
		l.readChar()
		if l.ch == '=' {
			tok.Type = TokLeftShiftEq
			tok.Value = "<<="
			l.readChar()
		} else {
			tok.Type = TokLeftShift
			tok.Value = "<<"
		}
	case '=':
		l.readChar()
		if l.ch == '>' {
			tok.Type = TokSpaceship
			tok.Value = "<=>"
			l.readChar()
		} else {
			tok.Type = TokLe
			tok.Value = "<="
		}
	case '>':
		// <> - read from ARGV/STDIN
		tok.Type = TokDiamond
		tok.Value = "<>"
		l.readChar()
	case '$':
		// <$fh> - read from filehandle variable
		tok.Type = TokReadLine
		var sb strings.Builder
		sb.WriteRune(l.ch)
		l.readChar()
		for l.ch != '>' && l.ch != 0 {
			sb.WriteRune(l.ch)
			l.readChar()
		}
		tok.Value = sb.String()
		if l.ch == '>' {
			l.readChar()
		}
	default:
		// Check if it's <FH> (bareword filehandle)
		if isIdentStart(l.ch) {
			tok.Type = TokReadLine
			var sb strings.Builder
			for l.ch != '>' && l.ch != 0 && !isSpace(l.ch) {
				sb.WriteRune(l.ch)
				l.readChar()
			}
			tok.Value = sb.String()
			if l.ch == '>' {
				l.readChar()
			}
		} else {
			tok.Type = TokLt
			tok.Value = "<"
		}
	}
	return tok
}

func (l *Lexer) readGreater() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '>':
		l.readChar()
		if l.ch == '=' {
			tok.Type = TokRightShiftEq
			tok.Value = ">>="
			l.readChar()
		} else {
			tok.Type = TokRightShift
			tok.Value = ">>"
		}
	case '=':
		tok.Type = TokGe
		tok.Value = ">="
		l.readChar()
	default:
		tok.Type = TokGt
		tok.Value = ">"
	}
	return tok
}

func (l *Lexer) readAmpersand() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()

	// Check for code reference &sub
	// Kod referansı &sub kontrol et
	if isIdentStart(l.ch) {
		name := l.readIdentName()
		tok.Type = TokCode
		tok.Value = "&" + name
		return tok
	}

	switch l.ch {
	case '&':
		l.readChar()
		if l.ch == '=' {
			tok.Type = TokAndEq
			tok.Value = "&&="
			l.readChar()
		} else {
			tok.Type = TokAnd
			tok.Value = "&&"
		}
	case '=':
		tok.Type = TokBitAndEq
		tok.Value = "&="
		l.readChar()
	default:
		tok.Type = TokBitAnd
		tok.Value = "&"
	}
	return tok
}

func (l *Lexer) readPipe() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	switch l.ch {
	case '|':
		l.readChar()
		if l.ch == '=' {
			tok.Type = TokOrEq
			tok.Value = "||="
			l.readChar()
		} else {
			tok.Type = TokOr
			tok.Value = "||"
		}
	case '=':
		tok.Type = TokBitOrEq
		tok.Value = "|="
		l.readChar()
	default:
		tok.Type = TokBitOr
		tok.Value = "|"
	}
	return tok
}

func (l *Lexer) readCaret() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	if l.ch == '=' {
		tok.Type = TokBitXorEq
		tok.Value = "^="
		l.readChar()
	} else {
		tok.Type = TokBitXor
		tok.Value = "^"
	}
	return tok
}

func (l *Lexer) readColon() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar()
	if l.ch == ':' {
		tok.Type = TokDoubleColon
		tok.Value = "::"
		l.readChar()
	} else {
		tok.Type = TokColon
		tok.Value = ":"
	}
	return tok
}

// ============================================================
// Variable readers
// Değişken okuyucuları
// ============================================================

func (l *Lexer) readScalar() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar() // Skip $

	// Special variables: $_, $@, $!, $$, etc.
	// Özel değişkenler: $_, $@, $!, $$, vb.
	switch l.ch {
	case '_', '@', '!', '$', '?', '"', '/', '\\', '&', '`', '\'', '+', '.', '|', '-', '^', '~', '=', '%', ':':
		tok.Type = TokSpecialVar
		tok.Value = "$" + string(l.ch)
		l.readChar()
		return tok
	case '#':
		// $#array - array length
		l.readChar()
		if isIdentStart(l.ch) {
			name := l.readIdentName()
			tok.Type = TokArrayLen
			tok.Value = "$#" + name
		} else {
			tok.Type = TokSpecialVar
			tok.Value = "$#"
		}
		return tok
	case '{':
		// ${var} - explicit variable name
		// ${var} - açık değişken adı
		l.readChar()
		name := l.readIdentName()
		if l.ch == '}' {
			l.readChar()
		}
		tok.Type = TokScalar
		tok.Value = "$" + name
		return tok
	}

	// Regular scalar $var or $Pkg::var
	// Normal skaler $var veya $Pkg::var
	if isIdentStart(l.ch) || isDigit(l.ch) {
		name := l.readIdentName()

		// Check for $1, $2, etc.
		// $1, $2, vb. kontrol et
		if len(name) > 0 && isDigit(rune(name[0])) {
			tok.Type = TokSpecialVar
		} else {
			tok.Type = TokScalar
		}
		tok.Value = "$" + name
	} else {
		tok.Type = TokError
		tok.Value = "expected variable name after $"
	}

	return tok
}

func (l *Lexer) readArray() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	l.readChar() // Skip @

	// Special: @_ @ARGV @ISA
	switch l.ch {
	case '_':
		tok.Type = TokSpecialVar
		tok.Value = "@_"
		l.readChar()
		return tok
	case '{':
		l.readChar()
		name := l.readIdentName()
		if l.ch == '}' {
			l.readChar()
		}
		tok.Type = TokArray
		tok.Value = "@" + name
		return tok
	}

	if isIdentStart(l.ch) {
		name := l.readIdentName()
		tok.Type = TokArray
		tok.Value = "@" + name
	} else {
		tok.Type = TokError
		tok.Value = "expected variable name after @"
	}

	return tok
}

// ============================================================
// String readers
// String okuyucuları
// ============================================================

func (l *Lexer) readDoubleQuotedString() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokString}
	l.readChar() // Skip opening "

	var sb strings.Builder
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '$':
				sb.WriteByte('$')
			case '@':
				sb.WriteByte('@')
			default:
				sb.WriteByte('\\')
				sb.WriteRune(l.ch)
			}
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}

	if l.ch == '"' {
		l.readChar() // Skip closing "
	}

	tok.Value = sb.String()
	return tok
}

func (l *Lexer) readSingleQuotedString() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokRawString}
	l.readChar() // Skip opening '

	var sb strings.Builder
	for l.ch != '\'' && l.ch != 0 {
		if l.ch == '\\' && l.peekChar() == '\'' {
			l.readChar()
			sb.WriteByte('\'')
		} else if l.ch == '\\' && l.peekChar() == '\\' {
			l.readChar()
			sb.WriteByte('\\')
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}

	if l.ch == '\'' {
		l.readChar() // Skip closing '
	}

	tok.Value = sb.String()
	return tok
}

func (l *Lexer) readBacktickString() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokString}
	l.readChar() // Skip opening `

	var sb strings.Builder
	for l.ch != '`' && l.ch != 0 {
		sb.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch == '`' {
		l.readChar()
	}

	tok.Value = sb.String()
	return tok
}

// ============================================================
// Number reader
// Sayı okuyucu
// ============================================================

func (l *Lexer) readNumber() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}

	var sb strings.Builder
	isFloat := false

	// Check for hex, octal, binary
	// Hex, octal, binary kontrol et
	if l.ch == '0' {
		sb.WriteRune(l.ch)
		l.readChar()

		switch l.ch {
		case 'x', 'X':
			// Hexadecimal
			sb.WriteRune(l.ch)
			l.readChar()
			for isHexDigit(l.ch) || l.ch == '_' {
				if l.ch != '_' {
					sb.WriteRune(l.ch)
				}
				l.readChar()
			}
			tok.Type = TokInteger
			tok.Value = sb.String()
			return tok

		case 'b', 'B':
			// Binary
			sb.WriteRune(l.ch)
			l.readChar()
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				if l.ch != '_' {
					sb.WriteRune(l.ch)
				}
				l.readChar()
			}
			tok.Type = TokInteger
			tok.Value = sb.String()
			return tok

		case 'o', 'O':
			// Octal (explicit)
			sb.WriteRune(l.ch)
			l.readChar()
			for isOctalDigit(l.ch) || l.ch == '_' {
				if l.ch != '_' {
					sb.WriteRune(l.ch)
				}
				l.readChar()
			}
			tok.Type = TokInteger
			tok.Value = sb.String()
			return tok
		}
	}

	// Read integer part
	// Tamsayı kısmını oku
	for isDigit(l.ch) || l.ch == '_' {
		if l.ch != '_' {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}

	// Check for decimal point
	// Ondalık nokta kontrol et
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		sb.WriteRune(l.ch)
		l.readChar()
		for isDigit(l.ch) || l.ch == '_' {
			if l.ch != '_' {
				sb.WriteRune(l.ch)
			}
			l.readChar()
		}
	}

	// Check for exponent
	// Üs kontrol et
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		sb.WriteRune(l.ch)
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			sb.WriteRune(l.ch)
			l.readChar()
		}
		for isDigit(l.ch) {
			sb.WriteRune(l.ch)
			l.readChar()
		}
	}

	if isFloat {
		tok.Type = TokFloat
	} else {
		tok.Type = TokInteger
	}
	tok.Value = sb.String()
	return tok
}

// ============================================================
// Identifier reader
// Tanımlayıcı okuyucu
// ============================================================

func (l *Lexer) readIdentifier() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file}
	name := l.readIdentName()

	tok.Type = LookupKeyword(name)
	tok.Value = name
	return tok
}

func (l *Lexer) readIdentName() string {
	var sb strings.Builder
	for isIdentChar(l.ch) {
		sb.WriteRune(l.ch)
		l.readChar()
	}

	// Handle Package::Name
	for l.ch == ':' && l.peekChar() == ':' {
		sb.WriteRune(l.ch)
		l.readChar()
		sb.WriteRune(l.ch)
		l.readChar()
		for isIdentChar(l.ch) {
			sb.WriteRune(l.ch)
			l.readChar()
		}
	}

	return sb.String()
}

// ============================================================
// Regex reader
// Regex okuyucu
// ============================================================

func (l *Lexer) readRegex(delim rune) Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokRegex}
	l.readChar() // Skip opening delimiter

	var sb strings.Builder
	for l.ch != delim && l.ch != 0 {
		if l.ch == '\\' {
			sb.WriteRune(l.ch)
			l.readChar()
			if l.ch != 0 {
				sb.WriteRune(l.ch)
				l.readChar()
			}
		} else {
			sb.WriteRune(l.ch)
			l.readChar()
		}
	}

	pattern := sb.String()

	if l.ch == delim {
		l.readChar()
	}

	// Read modifiers
	// Değiştiricileri oku
	var mods strings.Builder
	for l.ch == 'i' || l.ch == 'm' || l.ch == 's' || l.ch == 'x' || l.ch == 'g' || l.ch == 'o' {
		mods.WriteRune(l.ch)
		l.readChar()
	}

	if mods.Len() > 0 {
		tok.Value = pattern + "/" + mods.String()
	} else {
		tok.Value = pattern
	}

	return tok
}

// expectRegex returns true if the next / should be a regex.
// expectRegex, sonraki / regex olmalıysa true döndürür.
func (l *Lexer) expectRegex() bool {
	switch l.lastToken {
	case TokEOF, TokNewline, TokSemi, TokLParen, TokLBracket, TokLBrace,
		TokComma, TokAssign, TokMatch, TokNotMatch, TokAnd, TokOr,
		TokNot, TokQuestion, TokColon, TokIf, TokUnless, TokWhile,
		TokUntil, TokFor, TokForeach, TokAndWord, TokOrWord, TokNotWord:
		return true
	}
	return false
}

func (l *Lexer) readSubst() Token {
	tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokSubst}
	l.readChar() // skip 's'
	delim := l.ch
	l.readChar() // skip opening delimiter

	// Read pattern
	var pattern strings.Builder
	for l.ch != delim && l.ch != 0 {
		if l.ch == '\\' {
			pattern.WriteRune(l.ch)
			l.readChar()
			if l.ch != 0 {
				pattern.WriteRune(l.ch)
				l.readChar()
			}
		} else {
			pattern.WriteRune(l.ch)
			l.readChar()
		}
	}
	l.readChar() // skip middle delimiter

	// Read replacement
	var replacement strings.Builder
	for l.ch != delim && l.ch != 0 {
		if l.ch == '\\' {
			replacement.WriteRune(l.ch)
			l.readChar()
			if l.ch != 0 {
				replacement.WriteRune(l.ch)
				l.readChar()
			}
		} else {
			replacement.WriteRune(l.ch)
			l.readChar()
		}
	}
	l.readChar() // skip closing delimiter

	// Read flags
	var flags strings.Builder
	for l.ch == 'g' || l.ch == 'i' || l.ch == 'm' || l.ch == 's' || l.ch == 'x' || l.ch == 'e' {
		flags.WriteRune(l.ch)
		l.readChar()
	}

	// Format: pattern/replacement/flags
	tok.Value = pattern.String() + "/" + replacement.String() + "/" + flags.String()
	return tok
}

func (l *Lexer) readMatchOp() Token {
	//tok := Token{Line: l.line, Column: l.column, File: l.file, Type: TokRegex}
	l.readChar() // skip 'm'
	return l.readRegex(l.ch)
}

// ============================================================
// Helper functions
// Yardımcı fonksiyonlar
// ============================================================

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdentChar(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
