package av

import (
	"testing"

	"perlc/pkg/sv"
)

// TestPushPop tests stack operations.
// TestPushPop, yığın işlemlerini test eder.
func TestPushPop(t *testing.T) {
	arr := sv.NewArrayRef()

	Push(arr, sv.NewInt(1), sv.NewInt(2), sv.NewInt(3))

	if Len(arr).AsInt() != 3 {
		t.Errorf("Length should be 3, got %d", Len(arr).AsInt())
	}

	val := Pop(arr)
	if val.AsInt() != 3 {
		t.Errorf("Pop should return 3, got %d", val.AsInt())
	}

	if Len(arr).AsInt() != 2 {
		t.Errorf("Length should be 2 after pop, got %d", Len(arr).AsInt())
	}
}

// TestShiftUnshift tests queue operations.
// TestShiftUnshift, kuyruk işlemlerini test eder.
func TestShiftUnshift(t *testing.T) {
	arr := sv.NewArrayRef()

	Push(arr, sv.NewInt(2), sv.NewInt(3))
	Unshift(arr, sv.NewInt(1))

	if Len(arr).AsInt() != 3 {
		t.Errorf("Length should be 3, got %d", Len(arr).AsInt())
	}

	val := Shift(arr)
	if val.AsInt() != 1 {
		t.Errorf("Shift should return 1, got %d", val.AsInt())
	}
}

// TestFetchStore tests element access.
// TestFetchStore, eleman erişimini test eder.
func TestFetchStore(t *testing.T) {
	arr := sv.NewArrayRef()

	// Auto-extend on store
	// Store'da otomatik genişletme
	Store(arr, sv.NewInt(2), sv.NewString("hello"))

	if Len(arr).AsInt() != 3 {
		t.Errorf("Length should be 3, got %d", Len(arr).AsInt())
	}

	val := Fetch(arr, sv.NewInt(2))
	if val.AsString() != "hello" {
		t.Errorf("Fetch should return 'hello', got '%s'", val.AsString())
	}

	// Negative index
	// Negatif indeks
	val = Fetch(arr, sv.NewInt(-1))
	if val.AsString() != "hello" {
		t.Errorf("Fetch(-1) should return 'hello', got '%s'", val.AsString())
	}
}

// TestSplice tests the Swiss Army knife.
// TestSplice, İsviçre çakısını test eder.
func TestSplice(t *testing.T) {
	arr := sv.NewArrayRef()
	Push(arr, sv.NewInt(1), sv.NewInt(2), sv.NewInt(3), sv.NewInt(4), sv.NewInt(5))

	// Remove 2 elements starting at index 1, insert "a", "b"
	// İndeks 1'den başlayarak 2 öğe kaldır, "a", "b" ekle
	removed := Splice(arr, sv.NewInt(1), sv.NewInt(2), []*sv.SV{sv.NewString("a"), sv.NewString("b")})

	if len(removed) != 2 {
		t.Errorf("Should remove 2 elements, got %d", len(removed))
	}
	if removed[0].AsInt() != 2 || removed[1].AsInt() != 3 {
		t.Error("Removed wrong elements")
	}

	// Array should be [1, "a", "b", 4, 5]
	// Dizi [1, "a", "b", 4, 5] olmalı
	if Len(arr).AsInt() != 5 {
		t.Errorf("Length should be 5, got %d", Len(arr).AsInt())
	}
	if Fetch(arr, sv.NewInt(1)).AsString() != "a" {
		t.Error("Element at index 1 should be 'a'")
	}
}

// TestJoin tests joining array elements.
// TestJoin, dizi öğelerini birleştirmeyi test eder.
func TestJoin(t *testing.T) {
	arr := sv.NewArrayRef()
	Push(arr, sv.NewString("a"), sv.NewString("b"), sv.NewString("c"))

	result := Join(sv.NewString("-"), arr)
	if result.AsString() != "a-b-c" {
		t.Errorf("Join should be 'a-b-c', got '%s'", result.AsString())
	}
}

// TestSort tests sorting.
// TestSort, sıralamayı test eder.
func TestSort(t *testing.T) {
	arr := sv.NewArrayRef()
	Push(arr, sv.NewString("banana"), sv.NewString("apple"), sv.NewString("cherry"))

	Sort(arr, nil) // Default string sort

	if Fetch(arr, sv.NewInt(0)).AsString() != "apple" {
		t.Error("First element should be 'apple'")
	}
	if Fetch(arr, sv.NewInt(2)).AsString() != "cherry" {
		t.Error("Last element should be 'cherry'")
	}
}

// TestSortNumeric tests numeric sorting.
// TestSortNumeric, sayısal sıralamayı test eder.
func TestSortNumeric(t *testing.T) {
	arr := sv.NewArrayRef()
	Push(arr, sv.NewInt(10), sv.NewInt(2), sv.NewInt(100))

	SortNumeric(arr)

	if Fetch(arr, sv.NewInt(0)).AsInt() != 2 {
		t.Errorf("First should be 2, got %d", Fetch(arr, sv.NewInt(0)).AsInt())
	}
	if Fetch(arr, sv.NewInt(2)).AsInt() != 100 {
		t.Errorf("Last should be 100, got %d", Fetch(arr, sv.NewInt(2)).AsInt())
	}
}
