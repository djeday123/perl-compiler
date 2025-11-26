package main

import (
	"fmt"
	"strings"
	"math"
)

var _ = fmt.Sprint
var _ = strings.Join
var _ = math.Abs

// ============ Runtime ============

type SV struct {
	iv    int64
	nv    float64
	pv    string
	av    []*SV
	hv    map[string]*SV
	flags uint8
}

const (
	SVf_IOK uint8 = 1 << iota
	SVf_NOK
	SVf_POK
	SVf_AOK
	SVf_HOK
)

func svInt(i int64) *SV { return &SV{iv: i, flags: SVf_IOK} }
func svFloat(f float64) *SV { return &SV{nv: f, flags: SVf_NOK} }
func svStr(s string) *SV { return &SV{pv: s, flags: SVf_POK} }
func svUndef() *SV { return &SV{} }
func svArray(elems ...*SV) *SV { return &SV{av: elems, flags: SVf_AOK} }
func svHash() *SV { return &SV{hv: make(map[string]*SV), flags: SVf_HOK} }

func (sv *SV) AsInt() int64 {
	if sv == nil { return 0 }
	if sv.flags&SVf_IOK != 0 { return sv.iv }
	if sv.flags&SVf_NOK != 0 { return int64(sv.nv) }
	if sv.flags&SVf_POK != 0 { 
		var i int64
		fmt.Sscanf(sv.pv, "%d", &i)
		return i
	}
	return 0
}

func (sv *SV) AsFloat() float64 {
	if sv == nil { return 0 }
	if sv.flags&SVf_NOK != 0 { return sv.nv }
	if sv.flags&SVf_IOK != 0 { return float64(sv.iv) }
	if sv.flags&SVf_POK != 0 {
		var f float64
		fmt.Sscanf(sv.pv, "%f", &f)
		return f
	}
	return 0
}

func (sv *SV) AsString() string {
	if sv == nil { return "" }
	if sv.flags&SVf_POK != 0 { return sv.pv }
	if sv.flags&SVf_IOK != 0 { return fmt.Sprintf("%d", sv.iv) }
	if sv.flags&SVf_NOK != 0 { 
		if sv.nv == float64(int64(sv.nv)) {
			return fmt.Sprintf("%d", int64(sv.nv))
		}
		return fmt.Sprintf("%g", sv.nv)
	}
	return ""
}

func (sv *SV) IsTrue() bool {
	if sv == nil { return false }
	if sv.flags&SVf_IOK != 0 { return sv.iv != 0 }
	if sv.flags&SVf_NOK != 0 { return sv.nv != 0 }
	if sv.flags&SVf_POK != 0 { return sv.pv != "" && sv.pv != "0" }
	if sv.flags&SVf_AOK != 0 { return len(sv.av) > 0 }
	if sv.flags&SVf_HOK != 0 { return len(sv.hv) > 0 }
	return false
}

func svAdd(a, b *SV) *SV { 
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv + b.iv)
	}
	return svFloat(a.AsFloat() + b.AsFloat()) 
}
func svSub(a, b *SV) *SV {
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv - b.iv)
	}
	return svFloat(a.AsFloat() - b.AsFloat())
}
func svMul(a, b *SV) *SV {
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv * b.iv)
	}
	return svFloat(a.AsFloat() * b.AsFloat())
}
func svDiv(a, b *SV) *SV { return svFloat(a.AsFloat() / b.AsFloat()) }
func svMod(a, b *SV) *SV { return svInt(a.AsInt() % b.AsInt()) }
func svPow(a, b *SV) *SV { return svFloat(math.Pow(a.AsFloat(), b.AsFloat())) }
func svConcat(a, b *SV) *SV { return svStr(a.AsString() + b.AsString()) }
func svRepeat(s, n *SV) *SV { return svStr(strings.Repeat(s.AsString(), int(n.AsInt()))) }
func svNeg(a *SV) *SV { return svFloat(-a.AsFloat()) }
func svNot(a *SV) *SV { if a.IsTrue() { return svInt(0) }; return svInt(1) }

func svNumEq(a, b *SV) *SV { if a.AsFloat() == b.AsFloat() { return svInt(1) }; return svInt(0) }
func svNumNe(a, b *SV) *SV { if a.AsFloat() != b.AsFloat() { return svInt(1) }; return svInt(0) }
func svNumLt(a, b *SV) *SV { if a.AsFloat() < b.AsFloat() { return svInt(1) }; return svInt(0) }
func svNumLe(a, b *SV) *SV { if a.AsFloat() <= b.AsFloat() { return svInt(1) }; return svInt(0) }
func svNumGt(a, b *SV) *SV { if a.AsFloat() > b.AsFloat() { return svInt(1) }; return svInt(0) }
func svNumGe(a, b *SV) *SV { if a.AsFloat() >= b.AsFloat() { return svInt(1) }; return svInt(0) }
func svStrEq(a, b *SV) *SV { if a.AsString() == b.AsString() { return svInt(1) }; return svInt(0) }
func svStrNe(a, b *SV) *SV { if a.AsString() != b.AsString() { return svInt(1) }; return svInt(0) }
func svStrLt(a, b *SV) *SV { if a.AsString() < b.AsString() { return svInt(1) }; return svInt(0) }
func svStrLe(a, b *SV) *SV { if a.AsString() <= b.AsString() { return svInt(1) }; return svInt(0) }
func svStrGt(a, b *SV) *SV { if a.AsString() > b.AsString() { return svInt(1) }; return svInt(0) }
func svStrGe(a, b *SV) *SV { if a.AsString() >= b.AsString() { return svInt(1) }; return svInt(0) }

func svAGet(arr *SV, idx *SV) *SV {
	if arr == nil || arr.flags&SVf_AOK == 0 { return svUndef() }
	i := int(idx.AsInt())
	if i < 0 { i = len(arr.av) + i }
	if i < 0 || i >= len(arr.av) { return svUndef() }
	return arr.av[i]
}

func svASet(arr *SV, idx *SV, val *SV) *SV {
	if arr == nil { return val }
	i := int(idx.AsInt())
	for len(arr.av) <= i { arr.av = append(arr.av, svUndef()) }
	arr.av[i] = val
	return val
}

func svPush(arr *SV, vals ...*SV) *SV {
	arr.av = append(arr.av, vals...)
	return svInt(int64(len(arr.av)))
}

func svPop(arr *SV) *SV {
	if len(arr.av) == 0 { return svUndef() }
	val := arr.av[len(arr.av)-1]
	arr.av = arr.av[:len(arr.av)-1]
	return val
}

func svShift(arr *SV) *SV {
	if len(arr.av) == 0 { return svUndef() }
	val := arr.av[0]
	arr.av = arr.av[1:]
	return val
}

func svUnshift(arr *SV, vals ...*SV) *SV {
	arr.av = append(vals, arr.av...)
	return svInt(int64(len(arr.av)))
}

func svHGet(h *SV, key *SV) *SV {
	if h == nil || h.hv == nil { return svUndef() }
	if v, ok := h.hv[key.AsString()]; ok { return v }
	return svUndef()
}

func svHSet(h *SV, key *SV, val *SV) *SV {
	if h.hv == nil { h.hv = make(map[string]*SV); h.flags |= SVf_HOK }
	h.hv[key.AsString()] = val
	return val
}

func perlPrint(args ...*SV) *SV {
	for _, a := range args { fmt.Print(a.AsString()) }
	return svInt(1)
}

func perlSay(args ...*SV) *SV {
	for _, a := range args { fmt.Print(a.AsString()) }
	fmt.Println()
	return svInt(1)
}

func perlLength(s *SV) *SV { return svInt(int64(len(s.AsString()))) }
func perlUc(s *SV) *SV { return svStr(strings.ToUpper(s.AsString())) }
func perlLc(s *SV) *SV { return svStr(strings.ToLower(s.AsString())) }
func perlAbs(n *SV) *SV { return svFloat(math.Abs(n.AsFloat())) }
func perlInt(n *SV) *SV { return svInt(n.AsInt()) }
func perlSqrt(n *SV) *SV { return svFloat(math.Sqrt(n.AsFloat())) }
func perlChr(n *SV) *SV { return svStr(string(rune(n.AsInt()))) }
func perlOrd(s *SV) *SV { r := []rune(s.AsString()); if len(r) > 0 { return svInt(int64(r[0])) }; return svUndef() }

// ============ End Runtime ============

func perl_factorial(args ...*SV) *SV {
	_ = args
	v_n := svShift(svArray(args...))
	_ = v_n
	if (svNumLe(v_n, svInt(1))).IsTrue() {
		return svInt(1)
	}
	return svMul(v_n, perl_factorial(svSub(v_n, svInt(1))))
	return svUndef()
}

func main() {
	perlSay(func() *SV { var _s string; _s += "Factorial calculator"; return svStr(_s) }())
	for v_i := svInt(1); (svNumLe(v_i, svInt(10))).IsTrue(); func() *SV { _t := v_i; v_i = svAdd(v_i, svInt(1)); return _t }() {
		v_f := perl_factorial(v_i)
		_ = v_f
		perlSay(func() *SV { var _s string; _s += v_i.AsString(); _s += "! = "; _s += v_f.AsString(); return svStr(_s) }())
	}
}
