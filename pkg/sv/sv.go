// Package sv implements Perl's scalar value (SV) type system.
// This is the foundation of the entire runtime - every value in Perl is an SV.
package sv

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"
)

// Type represents the primary type of an SV
type Type uint8

const (
	TypeUndef  Type = iota
	TypeInt         // IV - integer value
	TypeFloat       // NV - numeric value (float64)
	TypeString      // PV - pointer value (string)
	TypeRef         // RV - reference to another SV
	TypeArray       // AV - array value ([]SV)
	TypeHash        // HV - hash value (map[string]SV)
	TypeCode        // CV - code value (subroutine)
	TypeGlob        // GV - glob value (*foo)
	TypeRegex       // Compiled regex
	TypeIO          // IO handle
)

// Flags для SV - определяют какие представления валидны
type Flags uint32

const (
	FlagIOK   Flags = 1 << iota // Integer value is valid
	FlagNOK                     // Numeric (float) value is valid
	FlagPOK                     // String value is valid
	FlagROK                     // Reference value is valid
	FlagUTF8                    // String is UTF-8 encoded
	FlagRO                      // Read-only
	FlagTemp                    // Temporary (mortal)
	FlagBless                   // Blessed into a package
	FlagWeak                    // Weak reference
	FlagTied                    // Tied variable
)

// SV is the core scalar value type, similar to Perl's internal SV structure.
// Every value in Perl (scalars, array elements, hash values) is an SV.
type SV struct {
	typ    Type
	flags  Flags
	refcnt uint32

	// Cached representations - Perl keeps multiple valid at once
	iv     int64   // Integer cache
	nv     float64 // Float cache
	pv     string  // String cache (Go strings are immutable, good for us)
	pvUTF8 bool    // pv contains valid UTF-8

	// For references and complex types
	rv *SV            // Referenced SV (when TypeRef)
	av []*SV          // Array storage (when TypeArray)
	hv map[string]*SV // Hash storage (when TypeHash)

	// For blessed references
	stash string // Package name if blessed

	// For tied variables
	tiedObj *SV // The tie object

	// Magic callbacks (simplified for now)
	magic []Magic
}

// Magic represents magical behavior attached to an SV
type Magic struct {
	Type  byte
	Flags uint32
	Obj   *SV // Associated object
	Get   func(*SV) *SV
	Set   func(*SV, *SV)
}

// ============================================================
// Constructors
// ============================================================

// NewUndef creates an undefined SV
func NewUndef() *SV {
	return &SV{typ: TypeUndef, refcnt: 1}
}

// NewInt creates an integer SV
func NewInt(v int64) *SV {
	return &SV{
		typ:    TypeInt,
		flags:  FlagIOK,
		refcnt: 1,
		iv:     v,
	}
}

// NewFloat creates a float SV
func NewFloat(v float64) *SV {
	return &SV{
		typ:    TypeFloat,
		flags:  FlagNOK,
		refcnt: 1,
		nv:     v,
	}
}

// NewString creates a string SV
func NewString(v string) *SV {
	flags := FlagPOK
	if utf8.ValidString(v) {
		flags |= FlagUTF8
	}
	return &SV{
		typ:    TypeString,
		flags:  flags,
		refcnt: 1,
		pv:     v,
		pvUTF8: utf8.ValidString(v),
	}
}

// NewRef creates a reference to another SV
func NewRef(target *SV) *SV {
	if target != nil {
		target.IncRef()
	}
	return &SV{
		typ:    TypeRef,
		flags:  FlagROK,
		refcnt: 1,
		rv:     target,
	}
}

// NewArrayRef creates a reference to a new array
func NewArrayRef(elements ...*SV) *SV {
	av := &SV{
		typ:    TypeArray,
		refcnt: 1,
		av:     make([]*SV, len(elements)),
	}
	for i, el := range elements {
		if el != nil {
			el.IncRef()
		}
		av.av[i] = el
	}
	return NewRef(av)
}

// NewHashRef creates a reference to a new hash
func NewHashRef() *SV {
	hv := &SV{
		typ:    TypeHash,
		refcnt: 1,
		hv:     make(map[string]*SV),
	}
	return NewRef(hv)
}

// ============================================================
// Reference Counting
// ============================================================

// IncRef increments reference count
func (sv *SV) IncRef() {
	if sv != nil {
		atomic.AddUint32(&sv.refcnt, 1)
	}
}

// DecRef decrements reference count and frees if zero
func (sv *SV) DecRef() {
	if sv == nil {
		return
	}
	if atomic.AddUint32(&sv.refcnt, ^uint32(0)) == 0 {
		sv.free()
	}
}

// RefCount returns current reference count
func (sv *SV) RefCount() uint32 {
	if sv == nil {
		return 0
	}
	return atomic.LoadUint32(&sv.refcnt)
}

// free releases all resources held by this SV
func (sv *SV) free() {
	// Decref any referenced SVs
	if sv.rv != nil {
		sv.rv.DecRef()
		sv.rv = nil
	}
	for _, el := range sv.av {
		if el != nil {
			el.DecRef()
		}
	}
	sv.av = nil
	for _, el := range sv.hv {
		if el != nil {
			el.DecRef()
		}
	}
	sv.hv = nil

	// Clear magic
	sv.magic = nil
}

// ============================================================
// Type Checking
// ============================================================

func (sv *SV) Type() Type      { return sv.typ }
func (sv *SV) IsUndef() bool   { return sv == nil || sv.typ == TypeUndef }
func (sv *SV) IsRef() bool     { return sv != nil && sv.typ == TypeRef }
func (sv *SV) IsArray() bool   { return sv != nil && sv.typ == TypeArray }
func (sv *SV) IsHash() bool    { return sv != nil && sv.typ == TypeHash }
func (sv *SV) IsCode() bool    { return sv != nil && sv.typ == TypeCode }
func (sv *SV) IsBlessed() bool { return sv != nil && sv.flags&FlagBless != 0 }

// Deref dereferences a reference, returns nil if not a ref
func (sv *SV) Deref() *SV {
	if sv == nil || sv.typ != TypeRef {
		return nil
	}
	return sv.rv
}

// ============================================================
// Value Coercion - The Heart of Perl's Type System
// ============================================================

// AsInt returns integer value, coercing if necessary (Perl's SvIV)
func (sv *SV) AsInt() int64 {
	if sv == nil {
		return 0
	}

	// Already have valid integer?
	if sv.flags&FlagIOK != 0 {
		return sv.iv
	}

	switch sv.typ {
	case TypeUndef:
		return 0
	case TypeInt:
		return sv.iv
	case TypeFloat:
		sv.iv = int64(sv.nv)
		sv.flags |= FlagIOK
		return sv.iv
	case TypeString:
		sv.iv = stringToInt(sv.pv)
		sv.flags |= FlagIOK
		return sv.iv
	case TypeRef:
		// Reference as integer = memory address (we fake it)
		return int64(uintptr(unsafe.Pointer(sv.rv)))
	case TypeArray:
		// Array in scalar context = length
		return int64(len(sv.av))
	case TypeHash:
		// Hash in scalar context = "used/buckets" fraction, we return count
		return int64(len(sv.hv))
	default:
		return 0
	}
}

// AsFloat returns float value, coercing if necessary (Perl's SvNV)
func (sv *SV) AsFloat() float64 {
	if sv == nil {
		return 0.0
	}

	if sv.flags&FlagNOK != 0 {
		return sv.nv
	}

	switch sv.typ {
	case TypeUndef:
		return 0.0
	case TypeInt:
		sv.nv = float64(sv.iv)
		sv.flags |= FlagNOK
		return sv.nv
	case TypeFloat:
		return sv.nv
	case TypeString:
		sv.nv = stringToFloat(sv.pv)
		sv.flags |= FlagNOK
		return sv.nv
	default:
		return 0.0
	}
}

// AsString returns string value, coercing if necessary (Perl's SvPV)
func (sv *SV) AsString() string {
	if sv == nil {
		return ""
	}

	if sv.flags&FlagPOK != 0 {
		return sv.pv
	}

	switch sv.typ {
	case TypeUndef:
		return ""
	case TypeInt:
		sv.pv = strconv.FormatInt(sv.iv, 10)
		sv.flags |= FlagPOK | FlagUTF8
		sv.pvUTF8 = true
		return sv.pv
	case TypeFloat:
		sv.pv = formatFloat(sv.nv)
		sv.flags |= FlagPOK | FlagUTF8
		sv.pvUTF8 = true
		return sv.pv
	case TypeRef:
		return sv.refString()
	case TypeArray:
		return fmt.Sprintf("ARRAY(0x%x)", uintptr(unsafe.Pointer(sv)))
	case TypeHash:
		return fmt.Sprintf("HASH(0x%x)", uintptr(unsafe.Pointer(sv)))
	case TypeCode:
		return fmt.Sprintf("CODE(0x%x)", uintptr(unsafe.Pointer(sv)))
	default:
		return ""
	}
}

// AsBool returns boolean value (Perl's SvTRUE)
func (sv *SV) AsBool() bool {
	if sv == nil || sv.typ == TypeUndef {
		return false
	}

	switch sv.typ {
	case TypeInt:
		return sv.iv != 0
	case TypeFloat:
		return sv.nv != 0.0 && !math.IsNaN(sv.nv)
	case TypeString:
		// Perl: "" and "0" are false, everything else is true
		return sv.pv != "" && sv.pv != "0"
	case TypeRef:
		// References are always true
		return true
	case TypeArray:
		return len(sv.av) > 0
	case TypeHash:
		return len(sv.hv) > 0
	default:
		return false
	}
}

// refString returns the string representation of a reference
func (sv *SV) refString() string {
	if sv.rv == nil {
		return "REF(0x0)"
	}

	target := sv.rv
	prefix := ""

	if sv.flags&FlagBless != 0 {
		prefix = sv.stash + "="
	}

	switch target.typ {
	case TypeArray:
		return fmt.Sprintf("%sARRAY(0x%x)", prefix, uintptr(unsafe.Pointer(target)))
	case TypeHash:
		return fmt.Sprintf("%sHASH(0x%x)", prefix, uintptr(unsafe.Pointer(target)))
	case TypeCode:
		return fmt.Sprintf("%sCODE(0x%x)", prefix, uintptr(unsafe.Pointer(target)))
	default:
		return fmt.Sprintf("%sSCALAR(0x%x)", prefix, uintptr(unsafe.Pointer(target)))
	}
}

// ============================================================
// String-to-Number Conversion (Perl semantics)
// ============================================================

// stringToInt converts string to int with Perl semantics
// "42abc" -> 42, "abc" -> 0, "  123  " -> 123
func stringToInt(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Find numeric prefix
	end := 0
	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		end = 1
	}
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}

	if end == 0 || (end == 1 && (s[0] == '-' || s[0] == '+')) {
		return 0
	}

	v, _ := strconv.ParseInt(s[:end], 10, 64)
	return v
}

// stringToFloat converts string to float with Perl semantics
func stringToFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0
	}

	// Try parsing as float, accepting partial matches
	// This is simplified - real Perl is more complex
	end := 0
	sawDot := false
	sawE := false

	if len(s) > 0 && (s[0] == '-' || s[0] == '+') {
		end = 1
	}

	for end < len(s) {
		c := s[end]
		if c >= '0' && c <= '9' {
			end++
		} else if c == '.' && !sawDot && !sawE {
			sawDot = true
			end++
		} else if (c == 'e' || c == 'E') && !sawE && end > 0 {
			sawE = true
			end++
			if end < len(s) && (s[end] == '+' || s[end] == '-') {
				end++
			}
		} else {
			break
		}
	}

	if end == 0 {
		return 0.0
	}

	v, _ := strconv.ParseFloat(s[:end], 64)
	return v
}

// formatFloat formats a float like Perl does
func formatFloat(v float64) string {
	if math.IsInf(v, 1) {
		return "Inf"
	}
	if math.IsInf(v, -1) {
		return "-Inf"
	}
	if math.IsNaN(v) {
		return "NaN"
	}

	// If it's a whole number, format without decimal
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return strconv.FormatInt(int64(v), 10)
	}

	// Otherwise use %g style formatting
	s := strconv.FormatFloat(v, 'g', -1, 64)
	return s
}

// ============================================================
// Setters - Modify SV value
// ============================================================

// SetInt sets integer value
func (sv *SV) SetInt(v int64) {
	sv.checkWritable()
	sv.typ = TypeInt
	sv.iv = v
	sv.flags = FlagIOK
	// Invalidate other caches
	sv.pv = ""
	sv.nv = 0
}

// SetFloat sets float value
func (sv *SV) SetFloat(v float64) {
	sv.checkWritable()
	sv.typ = TypeFloat
	sv.nv = v
	sv.flags = FlagNOK
	sv.pv = ""
	sv.iv = 0
}

// SetString sets string value
func (sv *SV) SetString(v string) {
	sv.checkWritable()
	sv.typ = TypeString
	sv.pv = v
	sv.flags = FlagPOK
	if utf8.ValidString(v) {
		sv.flags |= FlagUTF8
		sv.pvUTF8 = true
	}
	sv.iv = 0
	sv.nv = 0
}

// SetUndef sets to undefined
func (sv *SV) SetUndef() {
	sv.checkWritable()
	sv.typ = TypeUndef
	sv.flags = 0
	sv.iv = 0
	sv.nv = 0
	sv.pv = ""
	if sv.rv != nil {
		sv.rv.DecRef()
		sv.rv = nil
	}
}

// SetRef sets as reference to target
func (sv *SV) SetRef(target *SV) {
	sv.checkWritable()
	if sv.rv != nil {
		sv.rv.DecRef()
	}
	sv.typ = TypeRef
	sv.flags = FlagROK
	sv.rv = target
	if target != nil {
		target.IncRef()
	}
}

func (sv *SV) checkWritable() {
	if sv.flags&FlagRO != 0 {
		panic("Modification of a read-only value attempted")
	}
}

// ============================================================
// Blessing (OOP)
// ============================================================

// Bless blesses the reference into a package
func (sv *SV) Bless(pkg string) *SV {
	if sv.typ != TypeRef {
		panic("Can't bless non-reference value")
	}
	sv.stash = pkg
	sv.flags |= FlagBless
	return sv
}

// Package returns the package name if blessed, empty string otherwise
func (sv *SV) Package() string {
	if sv.flags&FlagBless == 0 {
		return ""
	}
	return sv.stash
}

// Isa checks if this blessed reference is-a given package
// TODO: Needs stash integration for inheritance checking
func (sv *SV) Isa(pkg string) bool {
	if sv.flags&FlagBless == 0 {
		return false
	}
	return sv.stash == pkg // Simple check, full implementation needs @ISA
}

// ============================================================
// Internal Data Access (for av/hv packages)
// ============================================================

// ArrayData returns the underlying array slice
func (sv *SV) ArrayData() []*SV {
	if sv == nil || sv.typ != TypeArray {
		return nil
	}
	return sv.av
}

// SetArrayData sets the underlying array slice
func (sv *SV) SetArrayData(data []*SV) {
	if sv == nil || sv.typ != TypeArray {
		return
	}
	sv.av = data
}

// HashData returns the underlying hash map
func (sv *SV) HashData() map[string]*SV {
	if sv == nil || sv.typ != TypeHash {
		return nil
	}
	return sv.hv
}

// SetHashData sets the underlying hash map
func (sv *SV) SetHashData(data map[string]*SV) {
	if sv == nil || sv.typ != TypeHash {
		return
	}
	sv.hv = data
}

// ============================================================
// Copy and Clone
// ============================================================

// Copy creates a shallow copy of the SV value (not references)
func (sv *SV) Copy() *SV {
	if sv == nil {
		return NewUndef()
	}

	cp := &SV{
		typ:    sv.typ,
		flags:  sv.flags &^ (FlagRO | FlagTemp), // Clear RO and Temp
		refcnt: 1,
		iv:     sv.iv,
		nv:     sv.nv,
		pv:     sv.pv,
		pvUTF8: sv.pvUTF8,
		stash:  sv.stash,
	}

	// For refs, copy the reference (not deep copy)
	if sv.rv != nil {
		cp.rv = sv.rv
		cp.rv.IncRef()
	}

	// For arrays and hashes, share the underlying storage (ref semantics)
	cp.av = sv.av
	cp.hv = sv.hv

	return cp
}

// ============================================================
// Debug
// ============================================================

func (sv *SV) String() string {
	if sv == nil {
		return "SV{nil}"
	}
	return fmt.Sprintf("SV{type=%v, flags=0x%x, val=%q}", sv.typ, sv.flags, sv.AsString())
}

func (t Type) String() string {
	names := []string{"undef", "int", "float", "string", "ref", "array", "hash", "code", "glob", "regex", "io"}
	if int(t) < len(names) {
		return names[t]
	}
	return fmt.Sprintf("unknown(%d)", t)
}
