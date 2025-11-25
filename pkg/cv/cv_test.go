package cv

import (
	"testing"

	"perlc/pkg/sv"
)

// ============================================================
// Basic Call Tests
// Temel Çağrı Testleri
// ============================================================

// TestBasicCall tests simple subroutine call.
// TestBasicCall, basit altyordam çağrısını test eder.
func TestBasicCall(t *testing.T) {
	add := New("main", "add", func(ctx *CallContext) *sv.SV {
		a := ctx.Arg(0).AsInt()
		b := ctx.Arg(1).AsInt()
		return sv.NewInt(a + b)
	})

	ctx := &CallContext{
		Args: []*sv.SV{sv.NewInt(2), sv.NewInt(3)},
	}

	result := add.Call(ctx)
	if result.AsInt() != 5 {
		t.Errorf("2 + 3 should be 5, got %d", result.AsInt())
	}
}

// TestCallNoArgs tests call without arguments.
// TestCallNoArgs, argümansız çağrıyı test eder.
func TestCallNoArgs(t *testing.T) {
	cv := New("main", "noargs", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(int64(ctx.NumArgs()))
	})

	result := cv.Call(&CallContext{})
	if result.AsInt() != 0 {
		t.Errorf("NumArgs should be 0, got %d", result.AsInt())
	}
}

// TestCallNilContext tests call with nil context.
// TestCallNilContext, nil bağlam ile çağrıyı test eder.
func TestCallNilContext(t *testing.T) {
	cv := New("main", "nilctx", func(ctx *CallContext) *sv.SV {
		if ctx == nil {
			return sv.NewString("nil")
		}
		return sv.NewString("ok")
	})

	result := cv.Call(nil)
	if result.AsString() != "ok" {
		t.Error("Context should be created if nil")
	}
}

// TestCallManyArgs tests call with many arguments.
// TestCallManyArgs, çok argümanla çağrıyı test eder.
func TestCallManyArgs(t *testing.T) {
	sum := New("main", "sum", func(ctx *CallContext) *sv.SV {
		total := int64(0)
		for i := 0; i < ctx.NumArgs(); i++ {
			total += ctx.Arg(i).AsInt()
		}
		return sv.NewInt(total)
	})

	args := make([]*sv.SV, 100)
	for i := range args {
		args[i] = sv.NewInt(int64(i + 1))
	}

	result := sum.Call(&CallContext{Args: args})
	// Sum of 1..100 = 5050
	if result.AsInt() != 5050 {
		t.Errorf("Sum of 1..100 should be 5050, got %d", result.AsInt())
	}
}

// TestArgOutOfBounds tests accessing non-existent argument.
// TestArgOutOfBounds, var olmayan argümana erişimi test eder.
func TestArgOutOfBounds(t *testing.T) {
	cv := New("main", "oob", func(ctx *CallContext) *sv.SV {
		return ctx.Arg(999) // Out of bounds
	})

	result := cv.Call(&CallContext{Args: []*sv.SV{sv.NewInt(1)}})
	if !result.IsUndef() {
		t.Error("Out of bounds arg should return undef")
	}

	// Negative index
	cv2 := New("main", "neg", func(ctx *CallContext) *sv.SV {
		return ctx.Arg(-1)
	})
	result2 := cv2.Call(&CallContext{Args: []*sv.SV{sv.NewInt(1)}})
	if !result2.IsUndef() {
		t.Error("Negative index should return undef")
	}
}

// ============================================================
// Anonymous Subroutine Tests
// Anonim Altyordam Testleri
// ============================================================

// TestAnonymousSub tests anonymous subroutine.
// TestAnonymousSub, anonim altyordamı test eder.
func TestAnonymousSub(t *testing.T) {
	anon := NewAnon("main", func(ctx *CallContext) *sv.SV {
		return sv.NewString("anonymous")
	})

	if !anon.IsAnon() {
		t.Error("Should be anonymous")
	}
	if anon.Name() != "" {
		t.Error("Anonymous sub should have empty name")
	}

	result := anon.Call(nil)
	if result.AsString() != "anonymous" {
		t.Error("Should return 'anonymous'")
	}
}

// TestMultipleAnonymous tests multiple anonymous subs don't interfere.
// TestMultipleAnonymous, birden fazla anonim alt yordamın karışmadığını test eder.
func TestMultipleAnonymous(t *testing.T) {
	anon1 := NewAnon("main", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(1)
	})
	anon2 := NewAnon("main", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(2)
	})

	if anon1.Call(nil).AsInt() != 1 {
		t.Error("anon1 should return 1")
	}
	if anon2.Call(nil).AsInt() != 2 {
		t.Error("anon2 should return 2")
	}
}

// ============================================================
// Closure Tests
// Closure Testleri
// ============================================================

// TestClosure tests closure capturing variables.
// TestClosure, değişken yakalayan closure'ı test eder.
func TestClosure(t *testing.T) {
	outer := New("main", "make_counter", func(ctx *CallContext) *sv.SV {
		val := ctx.GetPad(0).AsInt()
		val++
		ctx.SetPad(0, sv.NewInt(val))
		return sv.NewInt(val)
	})
	outer.SetPadNames([]string{"counter"})

	pad := []*sv.SV{sv.NewInt(0)}
	closure := NewClosure(outer, pad)

	if !closure.IsClosure() {
		t.Error("Should be closure")
	}

	ctx := &CallContext{Pad: make([]*sv.SV, 1)}
	ctx.Pad[0] = sv.NewInt(0)

	r1 := closure.Call(ctx)
	r2 := closure.Call(ctx)
	r3 := closure.Call(ctx)

	if r1.AsInt() != 1 || r2.AsInt() != 2 || r3.AsInt() != 3 {
		t.Errorf("Counter should be 1,2,3 got %d,%d,%d", r1.AsInt(), r2.AsInt(), r3.AsInt())
	}
}

// TestIndependentClosures tests that closures don't share state.
// TestIndependentClosures, closure'ların durumu paylaşmadığını test eder.
func TestIndependentClosures(t *testing.T) {
	makeCounter := func(start int64) *CV {
		outer := New("main", "counter", func(ctx *CallContext) *sv.SV {
			val := ctx.GetPad(0).AsInt()
			val++
			ctx.SetPad(0, sv.NewInt(val))
			return sv.NewInt(val)
		})
		outer.SetPadNames([]string{"n"})
		return NewClosure(outer, []*sv.SV{sv.NewInt(start)})
	}

	counter1 := makeCounter(0)
	counter2 := makeCounter(100)

	ctx1 := &CallContext{Pad: []*sv.SV{sv.NewInt(0)}}
	ctx2 := &CallContext{Pad: []*sv.SV{sv.NewInt(100)}}

	// Each closure should have independent state
	// Her closure bağımsız duruma sahip olmalı
	if counter1.Call(ctx1).AsInt() != 1 {
		t.Error("counter1 first call should be 1")
	}
	if counter2.Call(ctx2).AsInt() != 101 {
		t.Error("counter2 first call should be 101")
	}
	if counter1.Call(ctx1).AsInt() != 2 {
		t.Error("counter1 second call should be 2")
	}
}

// TestNestedClosures tests closures inside closures.
// TestNestedClosures, iç içe closure'ları test eder.
func TestNestedClosures(t *testing.T) {
	outer := New("main", "outer", func(ctx *CallContext) *sv.SV {
		x := ctx.GetPad(0).AsInt()
		y := ctx.GetPad(1).AsInt()
		return sv.NewInt(x + y)
	})
	outer.SetPadNames([]string{"x", "y"})

	closure := NewClosure(outer, []*sv.SV{sv.NewInt(10), sv.NewInt(20)})

	ctx := &CallContext{Pad: []*sv.SV{sv.NewInt(10), sv.NewInt(20)}}
	result := closure.Call(ctx)

	if result.AsInt() != 30 {
		t.Errorf("10 + 20 should be 30, got %d", result.AsInt())
	}
}

// ============================================================
// Prototype Tests
// Prototip Testleri
// ============================================================

// TestPrototype tests subroutine prototype.
// TestPrototype, altyordam prototipini test eder.
func TestPrototype(t *testing.T) {
	cv := NewWithProto("main", "mysub", "$$@", func(ctx *CallContext) *sv.SV {
		return sv.NewUndef()
	})

	if cv.Prototype() != "$$@" {
		t.Errorf("Prototype should be '$$@', got '%s'", cv.Prototype())
	}
	if !cv.HasProto() {
		t.Error("Should have prototype flag")
	}
}

// TestSetPrototype tests setting prototype after creation.
// TestSetPrototype, oluşturulduktan sonra prototip ayarlamayı test eder.
func TestSetPrototype(t *testing.T) {
	cv := New("main", "mysub", nil)

	if cv.HasProto() {
		t.Error("Should not have prototype initially")
	}

	cv.SetPrototype(`\@`)
	if cv.Prototype() != `\@` {
		t.Errorf("Prototype should be '\\@', got '%s'", cv.Prototype())
	}
	if !cv.HasProto() {
		t.Error("Should have prototype after set")
	}
}

// TestEmptyPrototype tests empty prototype (no args).
// TestEmptyPrototype, boş prototipi (argümansız) test eder.
func TestEmptyPrototype(t *testing.T) {
	cv := NewWithProto("main", "noargs", "", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(42)
	})

	// Empty string is still a prototype
	// Boş string yine de bir prototiptir
	if cv.Prototype() != "" {
		t.Error("Prototype should be empty string")
	}
}

// ============================================================
// Constant Subroutine Tests
// Sabit Altyordam Testleri
// ============================================================

// TestConst tests constant subroutine.
// TestConst, sabit altyordamı test eder.
func TestConst(t *testing.T) {
	pi := NewConst("main", "PI", sv.NewFloat(3.14159))

	if !pi.IsConst() {
		t.Error("Should be constant")
	}

	result := pi.Call(nil)
	if result.AsFloat() != 3.14159 {
		t.Errorf("PI should be 3.14159, got %f", result.AsFloat())
	}

	// Calling multiple times should return same value
	// Birden fazla çağrı aynı değeri döndürmeli
	result2 := pi.Call(nil)
	if result2.AsFloat() != 3.14159 {
		t.Error("Constant should always return same value")
	}
}

// TestConstString tests string constant.
// TestConstString, string sabiti test eder.
func TestConstString(t *testing.T) {
	version := NewConst("main", "VERSION", sv.NewString("1.0.0"))

	if version.Call(nil).AsString() != "1.0.0" {
		t.Error("VERSION should be '1.0.0'")
	}
}

// TestConstUndef tests undef constant.
// TestConstUndef, undef sabitini test eder.
func TestConstUndef(t *testing.T) {
	undef := NewConst("main", "UNDEF", sv.NewUndef())

	if !undef.Call(nil).IsUndef() {
		t.Error("UNDEF constant should return undef")
	}
}

// ============================================================
// Attribute Tests
// Özellik Testleri
// ============================================================

// TestAttributes tests subroutine attributes.
// TestAttributes, altyordam özelliklerini test eder.
func TestAttributes(t *testing.T) {
	cv := New("main", "mysub", func(ctx *CallContext) *sv.SV {
		return sv.NewUndef()
	})

	cv.SetAttr("lvalue", "")
	cv.SetAttr("method", "")

	if !cv.IsLvalue() {
		t.Error("Should have lvalue attribute")
	}
	if !cv.IsMethod() {
		t.Error("Should have method attribute")
	}
}

// TestGetAttr tests getting attribute values.
// TestGetAttr, özellik değerlerini almayı test eder.
func TestGetAttr(t *testing.T) {
	cv := New("main", "mysub", nil)

	cv.SetAttr("custom", "myvalue")

	val, ok := cv.GetAttr("custom")
	if !ok {
		t.Error("Should find 'custom' attribute")
	}
	if val != "myvalue" {
		t.Errorf("Attribute value should be 'myvalue', got '%s'", val)
	}

	_, ok = cv.GetAttr("nonexistent")
	if ok {
		t.Error("Should not find 'nonexistent' attribute")
	}
}

// TestLockedAttribute tests :locked attribute.
// TestLockedAttribute, :locked özelliğini test eder.
func TestLockedAttribute(t *testing.T) {
	cv := New("main", "mysub", nil)
	cv.SetAttr("locked", "")

	if cv.flags&CVLocked == 0 {
		t.Error("Should have locked flag")
	}
}

// ============================================================
// Name and Package Tests
// İsim ve Paket Testleri
// ============================================================

// TestFullName tests name formatting.
// TestFullName, isim biçimlendirmesini test eder.
func TestFullName(t *testing.T) {
	tests := []struct {
		pkg    string
		name   string
		expect string
	}{
		{"MyPkg", "foo", "MyPkg::foo"},
		{"main", "bar", "bar"},
		{"", "baz", "baz"},
		{"A::B::C", "method", "A::B::C::method"},
	}

	for _, tt := range tests {
		cv := New(tt.pkg, tt.name, nil)
		if cv.FullName() != tt.expect {
			t.Errorf("FullName(%s, %s) = '%s', want '%s'",
				tt.pkg, tt.name, cv.FullName(), tt.expect)
		}
	}
}

// TestAnonFullName tests anonymous sub full name.
// TestAnonFullName, anonim altyordam tam adını test eder.
func TestAnonFullName(t *testing.T) {
	anon := NewAnon("MyPkg", nil)
	if anon.FullName() != "MyPkg::__ANON__" {
		t.Errorf("Expected 'MyPkg::__ANON__', got '%s'", anon.FullName())
	}
}

// ============================================================
// Context Tests
// Bağlam Testleri
// ============================================================

// TestWantArray tests context detection.
// TestWantArray, bağlam algılamayı test eder.
func TestWantArray(t *testing.T) {
	cv := New("main", "context_test", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(int64(ctx.WantArrayVal()))
	})

	tests := []struct {
		wantArray int
		expect    int64
	}{
		{-1, -1}, // void
		{0, 0},   // scalar
		{1, 1},   // list
	}

	for _, tt := range tests {
		ctx := &CallContext{WantArray: tt.wantArray}
		result := cv.Call(ctx)
		if result.AsInt() != tt.expect {
			t.Errorf("WantArray %d should return %d, got %d",
				tt.wantArray, tt.expect, result.AsInt())
		}
	}
}

// TestCallList tests list context call.
// TestCallList, liste bağlamı çağrısını test eder.
func TestCallList(t *testing.T) {
	cv := New("main", "list_return", func(ctx *CallContext) *sv.SV {
		arr := sv.NewArrayRef(
			sv.NewInt(1),
			sv.NewInt(2),
			sv.NewInt(3),
		)
		return arr
	})

	results := cv.CallList(nil)
	if len(results) != 3 {
		t.Errorf("Should return 3 elements, got %d", len(results))
	}
}

// TestCallListScalar tests CallList with scalar return.
// TestCallListScalar, skaler dönüşlü CallList'i test eder.
func TestCallListScalar(t *testing.T) {
	cv := New("main", "scalar_return", func(ctx *CallContext) *sv.SV {
		return sv.NewInt(42)
	})

	results := cv.CallList(nil)
	if len(results) != 1 {
		t.Errorf("Should return 1 element, got %d", len(results))
	}
	if results[0].AsInt() != 42 {
		t.Error("Element should be 42")
	}
}

// TestCallListNil tests CallList with nil return.
// TestCallListNil, nil dönüşlü CallList'i test eder.
func TestCallListNil(t *testing.T) {
	cv := New("main", "nil_return", func(ctx *CallContext) *sv.SV {
		return nil
	})

	results := cv.CallList(nil)
	if len(results) != 0 {
		t.Errorf("Should return 0 elements for nil, got %d", len(results))
	}
}

// ============================================================
// Pad (Lexical Variables) Tests
// Pad (Leksikal Değişkenler) Testleri
// ============================================================

// TestPadNames tests lexical variable name management.
// TestPadNames, leksikal değişken isim yönetimini test eder.
func TestPadNames(t *testing.T) {
	cv := New("main", "lexicals", nil)

	idx1 := cv.AddPadName("$x")
	idx2 := cv.AddPadName("$y")
	idx3 := cv.AddPadName("@arr")

	if idx1 != 0 || idx2 != 1 || idx3 != 2 {
		t.Error("Pad indices should be sequential")
	}

	names := cv.PadNames()
	if len(names) != 3 {
		t.Errorf("Should have 3 pad names, got %d", len(names))
	}
}

// TestPadIndex tests looking up pad index by name.
// TestPadIndex, isme göre pad indeksi aramayı test eder.
func TestPadIndex(t *testing.T) {
	cv := New("main", "lookup", nil)
	cv.SetPadNames([]string{"$a", "$b", "$c"})

	if cv.PadIndex("$a") != 0 {
		t.Error("$a should be at index 0")
	}
	if cv.PadIndex("$c") != 2 {
		t.Error("$c should be at index 2")
	}
	if cv.PadIndex("$missing") != -1 {
		t.Error("Missing name should return -1")
	}
}

// TestCtxPadOperations tests context pad get/set.
// TestCtxPadOperations, bağlam pad get/set işlemlerini test eder.
func TestCtxPadOperations(t *testing.T) {
	ctx := &CallContext{}

	// Set should auto-extend
	// Set otomatik genişletmeli
	ctx.SetPad(5, sv.NewInt(42))

	if len(ctx.Pad) < 6 {
		t.Error("Pad should auto-extend")
	}

	val := ctx.GetPad(5)
	if val.AsInt() != 42 {
		t.Errorf("Pad[5] should be 42, got %d", val.AsInt())
	}

	// Out of bounds should return undef
	// Sınır dışı undef döndürmeli
	if !ctx.GetPad(100).IsUndef() {
		t.Error("Out of bounds should return undef")
	}
}

// ============================================================
// Caller Info Tests
// Çağıran Bilgisi Testleri
// ============================================================

// TestCallerInfo tests caller information retrieval.
// TestCallerInfo, çağıran bilgisi almayı test eder.
func TestCallerInfo(t *testing.T) {
	callerCtx := &CallContext{
		Package: "CallerPkg",
		File:    "test.pl",
		Line:    42,
	}

	ctx := &CallContext{
		Caller: callerCtx,
	}

	pkg, file, line := ctx.CallerInfo()
	if pkg != "CallerPkg" {
		t.Errorf("Caller package should be 'CallerPkg', got '%s'", pkg)
	}
	if file != "test.pl" {
		t.Errorf("Caller file should be 'test.pl', got '%s'", file)
	}
	if line != 42 {
		t.Errorf("Caller line should be 42, got %d", line)
	}
}

// TestCallerInfoNil tests caller info with no caller.
// TestCallerInfoNil, çağıran olmadan çağıran bilgisini test eder.
func TestCallerInfoNil(t *testing.T) {
	ctx := &CallContext{}
	pkg, file, line := ctx.CallerInfo()

	if pkg != "" || file != "" || line != 0 {
		t.Error("No caller should return empty values")
	}

	// Nil context
	var nilCtx *CallContext
	pkg, file, line = nilCtx.CallerInfo()
	if pkg != "" || file != "" || line != 0 {
		t.Error("Nil context should return empty values")
	}
}

// ============================================================
// XSUB Tests
// XSUB Testleri
// ============================================================

// TestXSUB tests external subroutine flag.
// TestXSUB, harici altyordam bayrağını test eder.
func TestXSUB(t *testing.T) {
	xsub := NewXSUB("main", "builtin", func(ctx *CallContext) *sv.SV {
		return sv.NewString("builtin result")
	})

	if !xsub.IsXSUB() {
		t.Error("Should be XSUB")
	}

	result := xsub.Call(nil)
	if result.AsString() != "builtin result" {
		t.Error("XSUB should execute correctly")
	}
}

// ============================================================
// Free/Cleanup Tests
// Serbest Bırakma/Temizlik Testleri
// ============================================================

// TestFree tests resource cleanup.
// TestFree, kaynak temizliğini test eder.
func TestFree(t *testing.T) {
	pad := []*sv.SV{sv.NewInt(1), sv.NewInt(2)}
	for _, v := range pad {
		v.IncRef() // Simulate extra reference
	}

	outer := New("main", "test", nil)
	closure := NewClosure(outer, pad)

	// Free should not panic
	// Free panic yapmamalı
	closure.Free()

	if closure.pad != nil {
		t.Error("Pad should be nil after free")
	}
}

// ============================================================
// Edge Cases
// Sınır Durumları
// ============================================================

// TestNilNativeFunction tests CV with nil native function.
// TestNilNativeFunction, nil native fonksiyonlu CV'yi test eder.
func TestNilNativeFunction(t *testing.T) {
	cv := New("main", "nil_native", nil)

	result := cv.Call(nil)
	if !result.IsUndef() {
		t.Error("Nil native function should return undef")
	}
}

// TestEmptyOps tests CV with empty ops slice.
// TestEmptyOps, boş ops dilimli CV'yi test eder.
func TestEmptyOps(t *testing.T) {
	cv := &CV{
		name: "empty_ops",
		pkg:  "main",
		ops:  []Op{},
	}

	result := cv.Call(nil)
	if result != nil && !result.IsUndef() {
		t.Error("Empty ops should return nil or undef")
	}
}

// TestMixedTypeArgs tests arguments with mixed types.
// TestMixedTypeArgs, karışık türlü argümanları test eder.
func TestMixedTypeArgs(t *testing.T) {
	cv := New("main", "mixed", func(ctx *CallContext) *sv.SV {
		result := ""
		for i := 0; i < ctx.NumArgs(); i++ {
			result += ctx.Arg(i).AsString() + ","
		}
		return sv.NewString(result)
	})

	ctx := &CallContext{
		Args: []*sv.SV{
			sv.NewInt(42),
			sv.NewString("hello"),
			sv.NewFloat(3.14),
			sv.NewUndef(),
		},
	}

	result := cv.Call(ctx)
	if result.AsString() != "42,hello,3.14,," {
		t.Errorf("Mixed args should concat correctly, got '%s'", result.AsString())
	}
}
