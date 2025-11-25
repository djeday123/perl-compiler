// Package av implements Perl's array operations.
// Paket av, Perl'in dizi işlemlerini uygular.
//
// Arrays in Perl are heterogeneous - can hold any mix of types.
// Perl'deki diziler heterojendir - herhangi bir tür karışımı içerebilir.
package av

import (
	"sort"

	"perlc/pkg/sv"
)

// ============================================================
// Array Access Operations
// Dizi Erişim İşlemleri
// ============================================================

// Fetch gets element at index (handles negative indices like Perl).
// Fetch, indeksteki öğeyi alır (Perl gibi negatif indeksleri destekler).
func Fetch(arr *sv.SV, idx *sv.SV) *sv.SV {
	if arr == nil {
		return sv.NewUndef()
	}

	// Dereference if it's a reference
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewUndef()
	}

	elements := target.ArrayData()
	if elements == nil {
		return sv.NewUndef()
	}

	i := int(idx.AsInt())
	length := len(elements)

	// Handle negative index
	if i < 0 {
		i = length + i
	}

	if i < 0 || i >= length {
		return sv.NewUndef()
	}

	if elements[i] == nil {
		return sv.NewUndef()
	}
	return elements[i]
}

// Store sets element at index (auto-extends array)
// Store, indeksteki öğeyi ayarlar (gerekirse diziyi otomatik genişletir).
func Store(arr *sv.SV, idx *sv.SV, val *sv.SV) {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		panic("Not an array / Dizi değil")
	}

	i := int(idx.AsInt())
	elements := target.ArrayData()
	length := len(elements)

	// Handle negative index
	// Negatif indeks işle
	if i < 0 {
		i = length + i
	}
	if i < 0 {
		panic("Modification of non-creatable array value attempted / Oluşturulamaz dizi değeri değiştirilmeye çalışıldı")
	}

	// Auto-extend if needed
	// Gerekirse otomatik genişlet
	if i >= length {
		newElements := make([]*sv.SV, i+1)
		copy(newElements, elements)
		// Fill gaps with undef
		// Boşlukları undef ile doldur
		for j := length; j < i; j++ {
			newElements[j] = sv.NewUndef()
		}
		elements = newElements
		target.SetArrayData(elements)
	}

	// Handle refcounts
	// Referans sayılarını yönet
	if elements[i] != nil {
		elements[i].DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	elements[i] = val
}

// Len returns scalar(@arr) - the length.
// Len, scalar(@arr) döndürür - uzunluk.
func Len(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewInt(0)
	}
	return sv.NewInt(int64(len(target.ArrayData())))
}

// MaxIndex returns $#arr - the last index.
// MaxIndex, $#arr döndürür - son indeks.
func MaxIndex(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewInt(-1)
	}
	return sv.NewInt(int64(len(target.ArrayData()) - 1))
}

// Exists checks if index exists (even if value is undef).
// Exists, indeksin var olup olmadığını kontrol eder (değer undef olsa bile).
func Exists(arr *sv.SV, idx *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewString("")
	}

	elements := target.ArrayData()
	i := int(idx.AsInt())

	if i < 0 {
		i = len(elements) + i
	}

	if i >= 0 && i < len(elements) {
		return sv.NewInt(1)
	}
	return sv.NewString("")
}

// Delete removes element at index, returns the removed value.
// Delete, indeksteki öğeyi kaldırır, kaldırılan değeri döndürür.
func Delete(arr *sv.SV, idx *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewUndef()
	}

	elements := target.ArrayData()
	i := int(idx.AsInt())

	if i < 0 {
		i = len(elements) + i
	}

	if i < 0 || i >= len(elements) {
		return sv.NewUndef()
	}

	val := elements[i]
	elements[i] = sv.NewUndef()

	// Shrink array if we deleted from the end
	// Sondan sildiyse diziyi küçült
	for len(elements) > 0 && elements[len(elements)-1].IsUndef() {
		elements = elements[:len(elements)-1]
	}
	target.SetArrayData(elements)

	return val
}

// Push appends elements to end, returns new length.
// Push, öğeleri sona ekler, yeni uzunluğu döndürür.
func Push(arr *sv.SV, values ...*sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		panic("Not an array / Dizi değil")
	}

	elements := target.ArrayData()
	for _, v := range values {
		if v != nil {
			v.IncRef()
		}
		elements = append(elements, v)
	}
	target.SetArrayData(elements)

	return sv.NewInt(int64(len(elements)))
}

// Pop removes and returns last element.
// Pop, son öğeyi kaldırır ve döndürür.
func Pop(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewUndef()
	}

	elements := target.ArrayData()
	if len(elements) == 0 {
		return sv.NewUndef()
	}

	last := elements[len(elements)-1]
	target.SetArrayData(elements[:len(elements)-1])

	// Don't decref - we're returning it
	// Decref yapma - döndürüyoruz
	return last
}

// Shift removes and returns first element.
// Shift, ilk öğeyi kaldırır ve döndürür.
func Shift(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewUndef()
	}

	elements := target.ArrayData()
	if len(elements) == 0 {
		return sv.NewUndef()
	}

	first := elements[0]
	target.SetArrayData(elements[1:])

	return first
}

// Unshift prepends elements, returns new length.
// Unshift, öğeleri başa ekler, yeni uzunluğu döndürür.
func Unshift(arr *sv.SV, values ...*sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		panic("Not an array / Dizi değil")
	}

	elements := target.ArrayData()

	// Incref new values
	// Yeni değerlerin referanslarını artır
	for _, v := range values {
		if v != nil {
			v.IncRef()
		}
	}

	newElements := make([]*sv.SV, len(values)+len(elements))
	copy(newElements, values)
	copy(newElements[len(values):], elements)
	target.SetArrayData(newElements)

	return sv.NewInt(int64(len(newElements)))
}

// ============================================================
// Splice Operation
// Splice İşlemi
// ============================================================

// Splice implements splice(@arr, $offset, $length, @list).
// Perl's Swiss Army knife for array manipulation.
//
// Splice, splice(@arr, $offset, $length, @list) işlevini uygular.
// Perl'in dizi manipülasyonu için İsviçre çakısı.
func Splice(arr *sv.SV, offset, length *sv.SV, replacement []*sv.SV) []*sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		panic("Not an array / Dizi değil")
	}

	elements := target.ArrayData()
	arrLen := len(elements)

	// Calculate offset
	// Ofseti hesapla
	off := 0
	if offset != nil && !offset.IsUndef() {
		off = int(offset.AsInt())
	}
	if off < 0 {
		off = arrLen + off
	}
	if off < 0 {
		off = 0
	}
	if off > arrLen {
		off = arrLen
	}

	// Calculate length to remove
	// Kaldırılacak uzunluğu hesapla
	removeLen := arrLen - off // Default: remove everything after offset
	if length != nil && !length.IsUndef() {
		removeLen = int(length.AsInt())
	}
	if removeLen < 0 {
		removeLen = arrLen - off + removeLen
	}
	if removeLen < 0 {
		removeLen = 0
	}
	if off+removeLen > arrLen {
		removeLen = arrLen - off
	}

	// Extract removed elements
	// Kaldırılan öğeleri çıkar
	removed := make([]*sv.SV, removeLen)
	copy(removed, elements[off:off+removeLen])

	// Incref replacement values
	// Yeni değerlerin referanslarını artır
	for _, v := range replacement {
		if v != nil {
			v.IncRef()
		}
	}

	// Build new array
	// Yeni diziyi oluştur
	newLen := arrLen - removeLen + len(replacement)
	newElements := make([]*sv.SV, newLen)

	copy(newElements, elements[:off])
	copy(newElements[off:], replacement)
	copy(newElements[off+len(replacement):], elements[off+removeLen:])

	target.SetArrayData(newElements)

	return removed
}

// ============================================================
// Transformation Operations
// Dönüşüm İşlemleri
// ============================================================

// Reverse reverses array in place and returns it.
// Reverse, diziyi yerinde ters çevirir ve döndürür.
func Reverse(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return arr
	}

	elements := target.ArrayData()
	for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
		elements[i], elements[j] = elements[j], elements[i]
	}

	return arr
}

// CmpFunc is the comparison function type for sorting.
// CmpFunc, sıralama için karşılaştırma fonksiyonu türüdür.
type CmpFunc func(a, b *sv.SV) int

// Sort sorts array using comparison function.
// If cmpFn is nil, uses string comparison (Perl default).
//
// Sort, karşılaştırma fonksiyonu kullanarak diziyi sıralar.
// cmpFn nil ise, string karşılaştırması kullanır (Perl varsayılanı).
func Sort(arr *sv.SV, cmpFn CmpFunc) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return arr
	}

	elements := target.ArrayData()

	if cmpFn == nil {
		// Default: string comparison
		// Varsayılan: string karşılaştırması
		cmpFn = func(a, b *sv.SV) int {
			return int(sv.StrCmp(a, b).AsInt())
		}
	}

	sort.SliceStable(elements, func(i, j int) bool {
		return cmpFn(elements[i], elements[j]) < 0
	})

	return arr
}

// SortNumeric sorts array numerically.
// SortNumeric, diziyi sayısal olarak sıralar.
func SortNumeric(arr *sv.SV) *sv.SV {
	return Sort(arr, func(a, b *sv.SV) int {
		return int(sv.NumCmp(a, b).AsInt())
	})
}

// ============================================================
// List Operations (grep, map, join)
// Liste İşlemleri (grep, map, join)
// ============================================================

// Grep filters array elements using predicate function.
// Returns elements for which predicate returns true.
//
// Grep, predicate fonksiyonu kullanarak dizi öğelerini filtreler.
// Predicate'in true döndürdüğü öğeleri döndürür.
func Grep(arr *sv.SV, predicate func(*sv.SV) bool) []*sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return []*sv.SV{}
	}

	var result []*sv.SV
	for _, el := range target.ArrayData() {
		if predicate(el) {
			el.IncRef()
			result = append(result, el)
		}
	}
	return result
}

// Map transforms array elements using transform function.
// Transform can return multiple values (like Perl's map).
//
// Map, transform fonksiyonu kullanarak dizi öğelerini dönüştürür.
// Transform birden fazla değer döndürebilir (Perl'in map'i gibi).
func Map(arr *sv.SV, transform func(*sv.SV) []*sv.SV) []*sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return []*sv.SV{}
	}

	var result []*sv.SV
	for _, el := range target.ArrayData() {
		transformed := transform(el)
		result = append(result, transformed...)
	}
	return result
}

// Join joins array elements with separator string.
// Join, dizi öğelerini ayırıcı string ile birleştirir.
func Join(sep *sv.SV, arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewString("")
	}

	elements := target.ArrayData()
	if len(elements) == 0 {
		return sv.NewString("")
	}

	sepStr := sep.AsString()
	result := elements[0].AsString()

	for i := 1; i < len(elements); i++ {
		result += sepStr + elements[i].AsString()
	}

	return sv.NewString(result)
}

// ============================================================
// Utility Operations
// Yardımcı İşlemler
// ============================================================

// Clear empties the array.
// Clear, diziyi boşaltır.
func Clear(arr *sv.SV) {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return
	}

	// Decref all elements
	// Tüm öğelerin referanslarını azalt
	for _, el := range target.ArrayData() {
		if el != nil {
			el.DecRef()
		}
	}
	target.SetArrayData([]*sv.SV{})
}

// Clone creates a shallow copy of the array.
// Clone, dizinin sığ bir kopyasını oluşturur.
func Clone(arr *sv.SV) *sv.SV {
	target := arr
	if arr.IsRef() {
		target = arr.Deref()
	}
	if target == nil || !target.IsArray() {
		return sv.NewArrayRef()
	}

	elements := target.ArrayData()
	newElements := make([]*sv.SV, len(elements))
	for i, el := range elements {
		if el != nil {
			el.IncRef()
		}
		newElements[i] = el
	}

	return sv.NewArrayRef(newElements...)
}

// Flatten takes nested arrayrefs and flattens one level.
// Flatten, iç içe dizi referanslarını bir seviye düzleştirir.
func Flatten(elements []*sv.SV) []*sv.SV {
	var result []*sv.SV
	for _, el := range elements {
		if el.IsRef() && el.Deref().IsArray() {
			inner := el.Deref().ArrayData()
			for _, item := range inner {
				if item != nil {
					item.IncRef()
				}
				result = append(result, item)
			}
		} else {
			if el != nil {
				el.IncRef()
			}
			result = append(result, el)
		}
	}
	return result
}

// ============================================================
// Slice Operations
// Dilim İşlemleri
// ============================================================

// Slice gets multiple elements: @arr[@indices]
// Slice, birden fazla öğe alır: @arr[@indices]
func Slice(arr *sv.SV, indices []*sv.SV) []*sv.SV {
	result := make([]*sv.SV, len(indices))
	for i, idx := range indices {
		result[i] = Fetch(arr, idx)
		result[i].IncRef()
	}
	return result
}

// SliceStore sets multiple elements: @arr[@indices] = @values
// SliceStore, birden fazla öğe ayarlar: @arr[@indices] = @values
func SliceStore(arr *sv.SV, indices []*sv.SV, values []*sv.SV) {
	for i, idx := range indices {
		var v *sv.SV
		if i < len(values) {
			v = values[i]
		} else {
			v = sv.NewUndef()
		}
		Store(arr, idx, v)
	}
}

// ============================================================
// Context Detection
// Bağlam Algılama
// ============================================================

// Context represents Perl's wantarray() - calling context.
// Context, Perl'in wantarray() fonksiyonunu temsil eder - çağrı bağlamı.
type Context int

const (
	ContextVoid   Context = iota // No return value expected / Dönüş değeri beklenmiyor
	ContextScalar                // Scalar value expected / Skaler değer bekleniyor
	ContextList                  // List of values expected / Değer listesi bekleniyor
)
