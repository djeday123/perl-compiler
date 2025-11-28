// Package eval - builtin functions
package eval

import (
	"fmt"
	"math"
	"os"
	"strings"

	"perlc/pkg/ast"
	"perlc/pkg/av"
	"perlc/pkg/hv"
	"perlc/pkg/sv"
)

func (i *Interpreter) builtinPrint(expr *ast.CallExpr) *sv.SV {
	// Check if first arg is filehandle
	if len(expr.Args) >= 2 {
		if fhVar, ok := expr.Args[0].(*ast.ScalarVar); ok {
			fhName := i.ctx.GetVar(fhVar.Name)
			if fhName != nil {
				fh := i.ctx.GetFileHandle(fhName.AsString())
				if fh != nil && fh.Writer != nil {
					for _, arg := range expr.Args[1:] {
						val := i.evalExpression(arg)
						fh.Writer.WriteString(val.AsString())
					}
					return sv.NewInt(1)
				}
			}
		}
	}
	// Normal print to stdout
	for _, arg := range expr.Args {
		val := i.evalExpression(arg)
		fmt.Fprint(i.stdout, val.AsString())
	}
	return sv.NewInt(1)
}

func (i *Interpreter) builtinSay(expr *ast.CallExpr) *sv.SV {
	// Check if first arg is filehandle
	if len(expr.Args) >= 2 {
		if fhVar, ok := expr.Args[0].(*ast.ScalarVar); ok {
			fhName := i.ctx.GetVar(fhVar.Name)
			if fhName != nil {
				fh := i.ctx.GetFileHandle(fhName.AsString())
				if fh != nil && fh.Writer != nil {
					for _, arg := range expr.Args[1:] {
						val := i.evalExpression(arg)
						fh.Writer.WriteString(val.AsString())
					}
					fh.Writer.WriteString("\n")
					return sv.NewInt(1)
				}
			}
		}
	}
	// Normal say to stdout
	for _, arg := range expr.Args {
		val := i.evalExpression(arg)
		fmt.Fprint(i.stdout, val.AsString())
	}
	fmt.Fprintln(i.stdout)
	return sv.NewInt(1)
}

func (i *Interpreter) builtinOpen(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) < 2 {
		return sv.NewInt(0)
	}

	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value
	}

	mode := i.evalExpression(expr.Args[1]).AsString()
	var filename string

	if len(expr.Args) >= 3 && expr.Args[2] != nil {
		filename = i.evalExpression(expr.Args[2]).AsString()
	} else {
		// 2-arg form: extract filename from mode
		if len(mode) > 0 {
			switch mode[0] {
			case '<':
				filename = strings.TrimSpace(mode[1:])
				mode = "<"
			case '>':
				if len(mode) > 1 && mode[1] == '>' {
					filename = strings.TrimSpace(mode[2:])
					mode = ">>"
				} else {
					filename = strings.TrimSpace(mode[1:])
					mode = ">"
				}
			}
		}
	}

	err := i.ctx.OpenFile(fhName, mode, filename)
	if err != nil {
		return sv.NewInt(0)
	}
	i.ctx.SetVar(fhName, sv.NewString(fhName))
	return sv.NewInt(1)
}

func (i *Interpreter) builtinClose(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) < 1 {
		return sv.NewInt(0)
	}

	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value
	default:
		fhName = i.evalExpression(expr.Args[0]).AsString()
	}

	err := i.ctx.CloseFile(fhName)
	if err != nil {
		return sv.NewInt(0)
	}
	return sv.NewInt(1)
}

func (i *Interpreter) builtinPush(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	if len(exprs) < 2 {
		return sv.NewInt(0)
	}

	// Get the array variable
	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		for _, val := range args[1:] {
			av.Push(arrSV, val)
		}
		return av.Len(arrSV)
	}
	return sv.NewInt(0)
}

func (i *Interpreter) builtinPop(exprs []ast.Expression) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewUndef()
	}

	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		return av.Pop(arrSV)
	}
	return sv.NewUndef()
}

func (i *Interpreter) builtinShift(exprs []ast.Expression) *sv.SV {
	if len(exprs) == 0 {
		// shift without args shifts @_
		args := i.ctx.GetArgs()
		return av.Shift(args)
	}

	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		return av.Shift(arrSV)
	}
	return sv.NewUndef()
}

func (i *Interpreter) builtinUnshift(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	if len(exprs) < 2 {
		return sv.NewInt(0)
	}

	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		return av.Unshift(arrSV, args[1:]...)
	}
	return sv.NewInt(0)
}

func (i *Interpreter) builtinKeys(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewArrayRef()
	}
	keys := hv.Keys(args[0])
	return sv.NewArrayRef(keys...)
}

func (i *Interpreter) builtinValues(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewArrayRef()
	}
	vals := hv.Values(args[0])
	return sv.NewArrayRef(vals...)
}

func (i *Interpreter) builtinJoin(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewString("")
	}
	return av.Join(args[0], args[1])
}

func (i *Interpreter) builtinSplit(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewArrayRef()
	}
	pattern := args[0].AsString()
	str := args[1].AsString()
	parts := strings.Split(str, pattern)
	elements := make([]*sv.SV, len(parts))
	for idx, p := range parts {
		elements[idx] = sv.NewString(p)
	}
	return sv.NewArrayRef(elements...)
}

func (i *Interpreter) builtinSubstr(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewUndef()
	}
	var length *sv.SV
	if len(args) >= 3 {
		length = args[2]
	}
	return sv.Substr(args[0], args[1], length)
}

func (i *Interpreter) builtinAbs(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewFloat(0)
	}
	return sv.NewFloat(math.Abs(args[0].AsFloat()))
}

func (i *Interpreter) builtinSqrt(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewFloat(0)
	}
	return sv.NewFloat(math.Sqrt(args[0].AsFloat()))
}

func (i *Interpreter) builtinChr(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}
	return sv.NewString(string(rune(args[0].AsInt())))
}

func (i *Interpreter) builtinOrd(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewUndef()
	}
	s := args[0].AsString()
	if len(s) == 0 {
		return sv.NewUndef()
	}
	return sv.NewInt(int64([]rune(s)[0]))
}

func (i *Interpreter) builtinChomp(exprs []ast.Expression) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewInt(0)
	}

	count := int64(0)
	for _, expr := range exprs {
		if v, ok := expr.(*ast.ScalarVar); ok {
			val := i.ctx.GetVar(v.Name)
			s := val.AsString()
			if strings.HasSuffix(s, "\n") {
				s = strings.TrimSuffix(s, "\n")
				i.ctx.SetVar(v.Name, sv.NewString(s))
				count++
			}
		}
	}
	return sv.NewInt(count)
}

func (i *Interpreter) builtinDie(args []*sv.SV) *sv.SV {
	msg := ""
	for _, arg := range args {
		msg += arg.AsString()
	}
	if msg == "" {
		msg = "Died"
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(i.stderr, msg)
	os.Exit(1)
	return sv.NewUndef()
}

func (i *Interpreter) builtinWarn(args []*sv.SV) *sv.SV {
	msg := ""
	for _, arg := range args {
		msg += arg.AsString()
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(i.stderr, msg)
	return sv.NewInt(1)
}

func (i *Interpreter) builtinExit(args []*sv.SV) *sv.SV {
	code := 0
	if len(args) > 0 {
		code = int(args[0].AsInt())
	}
	os.Exit(code)
	return sv.NewUndef()
}

func (i *Interpreter) builtinScalar(args []*sv.SV) *sv.SV {

	if len(args) == 0 {
		return sv.NewUndef()
	}
	// If array ref, return length
	if args[0].IsRef() {
		target := args[0].Deref()
		if target != nil && target.IsArray() {
			return sv.NewInt(int64(len(target.ArrayData())))
		}
	}
	if args[0].IsArray() {
		return sv.NewInt(int64(len(args[0].ArrayData())))
	}
	return args[0]
}

// ============================================================
// OOP Built-ins
// ============================================================

// builtinBless implements bless($ref, $class)
// Returns the blessed reference
func (i *Interpreter) builtinBless(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	_ = exprs
	if len(args) == 0 {
		return sv.NewUndef()
	}

	ref := args[0]
	if !ref.IsRef() {
		// Can only bless references
		return sv.NewUndef()
	}

	// Get package name - default to current package or caller's package
	pkgName := "main"
	if len(args) >= 2 {
		pkgName = args[1].AsString()
	}

	// Bless the reference into the package
	ref.Bless(pkgName)
	return ref
}

// builtinIsa implements $obj->isa('ClassName') or UNIVERSAL::isa($obj, 'ClassName')
// Returns true if $obj is a member of ClassName
func (i *Interpreter) builtinIsa(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewInt(0)
	}

	obj := args[0]
	className := args[1].AsString()

	// Check if object is blessed
	if !obj.IsRef() || !obj.IsBlessed() {
		return sv.NewInt(0)
	}

	// Direct class check
	if obj.Package() == className {
		return sv.NewInt(1)
	}

	// TODO: Check @ISA inheritance chain
	return sv.NewInt(0)
}

// builtinCan implements $obj->can('method') or UNIVERSAL::can($obj, 'method')
// Returns coderef if $obj can do method, undef otherwise
func (i *Interpreter) builtinCan(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewUndef()
	}

	obj := args[0]
	methodName := args[1].AsString()

	var pkgName string
	if obj.IsRef() && obj.IsBlessed() {
		pkgName = obj.Package()
	} else {
		// Assume it's a class name
		pkgName = obj.AsString()
	}

	// Try to find the method using FindMethod (includes @ISA)
	if found := i.ctx.FindMethod(pkgName, methodName); found != "" {
		return sv.NewInt(1)
	}

	// Try just the method name
	if i.ctx.GetSub(methodName) != nil {
		return sv.NewInt(1)
	}

	return sv.NewUndef()
}

// builtinSetIsa sets the @ISA for a package
// set_isa('Child', 'Parent1', 'Parent2', ...)
func (i *Interpreter) builtinSetIsa(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewInt(0)
	}

	pkg := args[0].AsString()
	parents := make([]string, len(args)-1)
	for idx, arg := range args[1:] {
		parents[idx] = arg.AsString()
	}

	i.ctx.SetPackageISA(pkg, parents)
	return sv.NewInt(1)
}
