// Package token defines constants representing the lexical tokens of the Perl language.
package token

// TokenType represents the type of a token.
type TokenType string

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Token types
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers and literals
	IDENT   TokenType = "IDENT"   // variable names, subroutine names
	INT     TokenType = "INT"     // integer literals
	FLOAT   TokenType = "FLOAT"   // floating-point literals
	STRING  TokenType = "STRING"  // string literals
	REGEX   TokenType = "REGEX"   // regular expression literals
	SCALAR  TokenType = "SCALAR"  // $variable
	ARRAY   TokenType = "ARRAY"   // @variable
	HASH    TokenType = "HASH"    // %variable
	COMMENT TokenType = "COMMENT" // # comment

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	MODULO   TokenType = "%"
	POWER    TokenType = "**"

	// Comparison operators
	EQ     TokenType = "=="
	NE     TokenType = "!="
	LT     TokenType = "<"
	GT     TokenType = ">"
	LE     TokenType = "<="
	GE     TokenType = ">="
	NUMEQ  TokenType = "<=>"
	STREQ  TokenType = "eq"
	STRNE  TokenType = "ne"
	STRLT  TokenType = "lt"
	STRGT  TokenType = "gt"
	STRLE  TokenType = "le"
	STRGE  TokenType = "ge"
	STRCMP TokenType = "cmp"

	// Logical operators
	AND    TokenType = "&&"
	OR     TokenType = "||"
	NOT    TokenType = "!"
	ANDKW  TokenType = "and"
	ORKW   TokenType = "or"
	NOTKW  TokenType = "not"
	BITNOT TokenType = "~"
	BITAND TokenType = "&"
	BITOR  TokenType = "|"
	BITXOR TokenType = "^"
	LSHIFT TokenType = "<<"
	RSHIFT TokenType = ">>"

	// Assignment operators
	PLUSEQ  TokenType = "+="
	MINUSEQ TokenType = "-="
	MULEQ   TokenType = "*="
	DIVEQ   TokenType = "/="
	MODEQ   TokenType = "%="
	POWEQ   TokenType = "**="
	CONCAT  TokenType = "."
	CONCATEQ TokenType = ".="
	REPEAT  TokenType = "x"
	REPEATEQ TokenType = "x="

	// Increment/Decrement
	INC TokenType = "++"
	DEC TokenType = "--"

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
	ARROW     TokenType = "->"
	FATCOMMA  TokenType = "=>"
	RANGE     TokenType = ".."
	ELLIPSIS  TokenType = "..."
	TERNARY   TokenType = "?"
	MATCH     TokenType = "=~"
	NOTMATCH  TokenType = "!~"

	// Keywords
	MY       TokenType = "my"
	OUR      TokenType = "our"
	LOCAL    TokenType = "local"
	STATE    TokenType = "state"
	SUB      TokenType = "sub"
	IF       TokenType = "if"
	ELSIF    TokenType = "elsif"
	ELSE     TokenType = "else"
	UNLESS   TokenType = "unless"
	WHILE    TokenType = "while"
	UNTIL    TokenType = "until"
	FOR      TokenType = "for"
	FOREACH  TokenType = "foreach"
	DO       TokenType = "do"
	LAST     TokenType = "last"
	NEXT     TokenType = "next"
	REDO     TokenType = "redo"
	RETURN   TokenType = "return"
	PACKAGE  TokenType = "package"
	USE      TokenType = "use"
	REQUIRE  TokenType = "require"
	BEGIN    TokenType = "BEGIN"
	END      TokenType = "END"
	PRINT    TokenType = "print"
	SAY      TokenType = "say"
	DIE      TokenType = "die"
	WARN     TokenType = "warn"
	DEFINED  TokenType = "defined"
	UNDEF    TokenType = "undef"
	QWSTART  TokenType = "qw"
)

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"my":       MY,
	"our":      OUR,
	"local":    LOCAL,
	"state":    STATE,
	"sub":      SUB,
	"if":       IF,
	"elsif":    ELSIF,
	"else":     ELSE,
	"unless":   UNLESS,
	"while":    WHILE,
	"until":    UNTIL,
	"for":      FOR,
	"foreach":  FOREACH,
	"do":       DO,
	"last":     LAST,
	"next":     NEXT,
	"redo":     REDO,
	"return":   RETURN,
	"package":  PACKAGE,
	"use":      USE,
	"require":  REQUIRE,
	"BEGIN":    BEGIN,
	"END":      END,
	"print":    PRINT,
	"say":      SAY,
	"die":      DIE,
	"warn":     WARN,
	"defined":  DEFINED,
	"undef":    UNDEF,
	"eq":       STREQ,
	"ne":       STRNE,
	"lt":       STRLT,
	"gt":       STRGT,
	"le":       STRLE,
	"ge":       STRGE,
	"cmp":      STRCMP,
	"and":      ANDKW,
	"or":       ORKW,
	"not":      NOTKW,
	"x":        REPEAT,
	"qw":       QWSTART,
}

// LookupIdent returns the token type for the given identifier.
// If the identifier is a keyword, it returns the corresponding keyword token type.
// Otherwise, it returns IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsKeyword returns true if the given token type is a keyword.
func IsKeyword(t TokenType) bool {
	switch t {
	case MY, OUR, LOCAL, STATE, SUB, IF, ELSIF, ELSE, UNLESS,
		WHILE, UNTIL, FOR, FOREACH, DO, LAST, NEXT, REDO, RETURN,
		PACKAGE, USE, REQUIRE, BEGIN, END, PRINT, SAY, DIE, WARN,
		DEFINED, UNDEF, STREQ, STRNE, STRLT, STRGT, STRLE, STRGE,
		STRCMP, ANDKW, ORKW, NOTKW, REPEAT, QWSTART:
		return true
	}
	return false
}
