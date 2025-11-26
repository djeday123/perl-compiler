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

func (i *Interpreter) builtinPrint(args []*sv.SV) *sv.SV {
	for _, arg := range args {
		fmt.Fprint(i.stdout, arg.AsString())
	}
	return sv.NewInt(1)
}

func (i *Interpreter) builtinSay(args []*sv.SV) *sv.SV {
	for _, arg := range args {
		fmt.Fprint(i.stdout, arg.AsString())
	}
	fmt.Fprintln(i.stdout)
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
