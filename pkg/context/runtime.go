// Package context implements Perl's runtime context and interpreter state.
// Paket context, Perl'in çalışma zamanı bağlamını ve yorumlayıcı durumunu uygular.
//
// This includes call stack, dynamic scoping, special variables, and error handling.
// Bu, çağrı yığını, dinamik kapsam, özel değişkenler ve hata işlemeyi içerir.
package context

import (
	"fmt"
	"os"
	"sync"

	"perlc/pkg/cv"
	"perlc/pkg/stash"
	"perlc/pkg/sv"
)

// Runtime represents the Perl interpreter state.
// Runtime, Perl yorumlayıcı durumunu temsil eder.
type Runtime struct {
	mu sync.RWMutex

	// Call stack for caller() and stack traces
	// caller() ve yığın izleri için çağrı yığını
	callStack []*StackFrame

	// Dynamic scope stack for local() variables
	// local() değişkenleri için dinamik kapsam yığını
	localStack []*LocalFrame

	// Current package
	// Geçerli paket
	curPackage string

	// Special variables (most are per-interpreter)
	// Özel değişkenler (çoğu yorumlayıcı başına)
	specials *SpecialVars

	// Hints and warnings state
	// İpuçları ve uyarı durumu
	hints *Hints

	// Error handling state
	// Hata işleme durumu
	evalError *sv.SV // $@
	osError   *sv.SV // $!
	childErr  *sv.SV // $?

	// Eval depth (for nested eval detection)
	// Eval derinliği (iç içe eval tespiti için)
	evalDepth int

	// Die/warn handlers
	// Die/warn işleyicileri
	dieHandler  *sv.SV // $SIG{__DIE__}
	warnHandler *sv.SV // $SIG{__WARN__}
}

// StackFrame represents a single call stack entry.
// StackFrame, tek bir çağrı yığını girdisini temsil eder.
type StackFrame struct {
	CV        *cv.CV   // The subroutine being called / Çağrılan altyordam
	Package   string   // Package context / Paket bağlamı
	File      string   // Source file / Kaynak dosya
	Line      int      // Source line / Kaynak satır
	Sub       string   // Subroutine name / Altyordam adı
	Args      []*sv.SV // @_ for this call / Bu çağrı için @_
	WantArray int      // Context: -1=void, 0=scalar, 1=list / Bağlam
	HasArgs   bool     // Whether @_ is available / @_ mevcut mu
	IsEval    bool     // Is this an eval frame? / Bu bir eval çerçevesi mi?
	EvalText  string   // For eval STRING / eval STRING için
}

// LocalFrame holds local() variable saves for one scope.
// LocalFrame, bir kapsam için local() değişken kayıtlarını tutar.
type LocalFrame struct {
	Saves []LocalSave
}

// LocalSave represents one local() save.
// LocalSave, bir local() kaydını temsil eder.
type LocalSave struct {
	GlobName string // Full glob name (Pkg::name) / Tam glob adı
	Slot     string // "SCALAR", "ARRAY", "HASH", "CODE" / Slot türü
	Value    *sv.SV // Saved value / Kaydedilen değer
}

// SpecialVars holds Perl's special variables.
// SpecialVars, Perl'in özel değişkenlerini tutar.
type SpecialVars struct {
	mu sync.RWMutex

	// Input/Output
	underscore *sv.SV // $_
	inputRS    *sv.SV // $/ (input record separator)
	outputRS   *sv.SV // $\ (output record separator)
	outputFS   *sv.SV // $, (output field separator)
	listSep    *sv.SV // $" (list separator)

	// Regex match results
	// Regex eşleşme sonuçları
	match     *sv.SV   // $& (entire match)
	preMath   *sv.SV   // $` (before match)
	postMatch *sv.SV   // $' (after match)
	lastParen *sv.SV   // $+ (last bracket)
	captures  []*sv.SV // $1, $2, $3... (capture groups)

	// Process info
	// Süreç bilgisi
	pid      *sv.SV // $$
	uid      *sv.SV // $
	euid     *sv.SV // $>
	gid      *sv.SV // $(
	egid     *sv.SV // $)
	progName *sv.SV // $0

	// Misc
	// Çeşitli
	subsep      *sv.SV // $; (subscript separator)
	format      *sv.SV // $~ (format name)
	accumulator *sv.SV // $^A (format accumulator)
}

// Hints holds pragma/hints state.
// Hints, pragma/ipucu durumunu tutar.
type Hints struct {
	Strict   StrictFlags
	Warnings WarningFlags
	Features FeatureFlags
	HintBits uint32
}

// StrictFlags for 'use strict'.
// 'use strict' için StrictFlags.
type StrictFlags uint8

const (
	StrictRefs StrictFlags = 1 << iota // strict 'refs'
	StrictVars                         // strict 'vars'
	StrictSubs                         // strict 'subs'
)

// WarningFlags for 'use warnings'.
// 'use warnings' için WarningFlags.
type WarningFlags uint32

const (
	WarnAll WarningFlags = 1 << iota
	WarnClosure
	WarnDeprecated
	WarnExiting
	WarnGlob
	WarnIO
	WarnMisc
	WarnNumeric
	WarnOnce
	WarnOverflow
	WarnPack
	WarnPortable
	WarnRecursion
	WarnRedefine
	WarnRegexp
	WarnSevere
	WarnSignal
	WarnSubstr
	WarnSyntax
	WarnTaint
	WarnUninitialized
	WarnUnpack
	WarnUntie
	WarnUtf8
	WarnVoid
)

// FeatureFlags for 'use feature'.
// 'use feature' için FeatureFlags.
type FeatureFlags uint32

const (
	FeatureSay FeatureFlags = 1 << iota
	FeatureState
	FeatureSwitch
	FeatureUnicode
	FeatureFC
	FeatureSignatures
)

// ============================================================
// Global Runtime Instance
// Global Runtime Örneği
// ============================================================

var (
	globalRuntime *Runtime
	runtimeOnce   sync.Once
)

// GetRuntime returns the global runtime instance.
// GetRuntime, global runtime örneğini döndürür.
func GetRuntime() *Runtime {
	runtimeOnce.Do(func() {
		globalRuntime = NewRuntime()
	})
	return globalRuntime
}

// NewRuntime creates a new runtime instance.
// NewRuntime, yeni bir runtime örneği oluşturur.
func NewRuntime() *Runtime {
	rt := &Runtime{
		curPackage: "main",
		specials:   newSpecialVars(),
		hints:      &Hints{},
		evalError:  sv.NewUndef(),
		osError:    sv.NewUndef(),
		childErr:   sv.NewUndef(),
	}
	return rt
}

func newSpecialVars() *SpecialVars {
	sp := &SpecialVars{
		underscore: sv.NewUndef(),
		inputRS:    sv.NewString("\n"),
		outputRS:   sv.NewString(""),
		outputFS:   sv.NewString(""),
		listSep:    sv.NewString(" "),
		subsep:     sv.NewString("\034"),
		pid:        sv.NewInt(int64(os.Getpid())),
		progName:   sv.NewString(os.Args[0]),
		captures:   make([]*sv.SV, 0),
	}
	return sp
}

// ============================================================
// Call Stack Management
// Çağrı Yığını Yönetimi
// ============================================================

// PushCall pushes a new call frame onto the stack.
// PushCall, yığına yeni bir çağrı çerçevesi ekler.
func (rt *Runtime) PushCall(frame *StackFrame) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.callStack = append(rt.callStack, frame)
}

// PopCall removes and returns the top call frame.
// PopCall, en üstteki çağrı çerçevesini kaldırır ve döndürür.
func (rt *Runtime) PopCall() *StackFrame {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.callStack) == 0 {
		return nil
	}

	frame := rt.callStack[len(rt.callStack)-1]
	rt.callStack = rt.callStack[:len(rt.callStack)-1]
	return frame
}

// CurrentFrame returns the current (top) call frame.
// CurrentFrame, geçerli (en üst) çağrı çerçevesini döndürür.
func (rt *Runtime) CurrentFrame() *StackFrame {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if len(rt.callStack) == 0 {
		return nil
	}
	return rt.callStack[len(rt.callStack)-1]
}

// Caller returns call frame N levels up (0 = current).
// Perl's caller() function.
//
// Caller, N seviye üstteki çağrı çerçevesini döndürür (0 = geçerli).
// Perl'in caller() fonksiyonu.
func (rt *Runtime) Caller(level int) *StackFrame {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	idx := len(rt.callStack) - 1 - level
	if idx < 0 || idx >= len(rt.callStack) {
		return nil
	}
	return rt.callStack[idx]
}

// CallDepth returns current call stack depth.
// CallDepth, geçerli çağrı yığını derinliğini döndürür.
func (rt *Runtime) CallDepth() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.callStack)
}

// StackTrace returns full stack trace as string.
// StackTrace, tam yığın izini string olarak döndürür.
func (rt *Runtime) StackTrace() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	result := ""
	for i := len(rt.callStack) - 1; i >= 0; i-- {
		f := rt.callStack[i]
		result += fmt.Sprintf("  %s::%s at %s line %d\n",
			f.Package, f.Sub, f.File, f.Line)
	}
	return result
}

// ============================================================
// Dynamic Scope (local)
// Dinamik Kapsam (local)
// ============================================================

// PushLocal creates a new local scope.
// PushLocal, yeni bir local kapsamı oluşturur.
func (rt *Runtime) PushLocal() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.localStack = append(rt.localStack, &LocalFrame{})
}

// PopLocal restores all local() variables for current scope.
// PopLocal, geçerli kapsam için tüm local() değişkenlerini geri yükler.
func (rt *Runtime) PopLocal() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.localStack) == 0 {
		return
	}

	frame := rt.localStack[len(rt.localStack)-1]
	rt.localStack = rt.localStack[:len(rt.localStack)-1]

	// Restore saved values in reverse order
	// Kaydedilen değerleri ters sırada geri yükle
	for i := len(frame.Saves) - 1; i >= 0; i-- {
		save := frame.Saves[i]
		rt.restoreLocal(save)
	}
}

// LocalScalar implements local($var) - saves current value.
// LocalScalar, local($var) uygular - geçerli değeri kaydeder.
func (rt *Runtime) LocalScalar(fullName string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.localStack) == 0 {
		rt.localStack = append(rt.localStack, &LocalFrame{})
	}

	frame := rt.localStack[len(rt.localStack)-1]
	gv := stash.Resolve(fullName)

	save := LocalSave{
		GlobName: fullName,
		Slot:     "SCALAR",
		Value:    gv.Scalar().Copy(),
	}
	frame.Saves = append(frame.Saves, save)

	// Set to undef for local
	// local için undef'e ayarla
	gv.SetScalar(sv.NewUndef())
}

// LocalArray implements local(@var).
// LocalArray, local(@var) uygular.
func (rt *Runtime) LocalArray(fullName string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.localStack) == 0 {
		rt.localStack = append(rt.localStack, &LocalFrame{})
	}

	frame := rt.localStack[len(rt.localStack)-1]
	gv := stash.Resolve(fullName)

	save := LocalSave{
		GlobName: fullName,
		Slot:     "ARRAY",
		Value:    gv.Array(),
	}
	if save.Value != nil {
		save.Value.IncRef()
	}
	frame.Saves = append(frame.Saves, save)

	// Set to empty array
	// Boş diziye ayarla
	gv.SetArray(sv.NewArrayRef().Deref())
}

// LocalHash implements local(%var).
// LocalHash, local(%var) uygular.
func (rt *Runtime) LocalHash(fullName string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.localStack) == 0 {
		rt.localStack = append(rt.localStack, &LocalFrame{})
	}

	frame := rt.localStack[len(rt.localStack)-1]
	gv := stash.Resolve(fullName)

	save := LocalSave{
		GlobName: fullName,
		Slot:     "HASH",
		Value:    gv.Hash(),
	}
	if save.Value != nil {
		save.Value.IncRef()
	}
	frame.Saves = append(frame.Saves, save)

	// Set to empty hash
	// Boş hash'e ayarla
	gv.SetHash(sv.NewHashRef().Deref())
}

func (rt *Runtime) restoreLocal(save LocalSave) {
	gv := stash.Resolve(save.GlobName)

	switch save.Slot {
	case "SCALAR":
		gv.SetScalar(save.Value)
	case "ARRAY":
		gv.SetArray(save.Value)
	case "HASH":
		gv.SetHash(save.Value)
	case "CODE":
		gv.SetCode(save.Value)
	}
}

// ============================================================
// Special Variables Access
// Özel Değişken Erişimi
// ============================================================

// Underscore returns $_.
// Underscore, $_ döndürür.
func (rt *Runtime) Underscore() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.underscore
}

// SetUnderscore sets $_.
// SetUnderscore, $_ ayarlar.
func (rt *Runtime) SetUnderscore(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.underscore = v
}

// InputRS returns $/ (input record separator).
// InputRS, $/ (girdi kayıt ayırıcı) döndürür.
func (rt *Runtime) InputRS() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.inputRS
}

// SetInputRS sets $/.
// SetInputRS, $/ ayarlar.
func (rt *Runtime) SetInputRS(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.inputRS = v
}

// OutputRS returns $\ (output record separator).
// OutputRS, $\ (çıktı kayıt ayırıcı) döndürür.
func (rt *Runtime) OutputRS() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.outputRS
}

// SetOutputRS sets $\.
// SetOutputRS, $\ ayarlar.
func (rt *Runtime) SetOutputRS(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.outputRS = v
}

// OutputFS returns $, (output field separator).
// OutputFS, $, (çıktı alan ayırıcı) döndürür.
func (rt *Runtime) OutputFS() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.outputFS
}

// ListSep returns $" (list separator for interpolation).
// ListSep, $" (interpolasyon için liste ayırıcı) döndürür.
func (rt *Runtime) ListSep() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.listSep
}

// PID returns $$.
// PID, $$ döndürür.
func (rt *Runtime) PID() *sv.SV {
	return rt.specials.pid
}

// ProgName returns $0.
// ProgName, $0 döndürür.
func (rt *Runtime) ProgName() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	return rt.specials.progName
}

// SetProgName sets $0.
// SetProgName, $0 ayarlar.
func (rt *Runtime) SetProgName(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.progName = v
}

// ============================================================
// Regex Match Variables
// Regex Eşleşme Değişkenleri
// ============================================================

// SetMatchVars sets regex match result variables.
// SetMatchVars, regex eşleşme sonuç değişkenlerini ayarlar.
func (rt *Runtime) SetMatchVars(match, preMath, postMatch string, captures []string) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()

	rt.specials.match = sv.NewString(match)
	rt.specials.preMath = sv.NewString(preMath)
	rt.specials.postMatch = sv.NewString(postMatch)

	rt.specials.captures = make([]*sv.SV, len(captures))
	for i, c := range captures {
		rt.specials.captures[i] = sv.NewString(c)
	}

	if len(captures) > 0 {
		rt.specials.lastParen = sv.NewString(captures[len(captures)-1])
	} else {
		rt.specials.lastParen = sv.NewUndef()
	}
}

// Match returns $& (entire match).
// Match, $& (tüm eşleşme) döndürür.
func (rt *Runtime) Match() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.match == nil {
		return sv.NewUndef()
	}
	return rt.specials.match
}

// PreMatch returns $` (before match).
// PreMatch, $` (eşleşmeden önce) döndürür.
func (rt *Runtime) PreMatch() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.preMath == nil {
		return sv.NewUndef()
	}
	return rt.specials.preMath
}

// PostMatch returns $' (after match).
// PostMatch, $' (eşleşmeden sonra) döndürür.
func (rt *Runtime) PostMatch() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.postMatch == nil {
		return sv.NewUndef()
	}
	return rt.specials.postMatch
}

// LastParen returns $+ (last captured group).
// LastParen, $+ (son yakalanan grup) döndürür.
func (rt *Runtime) LastParen() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.lastParen == nil {
		return sv.NewUndef()
	}
	return rt.specials.lastParen
}

// Capture returns $N (Nth capture group, 1-indexed).
// Capture, $N (N. yakalama grubu, 1-indeksli) döndürür.
func (rt *Runtime) Capture(n int) *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()

	idx := n - 1 // Convert to 0-indexed
	if idx < 0 || idx >= len(rt.specials.captures) {
		return sv.NewUndef()
	}
	return rt.specials.captures[idx]
}

// ============================================================
// Error Handling
// Hata İşleme
// ============================================================

// EvalError returns $@.
// EvalError, $@ döndürür.
func (rt *Runtime) EvalError() *sv.SV {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.evalError
}

// SetEvalError sets $@.
// SetEvalError, $@ ayarlar.
func (rt *Runtime) SetEvalError(v *sv.SV) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.evalError = v
}

// ClearEvalError clears $@ (sets to empty string).
// ClearEvalError, $@ temizler (boş string'e ayarlar).
func (rt *Runtime) ClearEvalError() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.evalError = sv.NewString("")
}

// OSError returns $!.
// OSError, $! döndürür.
func (rt *Runtime) OSError() *sv.SV {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.osError
}

// SetOSError sets $!.
// SetOSError, $! ayarlar.
func (rt *Runtime) SetOSError(err error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if err != nil {
		rt.osError = sv.NewString(err.Error())
	} else {
		rt.osError = sv.NewString("")
	}
}

// ChildError returns $?.
// ChildError, $? döndürür.
func (rt *Runtime) ChildError() *sv.SV {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.childErr
}

// SetChildError sets $? (child exit status).
// SetChildError, $? ayarlar (çocuk çıkış durumu).
func (rt *Runtime) SetChildError(code int) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.childErr = sv.NewInt(int64(code))
}

// Die implements die() - throws an exception.
// Die, die() uygular - bir istisna fırlatır.
func (rt *Runtime) Die(msg string) {
	rt.mu.Lock()
	rt.evalError = sv.NewString(msg)
	handler := rt.dieHandler
	rt.mu.Unlock()

	// Call $SIG{__DIE__} if set
	// Ayarlandıysa $SIG{__DIE__} çağır
	if handler != nil && !handler.IsUndef() {
		// TODO: Call the handler CV
	}

	// If in eval, just set $@ and return
	// eval içindeyse, sadece $@ ayarla ve dön
	if rt.evalDepth > 0 {
		return
	}

	// Otherwise, panic (will be caught by top-level)
	// Aksi halde, panic (üst seviyede yakalanacak)
	panic(PerlDie{Message: msg})
}

// Warn implements warn() - prints a warning.
// Warn, warn() uygular - bir uyarı yazdırır.
func (rt *Runtime) Warn(msg string) {
	rt.mu.RLock()
	handler := rt.warnHandler
	rt.mu.RUnlock()

	// Call $SIG{__WARN__} if set
	// Ayarlandıysa $SIG{__WARN__} çağır
	if handler != nil && !handler.IsUndef() {
		// TODO: Call the handler CV
		return
	}

	// Default: print to STDERR
	// Varsayılan: STDERR'e yazdır
	fmt.Fprintln(os.Stderr, msg)
}

// PerlDie is the panic type for die().
// PerlDie, die() için panic türüdür.
type PerlDie struct {
	Message string
}

func (e PerlDie) Error() string {
	return e.Message
}

// ============================================================
// Eval Support
// Eval Desteği
// ============================================================

// EnterEval marks entry into an eval block.
// EnterEval, bir eval bloğuna girişi işaretler.
func (rt *Runtime) EnterEval() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.evalDepth++
	rt.evalError = sv.NewString("")
}

// LeaveEval marks exit from an eval block.
// LeaveEval, bir eval bloğundan çıkışı işaretler.
func (rt *Runtime) LeaveEval() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.evalDepth > 0 {
		rt.evalDepth--
	}
}

// InEval returns true if currently inside an eval.
// InEval, şu anda eval içindeyse true döndürür.
func (rt *Runtime) InEval() bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.evalDepth > 0
}

// TryEval executes a function with eval semantics.
// Returns true if successful, false if died.
//
// TryEval, eval semantiği ile bir fonksiyon çalıştırır.
// Başarılıysa true, die olduysa false döndürür.
func (rt *Runtime) TryEval(fn func()) bool {
	rt.EnterEval()
	defer rt.LeaveEval()

	defer func() {
		if r := recover(); r != nil {
			if die, ok := r.(PerlDie); ok {
				rt.SetEvalError(sv.NewString(die.Message))
			} else {
				rt.SetEvalError(sv.NewString(fmt.Sprintf("%v", r)))
			}
		}
	}()

	fn()
	return rt.EvalError().AsString() == ""
}

// ============================================================
// Package Context
// Paket Bağlamı
// ============================================================

// Package returns the current package name.
// Package, geçerli paket adını döndürür.
func (rt *Runtime) Package() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.curPackage
}

// SetPackage sets the current package.
// SetPackage, geçerli paketi ayarlar.
func (rt *Runtime) SetPackage(pkg string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.curPackage = pkg
}

// ============================================================
// Hints and Pragmas
// İpuçları ve Pragmalar
// ============================================================

// UseStrict enables 'use strict' with given flags.
// UseStrict, verilen bayraklarla 'use strict' etkinleştirir.
func (rt *Runtime) UseStrict(flags StrictFlags) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.hints.Strict |= flags
}

// NoStrict disables strict flags.
// NoStrict, strict bayraklarını devre dışı bırakır.
func (rt *Runtime) NoStrict(flags StrictFlags) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.hints.Strict &^= flags
}

// IsStrict returns true if strict flag is set.
// IsStrict, strict bayrağı ayarlıysa true döndürür.
func (rt *Runtime) IsStrict(flag StrictFlags) bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.hints.Strict&flag != 0
}

// UseWarnings enables warnings with given flags.
// UseWarnings, verilen bayraklarla uyarıları etkinleştirir.
func (rt *Runtime) UseWarnings(flags WarningFlags) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.hints.Warnings |= flags
}

// NoWarnings disables warning flags.
// NoWarnings, uyarı bayraklarını devre dışı bırakır.
func (rt *Runtime) NoWarnings(flags WarningFlags) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.hints.Warnings &^= flags
}

// IsWarning returns true if warning flag is set.
// IsWarning, uyarı bayrağı ayarlıysa true döndürür.
func (rt *Runtime) IsWarning(flag WarningFlags) bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.hints.Warnings&flag != 0
}

// UseFeature enables feature flags.
// UseFeature, özellik bayraklarını etkinleştirir.
func (rt *Runtime) UseFeature(flags FeatureFlags) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.hints.Features |= flags
}

// HasFeature returns true if feature is enabled.
// HasFeature, özellik etkinse true döndürür.
func (rt *Runtime) HasFeature(flag FeatureFlags) bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.hints.Features&flag != 0
}

// ============================================================
// Signal Handlers
// Sinyal İşleyicileri
// ============================================================

// SetDieHandler sets $SIG{__DIE__}.
// SetDieHandler, $SIG{__DIE__} ayarlar.
func (rt *Runtime) SetDieHandler(handler *sv.SV) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.dieHandler = handler
}

// SetWarnHandler sets $SIG{__WARN__}.
// SetWarnHandler, $SIG{__WARN__} ayarlar.
func (rt *Runtime) SetWarnHandler(handler *sv.SV) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.warnHandler = handler
}

// ============================================================
// Process Info Variables
// Süreç Bilgisi Değişkenleri
// ============================================================

// UID returns $< (real uid).
// UID, $< (gerçek uid) döndürür.
func (rt *Runtime) UID() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.uid == nil {
		return sv.NewInt(int64(os.Getuid()))
	}
	return rt.specials.uid
}

// EUID returns $> (effective uid).
// EUID, $> (etkin uid) döndürür.
func (rt *Runtime) EUID() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.euid == nil {
		return sv.NewInt(int64(os.Geteuid()))
	}
	return rt.specials.euid
}

// GID returns $( (real gid).
// GID, $( (gerçek gid) döndürür.
func (rt *Runtime) GID() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.gid == nil {
		return sv.NewInt(int64(os.Getgid()))
	}
	return rt.specials.gid
}

// EGID returns $) (effective gid).
// EGID, $) (etkin gid) döndürür.
func (rt *Runtime) EGID() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.egid == nil {
		return sv.NewInt(int64(os.Getegid()))
	}
	return rt.specials.egid
}

// ============================================================
// Format Variables
// Format Değişkenleri
// ============================================================

// Format returns $~ (current format name).
// Format, $~ (geçerli format adı) döndürür.
func (rt *Runtime) Format() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.format == nil {
		return sv.NewUndef()
	}
	return rt.specials.format
}

// SetFormat sets $~.
// SetFormat, $~ ayarlar.
func (rt *Runtime) SetFormat(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.format = v
}

// Accumulator returns $^A (format accumulator).
// Accumulator, $^A (format akümülatör) döndürür.
func (rt *Runtime) Accumulator() *sv.SV {
	rt.specials.mu.RLock()
	defer rt.specials.mu.RUnlock()
	if rt.specials.accumulator == nil {
		return sv.NewString("")
	}
	return rt.specials.accumulator
}

// SetAccumulator sets $^A.
// SetAccumulator, $^A ayarlar.
func (rt *Runtime) SetAccumulator(v *sv.SV) {
	rt.specials.mu.Lock()
	defer rt.specials.mu.Unlock()
	rt.specials.accumulator = v
}
