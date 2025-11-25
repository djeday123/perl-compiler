package sv

import (
	"testing"
)

func floatEq(a, b float64) bool {
	const epsilon = 0.0001
	diff := a - b
	return diff < epsilon && diff > -epsilon
}

func TestArithmetic(t *testing.T) {
	a := NewInt(10)
	b := NewInt(3)

	// Add
	if Add(a, b).AsInt() != 13 {
		t.Errorf("10 + 3 should be 13")
	}

	// Sub
	if Sub(a, b).AsInt() != 7 {
		t.Errorf("10 - 3 should be 7")
	}

	// Mul
	if Mul(a, b).AsInt() != 30 {
		t.Errorf("10 * 3 should be 30")
	}

	// Div (always float)
	if Div(a, b).AsFloat() != 10.0/3.0 {
		t.Errorf("10 / 3 should be %f", 10.0/3.0)
	}

	// Mod
	if Mod(a, b).AsInt() != 1 {
		t.Errorf("10 %% 3 should be 1")
	}

	// Power
	if Pow(NewInt(2), NewInt(10)).AsInt() != 1024 {
		t.Errorf("2 ** 10 should be 1024")
	}
}

func TestFloatArithmetic(t *testing.T) {
	a := NewFloat(3.14)
	b := NewFloat(2.0)

	result := Add(a, b)
	if !floatEq(result.AsFloat(), 5.14) {
		t.Errorf("3.14 + 2.0 should be ~5.14, got %f", result.AsFloat())
	}

	s := NewString("3.5")
	result = Add(NewInt(1), s)
	if !floatEq(result.AsFloat(), 4.5) {
		t.Errorf("1 + '3.5' should be 4.5, got %f", result.AsFloat())
	}
}

func TestStringIncrement(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"aa", "ab"},
		{"az", "ba"},
		{"zz", "aaa"},
		{"a9", "b0"},
		{"z9", "aa0"},
		{"A", "B"},
		{"Z", "AA"},
		{"a1", "a2"},
	}

	for _, tt := range tests {
		s := NewString(tt.input)
		Inc(s)
		if got := s.AsString(); got != tt.want {
			t.Errorf("++%q = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNumericIncrement(t *testing.T) {
	// Numeric string increments numerically
	s := NewString("42")
	Inc(s)
	if s.AsInt() != 43 {
		t.Errorf("++'42' should be 43, got %d", s.AsInt())
	}

	// Int
	i := NewInt(10)
	Inc(i)
	if i.AsInt() != 11 {
		t.Errorf("++10 should be 11")
	}

	// Float
	f := NewFloat(1.5)
	Inc(f)
	if f.AsFloat() != 2.5 {
		t.Errorf("++1.5 should be 2.5")
	}
}

func TestConcat(t *testing.T) {
	a := NewString("Hello")
	b := NewString(" World")

	result := Concat(a, b)
	if result.AsString() != "Hello World" {
		t.Errorf("Concat failed: got '%s'", result.AsString())
	}

	// Numbers stringify
	num := NewInt(42)
	result = Concat(NewString("Answer: "), num)
	if result.AsString() != "Answer: 42" {
		t.Errorf("Concat with number failed: got '%s'", result.AsString())
	}
}

func TestRepeat(t *testing.T) {
	s := NewString("ab")
	n := NewInt(3)

	result := Repeat(s, n)
	if result.AsString() != "ababab" {
		t.Errorf("'ab' x 3 should be 'ababab', got '%s'", result.AsString())
	}

	// Zero times
	result = Repeat(s, NewInt(0))
	if result.AsString() != "" {
		t.Errorf("'ab' x 0 should be '', got '%s'", result.AsString())
	}
}

func TestLength(t *testing.T) {
	// ASCII
	s := NewString("hello")
	if Length(s).AsInt() != 5 {
		t.Errorf("length('hello') should be 5")
	}

	// Unicode - length counts characters, not bytes
	u := NewString("привет") // 6 Cyrillic characters
	if Length(u).AsInt() != 6 {
		t.Errorf("length('привет') should be 6, got %d", Length(u).AsInt())
	}

	// Emoji
	e := NewString("👍🎉") // 2 emoji
	if Length(e).AsInt() != 2 {
		t.Errorf("length of 2 emoji should be 2, got %d", Length(e).AsInt())
	}
}

func TestSubstr(t *testing.T) {
	s := NewString("Hello World")

	// Basic substr
	result := Substr(s, NewInt(0), NewInt(5))
	if result.AsString() != "Hello" {
		t.Errorf("substr('Hello World', 0, 5) = '%s', want 'Hello'", result.AsString())
	}

	// Negative offset
	result = Substr(s, NewInt(-5), nil)
	if result.AsString() != "World" {
		t.Errorf("substr('Hello World', -5) = '%s', want 'World'", result.AsString())
	}

	// Negative length
	result = Substr(s, NewInt(0), NewInt(-6))
	if result.AsString() != "Hello" {
		t.Errorf("substr with negative length failed: '%s'", result.AsString())
	}
}

func TestNumericComparison(t *testing.T) {
	a := NewInt(10)
	b := NewInt(20)

	if !NumLt(a, b).AsBool() {
		t.Error("10 < 20 should be true")
	}
	if NumLt(b, a).AsBool() {
		t.Error("20 < 10 should be false")
	}
	if !NumEq(a, NewInt(10)).AsBool() {
		t.Error("10 == 10 should be true")
	}

	// Spaceship
	if NumCmp(a, b).AsInt() != -1 {
		t.Error("10 <=> 20 should be -1")
	}
	if NumCmp(b, a).AsInt() != 1 {
		t.Error("20 <=> 10 should be 1")
	}
	if NumCmp(a, a).AsInt() != 0 {
		t.Error("10 <=> 10 should be 0")
	}
}

func TestStringComparison(t *testing.T) {
	a := NewString("apple")
	b := NewString("banana")

	if !StrLt(a, b).AsBool() {
		t.Error("'apple' lt 'banana' should be true")
	}
	if !StrEq(a, NewString("apple")).AsBool() {
		t.Error("'apple' eq 'apple' should be true")
	}

	// cmp
	if StrCmp(a, b).AsInt() >= 0 {
		t.Error("'apple' cmp 'banana' should be negative")
	}
}

func TestLogical(t *testing.T) {
	tr := NewInt(1)
	fa := NewString("")

	// Not
	if Not(tr).AsBool() {
		t.Error("!1 should be false")
	}
	if !Not(fa).AsBool() {
		t.Error("!'' should be true")
	}

	// And - returns last evaluated
	if And(tr, NewInt(42)).AsInt() != 42 {
		t.Error("1 && 42 should return 42")
	}
	if And(fa, tr).AsString() != "" {
		t.Error("'' && 1 should return ''")
	}

	// Or - returns first true
	if Or(tr, NewInt(42)).AsInt() != 1 {
		t.Error("1 || 42 should return 1")
	}
	if Or(fa, NewInt(42)).AsInt() != 42 {
		t.Error("'' || 42 should return 42")
	}

	// Defined-or
	if DefinedOr(NewUndef(), NewInt(42)).AsInt() != 42 {
		t.Error("undef // 42 should return 42")
	}
	if DefinedOr(NewInt(0), NewInt(42)).AsInt() != 0 {
		t.Error("0 // 42 should return 0 (defined but false)")
	}
}

func TestRange(t *testing.T) {
	// Numeric range
	result := Range(NewInt(1), NewInt(5))
	if len(result) != 5 {
		t.Errorf("1..5 should have 5 elements, got %d", len(result))
	}
	for i, v := range result {
		if v.AsInt() != int64(i+1) {
			t.Errorf("Element %d should be %d, got %d", i, i+1, v.AsInt())
		}
	}

	// String range
	result = Range(NewString("aa"), NewString("ad"))
	if len(result) != 4 {
		t.Errorf("'aa'..'ad' should have 4 elements, got %d", len(result))
	}
	expected := []string{"aa", "ab", "ac", "ad"}
	for i, v := range result {
		if v.AsString() != expected[i] {
			t.Errorf("Element %d should be '%s', got '%s'", i, expected[i], v.AsString())
		}
	}
}

func TestRef(t *testing.T) {
	// Not a reference
	if Ref(NewInt(42)).AsString() != "" {
		t.Error("ref(42) should return empty string")
	}

	// Array ref
	arr := NewArrayRef()
	if Ref(arr).AsString() != "ARRAY" {
		t.Errorf("ref([]) should be 'ARRAY', got '%s'", Ref(arr).AsString())
	}

	// Hash ref
	hash := NewHashRef()
	if Ref(hash).AsString() != "HASH" {
		t.Errorf("ref({}) should be 'HASH', got '%s'", Ref(hash).AsString())
	}

	// Blessed
	hash.Bless("MyClass")
	if Ref(hash).AsString() != "MyClass" {
		t.Errorf("ref(blessed) should be 'MyClass', got '%s'", Ref(hash).AsString())
	}

	// Reftype ignores blessing
	if Reftype(hash).AsString() != "HASH" {
		t.Errorf("reftype(blessed hash) should be 'HASH', got '%s'", Reftype(hash).AsString())
	}
}

func TestBitwiseOps(t *testing.T) {
	if BitAnd(NewInt(0b1100), NewInt(0b1010)).AsInt() != 0b1000 {
		t.Error("Bitwise AND failed")
	}
	if BitOr(NewInt(0b1100), NewInt(0b1010)).AsInt() != 0b1110 {
		t.Error("Bitwise OR failed")
	}
	if BitXor(NewInt(0b1100), NewInt(0b1010)).AsInt() != 0b0110 {
		t.Error("Bitwise XOR failed")
	}
	if LeftShift(NewInt(1), NewInt(4)).AsInt() != 16 {
		t.Error("Left shift failed")
	}
	if RightShift(NewInt(16), NewInt(2)).AsInt() != 4 {
		t.Error("Right shift failed")
	}
}
