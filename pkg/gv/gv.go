// Package gv implements Perl's glob values (*foo).
// Paket gv, Perl'in glob değerlerini (*foo) uygular.
//
// A glob is a symbol table entry that can hold multiple types:
// $foo, @foo, %foo, &foo, and filehandles - all under one name.
//
// Glob, birden fazla türü tutabilen bir sembol tablosu girdisidir:
// $foo, @foo, %foo, &foo ve dosya tanıtıcıları - hepsi tek isim altında.
package gv

import (
	"fmt"

	"perlc/pkg/sv"
)

// GV represents a glob value - a container for all variable types
// sharing the same name in a package.
//
// GV, bir glob değerini temsil eder - bir pakette aynı adı paylaşan
// tüm değişken türleri için bir kapsayıcı.
type GV struct {
	name   string // Variable name (without sigil) / Değişken adı (sigil olmadan)
	pkg    string // Package name / Paket adı
	scalar *sv.SV // $name
	array  *sv.SV // @name (TypeArray)
	hash   *sv.SV // %name (TypeHash)
	code   *sv.SV // &name (TypeCode)
	io     *sv.SV // Filehandle / Dosya tanıtıcı
	//format *sv.SV // Format (for write) / Format (write için)
	//flags  uint32
}

// GV flags
// GV bayrakları
const (
	GVImported uint32 = 1 << iota // Imported from another package / Başka paketten içe aktarıldı
	GVIntro                       // First declaration / İlk bildirim
	GVMulti                       // Used in multiple ways / Birden fazla şekilde kullanıldı
	GVAssumecv                    // Assume it's a subroutine / Altyordam olduğunu varsay
	GVConst                       // Constant subroutine / Sabit altyordam
)

// New creates a new glob value.
// New, yeni bir glob değeri oluşturur.
func New(pkg, name string) *GV {
	return &GV{
		name: name,
		pkg:  pkg,
	}
}

// ============================================================
// Getters - Get slot values
// Getter'lar - Slot değerlerini al
// ============================================================

// Name returns the variable name.
// Name, değişken adını döndürür.
func (gv *GV) Name() string {
	return gv.name
}

// Package returns the package name.
// Package, paket adını döndürür.
func (gv *GV) Package() string {
	return gv.pkg
}

// FullName returns "Package::name".
// FullName, "Package::name" döndürür.
func (gv *GV) FullName() string {
	if gv.pkg == "" || gv.pkg == "main" {
		return gv.name
	}
	return gv.pkg + "::" + gv.name
}

// Scalar returns $name (creates if not exists).
// Scalar, $name döndürür (yoksa oluşturur).
func (gv *GV) Scalar() *sv.SV {
	if gv.scalar == nil {
		gv.scalar = sv.NewUndef()
	}
	return gv.scalar
}

// Array returns @name (creates if not exists).
// Array, @name döndürür (yoksa oluşturur).
func (gv *GV) Array() *sv.SV {
	if gv.array == nil {
		// Create the actual array (not a reference)
		// Gerçek diziyi oluştur (referans değil)
		gv.array = sv.NewArrayRef().Deref()
	}
	return gv.array
}

// Hash returns %name (creates if not exists).
// Hash, %name döndürür (yoksa oluşturur).
func (gv *GV) Hash() *sv.SV {
	if gv.hash == nil {
		gv.hash = sv.NewHashRef().Deref()
	}
	return gv.hash
}

// Code returns &name (may be nil).
// Code, &name döndürür (nil olabilir).
func (gv *GV) Code() *sv.SV {
	return gv.code
}

// IO returns the filehandle (may be nil).
// IO, dosya tanıtıcıyı döndürür (nil olabilir).
func (gv *GV) IO() *sv.SV {
	return gv.io
}

// ============================================================
// Setters - Set slot values
// Setter'lar - Slot değerlerini ayarla
// ============================================================

// SetScalar sets $name.
// SetScalar, $name'i ayarlar.
func (gv *GV) SetScalar(val *sv.SV) {
	if gv.scalar != nil {
		gv.scalar.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	gv.scalar = val
}

// SetArray sets @name.
// SetArray, @name'i ayarlar.
func (gv *GV) SetArray(val *sv.SV) {
	if gv.array != nil {
		gv.array.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	gv.array = val
}

// SetHash sets %name.
// SetHash, %name'i ayarlar.
func (gv *GV) SetHash(val *sv.SV) {
	if gv.hash != nil {
		gv.hash.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	gv.hash = val
}

// SetCode sets &name.
// SetCode, &name'i ayarlar.
func (gv *GV) SetCode(val *sv.SV) {
	if gv.code != nil {
		gv.code.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	gv.code = val
}

// SetIO sets the filehandle.
// SetIO, dosya tanıtıcıyı ayarlar.
func (gv *GV) SetIO(val *sv.SV) {
	if gv.io != nil {
		gv.io.DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	gv.io = val
}

// ============================================================
// Slot checking
// Slot kontrolü
// ============================================================

// HasScalar returns true if $name is defined.
// HasScalar, $name tanımlıysa true döndürür.
func (gv *GV) HasScalar() bool {
	return gv.scalar != nil && !gv.scalar.IsUndef()
}

// HasArray returns true if @name exists.
// HasArray, @name varsa true döndürür.
func (gv *GV) HasArray() bool {
	return gv.array != nil
}

// HasHash returns true if %name exists.
// HasHash, %name varsa true döndürür.
func (gv *GV) HasHash() bool {
	return gv.hash != nil
}

// HasCode returns true if &name exists.
// HasCode, &name varsa true döndürür.
func (gv *GV) HasCode() bool {
	return gv.code != nil
}

// HasIO returns true if filehandle exists.
// HasIO, dosya tanıtıcı varsa true döndürür.
func (gv *GV) HasIO() bool {
	return gv.io != nil
}

// IsEmpty returns true if glob has no defined slots.
// IsEmpty, glob'un tanımlı slotu yoksa true döndürür.
func (gv *GV) IsEmpty() bool {
	return !gv.HasScalar() && !gv.HasArray() && !gv.HasHash() && !gv.HasCode() && !gv.HasIO()
}

// ============================================================
// Glob assignment (*foo = \$bar, *foo = *bar, etc.)
// Glob ataması (*foo = \$bar, *foo = *bar, vb.)
// ============================================================

// Assign assigns a value to the appropriate slot based on type.
// If val is a reference, assigns to the slot matching the referenced type.
// If val is a glob, copies all slots.
//
// Assign, türe göre değeri uygun slota atar.
// val bir referans ise, referans verilen türle eşleşen slota atar.
// val bir glob ise, tüm slotları kopyalar.
func (gv *GV) Assign(val *sv.SV) {
	if val == nil {
		return
	}

	// If it's a reference, assign to matching slot
	// Referans ise, eşleşen slota ata
	if val.IsRef() {
		target := val.Deref()
		if target == nil {
			return
		}

		switch target.Type() {
		case sv.TypeInt, sv.TypeFloat, sv.TypeString, sv.TypeUndef:
			// Scalar reference -> $slot
			// Skaler referans -> $slot
			gv.SetScalar(target)
		case sv.TypeArray:
			// Array reference -> @slot
			// Dizi referansı -> @slot
			gv.SetArray(target)
		case sv.TypeHash:
			// Hash reference -> %slot
			// Hash referansı -> %slot
			gv.SetHash(target)
		case sv.TypeCode:
			// Code reference -> &slot
			// Kod referansı -> &slot
			gv.SetCode(target)
		case sv.TypeIO:
			// IO reference -> filehandle
			// IO referansı -> dosya tanıtıcı
			gv.SetIO(target)
		}
		return
	}

	// If it's a GV (via TypeGlob), copy all slots
	// GV ise (TypeGlob üzerinden), tüm slotları kopyala
	if val.Type() == sv.TypeGlob {
		// TODO: Extract GV from TypeGlob SV and copy slots
		// TODO: TypeGlob SV'den GV'yi çıkar ve slotları kopyala
		return
	}

	// Otherwise, assign to scalar slot
	// Aksi takdirde, skaler slota ata
	gv.SetScalar(val)
}

// ============================================================
// Utility
// Yardımcı
// ============================================================

// String returns a debug representation.
// String, hata ayıklama temsilini döndürür.
func (gv *GV) String() string {
	return fmt.Sprintf("*%s::%s", gv.pkg, gv.name)
}

// Free releases all slot references.
// Free, tüm slot referanslarını serbest bırakır.
func (gv *GV) Free() {
	if gv.scalar != nil {
		gv.scalar.DecRef()
		gv.scalar = nil
	}
	if gv.array != nil {
		gv.array.DecRef()
		gv.array = nil
	}
	if gv.hash != nil {
		gv.hash.DecRef()
		gv.hash = nil
	}
	if gv.code != nil {
		gv.code.DecRef()
		gv.code = nil
	}
	if gv.io != nil {
		gv.io.DecRef()
		gv.io = nil
	}
}
