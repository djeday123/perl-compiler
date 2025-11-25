// Package cv implements Perl's code values (subroutines).
// Paket cv, Perl'in kod değerlerini (altyordamları) uygular.
//
// A CV represents a subroutine - named or anonymous, with optional closure.
// CV bir altyordamı temsil eder - isimli veya anonim, opsiyonel closure ile.
package cv

import (
	"perlc/pkg/sv"
)

// CV represents a code value (subroutine).
// CV, bir kod değerini (altyordam) temsil eder.
type CV struct {
	name  string // Subroutine name (empty for anonymous) / Altyordam adı (anonim için boş)
	pkg   string // Package where defined / Tanımlandığı paket
	proto string // Prototype e.g. "$$@" / Prototip örn. "$$@"

	// The actual code - one of these will be set
	// Gerçek kod - bunlardan biri ayarlanacak
	native func(*CallContext) *sv.SV // Native Go implementation / Native Go implementasyonu
	ops    []Op                      // Compiled opcodes (for eval'd code) / Derlenmiş opcode'lar (eval'd kod için)

	// Closure support
	// Closure desteği
	outer    *CV      // Enclosing subroutine (for closures) / Kapsayan altyordam (closure'lar için)
	padnames []string // Names of lexical variables / Leksikal değişken isimleri
	pad      []*sv.SV // Captured lexical values (closure) / Yakalanan leksikal değerler (closure)

	// Attributes
	// Özellikler
	flags CVFlags
	attrs map[string]string // :lvalue, :method, etc. / :lvalue, :method, vb.

	// For XS/external code
	// XS/harici kod için
	//xsub interface{} // External function pointer / Harici fonksiyon işaretçisi
}

// CVFlags represents subroutine flags.
// CVFlags, altyordam bayraklarını temsil eder.
type CVFlags uint32

const (
	CVAnon    CVFlags = 1 << iota // Anonymous sub / Anonim altyordam
	CVClone                       // Cloned for closure / Closure için klonlanmış
	CVLvalue                      // :lvalue attribute / :lvalue özelliği
	CVMethod                      // :method attribute / :method özelliği
	CVLocked                      // :locked attribute / :locked özelliği
	CVConst                       // Constant subroutine / Sabit altyordam
	CVIsXSUB                      // Is XSUB (external) / XSUB'dur (harici)
	CVNodebug                     // Skip in debugger / Debugger'da atla
	CVProto                       // Has prototype / Prototipi var
)

// Op represents a compiled operation (for eval).
// Op, derlenmiş bir işlemi temsil eder (eval için).
type Op interface {
	Execute(ctx *CallContext) *sv.SV
}

// CallContext holds the runtime context for a subroutine call.
// CallContext, bir altyordam çağrısı için çalışma zamanı bağlamını tutar.
type CallContext struct {
	Args      []*sv.SV     // @_ arguments / @_ argümanlar
	Pad       []*sv.SV     // Lexical variables / Leksikal değişkenler
	WantArray int          // Context: -1=void, 0=scalar, 1=list / Bağlam: -1=void, 0=skaler, 1=liste
	Package   string       // Current package / Geçerli paket
	File      string       // Source file / Kaynak dosya
	Line      int          // Source line / Kaynak satır
	Caller    *CallContext // Caller's context / Çağıranın bağlamı
	CV        *CV          // Current CV / Geçerli CV
	Error     *sv.SV       // $@ error / $@ hata
}

// ============================================================
// Constructors
// Yapıcılar
// ============================================================

// New creates a new CV with a native Go function.
// New, native Go fonksiyonu ile yeni bir CV oluşturur.
func New(pkg, name string, fn func(*CallContext) *sv.SV) *CV {
	return &CV{
		name:   name,
		pkg:    pkg,
		native: fn,
	}
}

// NewAnon creates an anonymous subroutine.
// NewAnon, anonim bir altyordam oluşturur.
func NewAnon(pkg string, fn func(*CallContext) *sv.SV) *CV {
	return &CV{
		pkg:    pkg,
		native: fn,
		flags:  CVAnon,
	}
}

// NewWithProto creates a CV with prototype.
// NewWithProto, prototip ile bir CV oluşturur.
func NewWithProto(pkg, name, proto string, fn func(*CallContext) *sv.SV) *CV {
	return &CV{
		name:   name,
		pkg:    pkg,
		proto:  proto,
		native: fn,
		flags:  CVProto,
	}
}

// NewClosure creates a closure that captures lexical variables.
// NewClosure, leksikal değişkenleri yakalayan bir closure oluşturur.
func NewClosure(outer *CV, pad []*sv.SV) *CV {
	// Clone pad values with incref
	// Pad değerlerini incref ile klonla
	clonedPad := make([]*sv.SV, len(pad))
	for i, v := range pad {
		if v != nil {
			v.IncRef()
		}
		clonedPad[i] = v
	}

	return &CV{
		pkg:      outer.pkg,
		native:   outer.native,
		ops:      outer.ops,
		outer:    outer,
		padnames: outer.padnames,
		pad:      clonedPad,
		flags:    outer.flags | CVClone,
	}
}

// ============================================================
// Execution
// Çalıştırma
// ============================================================

// Call executes the subroutine with given arguments.
// Call, altyordamı verilen argümanlarla çalıştırır.
func (cv *CV) Call(ctx *CallContext) *sv.SV {
	if ctx == nil {
		ctx = &CallContext{}
	}
	ctx.CV = cv
	ctx.Package = cv.pkg

	// Merge closure pad with call pad
	// Closure pad'i çağrı pad'i ile birleştir
	if len(cv.pad) > 0 {
		if ctx.Pad == nil {
			ctx.Pad = make([]*sv.SV, len(cv.pad))
		}
		for i, v := range cv.pad {
			if v != nil && (i >= len(ctx.Pad) || ctx.Pad[i] == nil) {
				if i >= len(ctx.Pad) {
					newPad := make([]*sv.SV, i+1)
					copy(newPad, ctx.Pad)
					ctx.Pad = newPad
				}
				ctx.Pad[i] = v
			}
		}
	}

	// Execute native function
	// Native fonksiyonu çalıştır
	if cv.native != nil {
		return cv.native(ctx)
	}

	// Execute compiled ops (for eval)
	// Derlenmiş op'ları çalıştır (eval için)
	if cv.ops != nil {
		var result *sv.SV
		for _, op := range cv.ops {
			result = op.Execute(ctx)
		}
		return result
	}

	return sv.NewUndef()
}

// CallList calls subroutine in list context, returns multiple values.
// CallList, altyordamı liste bağlamında çağırır, birden fazla değer döndürür.
func (cv *CV) CallList(ctx *CallContext) []*sv.SV {
	if ctx == nil {
		ctx = &CallContext{}
	}
	ctx.WantArray = 1

	result := cv.Call(ctx)

	// If result is an array, return its elements
	// Sonuç dizi ise, öğelerini döndür
	if result != nil && result.IsRef() {
		deref := result.Deref()
		if deref != nil && deref.IsArray() {
			return deref.ArrayData()
		}
	}

	if result == nil {
		return []*sv.SV{}
	}
	return []*sv.SV{result}
}

// ============================================================
// Properties
// Özellikler
// ============================================================

// Name returns the subroutine name.
// Name, altyordam adını döndürür.
func (cv *CV) Name() string {
	return cv.name
}

// Package returns the package name.
// Package, paket adını döndürür.
func (cv *CV) Package() string {
	return cv.pkg
}

// FullName returns "Package::name".
// FullName, "Package::name" döndürür.
func (cv *CV) FullName() string {
	if cv.name == "" {
		return cv.pkg + "::__ANON__"
	}
	if cv.pkg == "" || cv.pkg == "main" {
		return cv.name
	}
	return cv.pkg + "::" + cv.name
}

// Prototype returns the prototype string.
// Prototype, prototip dizesini döndürür.
func (cv *CV) Prototype() string {
	return cv.proto
}

// SetPrototype sets the prototype.
// SetPrototype, prototipi ayarlar.
func (cv *CV) SetPrototype(proto string) {
	cv.proto = proto
	if proto != "" {
		cv.flags |= CVProto
	}
}

// IsAnon returns true if anonymous sub.
// IsAnon, anonim altyordam ise true döndürür.
func (cv *CV) IsAnon() bool {
	return cv.flags&CVAnon != 0
}

// IsClosure returns true if this is a closure clone.
// IsClosure, bu bir closure klonu ise true döndürür.
func (cv *CV) IsClosure() bool {
	return cv.flags&CVClone != 0
}

// HasProto returns true if has prototype.
// HasProto, prototipi varsa true döndürür.
func (cv *CV) HasProto() bool {
	return cv.flags&CVProto != 0
}

// ============================================================
// Attributes
// Özellikler
// ============================================================

// SetAttr sets a subroutine attribute.
// SetAttr, bir altyordam özelliği ayarlar.
func (cv *CV) SetAttr(name, value string) {
	if cv.attrs == nil {
		cv.attrs = make(map[string]string)
	}
	cv.attrs[name] = value

	// Set corresponding flags
	// İlgili bayrakları ayarla
	switch name {
	case "lvalue":
		cv.flags |= CVLvalue
	case "method":
		cv.flags |= CVMethod
	case "locked":
		cv.flags |= CVLocked
	}
}

// GetAttr returns an attribute value.
// GetAttr, bir özellik değeri döndürür.
func (cv *CV) GetAttr(name string) (string, bool) {
	if cv.attrs == nil {
		return "", false
	}
	v, ok := cv.attrs[name]
	return v, ok
}

// IsLvalue returns true if :lvalue attribute set.
// IsLvalue, :lvalue özelliği ayarlıysa true döndürür.
func (cv *CV) IsLvalue() bool {
	return cv.flags&CVLvalue != 0
}

// IsMethod returns true if :method attribute set.
// IsMethod, :method özelliği ayarlıysa true döndürür.
func (cv *CV) IsMethod() bool {
	return cv.flags&CVMethod != 0
}

// ============================================================
// Lexical Pad
// Leksikal Pad
// ============================================================

// PadNames returns the names of lexical variables.
// PadNames, leksikal değişken isimlerini döndürür.
func (cv *CV) PadNames() []string {
	return cv.padnames
}

// SetPadNames sets the lexical variable names.
// SetPadNames, leksikal değişken isimlerini ayarlar.
func (cv *CV) SetPadNames(names []string) {
	cv.padnames = names
}

// AddPadName adds a lexical variable name, returns its index.
// AddPadName, bir leksikal değişken adı ekler, indeksini döndürür.
func (cv *CV) AddPadName(name string) int {
	idx := len(cv.padnames)
	cv.padnames = append(cv.padnames, name)
	return idx
}

// PadIndex returns the index of a lexical variable by name (-1 if not found).
// PadIndex, isme göre leksikal değişkenin indeksini döndürür (bulunamazsa -1).
func (cv *CV) PadIndex(name string) int {
	for i, n := range cv.padnames {
		if n == name {
			return i
		}
	}
	return -1
}

// ============================================================
// Constant Subroutines
// Sabit Altyordamlar
// ============================================================

// NewConst creates a constant subroutine that always returns the same value.
// Used for: use constant FOO => 42;
//
// NewConst, her zaman aynı değeri döndüren sabit bir altyordam oluşturur.
// Kullanım: use constant FOO => 42;
func NewConst(pkg, name string, value *sv.SV) *CV {
	if value != nil {
		value.IncRef()
	}
	return &CV{
		name:  name,
		pkg:   pkg,
		flags: CVConst,
		native: func(ctx *CallContext) *sv.SV {
			return value
		},
	}
}

// IsConst returns true if this is a constant subroutine.
// IsConst, bu sabit bir altyordam ise true döndürür.
func (cv *CV) IsConst() bool {
	return cv.flags&CVConst != 0
}

// ============================================================
// XSUB support (for built-in functions)
// XSUB desteği (yerleşik fonksiyonlar için)
// ============================================================

// NewXSUB creates a CV for an external/built-in function.
// NewXSUB, harici/yerleşik bir fonksiyon için CV oluşturur.
func NewXSUB(pkg, name string, fn func(*CallContext) *sv.SV) *CV {
	return &CV{
		name:   name,
		pkg:    pkg,
		native: fn,
		flags:  CVIsXSUB,
	}
}

// IsXSUB returns true if this is an XSUB.
// IsXSUB, bu bir XSUB ise true döndürür.
func (cv *CV) IsXSUB() bool {
	return cv.flags&CVIsXSUB != 0
}

// ============================================================
// Cleanup
// Temizlik
// ============================================================

// Free releases all resources.
// Free, tüm kaynakları serbest bırakır.
func (cv *CV) Free() {
	for _, v := range cv.pad {
		if v != nil {
			v.DecRef()
		}
	}
	cv.pad = nil
	cv.outer = nil
	cv.ops = nil
	cv.native = nil
}

// ============================================================
// Context Helpers
// Bağlam Yardımcıları
// ============================================================

// WantArray returns context from CallContext (-1=void, 0=scalar, 1=list).
// WantArray, CallContext'ten bağlamı döndürür (-1=void, 0=skaler, 1=liste).
func (ctx *CallContext) WantArrayVal() int {
	if ctx == nil {
		return 0
	}
	return ctx.WantArray
}

// Arg returns argument at index, or undef if out of bounds.
// Arg, indeksteki argümanı döndürür, sınır dışıysa undef.
func (ctx *CallContext) Arg(i int) *sv.SV {
	if ctx == nil || i < 0 || i >= len(ctx.Args) {
		return sv.NewUndef()
	}
	return ctx.Args[i]
}

// NumArgs returns number of arguments.
// NumArgs, argüman sayısını döndürür.
func (ctx *CallContext) NumArgs() int {
	if ctx == nil {
		return 0
	}
	return len(ctx.Args)
}

// SetPad sets a lexical variable value.
// SetPad, bir leksikal değişken değeri ayarlar.
func (ctx *CallContext) SetPad(idx int, val *sv.SV) {
	if ctx == nil {
		return
	}
	if idx >= len(ctx.Pad) {
		newPad := make([]*sv.SV, idx+1)
		copy(newPad, ctx.Pad)
		ctx.Pad = newPad
	}
	if ctx.Pad[idx] != nil {
		ctx.Pad[idx].DecRef()
	}
	if val != nil {
		val.IncRef()
	}
	ctx.Pad[idx] = val
}

// GetPad returns a lexical variable value.
// GetPad, bir leksikal değişken değeri döndürür.
func (ctx *CallContext) GetPad(idx int) *sv.SV {
	if ctx == nil || idx < 0 || idx >= len(ctx.Pad) {
		return sv.NewUndef()
	}
	if ctx.Pad[idx] == nil {
		return sv.NewUndef()
	}
	return ctx.Pad[idx]
}

// CallerInfo returns (package, file, line) of caller.
// CallerInfo, çağıranın (paket, dosya, satır) bilgisini döndürür.
func (ctx *CallContext) CallerInfo() (string, string, int) {
	if ctx == nil || ctx.Caller == nil {
		return "", "", 0
	}
	return ctx.Caller.Package, ctx.Caller.File, ctx.Caller.Line
}
