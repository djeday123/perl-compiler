package sv

import (
	"testing"
)

func TestNewTypes(t *testing.T) {
	// Test undef
	u := NewUndef()
	if !u.IsUndef() {
		t.Error("NewUndef should be undef")
	}
	if u.AsBool() {
		t.Error("undef should be false")
	}

	// Test int
	i := NewInt(42)
	if i.AsInt() != 42 {
		t.Errorf("Expected 42, got %d", i.AsInt())
	}
	if i.AsString() != "42" {
		t.Errorf("Expected '42', got '%s'", i.AsString())
	}
	if !i.AsBool() {
		t.Error("42 should be true")
	}

	// Test zero
	z := NewInt(0)
	if z.AsBool() {
		t.Error("0 should be false")
	}

	// Test float
	f := NewFloat(3.14)
	if f.AsFloat() != 3.14 {
		t.Errorf("Expected 3.14, got %f", f.AsFloat())
	}

	// Test string
	s := NewString("hello")
	if s.AsString() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", s.AsString())
	}

	// Empty string is false
	es := NewString("")
	if es.AsBool() {
		t.Error("Empty string should be false")
	}

	// "0" is false!
	zs := NewString("0")
	if zs.AsBool() {
		t.Error("String '0' should be false in Perl")
	}
}

func TestStringCoercion(t *testing.T) {
	tests := []struct {
		input     string
		wantInt   int64
		wantFloat float64
	}{
		{"42", 42, 42.0},
		{"  42  ", 42, 42.0},
		{"42abc", 42, 42.0},
		{"abc", 0, 0.0},
		{"-17", -17, -17.0},
		{"3.14", 3, 3.14},
		{"1e5", 1, 100000.0},
		{"", 0, 0.0},
	}

	for _, tt := range tests {
		s := NewString(tt.input)
		if got := s.AsInt(); got != tt.wantInt {
			t.Errorf("AsInt(%q) = %d, want %d", tt.input, got, tt.wantInt)
		}
		if got := s.AsFloat(); got != tt.wantFloat {
			t.Errorf("AsFloat(%q) = %f, want %f", tt.input, got, tt.wantFloat)
		}
	}
}

func TestReferences(t *testing.T) {
	// Scalar ref
	scalar := NewInt(42)
	ref := NewRef(scalar)

	if !ref.IsRef() {
		t.Error("Should be a reference")
	}

	deref := ref.Deref()
	if deref.AsInt() != 42 {
		t.Errorf("Deref should give 42, got %d", deref.AsInt())
	}

	// Modifying through reference
	deref.SetInt(100)
	if scalar.AsInt() != 100 {
		t.Errorf("Original should be modified to 100, got %d", scalar.AsInt())
	}
}

func TestArrayRef(t *testing.T) {
	arr := NewArrayRef(NewInt(1), NewString("hello"), NewFloat(3.14))

	if !arr.IsRef() {
		t.Error("Should be a reference")
	}

	deref := arr.Deref()
	if !deref.IsArray() {
		t.Error("Deref should be array")
	}

	data := deref.ArrayData()
	if len(data) != 3 {
		t.Errorf("Array should have 3 elements, got %d", len(data))
	}

	if data[0].AsInt() != 1 {
		t.Errorf("First element should be 1, got %d", data[0].AsInt())
	}
	if data[1].AsString() != "hello" {
		t.Errorf("Second element should be 'hello', got '%s'", data[1].AsString())
	}
}

func TestBlessing(t *testing.T) {
	hash := NewHashRef()
	hash.Bless("MyClass")

	if !hash.IsBlessed() {
		t.Error("Should be blessed")
	}
	if hash.Package() != "MyClass" {
		t.Errorf("Package should be 'MyClass', got '%s'", hash.Package())
	}
	if !hash.Isa("MyClass") {
		t.Error("Should isa MyClass")
	}

	// String representation should include class
	s := hash.AsString()
	if s[:8] != "MyClass=" {
		t.Errorf("Blessed ref string should start with class name, got '%s'", s)
	}
}

func TestRefCount(t *testing.T) {
	sv := NewInt(42)
	if sv.RefCount() != 1 {
		t.Errorf("Initial refcount should be 1, got %d", sv.RefCount())
	}

	sv.IncRef()
	if sv.RefCount() != 2 {
		t.Errorf("After incref should be 2, got %d", sv.RefCount())
	}

	sv.DecRef()
	if sv.RefCount() != 1 {
		t.Errorf("After decref should be 1, got %d", sv.RefCount())
	}
}
