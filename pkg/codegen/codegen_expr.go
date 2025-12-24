package codegen

import (
	"fmt"
	"perlc/pkg/ast"
	"strings"
)

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
	case *ast.RefExpr:
		g.generateRefExpr(e)
	case *ast.DerefExpr:
		g.generateDerefExpr(e)
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
		case "delete":
			// delete $h{key} - нужно получить хеш и ключ
			if len(expr.Args) >= 1 {
				if ha, ok := expr.Args[0].(*ast.HashAccess); ok {
					g.write("func() *SV { ")
					// Получаем хеш
					hashName := ""
					if sv, ok := ha.Hash.(*ast.ScalarVar); ok {
						hashName = g.hashName(sv.Name)
					} else {
						g.tempCount++
						hashName = fmt.Sprintf("_htmp%d", g.tempCount)
						g.write(hashName + " := ")
						g.generateExpression(ha.Hash)
						g.write("; ")
					}
					// Получаем ключ
					g.write("_k := ")
					g.generateExpression(ha.Key)
					g.write(".AsString(); ")
					// Сохраняем старое значение
					g.write("_v := " + hashName + ".hv[_k]; ")
					// Удаляем
					g.write("delete(" + hashName + ".hv, _k); ")
					// Возвращаем старое значение
					g.write("return _v }()")
					return
				}
			}
			g.write("svUndef()")
		case "index":
			g.write("perl_index(")
			g.generateExpression(expr.Args[0])
			g.write(", ")
			g.generateExpression(expr.Args[1])
			if len(expr.Args) >= 3 {
				g.write(", ")
				g.generateExpression(expr.Args[2])
			}
			g.write(")")
		case "rindex":
			g.write("perl_rindex(")
			g.generateExpression(expr.Args[0])
			g.write(", ")
			g.generateExpression(expr.Args[1])
			if len(expr.Args) >= 3 {
				g.write(", ")
				g.generateExpression(expr.Args[2])
			}
			g.write(")")
		case "lcfirst":
			g.write("perl_lcfirst(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "ucfirst":
			g.write("perl_ucfirst(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "chop":
			if len(expr.Args) >= 1 {
				if sv, ok := expr.Args[0].(*ast.ScalarVar); ok {
					g.write("perl_chop(" + g.scalarName(sv.Name) + ")")
				} else {
					g.write("perl_chop(")
					g.generateExpression(expr.Args[0])
					g.write(")")
				}
			} else {
				g.write("svStr(\"\")")
			}
		case "sprintf":
			g.write("perl_sprintf(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		case "quotemeta":
			g.write("perl_quotemeta(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "hex":
			g.write("perl_hex(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "oct":
			g.write("perl_oct(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "fc":
			g.write("perl_fc(")
			g.generateExpression(expr.Args[0])
			g.write(")")
		case "pack":
			g.write("perl_pack(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		case "unpack":
			g.write("perl_unpack(")
			for i, a := range expr.Args {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(a)
			}
			g.write(")")
		case "grep":
			g.write("perl_grep(")
			if len(expr.Args) >= 2 {
				// Первый аргумент - блок или выражение
				if block, ok := expr.Args[0].(*ast.AnonSubExpr); ok {
					// Генерируем анонимную функцию
					g.write("func(_v *SV) *SV { ")
					// Устанавливаем $_ = _v
					g.write("v__ := _v; _ = v__; ")
					// Генерируем тело блока
					for _, stmt := range block.Body.Statements {
						g.write("return ")
						if es, ok := stmt.(*ast.ExprStmt); ok {
							g.generateExpression(es.Expression)
						}
						break
					}
					g.write(" }")
				} else {
					g.write("func(_v *SV) *SV { v__ := _v; _ = v__; return ")
					g.generateExpression(expr.Args[0])
					g.write(" }")
				}
				g.write(", ")
				g.generateExpression(expr.Args[1])
			}
			g.write(")")

		case "map":
			g.write("perl_map(")
			if len(expr.Args) >= 2 {
				// Первый аргумент - блок или выражение
				if block, ok := expr.Args[0].(*ast.AnonSubExpr); ok {
					// Генерируем анонимную функцию
					g.write("func(_v *SV) *SV { ")
					g.write("v__ := _v; _ = v__; ")
					for _, stmt := range block.Body.Statements {
						g.write("return ")
						if es, ok := stmt.(*ast.ExprStmt); ok {
							g.generateExpression(es.Expression)
						}
						break
					}
					g.write(" }")
				} else {
					g.write("func(_v *SV) *SV { v__ := _v; _ = v__; return ")
					g.generateExpression(expr.Args[0])
					g.write(" }")
				}
				g.write(", ")
				g.generateExpression(expr.Args[1])
			}
			g.write(")")
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

func (g *Generator) generateRefExpr(expr *ast.RefExpr) {
	// \$scalar - ссылка на скаляр
	if sv, ok := expr.Value.(*ast.ScalarVar); ok {
		g.write("svRef(" + g.scalarName(sv.Name) + ")")
		return
	}

	// \@array - ссылка на массив
	if av, ok := expr.Value.(*ast.ArrayVar); ok {
		g.write(g.arrayName(av.Name))
		return
	}

	// \%hash - ссылка на хеш
	if hv, ok := expr.Value.(*ast.HashVar); ok {
		g.write(g.hashName(hv.Name))
		return
	}

	// Для других выражений
	g.write("svUndef()")
}

func (g *Generator) generateDerefExpr(expr *ast.DerefExpr) {
	switch expr.Sigil {
	case "$":
		// $$ref - разыменование скаляра
		g.write("svDeref(")
		g.generateExpression(expr.Value)
		g.write(")")
	case "@":
		// @$ref - разыменование массива
		g.generateExpression(expr.Value)
	case "%":
		// %$ref - разыменование хеша
		g.generateExpression(expr.Value)
	default:
		g.write("svUndef()")
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
	case *ast.DerefExpr:
		// $$ref = value - присваивание через разыменование скаляра
		if left.Sigil == "$" {
			g.write("func() *SV { ")
			g.write("_ref := ")
			g.generateExpression(left.Value)
			g.write("; ")
			g.write("_val := ")
			g.generateExpression(expr.Right)
			g.write("; ")
			g.write("if _ref != nil && len(_ref.av) > 0 { ")
			g.write("_ref.av[0].iv = _val.iv; ")
			g.write("_ref.av[0].nv = _val.nv; ")
			g.write("_ref.av[0].pv = _val.pv; ")
			g.write("_ref.av[0].flags = _val.flags; ")
			g.write("}; return _val }()")
			return
		}
		g.generateExpression(expr.Right)
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

func (g *Generator) generateRangeExpr(expr *ast.RangeExpr) {
	g.write("func() *SV { var _r []*SV; for _i := int(")
	g.generateExpression(expr.Start)
	g.write(".AsInt()); _i <= int(")
	g.generateExpression(expr.End)
	g.write(".AsInt()); _i++ { _r = append(_r, svInt(int64(_i))) }; return svArray(_r...) }()")
}
