package gv

import (
	"testing"

	"perlc/pkg/sv"
)

// TestSlots tests glob slot access.
// TestSlots, glob slot erişimini test eder.
func TestSlots(t *testing.T) {
	g := New("main", "foo")

	// Scalar
	g.SetScalar(sv.NewInt(42))
	if g.Scalar().AsInt() != 42 {
		t.Error("$foo should be 42")
	}

	// Array
	arr := g.Array()
	if arr == nil || !arr.IsArray() {
		t.Error("@foo should be array")
	}

	// Hash
	hash := g.Hash()
	if hash == nil || !hash.IsHash() {
		t.Error("hash slot should be hash type")
	}

	// Code (initially nil)
	if g.HasCode() {
		t.Error("&foo should not exist yet")
	}
}

// TestAssign tests *foo = \$bar style assignment.
// TestAssign, *foo = \$bar tarzı atamayı test eder.
func TestAssign(t *testing.T) {
	g := New("main", "foo")

	// Assign scalar ref
	// Skaler referans ata
	scalar := sv.NewInt(100)
	g.Assign(sv.NewRef(scalar))

	if g.Scalar().AsInt() != 100 {
		t.Errorf("$foo should be 100, got %d", g.Scalar().AsInt())
	}

	// Assign array ref
	// Dizi referansı ata
	arr := sv.NewArrayRef(sv.NewInt(1), sv.NewInt(2))
	g.Assign(arr)

	if !g.HasArray() {
		t.Error("@foo should exist after array assign")
	}
}

// TestFullName tests name formatting.
// TestFullName, isim biçimlendirmesini test eder.
func TestFullName(t *testing.T) {
	g1 := New("main", "foo")
	if g1.FullName() != "foo" {
		t.Errorf("main::foo should display as 'foo', got '%s'", g1.FullName())
	}

	g2 := New("MyPackage", "bar")
	if g2.FullName() != "MyPackage::bar" {
		t.Errorf("Expected 'MyPackage::bar', got '%s'", g2.FullName())
	}
}
