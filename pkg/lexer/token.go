// Package lexer implements Perl tokenization.
// Paket lexer, Perl tokenizasyonunu uygular.
package lexer

// TokenType represents the type of a token.
// TokenType, bir tokenin türünü temsil eder.
type TokenType int

const (
	// Special
	TokEOF TokenType = iota
	TokError
	TokNewline

	// Literals
	TokInteger   // 42, 0x2A, 0b101010, 0o52
	TokFloat     // 3.14, 6.02e23
	TokString    // 'single', "double", q(), qq()
	TokRawString // Raw string (no interpolation)
	TokRegex     // /pattern/, m//, qr//
	TokHeredoc   // <<EOF
	TokVersion   // v5.36, 5.036

	// Identifiers and keywords
	TokIdent      // identifier
	TokPackageRef // Package::Name
	TokLabel      // LABEL:

	// Variables
	TokScalar   // $var
	TokArray    // @arr
	TokHash     // %hash
	TokCode     // &sub
	TokGlob     // *glob
	TokArrayLen // $#arr

	// Special variables
	TokSpecialVar // $_, $@, $!, $$, etc.

	// Operators - Arithmetic
	TokPlus     // +
	TokMinus    // -
	TokStar     // *
	TokSlash    // /
	TokPercent  // %
	TokStarStar // **

	// Operators - String
	TokDot // .
	TokX   // x (string repeat)

	// Operators - Comparison (numeric)
	TokEq        // ==
	TokNe        // !=
	TokLt        //
	TokLe        // <=
	TokGt        // >
	TokGe        // >=
	TokSpaceship // <=>

	// Operators - Comparison (string)
	TokStrEq // eq
	TokStrNe // ne
	TokStrLt // lt
	TokStrLe // le
	TokStrGt // gt
	TokStrGe // ge
	TokCmp   // cmp

	// Operators - Logical
	TokAnd       // &&
	TokOr        // ||
	TokNot       // !
	TokAndWord   // and
	TokOrWord    // or
	TokNotWord   // not
	TokDefinedOr // //

	// Operators - Bitwise
	TokBitAnd     // &
	TokBitOr      // |
	TokBitXor     // ^
	TokBitNot     // ~
	TokLeftShift  //
	TokRightShift // >>

	// Operators - Assignment
	TokAssign       // =
	TokPlusEq       // +=
	TokMinusEq      // -=
	TokStarEq       // *=
	TokSlashEq      // /=
	TokPercentEq    // %=
	TokStarStarEq   // **=
	TokDotEq        // .=
	TokXEq          // x=
	TokAndEq        // &&=
	TokOrEq         // ||=
	TokDefinedOrEq  // //=
	TokBitAndEq     // &=
	TokBitOrEq      // |=
	TokBitXorEq     // ^=
	TokLeftShiftEq  // <<=
	TokRightShiftEq // >>=

	// Operators - Increment/Decrement
	TokIncr // ++
	TokDecr // --

	// Operators - Range
	TokRange  // ..
	TokRange3 // ...

	// Operators - Misc
	TokArrow       // ->
	TokFatArrow    // =>
	TokQuestion    // ?
	TokColon       // :
	TokDoubleColon // ::
	TokBackslash   // \
	TokMatch       // =~
	TokNotMatch    // !~
	TokComma       // ,

	// Brackets
	TokLParen   // (
	TokRParen   // )
	TokLBracket // [
	TokRBracket // ]
	TokLBrace   // {
	TokRBrace   // }

	// Delimiters
	TokSemi // ;

	// Keywords - Control flow
	TokIf
	TokUnless
	TokElse
	TokElsif
	TokWhile
	TokUntil
	TokFor
	TokForeach
	TokDo
	TokLast
	TokNext
	TokRedo
	TokReturn
	TokGoto

	// Keywords - Declarations
	TokMy
	TokOur
	TokLocal
	TokState
	TokSub
	TokPackage
	TokUse
	TokNo
	TokRequire
	TokBEGIN
	TokEND
	TokCHECK
	TokINIT
	TokUNITCHECK

	// Keywords - Misc
	TokQw // qw()
	TokEval
	TokDie
	TokWarn
	TokPrint
	TokSay
	TokOpen
	TokClose
	TokRead
	TokDiamond  // <>
	TokReadLine // <$fh> or <FH>
	TokWrite
	TokDefined
	TokUndef
	TokRef
	TokBless
	TokTie
	TokUntie
	TokTied
	TokWantarray
	TokCaller

	// scalar (keyword, not sigil)
	// Özel
	TokScalarKw
	TokGiven
	TokWhen
	TokDefault

	// Array/Hash functions
	TokShift
	TokUnshift
	TokPop
	TokPush
	TokSplice
	TokKeys
	TokValues
	TokEach
	TokExists
	TokDelete
	TokSort
	TokReverse
	TokMap
	TokGrep
	TokJoin
	TokSplit

	// String functions
	TokLength
	TokSubstr
	TokIndex
	TokRindex
	TokLc
	TokUc
	TokLcfirst
	TokUcfirst
	TokChomp
	TokChop
	TokChr
	TokOrd
	TokHex
	TokOct
	TokPack
	TokUnpack
	TokSprintf

	// Numeric functions
	TokAbs
	TokInt
	TokSqrt
	TokRand
	TokSrand
	TokSin
	TokCos
	TokAtan2
	TokExp
	TokLog

	// Misc functions
	TokLocaltime
	TokGmtime
	TokTime
	TokSleep
	TokExit
	TokSystem
	TokExec
	TokFork
	TokWait
	TokKill

	TokSubst // s/pattern/replacement/
)

// Token represents a lexical token.
// Token, bir leksikal tokeni temsil eder.
type Token struct {
	Type   TokenType
	Value  string // Literal value / Literal değer
	Line   int    // Source line (1-indexed) / Kaynak satır
	Column int    // Source column (1-indexed) / Kaynak sütun
	File   string // Source filename / Kaynak dosya adı
}

// String returns a string representation of the token.
// String, tokenin string temsilini döndürür.
func (t Token) String() string {
	if t.Value != "" {
		return t.Value
	}
	return tokenNames[t.Type]
}

// tokenNames maps token types to names.
// tokenNames, token türlerini isimlere eşler.
var tokenNames = map[TokenType]string{
	TokEOF:       "EOF",
	TokError:     "ERROR",
	TokNewline:   "NEWLINE",
	TokInteger:   "INTEGER",
	TokFloat:     "FLOAT",
	TokString:    "STRING",
	TokRawString: "RAWSTRING",
	TokRegex:     "REGEX",
	TokHeredoc:   "HEREDOC",
	TokIdent:     "IDENT",
	TokScalar:    "SCALAR",
	TokArray:     "ARRAY",
	TokHash:      "HASH",
	TokCode:      "CODE",
	TokGlob:      "GLOB",
	TokPlus:      "+",
	TokMinus:     "-",
	TokStar:      "*",
	TokSlash:     "/",
	TokAssign:    "=",
	TokSemi:      ";",
	TokLParen:    "(",
	TokRParen:    ")",
	TokLBrace:    "{",
	TokRBrace:    "}",
	TokLBracket:  "[",
	TokRBracket:  "]",
	TokIf:        "if",
	TokElse:      "else",
	TokWhile:     "while",
	TokFor:       "for",
	TokForeach:   "foreach",
	TokMy:        "my",
	TokSub:       "sub",
	TokPackage:   "package",
	TokUse:       "use",
	TokReturn:    "return",
}

// keywords maps keyword strings to token types.
// keywords, anahtar kelime stringlerini token türlerine eşler.
var keywords = map[string]TokenType{
	// Control flow
	"if":      TokIf,
	"unless":  TokUnless,
	"else":    TokElse,
	"elsif":   TokElsif,
	"while":   TokWhile,
	"until":   TokUntil,
	"for":     TokFor,
	"foreach": TokForeach,
	"do":      TokDo,
	"last":    TokLast,
	"next":    TokNext,
	"redo":    TokRedo,
	"return":  TokReturn,
	"goto":    TokGoto,
	"given":   TokGiven,
	"when":    TokWhen,
	"default": TokDefault,

	// Declarations
	"my":        TokMy,
	"our":       TokOur,
	"local":     TokLocal,
	"state":     TokState,
	"sub":       TokSub,
	"package":   TokPackage,
	"use":       TokUse,
	"no":        TokNo,
	"require":   TokRequire,
	"BEGIN":     TokBEGIN,
	"END":       TokEND,
	"CHECK":     TokCHECK,
	"INIT":      TokINIT,
	"UNITCHECK": TokUNITCHECK,

	// String comparison operators
	"eq":  TokStrEq,
	"ne":  TokStrNe,
	"lt":  TokStrLt,
	"le":  TokStrLe,
	"gt":  TokStrGt,
	"ge":  TokStrGe,
	"cmp": TokCmp,

	// Logical operators (word form)
	"and": TokAndWord,
	"or":  TokOrWord,
	"not": TokNotWord,

	// String repeat
	"x": TokX,

	// Misc keywords
	"qw":        TokQw,
	"eval":      TokEval,
	"die":       TokDie,
	"warn":      TokWarn,
	"print":     TokPrint,
	"say":       TokSay,
	"open":      TokOpen,
	"close":     TokClose,
	"read":      TokRead,
	"write":     TokWrite,
	"defined":   TokDefined,
	"undef":     TokUndef,
	"ref":       TokRef,
	"bless":     TokBless,
	"tie":       TokTie,
	"untie":     TokUntie,
	"tied":      TokTied,
	"wantarray": TokWantarray,
	"caller":    TokCaller,
	"scalar":    TokScalarKw,

	// Array/Hash functions
	"shift":   TokShift,
	"unshift": TokUnshift,
	"pop":     TokPop,
	"push":    TokPush,
	"splice":  TokSplice,
	"keys":    TokKeys,
	"values":  TokValues,
	"each":    TokEach,
	"exists":  TokExists,
	"delete":  TokDelete,
	"sort":    TokSort,
	"reverse": TokReverse,
	"map":     TokMap,
	"grep":    TokGrep,
	"join":    TokJoin,
	"split":   TokSplit,

	// String functions
	"length":  TokLength,
	"substr":  TokSubstr,
	"index":   TokIndex,
	"rindex":  TokRindex,
	"lc":      TokLc,
	"uc":      TokUc,
	"lcfirst": TokLcfirst,
	"ucfirst": TokUcfirst,
	"chomp":   TokChomp,
	"chop":    TokChop,
	"chr":     TokChr,
	"ord":     TokOrd,
	"hex":     TokHex,
	"oct":     TokOct,
	"pack":    TokPack,
	"unpack":  TokUnpack,
	"sprintf": TokSprintf,

	// Numeric functions
	"abs":   TokAbs,
	"int":   TokInt,
	"sqrt":  TokSqrt,
	"rand":  TokRand,
	"srand": TokSrand,
	"sin":   TokSin,
	"cos":   TokCos,
	"atan2": TokAtan2,
	"exp":   TokExp,
	"log":   TokLog,

	// Misc functions
	"localtime": TokLocaltime,
	"gmtime":    TokGmtime,
	"time":      TokTime,
	"sleep":     TokSleep,
	"exit":      TokExit,
	"system":    TokSystem,
	"exec":      TokExec,
	"fork":      TokFork,
	"wait":      TokWait,
	"kill":      TokKill,
}

// LookupKeyword returns the token type for an identifier.
// If not a keyword, returns TokIdent.
//
// LookupKeyword, bir tanımlayıcı için token türünü döndürür.
// Anahtar kelime değilse, TokIdent döndürür.
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokIdent
}
