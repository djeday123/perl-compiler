package context

import (
	"fmt"
	"runtime"
	"testing"

	"perlc/pkg/cv"
	"perlc/pkg/stash"
	"perlc/pkg/sv"
)

// ============================================================
// Runtime Creation Tests
// Runtime Oluşturma Testleri
// ============================================================

// TestNewRuntime tests runtime initialization.
// TestNewRuntime, runtime başlatmayı test eder.
func TestNewRuntime(t *testing.T) {
	rt := NewRuntime()

	if rt == nil {
		t.Fatal("NewRuntime should not return nil")
	}

	if rt.Package() != "main" {
		t.Errorf("Default package should be 'main', got '%s'", rt.Package())
	}

	if rt.CallDepth() != 0 {
		t.Errorf("Initial call depth should be 0, got %d", rt.CallDepth())
	}
}

// TestGetRuntime tests global runtime singleton.
// TestGetRuntime, global runtime singleton'ı test eder.
func TestGetRuntime(t *testing.T) {
	rt1 := GetRuntime()
	rt2 := GetRuntime()

	if rt1 != rt2 {
		t.Error("GetRuntime should return same instance")
	}
}

// ============================================================
// Call Stack Tests
// Çağrı Yığını Testleri
// ============================================================

// TestPushPopCall tests basic call stack operations.
// TestPushPopCall, temel çağrı yığını işlemlerini test eder.
func TestPushPopCall(t *testing.T) {
	rt := NewRuntime()

	frame := &StackFrame{
		Package: "TestPkg",
		File:    "test.pl",
		Line:    10,
		Sub:     "test_sub",
	}

	rt.PushCall(frame)

	if rt.CallDepth() != 1 {
		t.Errorf("Call depth should be 1, got %d", rt.CallDepth())
	}

	popped := rt.PopCall()
	if popped != frame {
		t.Error("PopCall should return pushed frame")
	}

	if rt.CallDepth() != 0 {
		t.Errorf("Call depth should be 0 after pop, got %d", rt.CallDepth())
	}
}

// TestCurrentFrame tests getting current frame.
// TestCurrentFrame, geçerli çerçeveyi almayı test eder.
func TestCurrentFrame(t *testing.T) {
	rt := NewRuntime()

	if rt.CurrentFrame() != nil {
		t.Error("CurrentFrame should be nil when stack empty")
	}

	frame1 := &StackFrame{Sub: "sub1"}
	frame2 := &StackFrame{Sub: "sub2"}

	rt.PushCall(frame1)
	rt.PushCall(frame2)

	current := rt.CurrentFrame()
	if current.Sub != "sub2" {
		t.Error("CurrentFrame should return top frame")
	}
}

// TestCaller tests caller() functionality.
// TestCaller, caller() işlevselliğini test eder.
func TestCaller(t *testing.T) {
	rt := NewRuntime()

	frames := []*StackFrame{
		{Package: "main", File: "test.pl", Line: 1, Sub: "main"},
		{Package: "Foo", File: "Foo.pm", Line: 10, Sub: "foo"},
		{Package: "Bar", File: "Bar.pm", Line: 20, Sub: "bar"},
	}

	for _, f := range frames {
		rt.PushCall(f)
	}

	// caller(0) = current = bar
	c0 := rt.Caller(0)
	if c0.Sub != "bar" {
		t.Errorf("caller(0) should be 'bar', got '%s'", c0.Sub)
	}

	// caller(1) = foo
	c1 := rt.Caller(1)
	if c1.Sub != "foo" {
		t.Errorf("caller(1) should be 'foo', got '%s'", c1.Sub)
	}

	// caller(2) = main
	c2 := rt.Caller(2)
	if c2.Sub != "main" {
		t.Errorf("caller(2) should be 'main', got '%s'", c2.Sub)
	}

	// caller(3) = nil (out of bounds)
	c3 := rt.Caller(3)
	if c3 != nil {
		t.Error("caller(3) should be nil")
	}
}

// TestCallerNegative tests caller with negative level.
// TestCallerNegative, negatif seviye ile caller'ı test eder.
func TestCallerNegative(t *testing.T) {
	rt := NewRuntime()
	rt.PushCall(&StackFrame{Sub: "test"})

	if rt.Caller(-1) != nil {
		t.Error("Negative caller level should return nil")
	}
}

// TestDeepCallStack tests deep call stack.
// TestDeepCallStack, derin çağrı yığınını test eder.
func TestDeepCallStack(t *testing.T) {
	rt := NewRuntime()

	depth := 100
	for i := 0; i < depth; i++ {
		rt.PushCall(&StackFrame{Line: i})
	}

	if rt.CallDepth() != depth {
		t.Errorf("Call depth should be %d, got %d", depth, rt.CallDepth())
	}

	// Verify order
	for i := 0; i < depth; i++ {
		frame := rt.Caller(i)
		expectedLine := depth - 1 - i
		if frame.Line != expectedLine {
			t.Errorf("caller(%d) line should be %d, got %d", i, expectedLine, frame.Line)
		}
	}
}

// TestStackTrace tests stack trace generation.
// TestStackTrace, yığın izi oluşturmayı test eder.
func TestStackTrace(t *testing.T) {
	rt := NewRuntime()

	rt.PushCall(&StackFrame{Package: "main", Sub: "main", File: "test.pl", Line: 1})
	rt.PushCall(&StackFrame{Package: "Foo", Sub: "bar", File: "Foo.pm", Line: 42})

	trace := rt.StackTrace()

	if trace == "" {
		t.Error("StackTrace should not be empty")
	}

	// Should contain both frames
	if len(trace) < 20 {
		t.Error("StackTrace seems too short")
	}
}

// TestPopEmptyStack tests popping from empty stack.
// TestPopEmptyStack, boş yığından pop'u test eder.
func TestPopEmptyStack(t *testing.T) {
	rt := NewRuntime()

	frame := rt.PopCall()
	if frame != nil {
		t.Error("PopCall on empty stack should return nil")
	}
}

// ============================================================
// Dynamic Scope (local) Tests
// Dinamik Kapsam (local) Testleri
// ============================================================

// TestLocalScalar tests local($var).
// TestLocalScalar, local($var) test eder.
func TestLocalScalar(t *testing.T) {
	rt := NewRuntime()

	// Set initial value
	stash.Get("main").SetScalar("x", sv.NewInt(42))

	// Enter scope and localize
	rt.PushLocal()
	rt.LocalScalar("main::x")

	// Value should be undef now
	val := stash.Get("main").Scalar("x")
	if !val.IsUndef() {
		t.Error("local($x) should set to undef")
	}

	// Set new value
	stash.Get("main").SetScalar("x", sv.NewInt(100))

	// Leave scope - should restore
	rt.PopLocal()

	restored := stash.Get("main").Scalar("x")
	if restored.AsInt() != 42 {
		t.Errorf("After PopLocal, $x should be 42, got %d", restored.AsInt())
	}
}

// TestLocalArray tests local(@var).
// TestLocalArray, local(@var) test eder.
func TestLocalArray(t *testing.T) {
	rt := NewRuntime()

	// Set initial array
	arr := stash.Get("main").Array("arr")
	arr.SetArrayData([]*sv.SV{sv.NewInt(1), sv.NewInt(2), sv.NewInt(3)})

	rt.PushLocal()
	rt.LocalArray("main::arr")

	// Should be empty array now
	localArr := stash.Get("main").Array("arr")
	if len(localArr.ArrayData()) != 0 {
		t.Error("local(@arr) should be empty")
	}

	rt.PopLocal()
}

// TestLocalHash tests local(%var).
// TestLocalHash, local(%var) test eder.
func TestLocalHash(t *testing.T) {
	rt := NewRuntime()

	hash := stash.Get("main").Hash("hash")
	hash.HashData()["key"] = sv.NewString("value")

	rt.PushLocal()
	rt.LocalHash("main::hash")

	localHash := stash.Get("main").Hash("hash")
	if len(localHash.HashData()) != 0 {
		t.Error("local(hash) should be empty")
	}

	rt.PopLocal()
}

// TestNestedLocal tests nested local scopes.
// TestNestedLocal, iç içe local kapsamlarını test eder.
func TestNestedLocal(t *testing.T) {
	rt := NewRuntime()

	stash.Get("main").SetScalar("n", sv.NewInt(1))

	rt.PushLocal()
	rt.LocalScalar("main::n")
	stash.Get("main").SetScalar("n", sv.NewInt(2))

	rt.PushLocal()
	rt.LocalScalar("main::n")
	stash.Get("main").SetScalar("n", sv.NewInt(3))

	if stash.Get("main").Scalar("n").AsInt() != 3 {
		t.Error("Innermost value should be 3")
	}

	rt.PopLocal()
	if stash.Get("main").Scalar("n").AsInt() != 2 {
		t.Error("After first pop, value should be 2")
	}

	rt.PopLocal()
	if stash.Get("main").Scalar("n").AsInt() != 1 {
		t.Error("After second pop, value should be 1")
	}
}

// TestLocalWithoutPush tests local without explicit PushLocal.
// TestLocalWithoutPush, açık PushLocal olmadan local'i test eder.
func TestLocalWithoutPush(t *testing.T) {
	rt := NewRuntime()

	stash.Get("main").SetScalar("auto", sv.NewInt(99))

	// Should auto-create frame
	rt.LocalScalar("main::auto")

	// Cleanup
	rt.PopLocal()
}

// ============================================================
// Special Variables Tests
// Özel Değişken Testleri
// ============================================================

// TestUnderscore tests $_.
// TestUnderscore, $_ test eder.
func TestUnderscore(t *testing.T) {
	rt := NewRuntime()

	rt.SetUnderscore(sv.NewString("test value"))

	val := rt.Underscore()
	if val.AsString() != "test value" {
		t.Errorf("$_ should be 'test value', got '%s'", val.AsString())
	}
}

// TestInputRS tests $/ (input record separator).
// TestInputRS, $/ (girdi kayıt ayırıcı) test eder.
func TestInputRS(t *testing.T) {
	rt := NewRuntime()

	// Default is newline
	if rt.InputRS().AsString() != "\n" {
		t.Error("Default $/ should be newline")
	}

	rt.SetInputRS(sv.NewString("\r\n"))
	if rt.InputRS().AsString() != "\r\n" {
		t.Error("$/ should be CRLF after set")
	}
}

// TestOutputRS tests $\ (output record separator).
// TestOutputRS, $\ (çıktı kayıt ayırıcı) test eder.
func TestOutputRS(t *testing.T) {
	rt := NewRuntime()

	// Default is empty
	if rt.OutputRS().AsString() != "" {
		t.Error("Default $\\ should be empty")
	}

	rt.SetOutputRS(sv.NewString("\n"))
	if rt.OutputRS().AsString() != "\n" {
		t.Error("$\\ should be newline after set")
	}
}

// TestPID tests $$.
// TestPID, $$ test eder.
func TestPID(t *testing.T) {
	rt := NewRuntime()

	pid := rt.PID().AsInt()
	if pid <= 0 {
		t.Errorf("$$ should be positive, got %d", pid)
	}
}

// TestProgName tests $0.
// TestProgName, $0 test eder.
func TestProgName(t *testing.T) {
	rt := NewRuntime()

	// Should have some value
	name := rt.ProgName().AsString()
	if name == "" {
		t.Error("$0 should not be empty")
	}

	rt.SetProgName(sv.NewString("custom_name"))
	if rt.ProgName().AsString() != "custom_name" {
		t.Error("$0 should be 'custom_name' after set")
	}
}

// TestListSep tests $" (list separator).
// TestListSep, $" (liste ayırıcı) test eder.
func TestListSep(t *testing.T) {
	rt := NewRuntime()

	// Default is space
	if rt.ListSep().AsString() != " " {
		t.Error("Default $\" should be space")
	}
}

// TestProcessInfo tests uid/gid variables.
// TestProcessInfo, uid/gid değişkenlerini test eder.
func TestProcessInfo(t *testing.T) {
	rt := NewRuntime()

	// These should return integers (may be -1 on Windows)
	// Bunlar tamsayı döndürmeli (Windows'ta -1 olabilir)
	uid := rt.UID().AsInt()
	euid := rt.EUID().AsInt()
	gid := rt.GID().AsInt()
	egid := rt.EGID().AsInt()

	// Just verify they return something (not undef)
	// Sadece bir şey döndürdüklerini doğrula (undef değil)
	if rt.UID().IsUndef() {
		t.Error("UID should not be undef")
	}
	if rt.EUID().IsUndef() {
		t.Error("EUID should not be undef")
	}
	if rt.GID().IsUndef() {
		t.Error("GID should not be undef")
	}
	if rt.EGID().IsUndef() {
		t.Error("EGID should not be undef")
	}

	// Use variables to avoid "unused" warning
	// "unused" uyarısından kaçınmak için değişkenleri kullan
	_ = uid
	_ = euid
	_ = gid
	_ = egid
}

// TestProcessInfo tests uid/gid variables.
// TestProcessInfo, uid/gid değişkenlerini test eder.
func TestProcessInfoWithPlatform(t *testing.T) {
	rt := NewRuntime()

	uid := rt.UID().AsInt()
	euid := rt.EUID().AsInt()
	gid := rt.GID().AsInt()
	egid := rt.EGID().AsInt()

	// On Unix these are >= 0, on Windows they are -1
	// Unix'te bunlar >= 0, Windows'ta -1
	if runtime.GOOS != "windows" {
		if uid < 0 || euid < 0 || gid < 0 || egid < 0 {
			t.Error("Process IDs should be non-negative on Unix")
		}
	} else {
		// Windows returns -1 for these
		// Windows bunlar için -1 döndürür
		if uid != -1 || euid != -1 || gid != -1 || egid != -1 {
			t.Error("Process IDs should be -1 on Windows")
		}
	}
}

// ============================================================
// Regex Match Variables Tests
// Regex Eşleşme Değişkenleri Testleri
// ============================================================

// TestMatchVars tests regex match result variables.
// TestMatchVars, regex eşleşme sonuç değişkenlerini test eder.
func TestMatchVars(t *testing.T) {
	rt := NewRuntime()

	rt.SetMatchVars("hello", "say ", " world", []string{"hel", "lo"})

	// $& (entire match)
	if rt.Match().AsString() != "hello" {
		t.Errorf("$& should be 'hello', got '%s'", rt.Match().AsString())
	}

	// $` (prematch)
	if rt.PreMatch().AsString() != "say " {
		t.Errorf("$` should be 'say ', got '%s'", rt.PreMatch().AsString())
	}

	// $' (postmatch)
	if rt.PostMatch().AsString() != " world" {
		t.Errorf("$' should be ' world', got '%s'", rt.PostMatch().AsString())
	}

	// $+ (last paren)
	if rt.LastParen().AsString() != "lo" {
		t.Errorf("$+ should be 'lo', got '%s'", rt.LastParen().AsString())
	}
}

// TestCaptures tests $1, $2, etc.
// TestCaptures, $1, $2, vb. test eder.
func TestCaptures(t *testing.T) {
	rt := NewRuntime()

	rt.SetMatchVars("match", "", "", []string{"first", "second", "third"})

	// $1
	if rt.Capture(1).AsString() != "first" {
		t.Errorf("$1 should be 'first', got '%s'", rt.Capture(1).AsString())
	}

	// $2
	if rt.Capture(2).AsString() != "second" {
		t.Errorf("$2 should be 'second', got '%s'", rt.Capture(2).AsString())
	}

	// $3
	if rt.Capture(3).AsString() != "third" {
		t.Errorf("$3 should be 'third', got '%s'", rt.Capture(3).AsString())
	}

	// $4 (doesn't exist)
	if !rt.Capture(4).IsUndef() {
		t.Error("$4 should be undef")
	}

	// $0 (invalid)
	if !rt.Capture(0).IsUndef() {
		t.Error("$0 capture should be undef")
	}
}

// TestEmptyCaptures tests when no captures.
// TestEmptyCaptures, yakalama olmadığında test eder.
func TestEmptyCaptures(t *testing.T) {
	rt := NewRuntime()

	rt.SetMatchVars("match", "pre", "post", []string{})

	if !rt.LastParen().IsUndef() {
		t.Error("$+ should be undef with no captures")
	}
}

// TestMatchVarsDefault tests default match var values.
// TestMatchVarsDefault, varsayılan eşleşme değişken değerlerini test eder.
func TestMatchVarsDefault(t *testing.T) {
	rt := NewRuntime()

	// Before any match
	if !rt.Match().IsUndef() {
		t.Error("$& should be undef before match")
	}
	if !rt.PreMatch().IsUndef() {
		t.Error("$` should be undef before match")
	}
	if !rt.PostMatch().IsUndef() {
		t.Error("$' should be undef before match")
	}
}

// ============================================================
// Error Handling Tests
// Hata İşleme Testleri
// ============================================================

// TestEvalError tests $@.
// TestEvalError, $@ test eder.
func TestEvalError(t *testing.T) {
	rt := NewRuntime()

	rt.SetEvalError(sv.NewString("Something went wrong"))

	if rt.EvalError().AsString() != "Something went wrong" {
		t.Error("$@ not set correctly")
	}

	rt.ClearEvalError()
	if rt.EvalError().AsString() != "" {
		t.Error("$@ should be empty after clear")
	}
}

// TestOSError tests $!.
// TestOSError, $! test eder.
func TestOSError(t *testing.T) {
	rt := NewRuntime()

	rt.SetOSError(nil)
	if rt.OSError().AsString() != "" {
		t.Error("$! should be empty for nil error")
	}

	rt.SetOSError(fmt.Errorf("file not found"))
	if rt.OSError().AsString() != "file not found" {
		t.Error("$! should be 'file not found'")
	}
}

// TestChildError tests $?.
// TestChildError, $? test eder.
func TestChildError(t *testing.T) {
	rt := NewRuntime()

	rt.SetChildError(256) // Exit code 1 in shell
	if rt.ChildError().AsInt() != 256 {
		t.Errorf("$? should be 256, got %d", rt.ChildError().AsInt())
	}
}

// TestDie tests die() function.
// TestDie, die() fonksiyonunu test eder.
func TestDie(t *testing.T) {
	rt := NewRuntime()

	// Die outside eval should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Die should panic outside eval")
		}
		die, ok := r.(PerlDie)
		if !ok {
			t.Error("Should panic with PerlDie type")
		}
		if die.Message != "test error" {
			t.Errorf("Die message should be 'test error', got '%s'", die.Message)
		}
	}()

	rt.Die("test error")
}

// TestDieInEval tests die() inside eval.
// TestDieInEval, eval içinde die() test eder.
func TestDieInEval(t *testing.T) {
	rt := NewRuntime()

	rt.EnterEval()
	rt.Die("eval error")
	rt.LeaveEval()

	// Should not panic, just set $@
	if rt.EvalError().AsString() != "eval error" {
		t.Errorf("$@ should be 'eval error', got '%s'", rt.EvalError().AsString())
	}
}

// TestTryEval tests TryEval helper.
// TestTryEval, TryEval yardımcısını test eder.
func TestTryEval(t *testing.T) {
	rt := NewRuntime()

	// Successful eval
	success := rt.TryEval(func() {
		// Do nothing, no error
	})

	if !success {
		t.Error("TryEval should return true on success")
	}

	// Failed eval
	success = rt.TryEval(func() {
		rt.Die("failure")
	})

	if success {
		t.Error("TryEval should return false on die")
	}
	if rt.EvalError().AsString() != "failure" {
		t.Error("$@ should be set after failed eval")
	}
}

// TestTryEvalPanic tests TryEval with Go panic.
// TestTryEvalPanic, Go panic ile TryEval'i test eder.
func TestTryEvalPanic(t *testing.T) {
	rt := NewRuntime()

	success := rt.TryEval(func() {
		panic("go panic")
	})

	if success {
		t.Error("TryEval should catch Go panics")
	}
	if rt.EvalError().AsString() != "go panic" {
		t.Errorf("$@ should contain panic message, got '%s'", rt.EvalError().AsString())
	}
}

// TestNestedEval tests nested eval blocks.
// TestNestedEval, iç içe eval bloklarını test eder.
func TestNestedEval(t *testing.T) {
	rt := NewRuntime()

	if rt.InEval() {
		t.Error("Should not be in eval initially")
	}

	rt.EnterEval()
	if !rt.InEval() {
		t.Error("Should be in eval after EnterEval")
	}

	rt.EnterEval()
	rt.LeaveEval()

	if !rt.InEval() {
		t.Error("Should still be in outer eval")
	}

	rt.LeaveEval()
	if rt.InEval() {
		t.Error("Should not be in eval after all LeaveEval")
	}
}

// ============================================================
// Package Context Tests
// Paket Bağlamı Testleri
// ============================================================

// TestPackage tests current package.
// TestPackage, geçerli paketi test eder.
func TestPackage(t *testing.T) {
	rt := NewRuntime()

	if rt.Package() != "main" {
		t.Error("Default package should be 'main'")
	}

	rt.SetPackage("Foo::Bar")
	if rt.Package() != "Foo::Bar" {
		t.Errorf("Package should be 'Foo::Bar', got '%s'", rt.Package())
	}
}

// ============================================================
// Strict/Warnings Tests
// Strict/Warnings Testleri
// ============================================================

// TestUseStrict tests 'use strict'.
// TestUseStrict, 'use strict' test eder.
func TestUseStrict(t *testing.T) {
	rt := NewRuntime()

	if rt.IsStrict(StrictRefs) {
		t.Error("strict refs should be off initially")
	}

	rt.UseStrict(StrictRefs | StrictVars)

	if !rt.IsStrict(StrictRefs) {
		t.Error("strict refs should be on")
	}
	if !rt.IsStrict(StrictVars) {
		t.Error("strict vars should be on")
	}
	if rt.IsStrict(StrictSubs) {
		t.Error("strict subs should still be off")
	}

	rt.NoStrict(StrictRefs)
	if rt.IsStrict(StrictRefs) {
		t.Error("strict refs should be off after NoStrict")
	}
}

// TestUseWarnings tests 'use warnings'.
// TestUseWarnings, 'use warnings' test eder.
func TestUseWarnings(t *testing.T) {
	rt := NewRuntime()

	rt.UseWarnings(WarnUninitialized | WarnNumeric)

	if !rt.IsWarning(WarnUninitialized) {
		t.Error("uninitialized warning should be on")
	}
	if !rt.IsWarning(WarnNumeric) {
		t.Error("numeric warning should be on")
	}
	if rt.IsWarning(WarnIO) {
		t.Error("IO warning should be off")
	}

	rt.NoWarnings(WarnUninitialized)
	if rt.IsWarning(WarnUninitialized) {
		t.Error("uninitialized warning should be off")
	}
}

// TestUseFeature tests 'use feature'.
// TestUseFeature, 'use feature' test eder.
func TestUseFeature(t *testing.T) {
	rt := NewRuntime()

	rt.UseFeature(FeatureSay | FeatureState)

	if !rt.HasFeature(FeatureSay) {
		t.Error("say feature should be enabled")
	}
	if !rt.HasFeature(FeatureState) {
		t.Error("state feature should be enabled")
	}
	if rt.HasFeature(FeatureSignatures) {
		t.Error("signatures feature should be disabled")
	}
}

// ============================================================
// Signal Handler Tests
// Sinyal İşleyici Testleri
// ============================================================

// TestSetDieHandler tests $SIG{__DIE__}.
// TestSetDieHandler, $SIG{__DIE__} test eder.
func TestSetDieHandler(t *testing.T) {
	rt := NewRuntime()

	handler := sv.NewInt(1) // Placeholder CV
	rt.SetDieHandler(handler)

	// Handler is set (actual invocation tested elsewhere)
	// İşleyici ayarlandı (gerçek çağrı başka yerde test edilir)
}

// TestSetWarnHandler tests $SIG{__WARN__}.
// TestSetWarnHandler, $SIG{__WARN__} test eder.
func TestSetWarnHandler(t *testing.T) {
	rt := NewRuntime()

	handler := sv.NewInt(1)
	rt.SetWarnHandler(handler)
}

// ============================================================
// Format Variables Tests
// Format Değişkenleri Testleri
// ============================================================

// TestFormat tests $~ (format name).
// TestFormat, $~ (format adı) test eder.
func TestFormat(t *testing.T) {
	rt := NewRuntime()

	if !rt.Format().IsUndef() {
		t.Error("Default format should be undef")
	}

	rt.SetFormat(sv.NewString("STDOUT"))
	if rt.Format().AsString() != "STDOUT" {
		t.Error("Format should be 'STDOUT'")
	}
}

// TestAccumulator tests $^A (format accumulator).
// TestAccumulator, $^A (format akümülatör) test eder.
func TestAccumulator(t *testing.T) {
	rt := NewRuntime()

	if rt.Accumulator().AsString() != "" {
		t.Error("Default accumulator should be empty")
	}

	rt.SetAccumulator(sv.NewString("accumulated"))
	if rt.Accumulator().AsString() != "accumulated" {
		t.Error("Accumulator should be 'accumulated'")
	}
}

// ============================================================
// Stack Frame Tests
// Yığın Çerçevesi Testleri
// ============================================================

// TestStackFrameFields tests StackFrame field access.
// TestStackFrameFields, StackFrame alan erişimini test eder.
func TestStackFrameFields(t *testing.T) {
	cvObj := cv.New("Test", "method", nil)

	frame := &StackFrame{
		CV:        cvObj,
		Package:   "Test",
		File:      "test.pl",
		Line:      42,
		Sub:       "method",
		Args:      []*sv.SV{sv.NewInt(1), sv.NewInt(2)},
		WantArray: 1,
		HasArgs:   true,
		IsEval:    false,
		EvalText:  "",
	}

	if frame.CV != cvObj {
		t.Error("CV field wrong")
	}
	if frame.Package != "Test" {
		t.Error("Package field wrong")
	}
	if frame.File != "test.pl" {
		t.Error("File field wrong")
	}
	if frame.Line != 42 {
		t.Error("Line field wrong")
	}
	if frame.Sub != "method" {
		t.Error("Sub field wrong")
	}
	if len(frame.Args) != 2 {
		t.Error("Args field wrong")
	}
	if frame.WantArray != 1 {
		t.Error("WantArray field wrong")
	}
	if !frame.HasArgs {
		t.Error("HasArgs field wrong")
	}
	if frame.IsEval {
		t.Error("IsEval field wrong")
	}
	if frame.EvalText != "" {
		t.Error("EvalText field wrong")
	}

}

// TestEvalFrame tests eval stack frame.
// TestEvalFrame, eval yığın çerçevesini test eder.
func TestEvalFrame(t *testing.T) {
	frame := &StackFrame{
		Package:  "main",
		IsEval:   true,
		EvalText: "print 'hello'",
	}

	if frame.Package != "main" {
		t.Error("Package should be 'main'")
	}
	if !frame.IsEval {
		t.Error("Should be eval frame")
	}
	if frame.EvalText != "print 'hello'" {
		t.Error("EvalText wrong")
	}
}

// ============================================================
// Concurrency Tests
// Eşzamanlılık Testleri
// ============================================================

// TestConcurrentSpecialVars tests thread-safe special var access.
// TestConcurrentSpecialVars, iş parçacığı güvenli özel değişken erişimini test eder.
func TestConcurrentSpecialVars(t *testing.T) {
	rt := NewRuntime()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			rt.SetUnderscore(sv.NewInt(int64(n)))
			_ = rt.Underscore()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestConcurrentCallStack tests thread-safe call stack.
// TestConcurrentCallStack, iş parçacığı güvenli çağrı yığınını test eder.
func TestConcurrentCallStack(t *testing.T) {
	rt := NewRuntime()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			rt.PushCall(&StackFrame{Line: n})
			_ = rt.CallDepth()
			rt.PopCall()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
