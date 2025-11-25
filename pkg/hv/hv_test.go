package hv

import (
	"testing"

	"perlc/pkg/sv"
)

// TestFetchStore tests basic hash access.
// TestFetchStore, temel hash erişimini test eder.
func TestFetchStore(t *testing.T) {
	hash := sv.NewHashRef()

	Store(hash, sv.NewString("name"), sv.NewString("Perl"))
	Store(hash, sv.NewString("year"), sv.NewInt(1987))

	val := Fetch(hash, sv.NewString("name"))
	if val.AsString() != "Perl" {
		t.Errorf("Expected 'Perl', got '%s'", val.AsString())
	}

	val = Fetch(hash, sv.NewString("year"))
	if val.AsInt() != 1987 {
		t.Errorf("Expected 1987, got %d", val.AsInt())
	}

	// Non-existent key returns undef
	// Var olmayan anahtar undef döndürür
	val = Fetch(hash, sv.NewString("missing"))
	if !val.IsUndef() {
		t.Error("Missing key should return undef")
	}
}

// TestExists tests key existence check.
// TestExists, anahtar varlık kontrolünü test eder.
func TestExists(t *testing.T) {
	hash := sv.NewHashRef()
	Store(hash, sv.NewString("key"), sv.NewUndef())

	// Key exists even if value is undef
	// Değer undef olsa bile anahtar var
	if !Exists(hash, sv.NewString("key")).AsBool() {
		t.Error("Key should exist")
	}

	if Exists(hash, sv.NewString("missing")).AsBool() {
		t.Error("Missing key should not exist")
	}
}

// TestDelete tests key deletion.
// TestDelete, anahtar silmeyi test eder.
func TestDelete(t *testing.T) {
	hash := sv.NewHashRef()
	Store(hash, sv.NewString("key"), sv.NewInt(42))

	val := Delete(hash, sv.NewString("key"))
	if val.AsInt() != 42 {
		t.Errorf("Delete should return 42, got %d", val.AsInt())
	}

	if Exists(hash, sv.NewString("key")).AsBool() {
		t.Error("Key should not exist after delete")
	}
}

// TestKeysValues tests keys() and values().
// TestKeysValues, keys() ve values() fonksiyonlarını test eder.
func TestKeysValues(t *testing.T) {
	hash := sv.NewHashRef()
	Store(hash, sv.NewString("a"), sv.NewInt(1))
	Store(hash, sv.NewString("b"), sv.NewInt(2))
	Store(hash, sv.NewString("c"), sv.NewInt(3))

	keys := Keys(hash)
	if len(keys) != 3 {
		t.Errorf("Should have 3 keys, got %d", len(keys))
	}

	values := Values(hash)
	if len(values) != 3 {
		t.Errorf("Should have 3 values, got %d", len(values))
	}
}

// TestEach tests iteration.
// TestEach, iterasyonu test eder.
func TestEach(t *testing.T) {
	hash := sv.NewHashRef()
	Store(hash, sv.NewString("x"), sv.NewInt(10))
	Store(hash, sv.NewString("y"), sv.NewInt(20))

	count := 0
	for {
		pair := Each(hash)
		if len(pair) == 0 {
			break
		}
		count++
	}

	if count != 2 {
		t.Errorf("Should iterate 2 times, got %d", count)
	}
}

// TestFromList tests hash creation from list.
// TestFromList, listeden hash oluşturmayı test eder.
func TestFromList(t *testing.T) {
	list := []*sv.SV{
		sv.NewString("name"), sv.NewString("Perl"),
		sv.NewString("year"), sv.NewInt(1987),
	}

	hash := FromList(list)

	if Fetch(hash, sv.NewString("name")).AsString() != "Perl" {
		t.Error("name should be 'Perl'")
	}
	if Fetch(hash, sv.NewString("year")).AsInt() != 1987 {
		t.Error("year should be 1987")
	}
}

// TestMerge tests hash merging.
// TestMerge, hash birleştirmeyi test eder.
func TestMerge(t *testing.T) {
	h1 := sv.NewHashRef()
	Store(h1, sv.NewString("a"), sv.NewInt(1))
	Store(h1, sv.NewString("b"), sv.NewInt(2))

	h2 := sv.NewHashRef()
	Store(h2, sv.NewString("b"), sv.NewInt(20)) // Override
	Store(h2, sv.NewString("c"), sv.NewInt(3))

	merged := Merge(h1, h2)

	if Fetch(merged, sv.NewString("a")).AsInt() != 1 {
		t.Error("a should be 1")
	}
	if Fetch(merged, sv.NewString("b")).AsInt() != 20 {
		t.Error("b should be 20 (overridden)")
	}
	if Fetch(merged, sv.NewString("c")).AsInt() != 3 {
		t.Error("c should be 3")
	}
}

// TestSlice tests hash slicing.
// TestSlice, hash dilimlemeyi test eder.
func TestSlice(t *testing.T) {
	hash := sv.NewHashRef()
	Store(hash, sv.NewString("a"), sv.NewInt(1))
	Store(hash, sv.NewString("b"), sv.NewInt(2))
	Store(hash, sv.NewString("c"), sv.NewInt(3))

	keys := []*sv.SV{sv.NewString("a"), sv.NewString("c")}
	values := Slice(hash, keys)

	if len(values) != 2 {
		t.Errorf("Should get 2 values, got %d", len(values))
	}
	if values[0].AsInt() != 1 || values[1].AsInt() != 3 {
		t.Error("Got wrong values from slice")
	}
}
