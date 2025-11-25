// Package hv implements Perl's hash operations.
// Paket hv, Perl'in hash işlemlerini uygular.
//
// Hashes in Perl are heterogeneous - values can be any type.
// Perl'deki hash'ler heterojendir - değerler herhangi bir tür olabilir.
package hv

import (
	"sort"

	"perlc/pkg/sv"
)

// ============================================================
// Hash Access Operations
// Hash Erişim İşlemleri
// ============================================================

// Fetch gets value for key, returns undef if not found.
// Fetch, anahtar için değeri alır, bulunamazsa undef döndürür.
func Fetch(hash *sv.SV, key *sv.SV) *sv.SV {
	if hash == nil {
		return sv.NewUndef()
	}

	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewUndef()
	}

	data := target.HashData()
	if data == nil {
		return sv.NewUndef()
	}

	k := key.AsString()
	if val, ok := data[k]; ok {
		return val
	}
	return sv.NewUndef()
}

// Store sets value for key.
// Store, anahtar için değeri ayarlar.
func Store(hash *sv.SV, key *sv.SV, val *sv.SV) {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		panic("Not a hash / Hash değil")
	}

	data := target.HashData()
	if data == nil {
		data = make(map[string]*sv.SV)
		target.SetHashData(data)
	}

	k := key.AsString()

	// Handle refcounts
	// Referans sayılarını yönet
	if old, ok := data[k]; ok && old != nil {
		old.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	data[k] = val
}

// Exists checks if key exists (even if value is undef).
// Exists, anahtarın var olup olmadığını kontrol eder (değer undef olsa bile).
func Exists(hash *sv.SV, key *sv.SV) *sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewString("")
	}

	data := target.HashData()
	if data == nil {
		return sv.NewString("")
	}

	k := key.AsString()
	if _, ok := data[k]; ok {
		return sv.NewInt(1)
	}
	return sv.NewString("")
}

// Delete removes key and returns its value.
// Delete, anahtarı kaldırır ve değerini döndürür.
func Delete(hash *sv.SV, key *sv.SV) *sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewUndef()
	}

	data := target.HashData()
	if data == nil {
		return sv.NewUndef()
	}

	k := key.AsString()
	if val, ok := data[k]; ok {
		delete(data, k)
		// Don't decref - we're returning it
		// Decref yapma - döndürüyoruz
		return val
	}
	return sv.NewUndef()
}

// ============================================================
// Hash Information
// Hash Bilgisi
// ============================================================

// Scalar returns hash in scalar context.
// Empty hash returns false, non-empty returns element count.
//
// Scalar, hash'i skaler bağlamda döndürür.
// Boş hash false döndürür, dolu hash eleman sayısını döndürür.
func Scalar(hash *sv.SV) *sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewInt(0)
	}

	data := target.HashData()
	if len(data) == 0 {
		return sv.NewString("") // Empty hash is false / Boş hash false'tur
	}
	return sv.NewInt(int64(len(data)))
}

// ============================================================
// Keys, Values, Each
// Anahtarlar, Değerler, Her Biri
// ============================================================

// Keys returns all keys as a list.
// Keys, tüm anahtarları liste olarak döndürür.
func Keys(hash *sv.SV) []*sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return []*sv.SV{}
	}

	data := target.HashData()
	if data == nil {
		return []*sv.SV{}
	}

	result := make([]*sv.SV, 0, len(data))
	for k := range data {
		result = append(result, sv.NewString(k))
	}
	return result
}

// KeysSorted returns keys sorted alphabetically.
// KeysSorted, anahtarları alfabetik sıralı döndürür.
func KeysSorted(hash *sv.SV) []*sv.SV {
	keys := Keys(hash)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].AsString() < keys[j].AsString()
	})
	return keys
}

// Values returns all values as a list.
// Values, tüm değerleri liste olarak döndürür.
func Values(hash *sv.SV) []*sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return []*sv.SV{}
	}

	data := target.HashData()
	if data == nil {
		return []*sv.SV{}
	}

	result := make([]*sv.SV, 0, len(data))
	for _, v := range data {
		if v != nil {
			v.IncRef()
		}
		result = append(result, v)
	}
	return result
}

// HashIterator maintains state for each() function.
// HashIterator, each() fonksiyonu için durumu korur.
type HashIterator struct {
	keys  []string
	index int
}

// iterators stores per-hash iterator state.
// iterators, hash başına iteratör durumunu saklar.
var iterators = make(map[*sv.SV]*HashIterator)

// Each returns next (key, value) pair for iteration.
// Returns empty slice when exhausted.
//
// Each, iterasyon için sonraki (anahtar, değer) çiftini döndürür.
// Tükendiğinde boş dilim döndürür.
func Each(hash *sv.SV) []*sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return []*sv.SV{}
	}

	data := target.HashData()
	if data == nil {
		return []*sv.SV{}
	}

	// Get or create iterator
	// İteratörü al veya oluştur
	iter, ok := iterators[target]
	if !ok {
		iter = &HashIterator{
			keys:  make([]string, 0, len(data)),
			index: 0,
		}
		for k := range data {
			iter.keys = append(iter.keys, k)
		}
		iterators[target] = iter
	}

	// Return next pair
	// Sonraki çifti döndür
	if iter.index >= len(iter.keys) {
		// Reset for next iteration
		// Sonraki iterasyon için sıfırla
		delete(iterators, target)
		return []*sv.SV{}
	}

	key := iter.keys[iter.index]
	val := data[key]
	iter.index++

	if val != nil {
		val.IncRef()
	}
	return []*sv.SV{sv.NewString(key), val}
}

// ResetIterator resets the each() iterator.
// ResetIterator, each() iteratörünü sıfırlar.
func ResetIterator(hash *sv.SV) {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	delete(iterators, target)
}

// ============================================================
// Hash Manipulation
// Hash Manipülasyonu
// ============================================================

// Clear empties the hash.
// Clear, hash'i boşaltır.
func Clear(hash *sv.SV) {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return
	}

	data := target.HashData()
	if data == nil {
		return
	}

	// Decref all values
	// Tüm değerlerin referanslarını azalt
	for _, v := range data {
		if v != nil {
			v.DecRef()
		}
	}
	target.SetHashData(make(map[string]*sv.SV))
	delete(iterators, target)
}

// Clone creates a shallow copy of the hash.
// Clone, hash'in sığ bir kopyasını oluşturur.
func Clone(hash *sv.SV) *sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewHashRef()
	}

	data := target.HashData()
	newHash := sv.NewHashRef()
	newTarget := newHash.Deref()
	newData := make(map[string]*sv.SV, len(data))

	for k, v := range data {
		if v != nil {
			v.IncRef()
		}
		newData[k] = v
	}
	newTarget.SetHashData(newData)

	return newHash
}

// Merge merges multiple hashes, later values override earlier.
// Merge, birden fazla hash'i birleştirir, sonraki değerler öncekileri geçersiz kılar.
func Merge(hashes ...*sv.SV) *sv.SV {
	result := sv.NewHashRef()
	resultTarget := result.Deref()
	resultData := make(map[string]*sv.SV)

	for _, hash := range hashes {
		target := hash
		if hash.IsRef() {
			target = hash.Deref()
		}
		if target == nil || !target.IsHash() {
			continue
		}

		data := target.HashData()
		for k, v := range data {
			// Decref old value if overwriting
			// Üzerine yazılıyorsa eski değerin referansını azalt
			if old, ok := resultData[k]; ok && old != nil {
				old.DecRef()
			}
			if v != nil {
				v.IncRef()
			}
			resultData[k] = v
		}
	}

	resultTarget.SetHashData(resultData)
	return result
}

// ============================================================
// Hash Slice Operations
// Hash Dilim İşlemleri
// ============================================================

// Slice gets multiple values: @hash{@keys}
// Slice, birden fazla değer alır: @hash{@keys}
func Slice(hash *sv.SV, keys []*sv.SV) []*sv.SV {
	result := make([]*sv.SV, len(keys))
	for i, k := range keys {
		result[i] = Fetch(hash, k)
		result[i].IncRef()
	}
	return result
}

// SliceStore sets multiple values: @hash{@keys} = @values
// SliceStore, birden fazla değer ayarlar: @hash{@keys} = @values
func SliceStore(hash *sv.SV, keys []*sv.SV, values []*sv.SV) {
	for i, k := range keys {
		var v *sv.SV
		if i < len(values) {
			v = values[i]
		} else {
			v = sv.NewUndef()
		}
		Store(hash, k, v)
	}
}

// DeleteSlice deletes multiple keys, returns removed values.
// DeleteSlice, birden fazla anahtarı siler, kaldırılan değerleri döndürür.
func DeleteSlice(hash *sv.SV, keys []*sv.SV) []*sv.SV {
	result := make([]*sv.SV, len(keys))
	for i, k := range keys {
		result[i] = Delete(hash, k)
	}
	return result
}

// ============================================================
// Hash to List Conversion
// Hash'ten Listeye Dönüşüm
// ============================================================

// Flatten returns hash as flat list: (key1, val1, key2, val2, ...)
// Flatten, hash'i düz liste olarak döndürür: (anahtar1, değer1, anahtar2, değer2, ...)
func Flatten(hash *sv.SV) []*sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return []*sv.SV{}
	}

	data := target.HashData()
	result := make([]*sv.SV, 0, len(data)*2)

	for k, v := range data {
		result = append(result, sv.NewString(k))
		if v != nil {
			v.IncRef()
		}
		result = append(result, v)
	}

	return result
}

// FromList creates hash from flat list.
// Odd number of elements: last key gets undef.
//
// FromList, düz listeden hash oluşturur.
// Tek sayıda eleman: son anahtar undef alır.
func FromList(list []*sv.SV) *sv.SV {
	hash := sv.NewHashRef()
	target := hash.Deref()
	data := make(map[string]*sv.SV)

	for i := 0; i < len(list)-1; i += 2 {
		k := list[i].AsString()
		v := list[i+1]
		if v != nil {
			v.IncRef()
		}
		data[k] = v
	}

	// Odd number of elements - last key gets undef
	// Tek sayıda eleman - son anahtar undef alır
	if len(list)%2 == 1 {
		k := list[len(list)-1].AsString()
		data[k] = sv.NewUndef()
	}

	target.SetHashData(data)
	return hash
}

// ============================================================
// Utility Functions
// Yardımcı Fonksiyonlar
// ============================================================

// Invert swaps keys and values: %inverse = reverse %hash
// Invert, anahtarları ve değerleri değiştirir: %inverse = reverse %hash
func Invert(hash *sv.SV) *sv.SV {
	target := hash
	if hash.IsRef() {
		target = hash.Deref()
	}
	if target == nil || !target.IsHash() {
		return sv.NewHashRef()
	}

	data := target.HashData()
	result := sv.NewHashRef()
	resultTarget := result.Deref()
	resultData := make(map[string]*sv.SV, len(data))

	for k, v := range data {
		newKey := ""
		if v != nil {
			newKey = v.AsString()
		}
		resultData[newKey] = sv.NewString(k)
	}

	resultTarget.SetHashData(resultData)
	return result
}
