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
	g.writeln(`"strings"`)
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
		g.declaredVars[name] = true
		g.write(strings.Repeat("\t", g.indent))

		// Check variable type for proper initialization
		switch decl.Names[0].(type) {
		case *ast.ArrayVar:
			if decl.Value != nil {
				g.write(name + " := ")
				g.generateExpression(decl.Value)
			} else {
				g.write(name + " := svArray()")
			}
		case *ast.HashVar:
			if decl.Value != nil {
				// Convert array to hash
				g.write(name + " := func() *SV { _arr := ")
				g.generateExpression(decl.Value)
				g.write("; _h := svHash(); for _i := 0; _i+1 < len(_arr.av); _i += 2 { svHSet(_h, _arr.av[_i], _arr.av[_i+1]) }; return _h }()")
			} else {
				g.write(name + " := svHash()")
			}
		default:
			if decl.Value != nil {
				g.write(name + " := ")
				g.generateExpression(decl.Value)
			} else {
				g.write(name + " := svUndef()")
			}
		}
		g.write("\n")
		g.writeln("_ = " + name)
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
	g.write("for (")
	g.generateExpression(stmt.Condition)
	g.write(").IsTrue() {\n")
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

func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		g.write(fmt.Sprintf("svInt(%d)", e.Value))
	case *ast.FloatLiteral:
		g.write(fmt.Sprintf("svFloat(%f)", e.Value))
	case *ast.StringLiteral:
		if e.Interpolated {
			g.generateInterpolatedString(e.Value)
		} else {
			g.write(fmt.Sprintf("svStr(%q)", e.Value))
		}
	case *ast.ScalarVar:
		g.write(g.scalarName(e.Name))
	case *ast.ArrayVar:
		g.write(g.arrayName(e.Name))
	case *ast.HashVar:
		g.write(g.hashName(e.Name))
	case *ast.SpecialVar:
		if e.Name == "@_" {
			g.write("svArray(args...)")
		} else if e.Name == "$_" {
			g.write("v__") // default variable
		} else if len(e.Name) >= 2 && e.Name[0] == '$' && e.Name[1] >= '1' && e.Name[1] <= '9' {
			// Capture group $1, $2, ..., $99, etc.
			g.write(fmt.Sprintf("svStr(_getCapture(%s))", e.Name[1:]))
		} else {
			g.write("svUndef()")
		}
	case *ast.PrefixExpr:
		g.generatePrefixExpr(e)
	case *ast.PostfixExpr:
		g.generatePostfixExpr(e)
	case *ast.InfixExpr:
		g.generateInfixExpr(e)
	case *ast.AssignExpr:
		g.generateAssignExpr(e)
	case *ast.TernaryExpr:
		g.write("func() *SV { if (")
		g.generateExpression(e.Condition)
		g.write(").IsTrue() { return ")
		g.generateExpression(e.Then)
		g.write(" } else { return ")
		g.generateExpression(e.Else)
		g.write(" } }()")
	case *ast.CallExpr:
		g.generateCallExpr(e)
	case *ast.ArrayExpr:
		g.write("svArray(")
		for i, el := range e.Elements {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(el)
		}
		g.write(")")
	case *ast.HashExpr:
		g.tempCount++
		hvar := fmt.Sprintf("_h%d", g.tempCount)
		g.write("func() *SV { " + hvar + " := svHash(); ")
		for _, p := range e.Pairs {
			g.write("svHSet(" + hvar + ", ")
			g.generateExpression(p.Key)
			g.write(", ")
			g.generateExpression(p.Value)
			g.write("); ")
		}
		g.write("return " + hvar + " }()")
	case *ast.ArrayAccess:
		g.write("svAGet(")
		// $arr[0] means access to @arr element
		if sv, ok := e.Array.(*ast.ScalarVar); ok {
			g.write(g.arrayName(sv.Name))
		} else {
			g.generateExpression(e.Array)
		}
		g.write(", ")
		g.generateExpression(e.Index)
		g.write(")")
	case *ast.HashAccess:
		g.write("svHGet(")
		// $h{key} means access to %h element
		if sv, ok := e.Hash.(*ast.ScalarVar); ok {
			g.write(g.hashName(sv.Name))
		} else {
			g.generateExpression(e.Hash)
		}
		g.write(", ")
		g.generateExpression(e.Key)
		g.write(")")
	case *ast.ArrowAccess:
		g.generateArrowAccess(e)
	case *ast.MethodCall:
		g.generateMethodCall(e)
	case *ast.Identifier:
		g.write(fmt.Sprintf("svStr(%q)", e.Value))
	case *ast.RangeExpr:
		g.generateRangeExpr(e)
	case *ast.UndefLiteral:
		g.write("svUndef()")
	case *ast.MatchExpr:
		g.generateMatchExpr(e)
	case *ast.SubstExpr:
		g.generateSubstExpr(e)
	case *ast.ReadLineExpr:
		g.generateReadLineExpr(e)
	default:
		g.write("svUndef()")
	}
}

func (g *Generator) generatePrefixExpr(expr *ast.PrefixExpr) {
	switch expr.Operator {
	case "-":
		g.write("svNeg(")
		g.generateExpression(expr.Right)
		g.write(")")
	case "!":
		g.write("svNot(")
		g.generateExpression(expr.Right)
		g.write(")")
	case "not":
		g.write("svNot(")
		g.generateExpression(expr.Right)
		g.write(")")
	case "++":
		// Pre-increment
		if v, ok := expr.Right.(*ast.ScalarVar); ok {
			name := g.scalarName(v.Name)
			g.write("func() *SV { " + name + " = svAdd(" + name + ", svInt(1)); return " + name + " }()")
		}
	case "--":
		if v, ok := expr.Right.(*ast.ScalarVar); ok {
			name := g.scalarName(v.Name)
			g.write("func() *SV { " + name + " = svSub(" + name + ", svInt(1)); return " + name + " }()")
		}
	default:
		g.generateExpression(expr.Right)
	}
}

func (g *Generator) generatePostfixExpr(expr *ast.PostfixExpr) {
	switch expr.Operator {
	case "++":
		if v, ok := expr.Left.(*ast.ScalarVar); ok {
			name := g.scalarName(v.Name)
			g.write("func() *SV { _t := " + name + "; " + name + " = svAdd(" + name + ", svInt(1)); return _t }()")
		}
	case "--":
		if v, ok := expr.Left.(*ast.ScalarVar); ok {
			name := g.scalarName(v.Name)
			g.write("func() *SV { _t := " + name + "; " + name + " = svSub(" + name + ", svInt(1)); return _t }()")
		}
	}
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

func (g *Generator) generateInfixExpr(expr *ast.InfixExpr) {
	op := expr.Operator
	switch op {
	case "+":
		g.write("svAdd(")
	case "-":
		g.write("svSub(")
	case "*":
		g.write("svMul(")
	case "/":
		g.write("svDiv(")
	case "%":
		g.write("svMod(")
	case "**":
		g.write("svPow(")
	case ".":
		g.write("svConcat(")
	case "x":
		g.write("svRepeat(")
	case "==":
		g.write("svNumEq(")
	case "!=":
		g.write("svNumNe(")
	case "<":
		g.write("svNumLt(")
	case "<=":
		g.write("svNumLe(")
	case ">":
		g.write("svNumGt(")
	case ">=":
		g.write("svNumGe(")
	case "eq":
		g.write("svStrEq(")
	case "ne":
		g.write("svStrNe(")
	case "lt":
		g.write("svStrLt(")
	case "le":
		g.write("svStrLe(")
	case "gt":
		g.write("svStrGt(")
	case "ge":
		g.write("svStrGe(")
	case "&&", "and":
		g.write("func() *SV { if (")
		g.generateExpression(expr.Left)
		g.write(").IsTrue() { return ")
		g.generateExpression(expr.Right)
		g.write(" }; return svInt(0) }()")
		return
	case "||", "or":
		g.write("func() *SV { if _v := ")
		g.generateExpression(expr.Left)
		g.write("; _v.IsTrue() { return _v }; return ")
		g.generateExpression(expr.Right)
		g.write(" }()")
		return
	case "//":
		g.write("func() *SV { if _v := ")
		g.generateExpression(expr.Left)
		g.write("; _v != nil && _v.flags != 0 { return _v }; return ")
		g.generateExpression(expr.Right)
		g.write(" }()")
		return
	default:
		g.write("svUndef(")
	}
	g.generateExpression(expr.Left)
	g.write(", ")
	g.generateExpression(expr.Right)
	g.write(")")
}

func (g *Generator) generateAssignExpr(expr *ast.AssignExpr) {
	switch left := expr.Left.(type) {
	case *ast.ScalarVar:
		name := g.scalarName(left.Name)
		switch expr.Operator {
		case "=":
			g.write(name + " = ")
			g.generateExpression(expr.Right)
		case "+=":
			g.write(name + " = svAdd(" + name + ", ")
			g.generateExpression(expr.Right)
			g.write(")")
		case "-=":
			g.write(name + " = svSub(" + name + ", ")
			g.generateExpression(expr.Right)
			g.write(")")
		case "*=":
			g.write(name + " = svMul(" + name + ", ")
			g.generateExpression(expr.Right)
			g.write(")")
		case "/=":
			g.write(name + " = svDiv(" + name + ", ")
			g.generateExpression(expr.Right)
			g.write(")")
		case ".=":
			g.write(name + " = svConcat(" + name + ", ")
			g.generateExpression(expr.Right)
			g.write(")")
		}
	case *ast.ArrayAccess:
		g.write("svASet(")
		if sv, ok := left.Array.(*ast.ScalarVar); ok {
			g.write(g.arrayName(sv.Name))
		} else {
			g.generateExpression(left.Array)
		}
		g.write(", ")
		g.generateExpression(left.Index)
		g.write(", ")
		g.generateExpression(expr.Right)
		g.write(")")
	case *ast.HashAccess:
		g.write("svHSet(")
		if sv, ok := left.Hash.(*ast.ScalarVar); ok {
			g.write(g.hashName(sv.Name))
		} else {
			g.generateExpression(left.Hash)
		}
		g.write(", ")
		g.generateExpression(left.Key)
		g.write(", ")
		g.generateExpression(expr.Right)
		g.write(")")
	case *ast.ArrowAccess:
		// $ref->{"key"} = value or $ref->[idx] = value
		switch acc := left.Right.(type) {
		case *ast.HashAccess:
			g.write("svHSet(")
			g.generateExpression(left.Left)
			g.write(", ")
			g.generateExpression(acc.Key)
			g.write(", ")
			g.generateExpression(expr.Right)
			g.write(")")
		case *ast.ArrayAccess:
			g.write("svASet(")
			g.generateExpression(left.Left)
			g.write(", ")
			g.generateExpression(acc.Index)
			g.write(", ")
			g.generateExpression(expr.Right)
			g.write(")")
		}
	}
}

func (g *Generator) generateCallExpr(expr *ast.CallExpr) {
	if ident, ok := expr.Function.(*ast.Identifier); ok {
		name := ident.Value
		switch name {
		case "print":
			// Check if first arg is filehandle
			if len(expr.Args) >= 2 {
				if _, ok := expr.Args[0].(*ast.ScalarVar); ok {
					// print $fh "text" form
					g.write("perlPrintFH(")
					g.generateExpression(expr.Args[0])
					g.write(".AsString()")
					for _, a := range expr.Args[1:] {
						g.write(", ")
						g.generateExpression(a)
					}
					g.write(")")
					return
				}
			}
			g.write("perlPrint(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		case "say":
			// Check if first arg is filehandle
			if len(expr.Args) >= 2 {
				if _, ok := expr.Args[0].(*ast.ScalarVar); ok {
					// say $fh "text" form
					g.write("perlSayFH(")
					g.generateExpression(expr.Args[0])
					g.write(".AsString()")
					for _, a := range expr.Args[1:] {
						g.write(", ")
						g.generateExpression(a)
					}
					g.write(")")
					return
				}
			}
			g.write("perlSay(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		case "push":
			if len(expr.Args) >= 1 {
				if av, ok := expr.Args[0].(*ast.ArrayVar); ok {
					g.write("svPush(" + g.arrayName(av.Name))
					for _, a := range expr.Args[1:] {
						g.write(", ")
						g.generateExpression(a)
					}
					g.write(")")
					return
				}
			}
			g.write("svUndef()")
		case "pop":
			if len(expr.Args) >= 1 {
				if av, ok := expr.Args[0].(*ast.ArrayVar); ok {
					g.write("svPop(" + g.arrayName(av.Name) + ")")
					return
				}
			}
			g.write("svUndef()")
		case "shift":
			if len(expr.Args) >= 1 {
				if av, ok := expr.Args[0].(*ast.ArrayVar); ok {
					g.write("svShift(" + g.arrayName(av.Name) + ")")
					return
				}
			}
			g.write("svShift(_args)")
		case "unshift":
			if len(expr.Args) >= 1 {
				if av, ok := expr.Args[0].(*ast.ArrayVar); ok {
					g.write("svUnshift(" + g.arrayName(av.Name))
					for _, a := range expr.Args[1:] {
						g.write(", ")
						g.generateExpression(a)
					}
					g.write(")")
					return
				}
			}
			g.write("svUndef()")
		case "length":
			if len(expr.Args) >= 1 {
				g.write("perlLength(")
				g.generateExpression(expr.Args[0])
				g.write(")")
			} else {
				g.write("svInt(0)")
			}
		case "uc":
			g.write("perlUc(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "lc":
			g.write("perlLc(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "abs":
			g.write("perlAbs(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "int":
			g.write("perlInt(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "sqrt":
			g.write("perlSqrt(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "chr":
			g.write("perlChr(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "ord":
			g.write("perlOrd(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "scalar":
			if len(expr.Args) >= 1 {
				g.write("perl_scalar(")
				g.generateExpression(expr.Args[0])
				g.write(")")
			} else {
				g.write("svUndef()")
			}
		case "keys":
			if len(expr.Args) >= 1 {
				g.write("perl_keys(")
				g.generateExpression(expr.Args[0])
				g.write(")")
			} else {
				g.write("svArray()")
			}
		case "join":
			if len(expr.Args) >= 2 {
				g.write("perl_join(")
				g.generateExpression(expr.Args[0])
				g.write(", ")
				g.generateExpression(expr.Args[1])
				g.write(")")
			} else {
				g.write("svStr(\"\")")
			}
		case "ref":
			if len(expr.Args) >= 1 {
				g.write("perl_ref(")
				g.generateExpression(expr.Args[0])
				g.write(")")
			} else {
				g.write("svStr(\"\")")
			}
		case "open":
			if len(expr.Args) >= 2 {
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
				g.write(")")
			}
		case "close":
			if len(expr.Args) >= 1 {
				g.write("perlClose(")
				g.generateExpression(expr.Args[0])
				g.write(".AsString())")
			}
		default:
			// User-defined function
			//g.write("perl_" + name + "(")
			g.write("perl_" + strings.ReplaceAll(name, "::", "_") + "(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		}
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

func (g *Generator) generateRangeExpr(expr *ast.RangeExpr) {
	g.write("func() *SV { var _r []*SV; for _i := int(")
	g.generateExpression(expr.Start)
	g.write(".AsInt()); _i <= int(")
	g.generateExpression(expr.End)
	g.write(".AsInt()); _i++ { _r = append(_r, svInt(int64(_i))) }; return svArray(_r...) }()")
}

func (g *Generator) generateInterpolatedString(s string) {
	// Build concatenation expression
	g.write("func() *SV { var _s string; ")

	i := 0
	for i < len(s) {
		if s[i] == '$' {
			// Find variable name
			j := i + 1
			if j < len(s) && s[j] == '{' {
				// ${var}
				k := j + 1
				for k < len(s) && s[k] != '}' {
					k++
				}
				varName := s[j+1 : k]
				g.write("_s += " + g.scalarName(varName) + ".AsString(); ")
				i = k + 1
			} else {
				// $var
				for j < len(s) && (isAlnum(s[j]) || s[j] == '_') {
					j++
				}
				varName := s[i+1 : j]
				if varName != "" {
					// Check if it's a capture group $1, $2, etc.
					if len(varName) > 0 && varName[0] >= '1' && varName[0] <= '9' {
						g.write("_s += _getCapture(" + varName + "); ")
					} else {
						g.write("_s += " + g.scalarName(varName) + ".AsString(); ")
					}
				}
				i = j
			}
		} else {
			// Literal text
			j := i
			for j < len(s) && s[j] != '$' {
				j++
			}
			g.write(fmt.Sprintf("_s += %q; ", s[i:j]))
			i = j
		}
	}

	g.write("return svStr(_s) }()")
}

func (g *Generator) generateMatchExpr(expr *ast.MatchExpr) {
	pattern := expr.Pattern.Pattern
	flags := expr.Pattern.Flags

	// Add case-insensitive flag if needed
	rePattern := pattern
	if strings.Contains(flags, "i") {
		rePattern = "(?i)" + rePattern
	}

	if expr.Negate {
		g.write("func() *SV { re := regexp.MustCompile(`" + rePattern + "`); _m := re.FindStringSubmatch(")
		g.generateExpression(expr.Target)
		g.write(".AsString()); if _m != nil { _captures = _m[1:]; return svInt(0) }; return svInt(1) }()")
	} else {
		g.write("func() *SV { re := regexp.MustCompile(`" + rePattern + "`); _m := re.FindStringSubmatch(")
		g.generateExpression(expr.Target)
		g.write(".AsString()); if _m != nil { _captures = _m[1:]; return svInt(1) }; return svInt(0) }()")
	}
}

func (g *Generator) varName(expr ast.Expression) string {
	switch v := expr.(type) {
	case *ast.ScalarVar:
		return "v_" + v.Name
	case *ast.ArrayVar:
		return "a_" + v.Name
	case *ast.HashVar:
		return "h_" + v.Name
	}
	return "_"
}

func (g *Generator) scalarName(name string) string {
	return "v_" + name
}

func (g *Generator) arrayName(name string) string {
	return "a_" + name
}

func (g *Generator) hashName(name string) string {
	return "h_" + name
}

func (g *Generator) generateSubstExpr(expr *ast.SubstExpr) {
	pattern := expr.Pattern
	replacement := expr.Replacement
	flags := expr.Flags

	rePattern := pattern
	if strings.Contains(flags, "i") {
		rePattern = "(?i)" + rePattern
	}

	// Get variable name
	varName := ""
	if v, ok := expr.Target.(*ast.ScalarVar); ok {
		varName = g.scalarName(v.Name)
	}

	if strings.Contains(flags, "g") {
		// Global replace with capture support
		g.write("func() *SV { re := regexp.MustCompile(`" + rePattern + "`); ")
		g.write("_old := " + varName + ".AsString(); ")
		g.write("_new := re.ReplaceAllStringFunc(_old, func(_match string) string { ")
		g.write("_m := re.FindStringSubmatch(_match); _captures = _m[1:]; ")
		g.write("_r := `" + replacement + "`; ")
		// Replace $1, $2 etc in replacement
		g.write("for _i := len(_captures); _i >= 1; _i-- { _r = strings.ReplaceAll(_r, fmt.Sprintf(\"$%d\", _i), _getCapture(_i)) }; ")
		g.write("return _r }); ")
		g.write(varName + " = svStr(_new); ")
		g.write("if _old != _new { return svInt(1) }; return svInt(0) }()")
	} else {
		// Single replace with capture support
		g.write("func() *SV { re := regexp.MustCompile(`" + rePattern + "`); ")
		g.write("_old := " + varName + ".AsString(); ")
		g.write("_m := re.FindStringSubmatch(_old); ")
		g.write("if _m != nil { _captures = _m[1:]; ")
		g.write("_loc := re.FindStringIndex(_old); ")
		g.write("_r := `" + replacement + "`; ")
		g.write("for _i := len(_captures); _i >= 1; _i-- { _r = strings.ReplaceAll(_r, fmt.Sprintf(\"$%d\", _i), _getCapture(_i)) }; ")
		g.write(varName + " = svStr(_old[:_loc[0]] + _r + _old[_loc[1]:]); return svInt(1) }; ")
		g.write("return svInt(0) }()")
	}
}

func (g *Generator) generateReadLineExpr(expr *ast.ReadLineExpr) {
	var name string
	if expr.Filehandle != nil {
		switch fh := expr.Filehandle.(type) {
		case *ast.Identifier:
			name = fh.Value
		case *ast.ScalarVar:
			name = fh.Name // НЕ добавляем "v_" prefix!
		}
	}

	if name == "" {
		g.write("perlReadLine(\"\")")
	} else {
		g.write("perlReadLine(\"" + name + "\")")
	}
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

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
