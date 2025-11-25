package lexer

import (
	"testing"
)

// ============================================================
// Basic Token Tests
// Temel Token Testleri
// ============================================================

// TestSingleCharTokens tests single character tokens.
// TestSingleCharTokens, tek karakterli tokenleri test eder.
func TestSingleCharTokens(t *testing.T) {
	input := `( ) [ ] { } ; , ? ~ \`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokLParen, "("},
		{TokRParen, ")"},
		{TokLBracket, "["},
		{TokRBracket, "]"},
		{TokLBrace, "{"},
		{TokRBrace, "}"},
		{TokSemi, ";"},
		{TokComma, ","},
		{TokQuestion, "?"},
		{TokBitNot, "~"},
		{TokBackslash, "\\"},
		{TokEOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
		if tt.expectedValue != "" && tok.Value != tt.expectedValue {
			t.Errorf("tests[%d] - wrong value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

// TestNewlines tests newline handling.
// TestNewlines, satır sonu işlemeyi test eder.
func TestNewlines(t *testing.T) {
	input := "a\nb\nc"

	l := New(input)

	tok := l.NextToken() // a
	if tok.Line != 1 {
		t.Errorf("First token should be on line 1, got %d", tok.Line)
	}

	l.NextToken()       // newline
	tok = l.NextToken() // b
	if tok.Line != 2 {
		t.Errorf("Second token should be on line 2, got %d", tok.Line)
	}

	l.NextToken()       // newline
	tok = l.NextToken() // c
	if tok.Line != 3 {
		t.Errorf("Third token should be on line 3, got %d", tok.Line)
	}
}

// ============================================================
// Arithmetic Operator Tests
// Aritmetik Operatör Testleri
// ============================================================

// TestArithmeticOperators tests arithmetic operators.
// TestArithmeticOperators, aritmetik operatörleri test eder.
func TestArithmeticOperators(t *testing.T) {
	input := `+ - * / % **`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokPlus, "+"},
		{TokMinus, "-"},
		{TokStar, "*"},
		{TokSlash, "/"},
		{TokPercent, "%"},
		{TokStarStar, "**"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedValue {
			t.Errorf("tests[%d] - wrong value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

// TestIncrementDecrement tests ++ and --.
// TestIncrementDecrement, ++ ve -- test eder.
func TestIncrementDecrement(t *testing.T) {
	input := `++ --`

	l := New(input)

	tok := l.NextToken()
	if tok.Type != TokIncr || tok.Value != "++" {
		t.Errorf("Expected ++, got %v %q", tok.Type, tok.Value)
	}

	tok = l.NextToken()
	if tok.Type != TokDecr || tok.Value != "--" {
		t.Errorf("Expected --, got %v %q", tok.Type, tok.Value)
	}
}

// ============================================================
// Comparison Operator Tests
// Karşılaştırma Operatörü Testleri
// ============================================================

// TestNumericComparison tests numeric comparison operators.
// TestNumericComparison, sayısal karşılaştırma operatörlerini test eder.
func TestNumericComparison(t *testing.T) {
	input := `== != < <= > >= <=>`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokEq, "=="},
		{TokNe, "!="},
		{TokLt, "<"},
		{TokLe, "<="},
		{TokGt, ">"},
		{TokGe, ">="},
		{TokSpaceship, "<=>"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedValue {
			t.Errorf("tests[%d] - wrong value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

// TestStringComparison tests string comparison operators.
// TestStringComparison, string karşılaştırma operatörlerini test eder.
func TestStringComparison(t *testing.T) {
	input := `eq ne lt le gt ge cmp`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokStrEq, "eq"},
		{TokStrNe, "ne"},
		{TokStrLt, "lt"},
		{TokStrLe, "le"},
		{TokStrGt, "gt"},
		{TokStrGe, "ge"},
		{TokCmp, "cmp"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v (%q)",
				i, tt.expectedType, tok.Type, tok.Value)
		}
	}
}

// ============================================================
// Logical Operator Tests
// Mantıksal Operatör Testleri
// ============================================================

// TestLogicalOperators tests logical operators.
// TestLogicalOperators, mantıksal operatörleri test eder.
func TestLogicalOperators(t *testing.T) {
	input := `&& || ! // and or not`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokAnd, "&&"},
		{TokOr, "||"},
		{TokNot, "!"},
		{TokDefinedOr, "//"},
		{TokAndWord, "and"},
		{TokOrWord, "or"},
		{TokNotWord, "not"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
	}
}

// ============================================================
// Bitwise Operator Tests
// Bitsel Operatör Testleri
// ============================================================

// TestBitwiseOperators tests bitwise operators.
// TestBitwiseOperators, bitsel operatörleri test eder.
func TestBitwiseOperators(t *testing.T) {
	input := `<< >> ^ | ~`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokLeftShift, "<<"},
		{TokRightShift, ">>"},
		{TokBitXor, "^"},
		{TokBitOr, "|"},
		{TokBitNot, "~"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
	}
}

// ============================================================
// Assignment Operator Tests
// Atama Operatörü Testleri
// ============================================================

// TestAssignmentOperators tests assignment operators.
// TestAssignmentOperators, atama operatörlerini test eder.
func TestAssignmentOperators(t *testing.T) {
	input := `= += -= *= /= %= **= .= &&= ||= //= &= |= ^= <<= >>=`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokAssign, "="},
		{TokPlusEq, "+="},
		{TokMinusEq, "-="},
		{TokStarEq, "*="},
		{TokSlashEq, "/="},
		{TokPercentEq, "%="},
		{TokStarStarEq, "**="},
		{TokDotEq, ".="},
		{TokAndEq, "&&="},
		{TokOrEq, "||="},
		{TokDefinedOrEq, "//="},
		{TokBitAndEq, "&="},
		{TokBitOrEq, "|="},
		{TokBitXorEq, "^="},
		{TokLeftShiftEq, "<<="},
		{TokRightShiftEq, ">>="},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedValue {
			t.Errorf("tests[%d] - wrong value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

// ============================================================
// Other Operator Tests
// Diğer Operatör Testleri
// ============================================================

// TestMiscOperators tests miscellaneous operators.
// TestMiscOperators, çeşitli operatörleri test eder.
func TestMiscOperators(t *testing.T) {
	input := `. .. ... -> => =~ !~ :: :`

	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{TokDot, "."},
		{TokRange, ".."},
		{TokRange3, "..."},
		{TokArrow, "->"},
		{TokFatArrow, "=>"},
		{TokMatch, "=~"},
		{TokNotMatch, "!~"},
		{TokDoubleColon, "::"},
		{TokColon, ":"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - wrong type. expected=%v, got=%v",
				i, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedValue {
			t.Errorf("tests[%d] - wrong value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

// ============================================================
// Number Tests
// Sayı Testleri
// ============================================================

// TestIntegers tests integer literals.
// TestIntegers, tamsayı literallerini test eder.
func TestIntegers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"0", "0"},
		{"123456", "123456"},
		{"1_000_000", "1000000"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokInteger {
			t.Errorf("input %q - wrong type. expected=TokInteger, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestHexNumbers tests hexadecimal literals.
// TestHexNumbers, onaltılık literalleri test eder.
func TestHexNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0x2A", "0x2A"},
		{"0xFF", "0xFF"},
		{"0xDEAD_BEEF", "0xDEADBEEF"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokInteger {
			t.Errorf("input %q - wrong type. expected=TokInteger, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestBinaryNumbers tests binary literals.
// TestBinaryNumbers, ikili literalleri test eder.
func TestBinaryNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0b101010", "0b101010"},
		{"0b1111_0000", "0b11110000"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokInteger {
			t.Errorf("input %q - wrong type. expected=TokInteger, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestOctalNumbers tests octal literals.
// TestOctalNumbers, sekizlik literalleri test eder.
func TestOctalNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0o52", "0o52"},
		{"0o755", "0o755"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokInteger {
			t.Errorf("input %q - wrong type. expected=TokInteger, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestFloats tests floating point literals.
// TestFloats, kayan nokta literallerini test eder.
func TestFloats(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"1.0", "1.0"},
		{"6.02e23", "6.02e23"},
		{"1.5E-10", "1.5E-10"},
		{"1e+5", "1e+5"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokFloat {
			t.Errorf("input %q - wrong type. expected=TokFloat, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// ============================================================
// String Tests
// String Testleri
// ============================================================

// TestDoubleQuotedStrings tests double-quoted strings.
// TestDoubleQuotedStrings, çift tırnaklı stringleri test eder.
func TestDoubleQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"line1\nline2"`, "line1\nline2"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"here"`, `quote"here`},
		{`"back\\slash"`, `back\slash`},
		{`"dollar\$var"`, `dollar$var`},
		{`"at\@arr"`, `at@arr`},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokString {
			t.Errorf("input %q - wrong type. expected=TokString, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestSingleQuotedStrings tests single-quoted strings.
// TestSingleQuotedStrings, tek tırnaklı stringleri test eder.
func TestSingleQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'hello'`, "hello"},
		{`'no\ninterpolation'`, `no\ninterpolation`},
		{`'it\'s'`, `it's`},
		{`'back\\slash'`, `back\slash`},
		{`''`, ""},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokRawString {
			t.Errorf("input %q - wrong type. expected=TokRawString, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expected {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expected, tok.Value)
		}
	}
}

// TestBacktickStrings tests backtick strings.
// TestBacktickStrings, backtick stringleri test eder.
func TestBacktickStrings(t *testing.T) {
	input := "`ls -la`"
	l := New(input)
	tok := l.NextToken()

	if tok.Type != TokString {
		t.Errorf("wrong type. expected=TokString, got=%v", tok.Type)
	}
	if tok.Value != "ls -la" {
		t.Errorf("wrong value. expected=%q, got=%q", "ls -la", tok.Value)
	}
}

// ============================================================
// Variable Tests
// Değişken Testleri
// ============================================================

// TestScalarVariables tests scalar variables.
// TestScalarVariables, skaler değişkenleri test eder.
func TestScalarVariables(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedVal  string
	}{
		{"$x", TokScalar, "$x"},
		{"$foo", TokScalar, "$foo"},
		{"$foo_bar", TokScalar, "$foo_bar"},
		{"$Foo123", TokScalar, "$Foo123"},
		{"$_", TokSpecialVar, "$_"},
		{"$1", TokSpecialVar, "$1"},
		{"$123", TokSpecialVar, "$123"},
		{"${name}", TokScalar, "$name"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q - wrong type. expected=%v, got=%v",
				tt.input, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// TestArrayVariables tests array variables.
// TestArrayVariables, dizi değişkenlerini test eder.
func TestArrayVariables(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedVal  string
	}{
		{"@arr", TokArray, "@arr"},
		{"@foo_bar", TokArray, "@foo_bar"},
		{"@_", TokSpecialVar, "@_"},
		{"@{name}", TokArray, "@name"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q - wrong type. expected=%v, got=%v",
				tt.input, tt.expectedType, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// TestHashVariables tests hash variables.
// TestHashVariables, hash değişkenlerini test eder.
func TestHashVariables(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal string
	}{
		{"%hash", "%hash"},
		{"%foo_bar", "%foo_bar"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokHash {
			t.Errorf("input %q - wrong type. expected=TokHash, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// TestCodeVariables tests code references.
// TestCodeVariables, kod referanslarını test eder.
func TestCodeVariables(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal string
	}{
		{"&sub", "&sub"},
		{"&foo_bar", "&foo_bar"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokCode {
			t.Errorf("input %q - wrong type. expected=TokCode, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// TestArrayLength tests $#array.
// TestArrayLength, $#array test eder.
func TestArrayLength(t *testing.T) {
	input := "$#arr"
	l := New(input)
	tok := l.NextToken()

	if tok.Type != TokArrayLen {
		t.Errorf("wrong type. expected=TokArrayLen, got=%v", tok.Type)
	}
	if tok.Value != "$#arr" {
		t.Errorf("wrong value. expected=%q, got=%q", "$#arr", tok.Value)
	}
}

// TestSpecialVariables tests special variables.
// TestSpecialVariables, özel değişkenleri test eder.
func TestSpecialVariables(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal string
	}{
		{"$_", "$_"},
		{"$@", "$@"},
		{"$!", "$!"},
		{"$$", "$$"},
		{"$?", "$?"},
		{`$"`, `$"`},
		{"$/", "$/"},
		{`$\`, `$\`},
		{"$&", "$&"},
		{"$`", "$`"},
		{"$'", "$'"},
		{"$+", "$+"},
		{"$.", "$."},
		{"$|", "$|"},
		{"$-", "$-"},
		{"$^", "$^"},
		{"$~", "$~"},
		{"$=", "$="},
		{"$%", "$%"},
		{"$:", "$:"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokSpecialVar {
			t.Errorf("input %q - wrong type. expected=TokSpecialVar, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// TestPackageVariables tests package-qualified variables.
// TestPackageVariables, paket-nitelikli değişkenleri test eder.
func TestPackageVariables(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal string
	}{
		{"$Foo::bar", "$Foo::bar"},
		{"@Foo::Bar::arr", "@Foo::Bar::arr"},
		{"%A::B::C::hash", "%A::B::C::hash"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// ============================================================
// Keyword Tests
// Anahtar Kelime Testleri
// ============================================================

// TestKeywords tests keyword recognition.
// TestKeywords, anahtar kelime tanımayı test eder.
func TestKeywords(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
	}{
		{"if", TokIf},
		{"unless", TokUnless},
		{"else", TokElse},
		{"elsif", TokElsif},
		{"while", TokWhile},
		{"until", TokUntil},
		{"for", TokFor},
		{"foreach", TokForeach},
		{"do", TokDo},
		{"last", TokLast},
		{"next", TokNext},
		{"redo", TokRedo},
		{"return", TokReturn},
		{"goto", TokGoto},
		{"my", TokMy},
		{"our", TokOur},
		{"local", TokLocal},
		{"state", TokState},
		{"sub", TokSub},
		{"package", TokPackage},
		{"use", TokUse},
		{"no", TokNo},
		{"require", TokRequire},
		{"BEGIN", TokBEGIN},
		{"END", TokEND},
		{"eval", TokEval},
		{"die", TokDie},
		{"warn", TokWarn},
		{"print", TokPrint},
		{"say", TokSay},
		{"defined", TokDefined},
		{"undef", TokUndef},
		{"ref", TokRef},
		{"bless", TokBless},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q - wrong type. expected=%v, got=%v",
				tt.input, tt.expectedType, tok.Type)
		}
	}
}

// TestIdentifiers tests identifier recognition.
// TestIdentifiers, tanımlayıcı tanımayı test eder.
func TestIdentifiers(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal string
	}{
		{"foo", "foo"},
		{"foo_bar", "foo_bar"},
		{"FooBar", "FooBar"},
		{"_private", "_private"},
		{"foo123", "foo123"},
		{"Foo::Bar", "Foo::Bar"},
		{"A::B::C::D", "A::B::C::D"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TokIdent {
			t.Errorf("input %q - wrong type. expected=TokIdent, got=%v",
				tt.input, tok.Type)
		}
		if tok.Value != tt.expectedVal {
			t.Errorf("input %q - wrong value. expected=%q, got=%q",
				tt.input, tt.expectedVal, tok.Value)
		}
	}
}

// ============================================================
// Regex Tests
// Regex Testleri
// ============================================================

// TestRegex tests regex literals.
// TestRegex, regex literallerini test eder.
func TestRegex(t *testing.T) {
	// After =~ we expect regex
	input := `=~ /hello/`

	l := New(input)
	l.NextToken() // =~

	tok := l.NextToken()
	if tok.Type != TokRegex {
		t.Errorf("wrong type. expected=TokRegex, got=%v", tok.Type)
	}
	if tok.Value != "hello" {
		t.Errorf("wrong value. expected=%q, got=%q", "hello", tok.Value)
	}
}

// TestRegexWithModifiers tests regex with modifiers.
// TestRegexWithModifiers, değiştiricili regex test eder.
func TestRegexWithModifiers(t *testing.T) {
	input := `=~ /pattern/gimsxo`

	l := New(input)
	l.NextToken() // =~

	tok := l.NextToken()
	if tok.Type != TokRegex {
		t.Errorf("wrong type. expected=TokRegex, got=%v", tok.Type)
	}
	if tok.Value != "pattern/gimsxo" {
		t.Errorf("wrong value. expected=%q, got=%q", "pattern/gimsxo", tok.Value)
	}
}

// TestRegexWithEscapes tests regex with escaped delimiters.
// TestRegexWithEscapes, kaçışlı sınırlayıcılı regex test eder.
func TestRegexWithEscapes(t *testing.T) {
	input := `=~ /hello\/world/`

	l := New(input)
	l.NextToken() // =~

	tok := l.NextToken()
	if tok.Type != TokRegex {
		t.Errorf("wrong type. expected=TokRegex, got=%v", tok.Type)
	}
	if tok.Value != `hello\/world` {
		t.Errorf("wrong value. expected=%q, got=%q", `hello\/world`, tok.Value)
	}
}

// ============================================================
// Comment Tests
// Yorum Testleri
// ============================================================

// TestComments tests comment skipping.
// TestComments, yorum atlamayı test eder.
func TestComments(t *testing.T) {
	input := `# This is a comment
$x # inline comment
$y`

	l := New(input)

	// Comment is skipped, newline after comment returned
	// Yorum atlanır, yorumdan sonraki newline döndürülür
	tok := l.NextToken() // newline after first comment
	if tok.Type != TokNewline {
		t.Errorf("Expected newline after comment, got %v %q", tok.Type, tok.Value)
	}

	tok = l.NextToken() // $x
	if tok.Type != TokScalar || tok.Value != "$x" {
		t.Errorf("Expected $x, got %v %q", tok.Type, tok.Value)
	}

	tok = l.NextToken() // newline (inline comment skipped)
	if tok.Type != TokNewline {
		t.Errorf("Expected newline, got %v %q", tok.Type, tok.Value)
	}

	tok = l.NextToken() // $y
	if tok.Type != TokScalar || tok.Value != "$y" {
		t.Errorf("Expected $y, got %v %q", tok.Type, tok.Value)
	}
}

// ============================================================
// Complex Expression Tests
// Karmaşık İfade Testleri
// ============================================================

// TestSimpleExpression tests a simple expression.
// TestSimpleExpression, basit bir ifadeyi test eder.
func TestSimpleExpression(t *testing.T) {
	input := `$x + $y * 2`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokScalar, "$x"},
		{TokPlus, "+"},
		{TokScalar, "$y"},
		{TokStar, "*"},
		{TokInteger, "2"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestVariableDeclaration tests my $x = 10;
// TestVariableDeclaration, my $x = 10; test eder.
func TestVariableDeclaration(t *testing.T) {
	input := `my $x = 10;`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokMy, "my"},
		{TokScalar, "$x"},
		{TokAssign, "="},
		{TokInteger, "10"},
		{TokSemi, ";"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestSubroutineDeclaration tests sub foo { }
// TestSubroutineDeclaration, sub foo { } test eder.
func TestSubroutineDeclaration(t *testing.T) {
	input := `sub foo { return $x; }`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokSub, "sub"},
		{TokIdent, "foo"},
		{TokLBrace, "{"},
		{TokReturn, "return"},
		{TokScalar, "$x"},
		{TokSemi, ";"},
		{TokRBrace, "}"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestHashAccess tests hash access $hash{key}.
// TestHashAccess, hash erişimini test eder $hash{key}.
func TestHashAccess(t *testing.T) {
	input := `$hash{key}`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokScalar, "$hash"},
		{TokLBrace, "{"},
		{TokIdent, "key"},
		{TokRBrace, "}"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestArrayAccess tests array access $arr[0].
// TestArrayAccess, dizi erişimini test eder $arr[0].
func TestArrayAccess(t *testing.T) {
	input := `$arr[0]`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokScalar, "$arr"},
		{TokLBracket, "["},
		{TokInteger, "0"},
		{TokRBracket, "]"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestMethodCall tests method call $obj->method().
// TestMethodCall, method çağrısını test eder $obj->method().
func TestMethodCall(t *testing.T) {
	input := `$obj->method()`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokScalar, "$obj"},
		{TokArrow, "->"},
		{TokIdent, "method"},
		{TokLParen, "("},
		{TokRParen, ")"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// TestTernaryOperator tests ternary $x ? $y : $z.
// TestTernaryOperator, üçlü operatörü test eder $x ? $y : $z.
func TestTernaryOperator(t *testing.T) {
	input := `$x ? $y : $z`

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokScalar, "$x"},
		{TokQuestion, "?"},
		{TokScalar, "$y"},
		{TokColon, ":"},
		{TokScalar, "$z"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] - wrong type. expected=%v, got=%v",
				i, exp.typ, tok.Type)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] - wrong value. expected=%q, got=%q",
				i, exp.val, tok.Value)
		}
	}
}

// ============================================================
// Line/Column Tracking Tests
// Satır/Sütun İzleme Testleri
// ============================================================

// TestLineColumn tests line and column tracking.
// TestLineColumn, satır ve sütun izlemeyi test eder.
func TestLineColumn(t *testing.T) {
	input := "a b\nc d"

	l := New(input)

	tok := l.NextToken() // a
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("'a' should be at 1:1, got %d:%d", tok.Line, tok.Column)
	}

	tok = l.NextToken() // b
	if tok.Line != 1 || tok.Column != 3 {
		t.Errorf("'b' should be at 1:3, got %d:%d", tok.Line, tok.Column)
	}

	tok = l.NextToken() // newline
	tok = l.NextToken() // c
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("'c' should be at 2:1, got %d:%d", tok.Line, tok.Column)
	}
}

// TestFilename tests filename in tokens.
// TestFilename, tokenlerdeki dosya adını test eder.
func TestFilename(t *testing.T) {
	l := NewFile("$x", "test.pl")
	tok := l.NextToken()

	if tok.File != "test.pl" {
		t.Errorf("File should be 'test.pl', got %q", tok.File)
	}
}

// ============================================================
// Edge Cases
// Sınır Durumları
// ============================================================

// TestEmptyInput tests empty input.
// TestEmptyInput, boş girdiyi test eder.
func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()

	if tok.Type != TokEOF {
		t.Errorf("Empty input should return EOF, got %v", tok.Type)
	}
}

// TestWhitespaceOnly tests whitespace-only input.
// TestWhitespaceOnly, sadece boşluk içeren girdiyi test eder.
func TestWhitespaceOnly(t *testing.T) {
	l := New("   \t\t   ")
	tok := l.NextToken()

	if tok.Type != TokEOF {
		t.Errorf("Whitespace-only should return EOF, got %v", tok.Type)
	}
}

// TestUnexpectedCharacter tests unexpected character.
// TestUnexpectedCharacter, beklenmeyen karakteri test eder.
func TestUnexpectedCharacter(t *testing.T) {
	l := New("@") // @ without identifier is error
	tok := l.NextToken()

	if tok.Type != TokError {
		t.Errorf("Expected TokError, got %v", tok.Type)
	}
}

// TestTokenString tests Token.String() method.
// TestTokenString, Token.String() metodunu test eder.
func TestTokenString(t *testing.T) {
	tok := Token{Type: TokPlus, Value: "+"}
	if tok.String() != "+" {
		t.Errorf("String() should return '+', got %q", tok.String())
	}

	tok = Token{Type: TokEOF}
	if tok.String() != "EOF" {
		t.Errorf("String() for EOF should return 'EOF', got %q", tok.String())
	}
}

// TestLookupKeyword tests LookupKeyword function.
// TestLookupKeyword, LookupKeyword fonksiyonunu test eder.
func TestLookupKeyword(t *testing.T) {
	if LookupKeyword("if") != TokIf {
		t.Error("'if' should be TokIf")
	}
	if LookupKeyword("foo") != TokIdent {
		t.Error("'foo' should be TokIdent")
	}
}
