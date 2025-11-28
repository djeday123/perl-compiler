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
	// Debug
	//fmt.Fprintf(os.Stderr, "DEBUG print: %d args\n", len(expr.Args))
	// for idx, arg := range expr.Args {
	// 	fmt.Fprintf(os.Stderr, "  arg[%d]: %T = %s\n", idx, arg, arg.String())
	// }
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
	fmt.Printf("DEBUG builtinScalar: len(args)=%d\n", len(args))
	if len(args) > 0 && args[0] != nil {
		fmt.Printf("DEBUG builtinScalar: args[0].IsArray()=%v, args[0].IsRef()=%v\n",
			args[0].IsArray(), args[0].IsRef())
	} else if len(args) > 0 {
		fmt.Println("DEBUG builtinScalar: args[0] is nil!")
	}

	if len(args) > 0 && args[0] != nil {
		fmt.Printf("DEBUG builtinScalar: IsArray=%v, IsRef=%v, IsHash=%v, AsString=%q\n",
			args[0].IsArray(), args[0].IsRef(), args[0].IsHash(), args[0].AsString())
	}

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
