package sv

import (
	"math"
	"strings"
	"unicode/utf8"
)

// ============================================================
// Arithmetic Operations
// ============================================================

// Add performs $a + $b
func Add(a, b *SV) *SV {
	// Check if either operand wants float math
	if needsFloatMath(a) || needsFloatMath(b) {
		return NewFloat(a.AsFloat() + b.AsFloat())
	}
	return NewInt(a.AsInt() + b.AsInt())
}

// Sub performs $a - $b
func Sub(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		return NewFloat(a.AsFloat() - b.AsFloat())
	}
	return NewInt(a.AsInt() - b.AsInt())
}

// Mul performs $a * $b
func Mul(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		return NewFloat(a.AsFloat() * b.AsFloat())
	}
	return NewInt(a.AsInt() * b.AsInt())
}

// Div performs $a / $b (always returns float like Perl)
func Div(a, b *SV) *SV {
	bv := b.AsFloat()
	if bv == 0 {
		// Perl: division by zero is fatal error
		panic("Illegal division by zero")
	}
	return NewFloat(a.AsFloat() / bv)
}

// IntDiv performs int($a / $b) - integer division
func IntDiv(a, b *SV) *SV {
	bv := b.AsInt()
	if bv == 0 {
		panic("Illegal division by zero")
	}
	return NewInt(a.AsInt() / bv)
}

// Mod performs $a % $b
func Mod(a, b *SV) *SV {
	bv := b.AsInt()
	if bv == 0 {
		panic("Illegal modulus zero")
	}
	return NewInt(a.AsInt() % bv)
}

// Pow performs $a ** $b
func Pow(a, b *SV) *SV {
	result := math.Pow(a.AsFloat(), b.AsFloat())
	// Return int if result is whole and fits
	if result == math.Trunc(result) && result >= math.MinInt64 && result <= math.MaxInt64 {
		return NewInt(int64(result))
	}
	return NewFloat(result)
}

// Neg performs -$a (negation)
func Neg(a *SV) *SV {
	if a.typ == TypeFloat || a.flags&FlagNOK != 0 {
		return NewFloat(-a.AsFloat())
	}
	return NewInt(-a.AsInt())
}

// needsFloatMath checks if value requires float arithmetic
func needsFloatMath(sv *SV) bool {
	if sv == nil {
		return false
	}
	if sv.typ == TypeFloat {
		return true
	}
	if sv.typ == TypeString {
		s := sv.pv
		for _, c := range s {
			if c == '.' || c == 'e' || c == 'E' {
				return true
			}
		}
	}
	return false
}

// ============================================================
// Increment/Decrement (Perl's magical ++ and --)
// ============================================================

// Inc performs ++$a (pre-increment) - modifies in place
func Inc(a *SV) *SV {
	a.checkWritable()

	// Perl's magical string increment: "aa" -> "ab", "az" -> "ba", "z9" -> "aa0"
	if a.typ == TypeString && !hasNumericPrefix(a.pv) && isIncrementableString(a.pv) {
		a.pv = incrementString(a.pv)
		a.flags = FlagPOK | FlagUTF8
		return a
	}

	if a.typ == TypeFloat || a.flags&FlagNOK != 0 {
		a.nv = a.AsFloat() + 1
		a.flags = FlagNOK
		a.typ = TypeFloat
	} else {
		a.iv = a.AsInt() + 1
		a.flags = FlagIOK
		a.typ = TypeInt
	}
	return a
}

// Dec performs --$a (pre-decrement) - modifies in place
// Note: -- does NOT do magical string decrement in Perl
func Dec(a *SV) *SV {
	a.checkWritable()

	if a.typ == TypeFloat || a.flags&FlagNOK != 0 {
		a.nv = a.AsFloat() - 1
		a.flags = FlagNOK
		a.typ = TypeFloat
	} else {
		a.iv = a.AsInt() - 1
		a.flags = FlagIOK
		a.typ = TypeInt
	}
	return a
}

// hasNumericPrefix checks if string starts with a number
func hasNumericPrefix(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	c := s[0]
	return (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.'
}

// isIncrementableString checks if string can be magically incremented
func isIncrementableString(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// incrementString performs Perl's magical string increment
func incrementString(s string) string {
	if s == "" {
		return "1"
	}

	runes := []rune(s)
	carry := true

	for i := len(runes) - 1; i >= 0 && carry; i-- {
		c := runes[i]

		switch {
		case c >= 'a' && c < 'z':
			runes[i] = c + 1
			carry = false
		case c == 'z':
			runes[i] = 'a'
		case c >= 'A' && c < 'Z':
			runes[i] = c + 1
			carry = false
		case c == 'Z':
			runes[i] = 'A'
		case c >= '0' && c < '9':
			runes[i] = c + 1
			carry = false
		case c == '9':
			runes[i] = '0'
		default:
			carry = false
		}
	}

	if carry {
		// Need to prepend
		first := runes[0]
		var prefix rune
		switch {
		case first >= 'a' && first <= 'z':
			prefix = 'a'
		case first >= 'A' && first <= 'Z':
			prefix = 'A'
		default:
			prefix = '1'
		}
		runes = append([]rune{prefix}, runes...)
	}

	return string(runes)
}

// ============================================================
// Bitwise Operations
// ============================================================

// BitAnd performs $a & $b
func BitAnd(a, b *SV) *SV {
	return NewInt(a.AsInt() & b.AsInt())
}

// BitOr performs $a | $b
func BitOr(a, b *SV) *SV {
	return NewInt(a.AsInt() | b.AsInt())
}

// BitXor performs $a ^ $b
func BitXor(a, b *SV) *SV {
	return NewInt(a.AsInt() ^ b.AsInt())
}

// BitNot performs ~$a
func BitNot(a *SV) *SV {
	return NewInt(^a.AsInt())
}

// LeftShift performs $a << $b
func LeftShift(a, b *SV) *SV {
	shift := b.AsInt()
	if shift < 0 {
		panic("Negative shift count")
	}
	return NewInt(a.AsInt() << uint(shift))
}

// RightShift performs $a >> $b
func RightShift(a, b *SV) *SV {
	shift := b.AsInt()
	if shift < 0 {
		panic("Negative shift count")
	}
	return NewInt(a.AsInt() >> uint(shift))
}

// ============================================================
// String Operations
// ============================================================

// Concat performs $a . $b (string concatenation)
func Concat(a, b *SV) *SV {
	return NewString(a.AsString() + b.AsString())
}

// Repeat performs $a x $b (string repetition)
func Repeat(a, b *SV) *SV {
	s := a.AsString()
	n := b.AsInt()
	if n <= 0 {
		return NewString("")
	}
	return NewString(strings.Repeat(s, int(n)))
}

// Length returns length($a) - character count for strings
func Length(a *SV) *SV {
	if a == nil || a.typ == TypeUndef {
		return NewUndef()
	}
	s := a.AsString()
	// Perl's length() returns character count, not byte count
	return NewInt(int64(utf8.RuneCountInString(s)))
}

// Substr implements substr($str, $offset, $len)
func Substr(str, offset, length *SV) *SV {
	s := str.AsString()
	runes := []rune(s)
	runeLen := len(runes)

	off := int(offset.AsInt())
	// Negative offset counts from end
	if off < 0 {
		off = runeLen + off
	}
	if off < 0 {
		off = 0
	}
	if off > runeLen {
		return NewString("")
	}

	var ln int
	if length == nil || length.IsUndef() {
		ln = runeLen - off
	} else {
		ln = int(length.AsInt())
		if ln < 0 {
			// Negative length means leave that many chars at end
			ln = runeLen - off + ln
		}
	}

	if ln <= 0 {
		return NewString("")
	}
	if off+ln > runeLen {
		ln = runeLen - off
	}

	return NewString(string(runes[off : off+ln]))
}

// Index implements index($str, $substr, $pos)
func Index(str, substr, pos *SV) *SV {
	s := str.AsString()
	sub := substr.AsString()

	startPos := 0
	if pos != nil && !pos.IsUndef() {
		startPos = int(pos.AsInt())
	}

	// Work with runes for proper Unicode handling
	runes := []rune(s)
	subRunes := []rune(sub)

	if startPos < 0 {
		startPos = 0
	}
	if startPos > len(runes) {
		return NewInt(-1)
	}

	// Simple search
	for i := startPos; i <= len(runes)-len(subRunes); i++ {
		match := true
		for j := 0; j < len(subRunes); j++ {
			if runes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return NewInt(int64(i))
		}
	}

	return NewInt(-1)
}

// Rindex implements rindex($str, $substr, $pos)
func Rindex(str, substr, pos *SV) *SV {
	s := str.AsString()
	sub := substr.AsString()

	runes := []rune(s)
	subRunes := []rune(sub)

	endPos := len(runes) - len(subRunes)
	if pos != nil && !pos.IsUndef() {
		p := int(pos.AsInt())
		if p < endPos {
			endPos = p
		}
	}

	if endPos < 0 {
		return NewInt(-1)
	}

	// Search backwards
	for i := endPos; i >= 0; i-- {
		match := true
		for j := 0; j < len(subRunes); j++ {
			if runes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return NewInt(int64(i))
		}
	}

	return NewInt(-1)
}

// Uc implements uc($str) - uppercase
func Uc(a *SV) *SV {
	return NewString(strings.ToUpper(a.AsString()))
}

// Lc implements lc($str) - lowercase
func Lc(a *SV) *SV {
	return NewString(strings.ToLower(a.AsString()))
}

// Ucfirst implements ucfirst($str)
func Ucfirst(a *SV) *SV {
	s := a.AsString()
	if s == "" {
		return NewString("")
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return NewString(string(runes))
}

// Lcfirst implements lcfirst($str)
func Lcfirst(a *SV) *SV {
	s := a.AsString()
	if s == "" {
		return NewString("")
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return NewString(string(runes))
}

// Reverse implements reverse($str) for scalar context
func Reverse(a *SV) *SV {
	s := a.AsString()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return NewString(string(runes))
}

// ============================================================
// Numeric Comparison Operations
// ============================================================

// NumEq performs $a == $b
func NumEq(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		if a.AsFloat() == b.AsFloat() {
			return NewInt(1)
		}
	} else {
		if a.AsInt() == b.AsInt() {
			return NewInt(1)
		}
	}
	return NewString("") // False
}

// NumNe performs $a != $b
func NumNe(a, b *SV) *SV {
	return Not(NumEq(a, b))
}

// NumLt performs $a < $b
func NumLt(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		if a.AsFloat() < b.AsFloat() {
			return NewInt(1)
		}
	} else {
		if a.AsInt() < b.AsInt() {
			return NewInt(1)
		}
	}
	return NewString("")
}

// NumLe performs $a <= $b
func NumLe(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		if a.AsFloat() <= b.AsFloat() {
			return NewInt(1)
		}
	} else {
		if a.AsInt() <= b.AsInt() {
			return NewInt(1)
		}
	}
	return NewString("")
}

// NumGt performs $a > $b
func NumGt(a, b *SV) *SV {
	return NumLt(b, a)
}

// NumGe performs $a >= $b
func NumGe(a, b *SV) *SV {
	return NumLe(b, a)
}

// NumCmp performs $a <=> $b (spaceship operator)
func NumCmp(a, b *SV) *SV {
	if needsFloatMath(a) || needsFloatMath(b) {
		av, bv := a.AsFloat(), b.AsFloat()
		if math.IsNaN(av) || math.IsNaN(bv) {
			return NewUndef()
		}
		switch {
		case av < bv:
			return NewInt(-1)
		case av > bv:
			return NewInt(1)
		default:
			return NewInt(0)
		}
	}
	av, bv := a.AsInt(), b.AsInt()
	switch {
	case av < bv:
		return NewInt(-1)
	case av > bv:
		return NewInt(1)
	default:
		return NewInt(0)
	}
}

// ============================================================
// String Comparison Operations
// ============================================================

// StrEq performs $a eq $b
func StrEq(a, b *SV) *SV {
	if a.AsString() == b.AsString() {
		return NewInt(1)
	}
	return NewString("")
}

// StrNe performs $a ne $b
func StrNe(a, b *SV) *SV {
	return Not(StrEq(a, b))
}

// StrLt performs $a lt $b
func StrLt(a, b *SV) *SV {
	if a.AsString() < b.AsString() {
		return NewInt(1)
	}
	return NewString("")
}

// StrLe performs $a le $b
func StrLe(a, b *SV) *SV {
	if a.AsString() <= b.AsString() {
		return NewInt(1)
	}
	return NewString("")
}

// StrGt performs $a gt $b
func StrGt(a, b *SV) *SV {
	return StrLt(b, a)
}

// StrGe performs $a ge $b
func StrGe(a, b *SV) *SV {
	return StrLe(b, a)
}

// StrCmp performs $a cmp $b
func StrCmp(a, b *SV) *SV {
	cmp := strings.Compare(a.AsString(), b.AsString())
	return NewInt(int64(cmp))
}

// ============================================================
// Logical Operations
// ============================================================

// Not performs !$a (logical not)
func Not(a *SV) *SV {
	if a.AsBool() {
		return NewString("")
	}
	return NewInt(1)
}

// And performs $a && $b (returns last evaluated value)
func And(a, b *SV) *SV {
	if !a.AsBool() {
		return a
	}
	return b
}

// Or performs $a || $b (returns last evaluated value)
func Or(a, b *SV) *SV {
	if a.AsBool() {
		return a
	}
	return b
}

// DefinedOr performs $a // $b (returns $a if defined, else $b)
func DefinedOr(a, b *SV) *SV {
	if !a.IsUndef() {
		return a
	}
	return b
}

// Defined checks if value is defined
func Defined(a *SV) *SV {
	if a == nil || a.typ == TypeUndef {
		return NewString("")
	}
	return NewInt(1)
}

// ============================================================
// Range Operations
// ============================================================

// Range generates $a .. $b (list of values)
func Range(a, b *SV) []*SV {
	// Numeric range
	if !isStringRange(a, b) {
		start := a.AsInt()
		end := b.AsInt()

		if start > end {
			return []*SV{}
		}

		result := make([]*SV, end-start+1)
		for i := start; i <= end; i++ {
			result[i-start] = NewInt(i)
		}
		return result
	}

	// String range: "aa" .. "az"
	startStr := a.AsString()
	endStr := b.AsString()

	var result []*SV
	current := startStr

	for {
		result = append(result, NewString(current))
		if current == endStr {
			break
		}
		if len(result) > 1000000 { // Safety limit
			break
		}
		current = incrementString(current)
		// Prevent infinite loop if we passed the end
		if len(current) > len(endStr) {
			break
		}
	}

	return result
}

// isStringRange checks if this should be a string range
func isStringRange(a, b *SV) bool {
	if a.typ == TypeString && b.typ == TypeString {
		as, bs := a.pv, b.pv
		return isIncrementableString(as) && isIncrementableString(bs) &&
			!hasNumericPrefix(as) && !hasNumericPrefix(bs)
	}
	return false
}

// ============================================================
// Type Operations
// ============================================================

// Ref returns ref($a) - the reference type as string
func Ref(a *SV) *SV {
	if a == nil || a.typ != TypeRef {
		return NewString("")
	}

	// If blessed, return the class name
	if a.flags&FlagBless != 0 {
		return NewString(a.stash)
	}

	// Otherwise return the type
	if a.rv == nil {
		return NewString("SCALAR")
	}

	switch a.rv.typ {
	case TypeArray:
		return NewString("ARRAY")
	case TypeHash:
		return NewString("HASH")
	case TypeCode:
		return NewString("CODE")
	case TypeGlob:
		return NewString("GLOB")
	case TypeRegex:
		return NewString("Regexp")
	case TypeIO:
		return NewString("IO")
	default:
		return NewString("SCALAR")
	}
}

// Reftype returns reftype($a) - always the underlying type, ignoring blessing
func Reftype(a *SV) *SV {
	if a == nil || a.typ != TypeRef || a.rv == nil {
		return NewUndef()
	}

	switch a.rv.typ {
	case TypeArray:
		return NewString("ARRAY")
	case TypeHash:
		return NewString("HASH")
	case TypeCode:
		return NewString("CODE")
	case TypeGlob:
		return NewString("GLOB")
	case TypeRegex:
		return NewString("Regexp")
	case TypeIO:
		return NewString("IO")
	default:
		return NewString("SCALAR")
	}
}
