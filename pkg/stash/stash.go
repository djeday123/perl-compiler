// Package stash implements Perl's symbol tables (stashes).
// Paket stash, Perl'in sembol tablolarını (stash'leri) uygular.
//
// A stash is a hash that maps names to globs. Each package has one stash.
// The main:: stash contains all top-level symbols and nested package stashes.
//
// Stash, isimleri globlara eşleyen bir hash'tir. Her paketin bir stash'i vardır.
// main:: stash'i tüm üst düzey sembolleri ve iç içe paket stash'lerini içerir.
package stash

import (
	"strings"
	"sync"

	"perlc/pkg/gv"
	"perlc/pkg/sv"
)

// Stash represents a package's symbol table.
// Stash, bir paketin sembol tablosunu temsil eder.
type Stash struct {
	name    string            // Package name / Paket adı
	symbols map[string]*gv.GV // name -> glob mapping / isim -> glob eşlemesi
	isa     []*sv.SV          // @ISA for inheritance / Kalıtım için @ISA
	mu      sync.RWMutex      // Thread safety / İş parçacığı güvenliği
}

// Global stash registry - all packages.
// Global stash kaydı - tüm paketler.
var (
	stashes   = make(map[string]*Stash)
	stashesMu sync.RWMutex
)

// init creates the main:: stash.
// init, main:: stash'ini oluşturur.
func init() {
	stashes["main"] = &Stash{
		name:    "main",
		symbols: make(map[string]*gv.GV),
	}
}

// ============================================================
// Stash Registry
// Stash Kaydı
// ============================================================

// Get returns stash for package name (creates if not exists).
// Get, paket adı için stash döndürür (yoksa oluşturur).
func Get(pkgName string) *Stash {
	if pkgName == "" {
		pkgName = "main"
	}

	stashesMu.RLock()
	s, ok := stashes[pkgName]
	stashesMu.RUnlock()

	if ok {
		return s
	}

	// Create new stash
	// Yeni stash oluştur
	stashesMu.Lock()

	// Double-check after acquiring write lock
	// Yazma kilidi aldıktan sonra tekrar kontrol et
	if s, ok = stashes[pkgName]; ok {
		stashesMu.Unlock()
		return s
	}

	s = &Stash{
		name:    pkgName,
		symbols: make(map[string]*gv.GV),
	}
	stashes[pkgName] = s
	stashesMu.Unlock() // Release lock BEFORE recursive call / Özyinelemeli çağrıdan ÖNCE kilidi bırak

	// Register in parent stash (outside of lock to avoid deadlock)
	// Üst stash'e kaydet (deadlock'u önlemek için kilidin dışında)
	if idx := strings.LastIndex(pkgName, "::"); idx > 0 {
		parent := pkgName[:idx]
		child := pkgName[idx+2:] + "::"
		Get(parent).FetchGV(child)
	} else if pkgName != "main" {
		Get("main").FetchGV(pkgName + "::")
	}

	return s
}

// Exists checks if stash exists without creating it.
// Exists, stash'in oluşturmadan var olup olmadığını kontrol eder.
func Exists(pkgName string) bool {
	stashesMu.RLock()
	defer stashesMu.RUnlock()
	_, ok := stashes[pkgName]
	return ok
}

// All returns all registered stash names.
// All, tüm kayıtlı stash isimlerini döndürür.
func All() []string {
	stashesMu.RLock()
	defer stashesMu.RUnlock()

	names := make([]string, 0, len(stashes))
	for name := range stashes {
		names = append(names, name)
	}
	return names
}

// ============================================================
// Symbol Access
// Sembol Erişimi
// ============================================================

// FetchGV returns glob for name (creates if not exists).
// FetchGV, isim için glob döndürür (yoksa oluşturur).
func (s *Stash) FetchGV(name string) *gv.GV {
	s.mu.RLock()
	g, ok := s.symbols[name]
	s.mu.RUnlock()

	if ok {
		return g
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check
	// Tekrar kontrol
	if g, ok = s.symbols[name]; ok {
		return g
	}

	g = gv.New(s.name, name)
	s.symbols[name] = g
	return g
}

// LookupGV returns glob for name (returns nil if not exists).
// LookupGV, isim için glob döndürür (yoksa nil döndürür).
func (s *Stash) LookupGV(name string) *gv.GV {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.symbols[name]
}

// DeleteGV removes a glob from the stash.
// DeleteGV, stash'ten bir glob kaldırır.
func (s *Stash) DeleteGV(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if g, ok := s.symbols[name]; ok {
		g.Free()
		delete(s.symbols, name)
	}
}

// ============================================================
// Shortcut accessors for common operations
// Yaygın işlemler için kısayol erişimcileri
// ============================================================

// Scalar returns $Package::name.
// Scalar, $Package::name döndürür.
func (s *Stash) Scalar(name string) *sv.SV {
	return s.FetchGV(name).Scalar()
}

// SetScalar sets $Package::name.
// SetScalar, $Package::name'i ayarlar.
func (s *Stash) SetScalar(name string, val *sv.SV) {
	s.FetchGV(name).SetScalar(val)
}

// Array returns @Package::name.
// Array, @Package::name döndürür.
func (s *Stash) Array(name string) *sv.SV {
	return s.FetchGV(name).Array()
}

// SetArray sets @Package::name.
// SetArray, @Package::name'i ayarlar.
func (s *Stash) SetArray(name string, val *sv.SV) {
	s.FetchGV(name).SetArray(val)
}

// Hash returns %Package::name.
// Hash, %Package::name döndürür.
func (s *Stash) Hash(name string) *sv.SV {
	return s.FetchGV(name).Hash()
}

// SetHash sets %Package::name.
// SetHash, %Package::name'i ayarlar.
func (s *Stash) SetHash(name string, val *sv.SV) {
	s.FetchGV(name).SetHash(val)
}

// Code returns &Package::name.
// Code, &Package::name döndürür.
func (s *Stash) Code(name string) *sv.SV {
	return s.FetchGV(name).Code()
}

// SetCode sets &Package::name.
// SetCode, &Package::name'i ayarlar.
func (s *Stash) SetCode(name string, val *sv.SV) {
	s.FetchGV(name).SetCode(val)
}

// ============================================================
// Symbol enumeration
// Sembol numaralandırma
// ============================================================

// Symbols returns all symbol names in this stash.
// Symbols, bu stash'teki tüm sembol isimlerini döndürür.
func (s *Stash) Symbols() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.symbols))
	for name := range s.symbols {
		names = append(names, name)
	}
	return names
}

// SubStashes returns names of nested package stashes (entries ending with "::").
// SubStashes, iç içe paket stash'lerinin isimlerini döndürür ("::" ile bitenler).
func (s *Stash) SubStashes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []string
	for name := range s.symbols {
		if strings.HasSuffix(name, "::") {
			result = append(result, name[:len(name)-2])
		}
	}
	return result
}

// ============================================================
// @ISA and Inheritance
// @ISA ve Kalıtım
// ============================================================

// ISA returns the @ISA array for inheritance.
// ISA, kalıtım için @ISA dizisini döndürür.
func (s *Stash) ISA() []*sv.SV {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isa
}

// SetISA sets the @ISA array.
// SetISA, @ISA dizisini ayarlar.
func (s *Stash) SetISA(parents []*sv.SV) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isa = parents
}

// AddISA adds a parent class to @ISA.
// AddISA, @ISA'ya bir üst sınıf ekler.
func (s *Stash) AddISA(parent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isa = append(s.isa, sv.NewString(parent))
}

// ============================================================
// Method Resolution (for OOP)
// Metod Çözümleme (OOP için)
// ============================================================

// FindMethod searches for a method in this class and its @ISA hierarchy.
// Returns the code ref and the package where it was found.
//
// FindMethod, bu sınıfta ve @ISA hiyerarşisinde bir metod arar.
// Kod referansını ve bulunduğu paketi döndürür.
func (s *Stash) FindMethod(name string) (*sv.SV, string) {
	// Check this package first
	// Önce bu paketi kontrol et
	if g := s.LookupGV(name); g != nil && g.HasCode() {
		return g.Code(), s.name
	}

	// Search @ISA recursively (depth-first, left-to-right)
	// @ISA'yı özyinelemeli ara (önce derinlik, soldan sağa)
	visited := make(map[string]bool)
	return s.findMethodRecursive(name, visited)
}

func (s *Stash) findMethodRecursive(name string, visited map[string]bool) (*sv.SV, string) {
	if visited[s.name] {
		return nil, ""
	}
	visited[s.name] = true

	// Check this package
	// Bu paketi kontrol et
	if g := s.LookupGV(name); g != nil && g.HasCode() {
		return g.Code(), s.name
	}

	// Search parents
	// Üst sınıfları ara
	s.mu.RLock()
	parents := s.isa
	s.mu.RUnlock()

	for _, parentSV := range parents {
		parentName := parentSV.AsString()
		parentStash := Get(parentName)
		if code, pkg := parentStash.findMethodRecursive(name, visited); code != nil {
			return code, pkg
		}
	}

	// Try UNIVERSAL as last resort
	// Son çare olarak UNIVERSAL'ı dene
	if s.name != "UNIVERSAL" {
		universal := Get("UNIVERSAL")
		if g := universal.LookupGV(name); g != nil && g.HasCode() {
			return g.Code(), "UNIVERSAL"
		}
	}

	return nil, ""
}

// Can checks if this class or its parents have a method.
// Can, bu sınıfın veya üst sınıflarının bir metodu olup olmadığını kontrol eder.
func (s *Stash) Can(name string) bool {
	code, _ := s.FindMethod(name)
	return code != nil
}

// Isa checks if this class inherits from another.
// Isa, bu sınıfın başka bir sınıftan miras alıp almadığını kontrol eder.
func (s *Stash) Isa(parent string) bool {
	if s.name == parent {
		return true
	}

	visited := make(map[string]bool)
	return s.isaRecursive(parent, visited)
}

func (s *Stash) isaRecursive(target string, visited map[string]bool) bool {
	if visited[s.name] {
		return false
	}
	visited[s.name] = true

	if s.name == target {
		return true
	}

	s.mu.RLock()
	parents := s.isa
	s.mu.RUnlock()

	for _, parentSV := range parents {
		parentName := parentSV.AsString()
		if parentName == target {
			return true
		}
		if Get(parentName).isaRecursive(target, visited) {
			return true
		}
	}

	// Everything isa UNIVERSAL
	// Her şey UNIVERSAL'dır
	return target == "UNIVERSAL"
}

// ============================================================
// AUTOLOAD Support
// AUTOLOAD Desteği
// ============================================================

// FindAutoload searches for AUTOLOAD in the class hierarchy.
// Returns the AUTOLOAD code ref and the package where it was found.
//
// FindAutoload, sınıf hiyerarşisinde AUTOLOAD arar.
// AUTOLOAD kod referansını ve bulunduğu paketi döndürür.
func (s *Stash) FindAutoload() (*sv.SV, string) {
	return s.FindMethod("AUTOLOAD")
}

// ============================================================
// Symbolic References
// Sembolik Referanslar
// ============================================================

// Resolve resolves a fully qualified name like "Foo::Bar::baz".
// Returns the glob for the symbol.
//
// Resolve, "Foo::Bar::baz" gibi tam nitelikli bir adı çözümler.
// Sembol için glob döndürür.
func Resolve(fullName string) *gv.GV {
	// Split into package and name
	// Paket ve isme böl
	pkg := "main"
	name := fullName

	if idx := strings.LastIndex(fullName, "::"); idx >= 0 {
		pkg = fullName[:idx]
		name = fullName[idx+2:]
	}

	return Get(pkg).FetchGV(name)
}

// ResolveScalar resolves $$varname (symbolic scalar reference).
// ResolveScalar, $$varname'i çözümler (sembolik skaler referans).
func ResolveScalar(fullName string) *sv.SV {
	return Resolve(fullName).Scalar()
}

// ResolveArray resolves @$varname (symbolic array reference).
// ResolveArray, @$varname'i çözümler (sembolik dizi referansı).
func ResolveArray(fullName string) *sv.SV {
	return Resolve(fullName).Array()
}

// ResolveHash resolves %$varname (symbolic hash reference).
// ResolveHash, %$varname'i çözümler (sembolik hash referansı).
func ResolveHash(fullName string) *sv.SV {
	return Resolve(fullName).Hash()
}

// ResolveCode resolves &$varname (symbolic code reference).
// ResolveCode, &$varname'i çözümler (sembolik kod referansı).
func ResolveCode(fullName string) *sv.SV {
	return Resolve(fullName).Code()
}

// ============================================================
// Package Information
// Paket Bilgisi
// ============================================================

// Name returns the package name.
// Name, paket adını döndürür.
func (s *Stash) Name() string {
	return s.name
}

// VERSION returns the $VERSION scalar if defined.
// VERSION, tanımlıysa $VERSION skalerini döndürür.
func (s *Stash) VERSION() *sv.SV {
	if g := s.LookupGV("VERSION"); g != nil {
		return g.Scalar()
	}
	return sv.NewUndef()
}
