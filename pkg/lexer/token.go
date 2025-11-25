// Package lexer provides tokenization for Perl source code.
package lexer

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Identifiers and literals
	IDENT  // variable names, function names
	INT    // integer literals
	FLOAT  // floating point literals
	STRING // string literals
	REGEX  // regular expression literals

	// Operators
	ASSIGN     // =
	PLUS       // +
	MINUS      // -
	ASTERISK   // *
	SLASH      // /
	PERCENT    // %
	BANG       // !
	LT         // <
	GT         // >
	EQ         // ==
	NOT_EQ     // !=
	LT_EQ      // <=
	GT_EQ      // >=
	AND        // &&
	OR         // ||
	CONCAT     // .
	ARROW      // ->
	FATARROW   // =>
	PLUSPLUS   // ++
	MINUSMINUS // --
	MATCH      // =~
	NOTMATCH   // !~

	// Delimiters
	COMMA     // ,
	SEMICOLON // ;
	COLON     // :
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]

	// Sigils
	SCALAR // $
	ARRAY  // @
	HASH   // %

	// Keywords
	SUB
	MY
	OUR
	LOCAL
	IF
	ELSIF
	ELSE
	UNLESS
	WHILE
	UNTIL
	FOR
	FOREACH
	DO
	RETURN
	LAST
	NEXT
	REDO
	USE
	REQUIRE
	PACKAGE
	PRINT
	SAY
	DIE
	WARN
)

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"sub":     SUB,
	"my":      MY,
	"our":     OUR,
	"local":   LOCAL,
	"if":      IF,
	"elsif":   ELSIF,
	"else":    ELSE,
	"unless":  UNLESS,
	"while":   WHILE,
	"until":   UNTIL,
	"for":     FOR,
	"foreach": FOREACH,
	"do":      DO,
	"return":  RETURN,
	"last":    LAST,
	"next":    NEXT,
	"redo":    REDO,
	"use":     USE,
	"require": REQUIRE,
	"package": PACKAGE,
	"print":   PRINT,
	"say":     SAY,
	"die":     DIE,
	"warn":    WARN,
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

// String returns a string representation of the token type.
func (t TokenType) String() string {
	names := map[TokenType]string{
		ILLEGAL:    "ILLEGAL",
		EOF:        "EOF",
		COMMENT:    "COMMENT",
		IDENT:      "IDENT",
		INT:        "INT",
		FLOAT:      "FLOAT",
		STRING:     "STRING",
		REGEX:      "REGEX",
		ASSIGN:     "ASSIGN",
		PLUS:       "PLUS",
		MINUS:      "MINUS",
		ASTERISK:   "ASTERISK",
		SLASH:      "SLASH",
		PERCENT:    "PERCENT",
		BANG:       "BANG",
		LT:         "LT",
		GT:         "GT",
		EQ:         "EQ",
		NOT_EQ:     "NOT_EQ",
		LT_EQ:      "LT_EQ",
		GT_EQ:      "GT_EQ",
		AND:        "AND",
		OR:         "OR",
		CONCAT:     "CONCAT",
		ARROW:      "ARROW",
		FATARROW:   "FATARROW",
		PLUSPLUS:   "PLUSPLUS",
		MINUSMINUS: "MINUSMINUS",
		MATCH:      "MATCH",
		NOTMATCH:   "NOTMATCH",
		COMMA:      "COMMA",
		SEMICOLON:  "SEMICOLON",
		COLON:      "COLON",
		LPAREN:     "LPAREN",
		RPAREN:     "RPAREN",
		LBRACE:     "LBRACE",
		RBRACE:     "RBRACE",
		LBRACKET:   "LBRACKET",
		RBRACKET:   "RBRACKET",
		SCALAR:     "SCALAR",
		ARRAY:      "ARRAY",
		HASH:       "HASH",
		SUB:        "SUB",
		MY:         "MY",
		OUR:        "OUR",
		LOCAL:      "LOCAL",
		IF:         "IF",
		ELSIF:      "ELSIF",
		ELSE:       "ELSE",
		UNLESS:     "UNLESS",
		WHILE:      "WHILE",
		UNTIL:      "UNTIL",
		FOR:        "FOR",
		FOREACH:    "FOREACH",
		DO:         "DO",
		RETURN:     "RETURN",
		LAST:       "LAST",
		NEXT:       "NEXT",
		REDO:       "REDO",
		USE:        "USE",
		REQUIRE:    "REQUIRE",
		PACKAGE:    "PACKAGE",
		PRINT:      "PRINT",
		SAY:        "SAY",
		DIE:        "DIE",
		WARN:       "WARN",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "UNKNOWN"
}
