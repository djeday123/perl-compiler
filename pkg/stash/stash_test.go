package stash

import (
	"testing"

	"perlc/pkg/sv"
)

// TestBasicAccess tests symbol storage and retrieval.
// TestBasicAccess, sembol depolama ve almayı test eder.
func TestBasicAccess(t *testing.T) {
	s := Get("TestPkg")

	s.SetScalar("foo", sv.NewInt(42))

	if s.Scalar("foo").AsInt() != 42 {
		t.Error("$TestPkg::foo should be 42")
	}
}

// TestNestedPackages tests package hierarchy.
// TestNestedPackages, paket hiyerarşisini test eder.
func TestNestedPackages(t *testing.T) {
	s := Get("Foo::Bar::Baz")

	if s.Name() != "Foo::Bar::Baz" {
		t.Errorf("Package name wrong: %s", s.Name())
	}

	// Parent stashes should exist
	// Üst stash'ler var olmalı
	if !Exists("Foo::Bar") {
		t.Error("Foo::Bar should exist")
	}
	if !Exists("Foo") {
		t.Error("Foo should exist")
	}
}

// TestResolve tests symbolic reference resolution.
// TestResolve, sembolik referans çözümlemeyi test eder.
func TestResolve(t *testing.T) {
	Get("MyMod").SetScalar("version", sv.NewString("1.0"))

	val := ResolveScalar("MyMod::version")
	if val.AsString() != "1.0" {
		t.Errorf("Expected '1.0', got '%s'", val.AsString())
	}
}

// TestInheritance tests @ISA and method resolution.
// TestInheritance, @ISA ve metod çözümlemeyi test eder.
func TestInheritance(t *testing.T) {
	// Setup: Child inherits from Parent
	// Kurulum: Child, Parent'tan miras alır
	parent := Get("Parent")
	parent.SetCode("greet", sv.NewInt(1)) // Dummy code / Kukla kod

	child := Get("Child")
	child.AddISA("Parent")

	// Child should find parent's method
	// Child, parent'ın metodunu bulmalı
	code, pkg := child.FindMethod("greet")
	if code == nil {
		t.Error("Should find greet method")
	}
	if pkg != "Parent" {
		t.Errorf("Method should be from Parent, got %s", pkg)
	}

	// isa check
	// isa kontrolü
	if !child.Isa("Parent") {
		t.Error("Child should isa Parent")
	}
	if !child.Isa("UNIVERSAL") {
		t.Error("Everything should isa UNIVERSAL")
	}
}

// TestCan tests method existence check.
// TestCan, metod varlık kontrolünü test eder.
func TestCan(t *testing.T) {
	s := Get("CanTest")
	s.SetCode("exists", sv.NewInt(1))

	if !s.Can("exists") {
		t.Error("Should have 'exists' method")
	}
	if s.Can("missing") {
		t.Error("Should not have 'missing' method")
	}
}

// TestSymbols tests symbol enumeration.
// TestSymbols, sembol numaralandırmayı test eder.
func TestSymbols(t *testing.T) {
	s := Get("EnumTest")
	s.SetScalar("a", sv.NewInt(1))
	s.SetScalar("b", sv.NewInt(2))
	s.SetCode("func", sv.NewInt(3))

	symbols := s.Symbols()
	if len(symbols) < 3 {
		t.Errorf("Should have at least 3 symbols, got %d", len(symbols))
	}
}
