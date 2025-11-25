package stash

import (
	"testing"

	"perlc/pkg/sv"
)

// ============================================================
// Basic Access Tests
// Temel Erişim Testleri
// ============================================================

// TestBasicAccess tests symbol storage and retrieval.
// TestBasicAccess, sembol depolama ve almayı test eder.
func TestBasicAccess(t *testing.T) {
	s := Get("BasicTest")

	s.SetScalar("foo", sv.NewInt(42))

	if s.Scalar("foo").AsInt() != 42 {
		t.Error("$BasicTest::foo should be 42")
	}
}

// TestMultipleTypes tests storing different types under same name.
// TestMultipleTypes, aynı isim altında farklı türleri saklamayı test eder.
func TestMultipleTypes(t *testing.T) {
	s := Get("MultiType")

	// $foo, @foo, %foo can coexist
	// $foo, @foo, %foo bir arada var olabilir
	s.SetScalar("foo", sv.NewString("scalar"))

	arr := s.Array("foo")
	if arr == nil || !arr.IsArray() {
		t.Error("@foo should be array")
	}

	hash := s.Hash("foo")
	if hash == nil || !hash.IsHash() {
		t.Error("hash slot should be hash type")
	}

	// Scalar should still be accessible
	// Skaler hâlâ erişilebilir olmalı
	if s.Scalar("foo").AsString() != "scalar" {
		t.Error("$foo should still be 'scalar'")
	}
}

// TestSetAndGet tests all set/get combinations.
// TestSetAndGet, tüm set/get kombinasyonlarını test eder.
func TestSetAndGet(t *testing.T) {
	s := Get("SetGetTest")

	// Scalar
	s.SetScalar("s", sv.NewInt(1))
	if s.Scalar("s").AsInt() != 1 {
		t.Error("Scalar set/get failed")
	}

	// Array
	arr := sv.NewArrayRef(sv.NewInt(1), sv.NewInt(2)).Deref()
	s.SetArray("a", arr)
	if s.Array("a").ArrayData() == nil {
		t.Error("Array set/get failed")
	}

	// Hash
	hash := sv.NewHashRef().Deref()
	s.SetHash("h", hash)
	if s.Hash("h").HashData() == nil {
		t.Error("Hash set/get failed")
	}

	// Code
	code := sv.NewInt(999) // Placeholder for CV
	s.SetCode("c", code)
	if s.Code("c").AsInt() != 999 {
		t.Error("Code set/get failed")
	}
}

// ============================================================
// Package Hierarchy Tests
// Paket Hiyerarşisi Testleri
// ============================================================

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

// TestDeeplyNested tests very deep nesting.
// TestDeeplyNested, çok derin iç içeliği test eder.
func TestDeeplyNested(t *testing.T) {
	s := Get("A::B::C::D::E::F")

	if s.Name() != "A::B::C::D::E::F" {
		t.Error("Deep package name wrong")
	}

	// All parents should exist
	// Tüm üst paketler var olmalı
	parents := []string{"A", "A::B", "A::B::C", "A::B::C::D", "A::B::C::D::E"}
	for _, p := range parents {
		if !Exists(p) {
			t.Errorf("Parent %s should exist", p)
		}
	}
}

// TestMainPackage tests the main package.
// TestMainPackage, main paketini test eder.
func TestMainPackage(t *testing.T) {
	main := Get("main")

	if main.Name() != "main" {
		t.Error("main package name wrong")
	}

	// main should always exist
	// main her zaman var olmalı
	if !Exists("main") {
		t.Error("main should always exist")
	}
}

// TestEmptyPackageName tests empty string defaults to main.
// TestEmptyPackageName, boş string'in main'e varsayıldığını test eder.
func TestEmptyPackageName(t *testing.T) {
	s := Get("")

	if s.Name() != "main" {
		t.Error("Empty package name should default to main")
	}
}

// TestSubStashes tests getting nested package list.
// TestSubStashes, iç içe paket listesini almayı test eder.
func TestSubStashes(t *testing.T) {
	parent := Get("SubTest")
	Get("SubTest::Child1")
	Get("SubTest::Child2")
	Get("SubTest::Child3")

	subs := parent.SubStashes()
	if len(subs) < 3 {
		t.Errorf("Should have at least 3 sub-stashes, got %d", len(subs))
	}
}

// ============================================================
// Symbolic Reference Tests
// Sembolik Referans Testleri
// ============================================================

// TestResolve tests symbolic reference resolution.
// TestResolve, sembolik referans çözümlemeyi test eder.
func TestResolve(t *testing.T) {
	Get("MyMod").SetScalar("version", sv.NewString("1.0"))

	val := ResolveScalar("MyMod::version")
	if val.AsString() != "1.0" {
		t.Errorf("Expected '1.0', got '%s'", val.AsString())
	}
}

// TestResolveAllTypes tests resolving all sigil types.
// TestResolveAllTypes, tüm sigil türlerini çözümlemeyi test eder.
func TestResolveAllTypes(t *testing.T) {
	pkg := Get("ResolveTypes")

	pkg.SetScalar("var", sv.NewInt(42))

	// $ResolveTypes::var
	if ResolveScalar("ResolveTypes::var").AsInt() != 42 {
		t.Error("Scalar resolve failed")
	}

	// @ResolveTypes::arr
	arr := ResolveArray("ResolveTypes::arr")
	if arr == nil || !arr.IsArray() {
		t.Error("Array resolve failed")
	}

	// %ResolveTypes::hash
	hash := ResolveHash("ResolveTypes::hash")
	if hash == nil || !hash.IsHash() {
		t.Error("Hash resolve failed")
	}

	// &ResolveTypes::code
	code := ResolveCode("ResolveTypes::code")
	// Code might be nil, but shouldn't panic
	_ = code
}

// TestResolveMainPackage tests resolving without package prefix.
// TestResolveMainPackage, paket öneki olmadan çözümlemeyi test eder.
func TestResolveMainPackage(t *testing.T) {
	Get("main").SetScalar("mainvar", sv.NewInt(123))

	// "mainvar" should resolve to main::mainvar
	// "mainvar" main::mainvar olarak çözümlenmeli
	gv := Resolve("mainvar")
	if gv.Scalar().AsInt() != 123 {
		t.Error("Should resolve to main:: by default")
	}
}

// TestResolveCreatesGV tests that Resolve creates GV if not exists.
// TestResolveCreatesGV, Resolve'ın yoksa GV oluşturduğunu test eder.
func TestResolveCreatesGV(t *testing.T) {
	gv := Resolve("NewPkg::newvar")

	if gv == nil {
		t.Error("Resolve should create GV")
	}
	if gv.Name() != "newvar" {
		t.Errorf("GV name should be 'newvar', got '%s'", gv.Name())
	}
}

// ============================================================
// Inheritance Tests (@ISA)
// Kalıtım Testleri (@ISA)
// ============================================================

// TestInheritance tests @ISA and method resolution.
// TestInheritance, @ISA ve metod çözümlemeyi test eder.
func TestInheritance(t *testing.T) {
	parent := Get("Parent")
	parent.SetCode("greet", sv.NewInt(1))

	child := Get("Child")
	child.AddISA("Parent")

	code, pkg := child.FindMethod("greet")
	if code == nil {
		t.Error("Should find greet method")
	}
	if pkg != "Parent" {
		t.Errorf("Method should be from Parent, got %s", pkg)
	}
}

// TestMultipleInheritance tests multiple parent classes.
// TestMultipleInheritance, birden fazla üst sınıfı test eder.
func TestMultipleInheritance(t *testing.T) {
	Get("MixinA").SetCode("method_a", sv.NewInt(1))
	Get("MixinB").SetCode("method_b", sv.NewInt(2))

	child := Get("MultiChild")
	child.AddISA("MixinA")
	child.AddISA("MixinB")

	// Should find both methods
	// Her iki metodu da bulmalı
	codeA, _ := child.FindMethod("method_a")
	codeB, _ := child.FindMethod("method_b")

	if codeA == nil {
		t.Error("Should find method_a from MixinA")
	}
	if codeB == nil {
		t.Error("Should find method_b from MixinB")
	}
}

// TestInheritanceOrder tests left-to-right, depth-first search.
// TestInheritanceOrder, soldan sağa, derinlik öncelikli aramayı test eder.
func TestInheritanceOrder(t *testing.T) {
	// Diamond inheritance: D -> B, C -> A
	//      A
	//     / \
	//    B   C
	//     \ /
	//      D
	Get("DiamondA").SetCode("method", sv.NewInt(100))

	bPkg := Get("DiamondB")
	bPkg.AddISA("DiamondA")
	bPkg.SetCode("method", sv.NewInt(200)) // Override

	cPkg := Get("DiamondC")
	cPkg.AddISA("DiamondA")

	dPkg := Get("DiamondD")
	dPkg.AddISA("DiamondB")
	dPkg.AddISA("DiamondC")

	// Should find B's method first (left-to-right)
	// B'nin metodunu önce bulmalı (soldan sağa)
	code, pkg := dPkg.FindMethod("method")
	if pkg != "DiamondB" {
		t.Errorf("Should find method in DiamondB first, got %s", pkg)
	}
	if code.AsInt() != 200 {
		t.Error("Should get overridden value")
	}
}

// TestDeepInheritance tests deep inheritance chain.
// TestDeepInheritance, derin kalıtım zincirini test eder.
func TestDeepInheritance(t *testing.T) {
	// Chain: E -> D -> C -> B -> A
	Get("ChainA").SetCode("root_method", sv.NewInt(1))
	Get("ChainB").AddISA("ChainA")
	Get("ChainC").AddISA("ChainB")
	Get("ChainD").AddISA("ChainC")
	Get("ChainE").AddISA("ChainD")

	code, pkg := Get("ChainE").FindMethod("root_method")
	if code == nil {
		t.Error("Should find method through deep chain")
	}
	if pkg != "ChainA" {
		t.Errorf("Method should be from ChainA, got %s", pkg)
	}
}

// TestIsaCheck tests isa() method.
// TestIsaCheck, isa() metodunu test eder.
func TestIsaCheck(t *testing.T) {
	Get("IsaParent")
	child := Get("IsaChild")
	child.AddISA("IsaParent")

	if !child.Isa("IsaParent") {
		t.Error("Child should isa Parent")
	}
	if !child.Isa("IsaChild") {
		t.Error("Child should isa itself")
	}
	if !child.Isa("UNIVERSAL") {
		t.Error("Everything should isa UNIVERSAL")
	}
	if child.Isa("NonExistent") {
		t.Error("Should not isa NonExistent")
	}
}

// TestIsaSelf tests that class isa itself.
// TestIsaSelf, sınıfın kendisi olduğunu test eder.
func TestIsaSelf(t *testing.T) {
	s := Get("SelfTest")
	if !s.Isa("SelfTest") {
		t.Error("Class should isa itself")
	}
}

// TestCircularInheritance tests protection against circular @ISA.
// TestCircularInheritance, döngüsel @ISA'ya karşı korumayı test eder.
func TestCircularInheritance(t *testing.T) {
	// Create circular: A -> B -> A
	aPkg := Get("CircularA")
	bPkg := Get("CircularB")

	aPkg.AddISA("CircularB")
	bPkg.AddISA("CircularA")

	// Should not infinite loop - visited map prevents this
	// Sonsuz döngü olmamalı - visited map bunu önler
	code, _ := aPkg.FindMethod("nonexistent")
	if code != nil {
		t.Error("Should not find nonexistent method")
	}

	// Same for Isa
	// Isa için de aynı
	result := aPkg.Isa("CircularA")
	if !result {
		t.Error("Circular A should still isa itself")
	}
}

// ============================================================
// Can Tests
// Can Testleri
// ============================================================

// TestCan tests method existence check.
// TestCan, metod varlık kontrolünü test eder.
func TestCan(t *testing.T) {
	s := Get("CanTest2")
	s.SetCode("exists", sv.NewInt(1))

	if !s.Can("exists") {
		t.Error("Should have 'exists' method")
	}
	if s.Can("missing") {
		t.Error("Should not have 'missing' method")
	}
}

// TestCanInherited tests can() with inherited methods.
// TestCanInherited, kalıtılmış metodlarla can()'i test eder.
func TestCanInherited(t *testing.T) {
	Get("CanParent").SetCode("parent_method", sv.NewInt(1))

	child := Get("CanChild")
	child.AddISA("CanParent")

	if !child.Can("parent_method") {
		t.Error("Should find inherited method with can()")
	}
}

// ============================================================
// UNIVERSAL Tests
// UNIVERSAL Testleri
// ============================================================

// TestUniversalMethod tests UNIVERSAL method resolution.
// TestUniversalMethod, UNIVERSAL metod çözümlemeyi test eder.
func TestUniversalMethod(t *testing.T) {
	Get("UNIVERSAL").SetCode("universal_method", sv.NewInt(1))

	// Any class should find UNIVERSAL methods
	// Herhangi bir sınıf UNIVERSAL metodları bulmalı
	randomPkg := Get("RandomPackage")
	code, pkg := randomPkg.FindMethod("universal_method")

	if code == nil {
		t.Error("Should find UNIVERSAL method")
	}
	if pkg != "UNIVERSAL" {
		t.Errorf("Method should be from UNIVERSAL, got %s", pkg)
	}
}

// ============================================================
// AUTOLOAD Tests
// AUTOLOAD Testleri
// ============================================================

// TestFindAutoload tests AUTOLOAD resolution.
// TestFindAutoload, AUTOLOAD çözümlemeyi test eder.
func TestFindAutoload(t *testing.T) {
	Get("AutoParent").SetCode("AUTOLOAD", sv.NewInt(1))

	child := Get("AutoChild")
	child.AddISA("AutoParent")

	code, pkg := child.FindAutoload()
	if code == nil {
		t.Error("Should find AUTOLOAD")
	}
	if pkg != "AutoParent" {
		t.Errorf("AUTOLOAD should be from AutoParent, got %s", pkg)
	}
}

// TestNoAutoload tests when no AUTOLOAD exists.
// TestNoAutoload, AUTOLOAD yokken test eder.
func TestNoAutoload(t *testing.T) {
	s := Get("NoAutoload")
	code, _ := s.FindAutoload()

	if code != nil {
		t.Error("Should not find AUTOLOAD when none defined")
	}
}

// ============================================================
// Symbol Enumeration Tests
// Sembol Numaralandırma Testleri
// ============================================================

// TestSymbols tests symbol enumeration.
// TestSymbols, sembol numaralandırmayı test eder.
func TestSymbols(t *testing.T) {
	s := Get("EnumTest2")
	s.SetScalar("a", sv.NewInt(1))
	s.SetScalar("b", sv.NewInt(2))
	s.SetCode("func", sv.NewInt(3))

	symbols := s.Symbols()
	if len(symbols) < 3 {
		t.Errorf("Should have at least 3 symbols, got %d", len(symbols))
	}

	// Check specific symbols exist
	// Belirli sembollerin varlığını kontrol et
	found := make(map[string]bool)
	for _, sym := range symbols {
		found[sym] = true
	}

	if !found["a"] || !found["b"] || !found["func"] {
		t.Error("Should find all added symbols")
	}
}

// ============================================================
// GV Operations Tests
// GV İşlem Testleri
// ============================================================

// TestFetchGV tests GV creation and retrieval.
// TestFetchGV, GV oluşturma ve almayı test eder.
func TestFetchGV(t *testing.T) {
	s := Get("GVTest")

	gv1 := s.FetchGV("test")
	gv2 := s.FetchGV("test")

	// Should return same GV
	// Aynı GV döndürmeli
	if gv1 != gv2 {
		t.Error("FetchGV should return same GV for same name")
	}
}

// TestLookupGV tests GV lookup without creation.
// TestLookupGV, oluşturmadan GV aramayı test eder.
func TestLookupGV(t *testing.T) {
	s := Get("LookupTest")

	// Before creation
	// Oluşturmadan önce
	if s.LookupGV("nonexistent") != nil {
		t.Error("LookupGV should return nil for nonexistent")
	}

	// After creation
	// Oluşturduktan sonra
	s.FetchGV("created")
	if s.LookupGV("created") == nil {
		t.Error("LookupGV should find created GV")
	}
}

// TestDeleteGV tests GV deletion.
// TestDeleteGV, GV silmeyi test eder.
func TestDeleteGV(t *testing.T) {
	s := Get("DeleteTest")
	s.SetScalar("todelete", sv.NewInt(42))

	// Verify exists
	// Varlığını doğrula
	if s.LookupGV("todelete") == nil {
		t.Error("Should exist before delete")
	}

	s.DeleteGV("todelete")

	// Should be gone
	// Gitmiş olmalı
	if s.LookupGV("todelete") != nil {
		t.Error("Should not exist after delete")
	}
}

// ============================================================
// VERSION Tests
// VERSION Testleri
// ============================================================

// TestVERSION tests $VERSION retrieval.
// TestVERSION, $VERSION almayı test eder.
func TestVERSION(t *testing.T) {
	s := Get("VersionTest")
	s.SetScalar("VERSION", sv.NewString("2.0.0"))

	ver := s.VERSION()
	if ver.AsString() != "2.0.0" {
		t.Errorf("VERSION should be '2.0.0', got '%s'", ver.AsString())
	}
}

// TestVERSIONUndef tests undefined $VERSION.
// TestVERSIONUndef, tanımsız $VERSION'ı test eder.
func TestVERSIONUndef(t *testing.T) {
	s := Get("NoVersionTest")

	ver := s.VERSION()
	if !ver.IsUndef() {
		t.Error("VERSION should be undef when not set")
	}
}

// ============================================================
// Thread Safety Tests
// İş Parçacığı Güvenliği Testleri
// ============================================================

// TestConcurrentAccess tests concurrent stash access.
// TestConcurrentAccess, eşzamanlı stash erişimini test eder.
func TestConcurrentAccess(t *testing.T) {
	done := make(chan bool)

	// Concurrent reads and writes
	// Eşzamanlı okuma ve yazma
	for i := 0; i < 10; i++ {
		go func(n int) {
			s := Get("ConcurrentTest")
			s.SetScalar("var", sv.NewInt(int64(n)))
			_ = s.Scalar("var")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	// Tüm goroutine'leri bekle
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestConcurrentPackageCreation tests concurrent package creation.
// TestConcurrentPackageCreation, eşzamanlı paket oluşturmayı test eder.
func TestConcurrentPackageCreation(t *testing.T) {
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			// All trying to create same nested package
			// Hepsi aynı iç içe paketi oluşturmaya çalışıyor
			s := Get("Concurrent::Nested::Package")
			_ = s.Name()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have created successfully
	// Başarıyla oluşturulmuş olmalı
	if !Exists("Concurrent::Nested::Package") {
		t.Error("Concurrent creation should succeed")
	}
}

// ============================================================
// All Stashes Tests
// Tüm Stash'ler Testleri
// ============================================================

// TestAll tests listing all stashes.
// TestAll, tüm stash'leri listelemeyi test eder.
func TestAll(t *testing.T) {
	// Create some packages
	// Bazı paketler oluştur
	Get("AllTest1")
	Get("AllTest2")

	all := All()
	if len(all) < 3 { // At least main + our 2
		t.Errorf("Should have at least 3 stashes, got %d", len(all))
	}

	// Check main exists
	// main'in var olduğunu kontrol et
	foundMain := false
	for _, name := range all {
		if name == "main" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Error("All() should include 'main'")
	}
}

// ============================================================
// ISA Management Tests
// ISA Yönetimi Testleri
// ============================================================

// TestSetISA tests replacing @ISA entirely.
// TestSetISA, @ISA'yı tamamen değiştirmeyi test eder.
func TestSetISA(t *testing.T) {
	s := Get("SetISATest")
	s.AddISA("Parent1")
	s.AddISA("Parent2")

	// Replace with new list
	// Yeni liste ile değiştir
	s.SetISA([]*sv.SV{sv.NewString("NewParent")})

	isa := s.ISA()
	if len(isa) != 1 {
		t.Errorf("Should have 1 parent, got %d", len(isa))
	}
	if isa[0].AsString() != "NewParent" {
		t.Error("Parent should be NewParent")
	}
}

// TestEmptyISA tests class with no parents.
// TestEmptyISA, üst sınıfsız sınıfı test eder.
func TestEmptyISA(t *testing.T) {
	s := Get("EmptyISATest")

	isa := s.ISA()
	if len(isa) != 0 {
		t.Error("New class should have empty @ISA")
	}

	// Should still isa UNIVERSAL
	// Yine de UNIVERSAL olmalı
	if !s.Isa("UNIVERSAL") {
		t.Error("Even with empty @ISA, should isa UNIVERSAL")
	}
}
