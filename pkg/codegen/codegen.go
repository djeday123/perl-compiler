// Package codegen generates Go code from Perl AST.
package codegen

import (
	"fmt"
	"strings"

	"perlc/pkg/ast"
)

// Generator generates Go code from AST.
type Generator struct {
	output strings.Builder
	indent int
	//varCount  int
	tempCount    int
	declaredVars map[string]bool
}

// New creates a new Generator.
func New() *Generator {
	return &Generator{
		declaredVars: make(map[string]bool),
	}
}

// Generate generates Go code from a program.
func (g *Generator) Generate(program *ast.Program) string {
	g.output.Reset()

	// Header
	g.writeln("package main")
	g.writeln("")
	g.writeln("import (")
	g.indent++
	g.writeln(`"bufio"`)
	g.writeln(`"fmt"`)
	g.writeln(`"math"`)
	g.writeln(`"os"`)
	g.writeln(`"regexp"`)
	g.writeln(`"strconv"`)
	g.writeln(`"strings"`)
	g.writeln(`"unicode"`)
	g.indent--
	g.writeln(")")
	g.writeln("")

	// Suppress unused import errors
	g.writeln("var _ = fmt.Sprint")
	g.writeln("var _ = strings.Join")
	g.writeln("var _ = math.Abs")
	g.writeln("var _ = regexp.Compile")
	g.writeln("var _ = bufio.NewReader")
	g.writeln("var _ = os.Stdin")
	g.writeln("var _ = strconv.Atoi")
	g.writeln("var _ = unicode.ToLower")
	g.writeln("")

	// Runtime types and functions
	g.writeRuntime()

	// Collect subroutine declarations first
	var subs []*ast.SubDecl
	var stmts []ast.Statement
	for _, stmt := range program.Statements {
		if sub, ok := stmt.(*ast.SubDecl); ok {
			subs = append(subs, sub)
		} else {
			stmts = append(stmts, stmt)
		}
	}

	// Generate subroutines as Go functions
	for _, sub := range subs {
		g.generateSubDecl(sub)
		g.writeln("")
	}

	// Generate init function to register methods
	g.writeln("func init() {")
	g.indent++
	for _, sub := range subs {
		// Register each subroutine as a potential method
		funcName := "perl_" + strings.ReplaceAll(sub.Name, "::", "_")
		g.writeln(fmt.Sprintf("perl_register_method(%q, %s)", strings.ReplaceAll(sub.Name, "::", "_"), funcName))
	}
	g.indent--
	g.writeln("}")
	g.writeln("")

	// Generate main function
	g.writeln("func main() {")
	g.indent++

	for _, stmt := range stmts {
		g.generateStatement(stmt)
	}

	g.indent--
	g.writeln("}")

	return g.output.String()
}

func (g *Generator) writeRuntime() {
	g.writeln("// ============ Runtime ============")
	g.writeln("")

	// SV type
	g.writeln("type SV struct {")
	g.indent++
	g.writeln("iv    int64")
	g.writeln("nv    float64")
	g.writeln("pv    string")
	g.writeln("av    []*SV")
	g.writeln("hv    map[string]*SV")
	g.writeln("flags uint8")
	g.indent--
	g.writeln("}")
	g.writeln("")

	g.writeln("const (")
	g.indent++
	g.writeln("SVf_IOK uint8 = 1 << iota")
	g.writeln("SVf_NOK")
	g.writeln("SVf_POK")
	g.writeln("SVf_AOK")
	g.writeln("SVf_HOK")
	g.indent--
	g.writeln(")")
	g.writeln("")

	// g.writeln("var _ = bufio.NewReader") //- Move to generate

	// Constructors
	g.writeln("func svInt(i int64) *SV { return &SV{iv: i, flags: SVf_IOK} }")
	g.writeln("func svFloat(f float64) *SV { return &SV{nv: f, flags: SVf_NOK} }")
	g.writeln("func svStr(s string) *SV { return &SV{pv: s, flags: SVf_POK} }")
	g.writeln("func svUndef() *SV { return &SV{} }")
	g.writeln("func svArray(elems ...*SV) *SV { return &SV{av: elems, flags: SVf_AOK} }")
	g.writeln("func svHash() *SV { return &SV{hv: make(map[string]*SV), flags: SVf_HOK} }")
	g.writeln("")

	// Converters
	g.writeln(`func (sv *SV) AsInt() int64 {
	if sv == nil { return 0 }
	if sv.flags&SVf_IOK != 0 { return sv.iv }
	if sv.flags&SVf_NOK != 0 { return int64(sv.nv) }
	if sv.flags&SVf_POK != 0 { 
		var i int64
		fmt.Sscanf(sv.pv, "%d", &i)
		return i
	}
	return 0
}`)
	g.writeln("")

	g.writeln(`func (sv *SV) AsFloat() float64 {
	if sv == nil { return 0 }
	if sv.flags&SVf_NOK != 0 { return sv.nv }
	if sv.flags&SVf_IOK != 0 { return float64(sv.iv) }
	if sv.flags&SVf_POK != 0 {
		var f float64
		fmt.Sscanf(sv.pv, "%f", &f)
		return f
	}
	return 0
}`)
	g.writeln("")

	g.writeln(`func (sv *SV) AsString() string {
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
}`)
	g.writeln("")

	g.writeln(`func (sv *SV) IsTrue() bool {
	if sv == nil { return false }
	if sv.flags&SVf_IOK != 0 { return sv.iv != 0 }
	if sv.flags&SVf_NOK != 0 { return sv.nv != 0 }
	if sv.flags&SVf_POK != 0 { return sv.pv != "" && sv.pv != "0" }
	if sv.flags&SVf_AOK != 0 { return len(sv.av) > 0 }
	if sv.flags&SVf_HOK != 0 { return len(sv.hv) > 0 }
	return false
}`)
	g.writeln("")

	// Operations
	g.writeln(`func svAdd(a, b *SV) *SV { 
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv + b.iv)
	}
	return svFloat(a.AsFloat() + b.AsFloat()) 
}`)

	g.writeln(`func svSub(a, b *SV) *SV {
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv - b.iv)
	}
	return svFloat(a.AsFloat() - b.AsFloat())
}`)

	g.writeln(`func svMul(a, b *SV) *SV {
	if a.flags&SVf_IOK != 0 && b.flags&SVf_IOK != 0 {
		return svInt(a.iv * b.iv)
	}
	return svFloat(a.AsFloat() * b.AsFloat())
}`)

	g.writeln("func svDiv(a, b *SV) *SV { return svFloat(a.AsFloat() / b.AsFloat()) }")
	g.writeln("func svMod(a, b *SV) *SV { return svInt(a.AsInt() % b.AsInt()) }")
	g.writeln("func svPow(a, b *SV) *SV { return svFloat(math.Pow(a.AsFloat(), b.AsFloat())) }")
	g.writeln("func svConcat(a, b *SV) *SV { return svStr(a.AsString() + b.AsString()) }")
	g.writeln("func svRepeat(s, n *SV) *SV { return svStr(strings.Repeat(s.AsString(), int(n.AsInt()))) }")
	g.writeln("func svNeg(a *SV) *SV { return svFloat(-a.AsFloat()) }")
	g.writeln("func svNot(a *SV) *SV { if a.IsTrue() { return svInt(0) }; return svInt(1) }")
	g.writeln("")

	// Comparisons
	g.writeln("func svNumEq(a, b *SV) *SV { if a.AsFloat() == b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svNumNe(a, b *SV) *SV { if a.AsFloat() != b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svNumLt(a, b *SV) *SV { if a.AsFloat() < b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svNumLe(a, b *SV) *SV { if a.AsFloat() <= b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svNumGt(a, b *SV) *SV { if a.AsFloat() > b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svNumGe(a, b *SV) *SV { if a.AsFloat() >= b.AsFloat() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrEq(a, b *SV) *SV { if a.AsString() == b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrNe(a, b *SV) *SV { if a.AsString() != b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrLt(a, b *SV) *SV { if a.AsString() < b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrLe(a, b *SV) *SV { if a.AsString() <= b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrGt(a, b *SV) *SV { if a.AsString() > b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("func svStrGe(a, b *SV) *SV { if a.AsString() >= b.AsString() { return svInt(1) }; return svInt(0) }")
	g.writeln("")

	// Array ops
	g.writeln(`func svAGet(arr *SV, idx *SV) *SV {
	if arr == nil || arr.flags&SVf_AOK == 0 { return svUndef() }
	i := int(idx.AsInt())
	if i < 0 { i = len(arr.av) + i }
	if i < 0 || i >= len(arr.av) { return svUndef() }
	return arr.av[i]
}`)
	g.writeln("")

	g.writeln(`func svASet(arr *SV, idx *SV, val *SV) *SV {
	if arr == nil { return val }
	i := int(idx.AsInt())
	for len(arr.av) <= i { arr.av = append(arr.av, svUndef()) }
	arr.av[i] = val
	return val
}`)
	g.writeln("")

	g.writeln(`func svPush(arr *SV, vals ...*SV) *SV {
	arr.av = append(arr.av, vals...)
	return svInt(int64(len(arr.av)))
}`)
	g.writeln("")

	g.writeln(`func svPop(arr *SV) *SV {
	if len(arr.av) == 0 { return svUndef() }
	val := arr.av[len(arr.av)-1]
	arr.av = arr.av[:len(arr.av)-1]
	return val
}`)
	g.writeln("")

	g.writeln(`func svShift(arr *SV) *SV {
	if len(arr.av) == 0 { return svUndef() }
	val := arr.av[0]
	arr.av = arr.av[1:]
	return val
}`)
	g.writeln("")

	g.writeln(`func svUnshift(arr *SV, vals ...*SV) *SV {
	arr.av = append(vals, arr.av...)
	return svInt(int64(len(arr.av)))
}`)
	g.writeln("")

	// Hash ops
	g.writeln(`func svHGet(h *SV, key *SV) *SV {
	if h == nil || h.hv == nil { return svUndef() }
	if v, ok := h.hv[key.AsString()]; ok { return v }
	return svUndef()
}`)
	g.writeln("")

	g.writeln(`func svHSet(h *SV, key *SV, val *SV) *SV {
	if h.hv == nil { h.hv = make(map[string]*SV); h.flags |= SVf_HOK }
	h.hv[key.AsString()] = val
	return val
}`)
	g.writeln("")

	// Builtins
	g.writeln(`func perlPrint(args ...*SV) *SV {
	for _, a := range args { fmt.Print(a.AsString()) }
	return svInt(1)
}`)
	g.writeln("")

	g.writeln(`func perlSay(args ...*SV) *SV {
	for _, a := range args { fmt.Print(a.AsString()) }
	fmt.Println()
	return svInt(1)
}`)
	g.writeln("")

	g.writeln(`func perlLength(s *SV) *SV { return svInt(int64(len(s.AsString()))) }`)
	g.writeln(`func perlUc(s *SV) *SV { return svStr(strings.ToUpper(s.AsString())) }`)
	g.writeln(`func perlLc(s *SV) *SV { return svStr(strings.ToLower(s.AsString())) }`)
	g.writeln(`func perlAbs(n *SV) *SV { return svFloat(math.Abs(n.AsFloat())) }`)
	g.writeln(`func perlInt(n *SV) *SV { return svInt(n.AsInt()) }`)
	g.writeln(`func perlSqrt(n *SV) *SV { return svFloat(math.Sqrt(n.AsFloat())) }`)
	g.writeln(`func perlChr(n *SV) *SV { return svStr(string(rune(n.AsInt()))) }`)
	g.writeln(`func perlOrd(s *SV) *SV { r := []rune(s.AsString()); if len(r) > 0 { return svInt(int64(r[0])) }; return svUndef() }`)
	g.writeln("")

	g.writeln(`func perl_scalar(sv *SV) *SV {
		if sv == nil { return svInt(0) }
		if sv.flags&SVf_AOK != 0 { return svInt(int64(len(sv.av))) }
		if sv.flags&SVf_HOK != 0 { return svInt(int64(len(sv.hv))) }
		return sv
}`)
	g.writeln(`func perl_keys(h *SV) *SV {
		if h == nil || h.hv == nil { return svArray() }
		var keys []*SV
		for k := range h.hv { keys = append(keys, svStr(k)) }
		return svArray(keys...)
}`)
	g.writeln(`func perl_join(sep, arr *SV) *SV {
		if arr == nil { return svStr("") }
		var parts []string
		for _, el := range arr.av { parts = append(parts, el.AsString()) }
		return svStr(strings.Join(parts, sep.AsString()))
}`)
	g.writeln("")

	g.writeln(`// OOP Support
var _blessedPkg = make(map[*SV]string)
var _packageISA = make(map[string][]string)
var _methods = make(map[string]func(args ...*SV) *SV)

func perl_register_method(name string, fn func(args ...*SV) *SV) {
	_methods[name] = fn
}

func perl_bless(ref, class *SV) *SV {
	_blessedPkg[ref] = class.AsString()
	return ref
}

func perl_ref(sv *SV) *SV {
	if sv == nil { return svStr("") }
	if pkg, ok := _blessedPkg[sv]; ok { return svStr(pkg) }
	if sv.flags&0x80 != 0 { return svStr("SCALAR") }
	if sv.flags&SVf_AOK != 0 { return svStr("ARRAY") }
	if sv.flags&SVf_HOK != 0 { return svStr("HASH") }
	return svStr("")
}

func perl_set_isa(child *SV, parents ...*SV) *SV {
	childName := child.AsString()
	var parentNames []string
	for _, p := range parents {
		parentNames = append(parentNames, p.AsString())
	}
	_packageISA[childName] = parentNames
	return svInt(1)
}

func perl_method_call(obj *SV, method string, args ...*SV) *SV {
	var pkg string
	
	// Check if obj is a class name (string) or blessed reference
	if obj.flags&SVf_POK != 0 && _blessedPkg[obj] == "" {
		// Class method call: Point->new()
		pkg = obj.AsString()
	} else if p, ok := _blessedPkg[obj]; ok {
		// Instance method call: $obj->method()
		pkg = p
	} else {
		return svUndef()
	}
	
	// Search for method in class hierarchy
	fullArgs := append([]*SV{obj}, args...)
	return perl_find_and_call(pkg, method, fullArgs)
}

func perl_find_and_call(pkg, method string, args []*SV) *SV {
	// Try this package first
	key := pkg + "_" + method
	if fn, ok := _methods[key]; ok {
		return fn(args...)
	}
	
	// Try parent classes
	for _, parent := range _packageISA[pkg] {
		result := perl_find_and_call(parent, method, args)
		if result != nil {
			return result
		}
	}
	
	return svUndef()
}

func perl_isa(obj, class *SV) *SV {
	pkg, ok := _blessedPkg[obj]
	if !ok { return svInt(0) }
	target := class.AsString()
	if pkg == target { return svInt(1) }
	return perl_isa_check(pkg, target)
}

func perl_isa_check(pkg, target string) *SV {
	if pkg == target { return svInt(1) }
	for _, parent := range _packageISA[pkg] {
		if perl_isa_check(parent, target).IsTrue() { return svInt(1) }
	}
	return svInt(0)
}`)
	g.writeln("")
	// Regex captures
	g.writeln("var _captures []string")
	g.writeln("")
	g.writeln(`func _getCapture(n int) string {
	if n < 1 || n > len(_captures) { return "" }
	return _captures[n-1]
}`)

	g.writeln("")

	// File I/O
	// File I/O - добавь _filehandles ПЕРЕД perlOpen
	g.writeln("var _filehandles = make(map[string]*_FileHandle)")
	g.writeln("")
	g.writeln(`type _FileHandle struct {
	file    *os.File
	scanner *bufio.Scanner
	writer  *bufio.Writer
}`)
	g.writeln("")

	g.writeln(`func perlOpen(name, mode, filename string) *SV {
	var file *os.File
	var err error
	switch mode {
	case "<", "r":
		file, err = os.Open(filename)
	case ">", "w":
		file, err = os.Create(filename)
	case ">>", "a":
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	default:
		file, err = os.Open(filename)
	}
	if err != nil { return svInt(0) }
	fh := &_FileHandle{file: file}
	if mode == "<" || mode == "r" || mode == "" {
		fh.scanner = bufio.NewScanner(file)
	} else {
		fh.writer = bufio.NewWriter(file)
	}
	_filehandles[name] = fh
	return svInt(1)
}`)
	g.writeln("")
	g.writeln(`func perlClose(name string) *SV {
	if fh, ok := _filehandles[name]; ok {
		if fh.writer != nil { fh.writer.Flush() }
		fh.file.Close()
		delete(_filehandles, name)
		return svInt(1)
	}
	return svInt(0)
}`)
	g.writeln("")
	g.writeln(`func perlReadLine(name string) *SV {
	if name == "" {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() { return svStr(scanner.Text() + "\n") }
		return svUndef()
	}
	if fh, ok := _filehandles[name]; ok && fh.scanner != nil {
		if fh.scanner.Scan() { return svStr(fh.scanner.Text() + "\n") }
	}
	return svUndef()
}`)
	g.writeln("")

	g.writeln(`func perlPrintFH(fhName string, args ...*SV) *SV {
	if fh, ok := _filehandles[fhName]; ok && fh.writer != nil {
		for _, a := range args { fh.writer.WriteString(a.AsString()) }
		return svInt(1)
	}
	return svInt(0)
}`)
	g.writeln("")
	g.writeln(`func perlSayFH(fhName string, args ...*SV) *SV {
	if fh, ok := _filehandles[fhName]; ok && fh.writer != nil {
		for _, a := range args { fh.writer.WriteString(a.AsString()) }
		fh.writer.WriteString("\n")
		return svInt(1)
	}
	return svInt(0)
}`)
	g.writeln("")

	// === НАЧАЛО ПАТЧА - добавить в writeRuntime() ===

	// split
	g.writeln(`func perl_split(sep, str *SV) *SV {
	parts := strings.Split(str.AsString(), sep.AsString())
	var result []*SV
	for _, p := range parts {
		result = append(result, svStr(p))
	}
	return svArray(result...)
}`)
	g.writeln("")

	// reverse
	g.writeln(`func perl_reverse(arr *SV) *SV {
	if arr == nil || arr.flags&SVf_AOK == 0 { return svArray() }
	n := len(arr.av)
	result := make([]*SV, n)
	for i := 0; i < n; i++ {
		result[i] = arr.av[n-1-i]
	}
	return svArray(result...)
}`)
	g.writeln("")

	// sort
	g.writeln(`func perl_sort(arr *SV) *SV {
	if arr == nil || arr.flags&SVf_AOK == 0 { return svArray() }
	result := make([]*SV, len(arr.av))
	copy(result, arr.av)
	for i := 0; i < len(result)-1; i++ {
		for j := i+1; j < len(result); j++ {
			if result[i].AsString() > result[j].AsString() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return svArray(result...)
}`)
	g.writeln("")

	// values
	g.writeln(`func perl_values(h *SV) *SV {
	if h == nil || h.hv == nil { return svArray() }
	var vals []*SV
	for _, v := range h.hv { vals = append(vals, v) }
	return svArray(vals...)
}`)
	g.writeln("")

	// exists
	g.writeln(`func perl_exists(v *SV) *SV {
	if v == nil || v.flags == 0 { return svInt(0) }
	return svInt(1)
}`)
	g.writeln("")

	// delete (для хеша - нужно передавать хеш и ключ)
	g.writeln(`func perl_delete(v *SV) *SV {
	return svUndef()
}`)
	g.writeln("")

	// chomp
	g.writeln(`func perl_chomp(sv *SV) *SV {
	if sv == nil { return svInt(0) }
	s := sv.pv
	if len(s) > 0 && s[len(s)-1] == '\n' {
		sv.pv = s[:len(s)-1]
		return svInt(1)
	}
	return svInt(0)
}`)
	g.writeln("")

	// defined
	g.writeln(`func perl_defined(sv *SV) *SV {
	if sv == nil || sv.flags == 0 { return svInt(0) }
	return svInt(1)
}`)
	g.writeln("")

	g.writeln(`func svRef(sv *SV) *SV {
		return &SV{av: []*SV{sv}, flags: SVf_AOK | 0x80}
	}`)
	g.writeln("")

	g.writeln(`func svDeref(ref *SV) *SV {
		if ref != nil && len(ref.av) > 0 {
			return ref.av[0]
		}
		return svUndef()
	}`)
	g.writeln("")

	// === КОНЕЦ ПАТЧА ===

	// ============================================================
	// ПАТЧ 2: Добавить в writeHelperFunctions перед "// === КОНЕЦ ПАТЧА ==="
	// (в pkg/codegen/codegen.go)
	// ============================================================

	// index
	g.writeln(`func perl_index(str, substr *SV, args ...*SV) *SV {
	s := str.AsString()
	sub := substr.AsString()
	start := 0
	if len(args) > 0 {
		start = int(args[0].AsInt())
		if start < 0 { start = 0 }
		if start > len(s) { return svInt(-1) }
	}
	pos := strings.Index(s[start:], sub)
	if pos == -1 { return svInt(-1) }
	return svInt(int64(pos + start))
}`)
	g.writeln("")

	// rindex
	g.writeln(`func perl_rindex(str, substr *SV, args ...*SV) *SV {
	s := str.AsString()
	sub := substr.AsString()
	end := len(s)
	if len(args) > 0 {
		end = int(args[0].AsInt()) + len(sub)
		if end > len(s) { end = len(s) }
		if end < 0 { return svInt(-1) }
	}
	pos := strings.LastIndex(s[:end], sub)
	return svInt(int64(pos))
}`)
	g.writeln("")

	// lcfirst
	g.writeln(`func perl_lcfirst(sv *SV) *SV {
	s := sv.AsString()
	if len(s) == 0 { return svStr("") }
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return svStr(string(r))
}`)
	g.writeln("")

	// ucfirst
	g.writeln(`func perl_ucfirst(sv *SV) *SV {
	s := sv.AsString()
	if len(s) == 0 { return svStr("") }
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return svStr(string(r))
}`)
	g.writeln("")

	// chop
	g.writeln(`func perl_chop(sv *SV) *SV {
	if sv == nil { return svStr("") }
	s := sv.pv
	if len(s) == 0 { return svStr("") }
	r := []rune(s)
	last := string(r[len(r)-1])
	sv.pv = string(r[:len(r)-1])
	return svStr(last)
}`)
	g.writeln("")

	// sprintf
	g.writeln(`func perl_sprintf(args ...*SV) *SV {
	if len(args) == 0 { return svStr("") }
	format := args[0].AsString()
	fmtArgs := make([]interface{}, len(args)-1)
	fmtIdx := 0
	for idx, arg := range args[1:] {
		for fmtIdx < len(format) {
			if format[fmtIdx] == '%' {
				fmtIdx++
				if fmtIdx < len(format) && format[fmtIdx] == '%' {
					fmtIdx++
					continue
				}
				for fmtIdx < len(format) {
					c := format[fmtIdx]
					if c == '-' || c == '+' || c == ' ' || c == '#' || c == '0' ||
						(c >= '0' && c <= '9') || c == '.' || c == '*' {
						fmtIdx++
					} else {
						break
					}
				}
				if fmtIdx < len(format) {
					spec := format[fmtIdx]
					fmtIdx++
					switch spec {
					case 'd', 'i', 'o', 'x', 'X', 'b', 'c':
						fmtArgs[idx] = arg.AsInt()
					case 'e', 'E', 'f', 'F', 'g', 'G':
						fmtArgs[idx] = arg.AsFloat()
					default:
						fmtArgs[idx] = arg.AsString()
					}
					break
				}
			} else {
				fmtIdx++
			}
		}
		if fmtArgs[idx] == nil {
			fmtArgs[idx] = arg.AsString()
		}
	}
	return svStr(fmt.Sprintf(format, fmtArgs...))
}`)
	g.writeln("")

	// quotemeta
	g.writeln(`func perl_quotemeta(sv *SV) *SV {
	return svStr(regexp.QuoteMeta(sv.AsString()))
}`)
	g.writeln("")

	// hex
	g.writeln(`func perl_hex(sv *SV) *SV {
	s := sv.AsString()
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	v, _ := strconv.ParseInt(s, 16, 64)
	return svInt(v)
}`)
	g.writeln("")

	// oct
	g.writeln(`func perl_oct(sv *SV) *SV {
	s := strings.TrimSpace(sv.AsString())
	var v int64
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		v, _ = strconv.ParseInt(s[2:], 16, 64)
	} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		v, _ = strconv.ParseInt(s[2:], 2, 64)
	} else if strings.HasPrefix(s, "0") && len(s) > 1 {
		v, _ = strconv.ParseInt(s[1:], 8, 64)
	} else {
		v, _ = strconv.ParseInt(s, 8, 64)
	}
	return svInt(v)
}`)
	g.writeln("")

	// fc (fold case)
	g.writeln(`func perl_fc(sv *SV) *SV {
	return svStr(strings.ToLower(sv.AsString()))
}`)
	g.writeln("")

	// pack (simplified)
	g.writeln(`func perl_pack(args ...*SV) *SV {
	if len(args) == 0 { return svStr("") }
	template := args[0].AsString()
	values := args[1:]
	var buf []byte
	valIdx := 0
	for i := 0; i < len(template) && valIdx < len(values); i++ {
		ch := template[i]
		switch ch {
		case 'A', 'a':
			buf = append(buf, []byte(values[valIdx].AsString())...)
			valIdx++
		case 'C', 'c':
			buf = append(buf, byte(values[valIdx].AsInt()))
			valIdx++
		case 'Z':
			buf = append(buf, []byte(values[valIdx].AsString())...)
			buf = append(buf, 0)
			valIdx++
		}
	}
	return svStr(string(buf))
}`)
	g.writeln("")

	// unpack (simplified)
	g.writeln(`func perl_unpack(args ...*SV) *SV {
	if len(args) < 2 { return svArray() }
	template := args[0].AsString()
	data := []byte(args[1].AsString())
	var results []*SV
	offset := 0
	for i := 0; i < len(template) && offset < len(data); i++ {
		ch := template[i]
		// Check for count
		count := 1
		if i+1 < len(template) && template[i+1] >= '0' && template[i+1] <= '9' {
			countStr := ""
			for i+1 < len(template) && template[i+1] >= '0' && template[i+1] <= '9' {
				i++
				countStr += string(template[i])
			}
			count, _ = strconv.Atoi(countStr)
		}
		switch ch {
		case 'A', 'a':
			end := offset + count
			if end > len(data) { end = len(data) }
			results = append(results, svStr(string(data[offset:end])))
			offset = end
		case 'C', 'c':
			for c := 0; c < count && offset < len(data); c++ {
				results = append(results, svInt(int64(data[offset])))
				offset++
			}
		case 'Z':
			end := offset
			for end < len(data) && data[end] != 0 { end++ }
			results = append(results, svStr(string(data[offset:end])))
			offset = end + 1
		}
	}
	return svArray(results...)
}`)
	g.writeln("")

	// wantarray
	g.writeln(`func perl_wantarray() *SV {
		return svUndef()
	}`)
	g.writeln("")

	// printf
	g.writeln(`func perl_printf(args ...*SV) *SV {
		if len(args) == 0 { return svInt(0) }
		format := args[0].AsString()
		fmtArgs := make([]interface{}, len(args)-1)
		for i, arg := range args[1:] {
			fmtArgs[i] = arg.AsString()
		}
		n, _ := fmt.Printf(format, fmtArgs...)
		return svInt(int64(n))
	}`)
	g.writeln("")

	// each
	g.writeln(`var _hashIterators = make(map[*SV][]string)`)
	g.writeln("")

	g.writeln(`func perl_each(h *SV) *SV {
		if h == nil || h.hv == nil { return svArray() }
		
		// Получаем или создаём список ключей для итерации
		keys, ok := _hashIterators[h]
		if !ok || len(keys) == 0 {
			keys = make([]string, 0, len(h.hv))
			for k := range h.hv {
				keys = append(keys, k)
			}
			_hashIterators[h] = keys
		}
		
		// Если ключи закончились - сбрасываем
		if len(keys) == 0 {
			delete(_hashIterators, h)
			return svArray()
		}
		
		// Берём первый ключ
		k := keys[0]
		_hashIterators[h] = keys[1:]
		
		return svArray(svStr(k), h.hv[k])
	}`)

	// pos
	g.writeln(`func perl_pos(sv *SV) *SV {
		return svUndef()
	}`)
	g.writeln("")

	// grep
	g.writeln(`func perl_grep(block func(*SV) *SV, arr *SV) *SV {
		if arr == nil { return svArray() }
		var results []*SV
		for _, el := range arr.av {
			if block(el).IsTrue() {
				results = append(results, el)
			}
		}
		return svArray(results...)
	}`)
	g.writeln("")

	// map
	g.writeln(`func perl_map(block func(*SV) *SV, arr *SV) *SV {
		if arr == nil { return svArray() }
		var results []*SV
		for _, el := range arr.av {
			results = append(results, block(el))
		}
		return svArray(results...)
	}`)
	g.writeln("")

	// //grep - упрощённая версия без блока (пока не поддерживается в codegen)
	// g.writeln(`func perl_grep(block, arr *SV) *SV {
	// 	if arr == nil { return svArray() }
	// 	// block игнорируется - возвращаем пустой массив
	// 	// TODO: implement proper block support in codegen
	// 	return svArray()
	// }`)
	// g.writeln("")

	// // map - упрощённая версия
	// g.writeln(`func perl_map(block, arr *SV) *SV {
	// 	if arr == nil { return svArray() }
	// 	// block игнорируется - возвращаем пустой массив
	// 	// TODO: implement proper block support in codegen
	// 	return svArray()
	// }`)
	// g.writeln("")

	// eof - без аргументов или с SV
	g.writeln(`func perl_eof(args ...*SV) *SV {
		if len(args) == 0 { return svInt(1) }
		name := args[0].AsString()
		if fh, ok := _filehandles[name]; ok && fh.file != nil {
			pos, _ := fh.file.Seek(0, 1)
			fi, _ := fh.file.Stat()
			if pos >= fi.Size() { return svInt(1) }
			return svInt(0)
		}
		return svInt(1)
	}`)
	g.writeln("")

	// tell
	g.writeln(`func perl_tell(args ...*SV) *SV {
		if len(args) == 0 { return svInt(-1) }
		name := args[0].AsString()
		if fh, ok := _filehandles[name]; ok && fh.file != nil {
			pos, _ := fh.file.Seek(0, 1)
			return svInt(pos)
		}
		return svInt(-1)
	}`)
	g.writeln("")

	// seek
	g.writeln(`func perl_seek(fh, pos, whence *SV) *SV {
		name := fh.AsString()
		if h, ok := _filehandles[name]; ok && h.file != nil {
			_, err := h.file.Seek(pos.AsInt(), int(whence.AsInt()))
			if err == nil {
				if h.scanner != nil {
					h.scanner = bufio.NewScanner(h.file)
				}
				return svInt(1)
			}
		}
		return svInt(0)
	}`)
	g.writeln("")

	// read
	g.writeln(`func perl_read(fh, buf, length *SV) *SV {
		name := fh.AsString()
		if h, ok := _filehandles[name]; ok && h.file != nil {
			data := make([]byte, length.AsInt())
			n, _ := h.file.Read(data)
			buf.pv = string(data[:n])
			buf.flags = SVf_POK
			return svInt(int64(n))
		}
		return svInt(0)
	}`)
	g.writeln("")

	// binmode
	g.writeln(`func perl_binmode(args ...*SV) *SV {
		return svInt(1)
	}`)
	g.writeln("")

	g.writeln("// ============ End Runtime ============")
	g.writeln("")
}

func (g *Generator) write(s string) {
	g.output.WriteString(s)
}

func (g *Generator) writeln(s string) {
	for i := 0; i < g.indent; i++ {
		g.output.WriteString("\t")
	}
	g.output.WriteString(s)
	g.output.WriteString("\n")
}

func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		// Special handling for open() to declare filehandle variable
		if call, ok := s.Expression.(*ast.CallExpr); ok {
			if ident, ok := call.Function.(*ast.Identifier); ok && ident.Value == "open" {
				g.generateOpenStatement(call)
				return
			}
		}
		g.write(strings.Repeat("\t", g.indent))
		g.generateExpression(s.Expression)
		g.write("\n")
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.IfStmt:
		g.generateIfStmt(s)
	case *ast.WhileStmt:
		g.generateWhileStmt(s)
	case *ast.ForStmt:
		g.generateForStmt(s)
	case *ast.ForeachStmt:
		g.generateForeachStmt(s)
	case *ast.BlockStmt:
		g.generateBlockStmt(s)
	case *ast.ReturnStmt:
		g.generateReturnStmt(s)
	case *ast.LastStmt:
		g.writeln("break")
	case *ast.NextStmt:
		g.writeln("continue")
	case *ast.SubDecl:
		// Already handled at top level
	case *ast.UseDecl:
		// Ignore for now
	case *ast.PackageDecl:
		// Ignore for now
	}
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	// Handle list assignment: my ($a, $b) = @_
	if decl.IsList && decl.Value != nil {
		// Check if assigning from @_ (can be ArrayVar or SpecialVar)
		isArgsAssign := false
		if av, ok := decl.Value.(*ast.ArrayVar); ok && av.Name == "_" {
			isArgsAssign = true
		}
		if sv, ok := decl.Value.(*ast.SpecialVar); ok && sv.Name == "@_" {
			isArgsAssign = true
		}

		if isArgsAssign {
			// Unpack from args
			for i, v := range decl.Names {
				name := g.varName(v)
				g.declaredVars[name] = true
				g.write(strings.Repeat("\t", g.indent))
				g.write(fmt.Sprintf("%s := func() *SV { if %d < len(args) { return args[%d] }; return svUndef() }()\n", name, i, i))
				g.writeln("_ = " + name)
			}
			return
		}
		// Other list assignments - generate temp array and unpack
		g.tempCount++
		tmpVar := fmt.Sprintf("_tmp%d", g.tempCount)
		g.write(strings.Repeat("\t", g.indent))
		g.write(tmpVar + " := ")
		g.generateExpression(decl.Value)
		g.write("\n")
		for i, v := range decl.Names {
			name := g.varName(v)
			g.declaredVars[name] = true
			g.write(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("%s := svAGet(%s, svInt(%d))\n", name, tmpVar, i))
			g.writeln("_ = " + name)
		}
		return
	}

	if len(decl.Names) == 1 {
		name := g.varName(decl.Names[0])
		g.write(strings.Repeat("\t", g.indent))

		// Определяем оператор: := для нового, = для уже объявленного
		op := " := "
		if g.declaredVars[name] {
			op = " = "
		} else {
			g.declaredVars[name] = true
		}

		// Check variable type for proper initialization
		switch decl.Names[0].(type) {
		case *ast.ArrayVar:
			if decl.Value != nil {
				g.write(name + op)
				g.generateExpression(decl.Value)
			} else {
				g.write(name + op + "svArray()")
			}
		case *ast.HashVar:
			if decl.Value != nil {
				// Convert array to hash
				g.write(name + op + "func() *SV { _arr := ")
				g.generateExpression(decl.Value)
				g.write("; _h := svHash(); for _i := 0; _i+1 < len(_arr.av); _i += 2 { svHSet(_h, _arr.av[_i], _arr.av[_i+1]) }; return _h }()")
			} else {
				g.write(name + op + "svHash()")
			}
		default:
			if decl.Value != nil {
				g.write(name + op)
				g.generateExpression(decl.Value)
			} else {
				g.write(name + op + "svUndef()")
			}
		}
		g.write("\n")
		// _ = name только для новых переменных
		if op == " := " {
			g.writeln("_ = " + name)
		}
		return
	}

	for _, v := range decl.Names {
		name := g.varName(v)
		g.declaredVars[name] = true
		g.write(strings.Repeat("\t", g.indent))
		g.write(name + " := svUndef()")
		g.write("\n")
		g.writeln("_ = " + name)
	}
}

func (g *Generator) generateSubDecl(sub *ast.SubDecl) {
	// Очищаем declaredVars для нового scope функции
	g.declaredVars = make(map[string]bool)

	g.write("func perl_" + strings.ReplaceAll(sub.Name, "::", "_") + "(args ...*SV) *SV {\n")
	g.indent++
	g.writeln("_ = args")
	g.writeln("_args := svArray(args...)") // Создаём один массив для @_
	g.writeln("_ = _args")                 // Предотвращаем ошибку "declared and not used"

	// Generate body
	for _, stmt := range sub.Body.Statements {
		g.generateStatement(stmt)
	}

	g.writeln("return svUndef()")
	g.indent--
	g.writeln("}")
}

func (g *Generator) generateIfStmt(stmt *ast.IfStmt) {
	g.write(strings.Repeat("\t", g.indent))
	if stmt.Unless {
		g.write("if !(")
	} else {
		g.write("if (")
	}
	g.generateExpression(stmt.Condition)
	g.write(").IsTrue() {\n")
	g.indent++
	for _, s := range stmt.Then.Statements {
		g.generateStatement(s)
	}
	g.indent--

	for _, elsif := range stmt.Elsif {
		g.write(strings.Repeat("\t", g.indent))
		g.write("} else if (")
		g.generateExpression(elsif.Condition)
		g.write(").IsTrue() {\n")
		g.indent++
		for _, s := range elsif.Body.Statements {
			g.generateStatement(s)
		}
		g.indent--
	}

	if stmt.Else != nil {
		g.writeln("} else {")
		g.indent++
		for _, s := range stmt.Else.Statements {
			g.generateStatement(s)
		}
		g.indent--
	}
	g.writeln("}")
}

func (g *Generator) generateWhileStmt(stmt *ast.WhileStmt) {
	g.write(strings.Repeat("\t", g.indent))
	if stmt.Until {
		// until = пока НЕ выполняется условие
		g.write("for !(")
		g.generateExpression(stmt.Condition)
		g.write(").IsTrue() {\n")
	} else {
		// while = пока выполняется условие
		g.write("for (")
		g.generateExpression(stmt.Condition)
		g.write(").IsTrue() {\n")
	}
	g.indent++
	for _, s := range stmt.Body.Statements {
		g.generateStatement(s)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) generateForStmt(stmt *ast.ForStmt) {
	g.write(strings.Repeat("\t", g.indent))
	g.write("for ")

	// Init
	if stmt.Init != nil {
		if decl, ok := stmt.Init.(*ast.VarDecl); ok && len(decl.Names) > 0 {
			name := g.varName(decl.Names[0])
			g.write(name + " := ")
			if decl.Value != nil {
				g.generateExpression(decl.Value)
			} else {
				g.write("svUndef()")
			}
		}
	}
	g.write("; ")

	// Condition
	if stmt.Condition != nil {
		g.write("(")
		g.generateExpression(stmt.Condition)
		g.write(").IsTrue()")
	}
	g.write("; ")

	// Post
	if stmt.Post != nil {
		g.generateExpression(stmt.Post)
	}

	g.write(" {\n")
	g.indent++
	for _, s := range stmt.Body.Statements {
		g.generateStatement(s)
	}
	g.indent--
	g.writeln("}")
}
func (g *Generator) generateForeachStmt(stmt *ast.ForeachStmt) {
	iterVar := g.varName(stmt.Variable)
	g.tempCount++
	listVar := fmt.Sprintf("_list%d", g.tempCount)
	idxVar := fmt.Sprintf("_i%d", g.tempCount)

	g.write(strings.Repeat("\t", g.indent))
	g.write(listVar + " := ")
	g.generateExpression(stmt.List)
	g.write("\n")

	g.writeln(fmt.Sprintf("for %s := 0; %s < len(%s.av); %s++ {", idxVar, idxVar, listVar, idxVar))
	g.indent++
	g.writeln(fmt.Sprintf("%s := %s.av[%s]", iterVar, listVar, idxVar))
	g.writeln("_ = " + iterVar)
	for _, s := range stmt.Body.Statements {
		g.generateStatement(s)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) generateBlockStmt(stmt *ast.BlockStmt) {
	g.writeln("{")
	g.indent++
	for _, s := range stmt.Statements {
		g.generateStatement(s)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) generateReturnStmt(stmt *ast.ReturnStmt) {
	g.write(strings.Repeat("\t", g.indent))
	g.write("return ")
	if stmt.Value != nil {
		g.generateExpression(stmt.Value)
	} else {
		g.write("svUndef()")
	}
	g.write("\n")
}

func (g *Generator) generateMethodCall(e *ast.MethodCall) {
	g.write("perl_method_call(")
	g.generateExpression(e.Object)
	g.write(fmt.Sprintf(", %q", e.Method))
	for _, arg := range e.Args {
		g.write(", ")
		g.generateExpression(arg)
	}
	g.write(")")
}

func (g *Generator) GenerateMethodCall_test(e *ast.MethodCall) {
	// Get the class/object
	// For Class->method(): Object is Identifier with class name
	// For $obj->method(): Object is ScalarVar

	var className string
	var isClassMethod bool

	switch obj := e.Object.(type) {
	case *ast.Identifier:
		// Class->method() - class method call
		className = obj.Value
		isClassMethod = true
	case *ast.ScalarVar:
		// $obj->method() - instance method call
		isClassMethod = false
	}

	// Generate function call
	methodName := strings.ReplaceAll(e.Method, "::", "_")

	if isClassMethod {
		// Class->new() becomes perl_Class_new(svStr("Class"), args...)
		g.write("perl_" + strings.ReplaceAll(className, "::", "_") + "_" + methodName + "(")
		g.write(fmt.Sprintf("svStr(%q)", className))
		for _, arg := range e.Args {
			g.write(", ")
			g.generateExpression(arg)
		}
		g.write(")")
	} else {
		// $obj->method() - need to look up method based on blessed package
		// For simplicity, we'll need runtime method dispatch
		// For now, generate direct call if we know the type
		g.write("perl_method_call(")
		g.generateExpression(e.Object)
		g.write(fmt.Sprintf(", %q", e.Method))
		for _, arg := range e.Args {
			g.write(", ")
			g.generateExpression(arg)
		}
		g.write(")")
	}
}

func (g *Generator) generateArrowAccess(expr *ast.ArrowAccess) {
	switch right := expr.Right.(type) {
	case *ast.ArrayAccess:
		g.write("svAGet(")
		g.generateExpression(expr.Left)
		g.write(", ")
		g.generateExpression(right.Index)
		g.write(")")
	case *ast.HashAccess:
		g.write("svHGet(")
		g.generateExpression(expr.Left)
		g.write(", ")
		g.generateExpression(right.Key)
		g.write(")")
	default:
		g.generateExpression(expr.Left)
	}
}

func (g *Generator) generateInterpolatedString(s string) {
	g.write("func() *SV { var _s string; ")

	i := 0
	for i < len(s) {
		if s[i] == '$' {
			j := i + 1

			// ${var}
			if j < len(s) && s[j] == '{' {
				k := j + 1
				for k < len(s) && s[k] != '}' {
					k++
				}
				varName := s[j+1 : k]
				g.write("_s += " + g.scalarName(varName) + ".AsString(); ")
				i = k + 1
				continue
			}

			// $var[idx] - элемент массива
			// Сначала читаем имя переменной
			for j < len(s) && (isAlnum(s[j]) || s[j] == '_') {
				j++
			}
			varName := s[i+1 : j]

			if varName != "" && j < len(s) && s[j] == '[' {
				// Это $arr[idx]
				k := j + 1
				for k < len(s) && s[k] != ']' {
					k++
				}
				idxStr := s[j+1 : k]
				g.write("_s += svAGet(" + g.arrayName(varName) + ", svInt(" + idxStr + ")).AsString(); ")
				i = k + 1
				continue
			}

			if varName != "" && j < len(s) && s[j] == '{' {
				// Это $hash{key}
				k := j + 1
				for k < len(s) && s[k] != '}' {
					k++
				}
				keyStr := s[j+1 : k]
				g.write("_s += svHGet(" + g.hashName(varName) + ", svStr(\"" + keyStr + "\")).AsString(); ")
				i = k + 1
				continue
			}

			// Простая переменная $var
			if varName != "" {
				// Capture group $1, $2, etc.
				if len(varName) > 0 && varName[0] >= '1' && varName[0] <= '9' {
					g.write("_s += _getCapture(" + varName + "); ")
				} else {
					g.write("_s += " + g.scalarName(varName) + ".AsString(); ")
				}
			}
			i = j
		} else if s[i] == '@' {
			// @array
			j := i + 1
			for j < len(s) && (isAlnum(s[j]) || s[j] == '_') {
				j++
			}
			varName := s[i+1 : j]
			if varName != "" {
				g.write("_s += func() string { var _parts []string; for _, _el := range " + g.arrayName(varName) + ".av { _parts = append(_parts, _el.AsString()) }; return strings.Join(_parts, \" \") }(); ")
			}
			i = j
		} else {
			// Literal text
			j := i
			for j < len(s) && s[j] != '$' && s[j] != '@' {
				j++
			}
			g.write(fmt.Sprintf("_s += %q; ", s[i:j]))
			i = j
		}
	}

	g.write("return svStr(_s) }()")
}

func (g *Generator) generateOpenStatement(expr *ast.CallExpr) {
	if len(expr.Args) < 2 {
		return
	}

	// Declare or assign filehandle variable
	if sv, ok := expr.Args[0].(*ast.ScalarVar); ok {
		name := g.scalarName(sv.Name)
		if !g.declaredVars[name] {
			g.writeln(name + " := svStr(\"" + sv.Name + "\")")
			g.writeln("_ = " + name)
			g.declaredVars[name] = true
		} else {
			g.writeln(name + " = svStr(\"" + sv.Name + "\")")
		}
	}

	// Call perlOpen
	g.write(strings.Repeat("\t", g.indent))
	g.write("perlOpen(")
	g.generateExpression(expr.Args[0])
	g.write(".AsString(), ")
	g.generateExpression(expr.Args[1])
	g.write(".AsString(), ")
	if len(expr.Args) >= 3 && expr.Args[2] != nil {
		g.generateExpression(expr.Args[2])
		g.write(".AsString()")
	} else {
		g.write("\"\"")
	}
	g.write(")\n")
}
