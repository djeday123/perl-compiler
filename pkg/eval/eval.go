// Package eval implements AST interpretation.
package eval

import (
	"io"
	"os"
	"regexp"
	"strings"

	"perlc/pkg/ast"
	"perlc/pkg/av"
	"perlc/pkg/context"
	"perlc/pkg/hv"
	"perlc/pkg/sv"
)

// Interpreter executes Perl AST.
type Interpreter struct {
	ctx    *context.Context
	stdout io.Writer
	stderr io.Writer
}

// New creates a new interpreter.
func New() *Interpreter {
	return &Interpreter{
		ctx:    context.New(),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

var interpolateRe = regexp.MustCompile(`\$(\w+)|\$\{(\w+)\}|@(\w+)`)

// SetStdout sets the output writer.
func (i *Interpreter) SetStdout(w io.Writer) {
	i.stdout = w
}

// Eval evaluates a program and returns the last value.
func (i *Interpreter) Eval(program *ast.Program) *sv.SV {
	var result *sv.SV
	for _, stmt := range program.Statements {
		result = i.evalStatement(stmt)
		if i.ctx.HasReturn() {
			return i.ctx.ReturnValue()
		}
	}
	return result
}

// ============================================================
// Statement Evaluation
// ============================================================

func (i *Interpreter) evalStatement(stmt ast.Statement) *sv.SV {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return i.evalExpression(s.Expression)
	case *ast.VarDecl:
		return i.evalVarDecl(s)
	case *ast.IfStmt:
		return i.evalIfStmt(s)
	case *ast.WhileStmt:
		return i.evalWhileStmt(s)
	case *ast.ForStmt:
		return i.evalForStmt(s)
	case *ast.ForeachStmt:
		return i.evalForeachStmt(s)
	case *ast.SubDecl:
		return i.evalSubDecl(s)
	case *ast.ReturnStmt:
		return i.evalReturnStmt(s)
	case *ast.BlockStmt:
		return i.evalBlockStmt(s)
	case *ast.LastStmt:
		i.ctx.SetLast(s.Label)
		return sv.NewUndef()
	case *ast.NextStmt:
		i.ctx.SetNext(s.Label)
		return sv.NewUndef()
	case *ast.UseDecl, *ast.PackageDecl, *ast.NoDecl, *ast.RequireDecl:
		return sv.NewUndef()
	default:
		return sv.NewUndef()
	}
}

func (i *Interpreter) evalBlockStmt(block *ast.BlockStmt) *sv.SV {
	var result *sv.SV
	for _, stmt := range block.Statements {
		result = i.evalStatement(stmt)
		if i.ctx.HasReturn() || i.ctx.HasLast() || i.ctx.HasNext() {
			break
		}
	}
	return result
}

func (i *Interpreter) evalVarDecl(decl *ast.VarDecl) *sv.SV {
	var value *sv.SV
	if decl.Value != nil {
		value = i.evalExpression(decl.Value)
	} else {
		value = sv.NewUndef()
	}

	if len(decl.Names) > 1 {
		values := i.svToList(value)
		for idx, name := range decl.Names {
			var val *sv.SV
			if idx < len(values) {
				val = values[idx]
			} else {
				val = sv.NewUndef()
			}
			i.assignToVar(name, val, decl.Kind)
		}
		return value
	}

	if len(decl.Names) == 1 {
		i.assignToVar(decl.Names[0], value, decl.Kind)
	}
	return value
}

func (i *Interpreter) assignToVar(expr ast.Expression, value *sv.SV, kind string) {
	switch v := expr.(type) {
	case *ast.ScalarVar:
		i.ctx.DeclareVar(v.Name, value, kind)
	case *ast.ArrayVar:
		i.ctx.DeclareVar(v.Name, value, kind)
	case *ast.HashVar:
		i.ctx.DeclareVar(v.Name, value, kind)
	}
}

func (i *Interpreter) evalIfStmt(stmt *ast.IfStmt) *sv.SV {
	cond := i.evalExpression(stmt.Condition)
	testResult := cond.IsTrue()
	if stmt.Unless {
		testResult = !testResult
	}

	if testResult {
		return i.evalBlockStmt(stmt.Then)
	}

	for _, elsif := range stmt.Elsif {
		cond := i.evalExpression(elsif.Condition)
		if cond.IsTrue() {
			return i.evalBlockStmt(elsif.Body)
		}
	}

	if stmt.Else != nil {
		return i.evalBlockStmt(stmt.Else)
	}
	return sv.NewUndef()
}

func (i *Interpreter) evalWhileStmt(stmt *ast.WhileStmt) *sv.SV {
	var result *sv.SV
	for {
		cond := i.evalExpression(stmt.Condition)
		testResult := cond.IsTrue()
		if stmt.Until {
			testResult = !testResult
		}
		if !testResult {
			break
		}

		result = i.evalBlockStmt(stmt.Body)

		if i.ctx.HasLast() {
			i.ctx.ClearLast()
			break
		}
		if i.ctx.HasNext() {
			i.ctx.ClearNext()
			continue
		}
		if i.ctx.HasReturn() {
			break
		}
	}
	return result
}

func (i *Interpreter) evalForStmt(stmt *ast.ForStmt) *sv.SV {
	var result *sv.SV

	// Init - может быть VarDecl или ExprStmt
	if stmt.Init != nil {
		i.evalStatement(stmt.Init)
	}

	for {
		// Condition
		if stmt.Condition != nil {
			cond := i.evalExpression(stmt.Condition)
			if !cond.IsTrue() {
				break
			}
		}

		result = i.evalBlockStmt(stmt.Body)

		if i.ctx.HasLast() {
			i.ctx.ClearLast()
			break
		}
		if i.ctx.HasNext() {
			i.ctx.ClearNext()
			// Still need to execute Post before next iteration
		}
		if i.ctx.HasReturn() {
			break
		}

		// Post - execute after body, before next condition check
		if stmt.Post != nil {
			i.evalExpression(stmt.Post)
		}
	}
	return result
}

func (i *Interpreter) evalForeachStmt(stmt *ast.ForeachStmt) *sv.SV {
	var result *sv.SV
	list := i.evalExpression(stmt.List)
	values := i.svToList(list)

	varName := ""
	if v, ok := stmt.Variable.(*ast.ScalarVar); ok {
		varName = v.Name
	}

	for _, val := range values {
		i.ctx.SetVar(varName, val)
		result = i.evalBlockStmt(stmt.Body)

		if i.ctx.HasLast() {
			i.ctx.ClearLast()
			break
		}
		if i.ctx.HasNext() {
			i.ctx.ClearNext()
			continue
		}
		if i.ctx.HasReturn() {
			break
		}
	}
	return result
}

func (i *Interpreter) evalSubDecl(decl *ast.SubDecl) *sv.SV {
	i.ctx.DeclareSub(decl.Name, decl.Body)
	return sv.NewUndef()
}

func (i *Interpreter) evalReturnStmt(stmt *ast.ReturnStmt) *sv.SV {
	var value *sv.SV
	if stmt.Value != nil {
		value = i.evalExpression(stmt.Value)
	} else {
		value = sv.NewUndef()
	}
	i.ctx.SetReturn(value)
	return value
}

// ============================================================
// Expression Evaluation
// ============================================================

func (i *Interpreter) evalExpression(expr ast.Expression) *sv.SV {
	if expr == nil {
		return sv.NewUndef()
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return sv.NewInt(e.Value)
	case *ast.FloatLiteral:
		return sv.NewFloat(e.Value)
	case *ast.StringLiteral:
		if e.Interpolated {
			return sv.NewString(i.interpolateString(e.Value))
		}
		return sv.NewString(e.Value)
	case *ast.UndefLiteral:
		return sv.NewUndef()
	case *ast.ScalarVar:
		return i.ctx.GetVar(e.Name)
	case *ast.ArrayVar:
		return i.ctx.GetVar(e.Name)
	case *ast.HashVar:
		return i.ctx.GetVar(e.Name)
	case *ast.SpecialVar:
		return i.evalSpecialVar(e.Name)
	case *ast.PrefixExpr:
		return i.evalPrefixExpr(e)
	case *ast.InfixExpr:
		return i.evalInfixExpr(e)
	case *ast.PostfixExpr:
		return i.evalPostfixExpr(e)
	case *ast.AssignExpr:
		return i.evalAssignExpr(e)
	case *ast.TernaryExpr:
		return i.evalTernaryExpr(e)
	case *ast.ArrayExpr:
		return i.evalArrayExpr(e)
	case *ast.HashExpr:
		return i.evalHashExpr(e)
	case *ast.ArrayAccess:
		return i.evalArrayAccess(e)
	case *ast.HashAccess:
		return i.evalHashAccess(e)
	case *ast.CallExpr:
		return i.evalCallExpr(e)
	case *ast.MethodCall:
		return i.evalMethodCall(e)
	case *ast.RefExpr:
		return i.evalRefExpr(e)
	case *ast.Identifier:
		return sv.NewString(e.Value)
	case *ast.RangeExpr:
		return i.evalRangeExpr(e)
	case *ast.ArrowAccess:
		return i.evalArrowAccess(e)
	case *ast.MatchExpr:
		return i.evalMatchExpr(e)
	case *ast.SubstExpr:
		return i.evalSubstExpr(e)
	case *ast.ReadLineExpr:
		return i.evalReadLineExpr(e)
	default:
		return sv.NewUndef()
	}
}

func (i *Interpreter) evalPrefixExpr(expr *ast.PrefixExpr) *sv.SV {
	right := i.evalExpression(expr.Right)

	switch expr.Operator {
	case "-":
		return sv.NewFloat(-right.AsFloat())
	case "+":
		return sv.NewFloat(right.AsFloat())
	case "!":
		return boolToSV(!right.IsTrue())
	case "not":
		return boolToSV(!right.IsTrue())
	case "~":
		return sv.NewInt(^right.AsInt())
	case "++":
		val := sv.NewInt(right.AsInt() + 1)
		i.assignBack(expr.Right, val)
		return val
	case "--":
		val := sv.NewInt(right.AsInt() - 1)
		i.assignBack(expr.Right, val)
		return val
	default:
		return sv.NewUndef()
	}
}

func (i *Interpreter) evalInfixExpr(expr *ast.InfixExpr) *sv.SV {
	// Short-circuit
	if expr.Operator == "&&" || expr.Operator == "and" {
		left := i.evalExpression(expr.Left)
		if !left.IsTrue() {
			return left
		}
		return i.evalExpression(expr.Right)
	}
	if expr.Operator == "||" || expr.Operator == "or" {
		left := i.evalExpression(expr.Left)
		if left.IsTrue() {
			return left
		}
		return i.evalExpression(expr.Right)
	}
	if expr.Operator == "//" {
		left := i.evalExpression(expr.Left)
		if !left.IsUndef() {
			return left
		}
		return i.evalExpression(expr.Right)
	}

	left := i.evalExpression(expr.Left)
	right := i.evalExpression(expr.Right)

	switch expr.Operator {
	case "+":
		return sv.Add(left, right)
	case "-":
		return sv.Sub(left, right)
	case "*":
		return sv.Mul(left, right)
	case "/":
		return sv.Div(left, right)
	case "%":
		return sv.Mod(left, right)
	case "**":
		return sv.Pow(left, right)
	case ".":
		return sv.Concat(left, right)
	case "x":
		return sv.Repeat(left, right)
	case "==":
		return sv.NumEq(left, right)
	case "!=":
		return sv.NumNe(left, right)
	case "<":
		return sv.NumLt(left, right)
	case "<=":
		return sv.NumLe(left, right)
	case ">":
		return sv.NumGt(left, right)
	case ">=":
		return sv.NumGe(left, right)
	case "<=>":
		return sv.NumCmp(left, right)
	case "eq":
		return sv.StrEq(left, right)
	case "ne":
		return sv.StrNe(left, right)
	case "lt":
		return sv.StrLt(left, right)
	case "le":
		return sv.StrLe(left, right)
	case "gt":
		return sv.StrGt(left, right)
	case "ge":
		return sv.StrGe(left, right)
	case "cmp":
		return sv.StrCmp(left, right)
	case "&":
		return sv.BitAnd(left, right)
	case "|":
		return sv.BitOr(left, right)
	case "^":
		return sv.BitXor(left, right)
	case "<<":
		return sv.LeftShift(left, right)
	case ">>":
		return sv.RightShift(left, right)
	default:
		return sv.NewUndef()
	}
}

func (i *Interpreter) evalPostfixExpr(expr *ast.PostfixExpr) *sv.SV {
	left := i.evalExpression(expr.Left)
	oldVal := sv.NewInt(left.AsInt())

	switch expr.Operator {
	case "++":
		i.assignBack(expr.Left, sv.NewInt(left.AsInt()+1))
		return oldVal
	case "--":
		i.assignBack(expr.Left, sv.NewInt(left.AsInt()-1))
		return oldVal
	default:
		return oldVal
	}
}

func (i *Interpreter) evalAssignExpr(expr *ast.AssignExpr) *sv.SV {
	right := i.evalExpression(expr.Right)

	if expr.Operator != "=" {
		left := i.evalExpression(expr.Left)
		switch expr.Operator {
		case "+=":
			right = sv.Add(left, right)
		case "-=":
			right = sv.Sub(left, right)
		case "*=":
			right = sv.Mul(left, right)
		case "/=":
			right = sv.Div(left, right)
		case ".=":
			right = sv.Concat(left, right)
		case "||=":
			if left.IsTrue() {
				return left
			}
		case "//=":
			if !left.IsUndef() {
				return left
			}
		}
	}

	i.assignBack(expr.Left, right)
	return right
}

func (i *Interpreter) evalTernaryExpr(expr *ast.TernaryExpr) *sv.SV {
	cond := i.evalExpression(expr.Condition)
	if cond.IsTrue() {
		return i.evalExpression(expr.Then)
	}
	return i.evalExpression(expr.Else)
}

func (i *Interpreter) evalArrayExpr(expr *ast.ArrayExpr) *sv.SV {
	elements := make([]*sv.SV, len(expr.Elements))
	for idx, el := range expr.Elements {
		elements[idx] = i.evalExpression(el)
	}
	return sv.NewArrayRef(elements...)
}

func (i *Interpreter) evalHashExpr(expr *ast.HashExpr) *sv.SV {
	href := sv.NewHashRef()
	for _, pair := range expr.Pairs {
		key := i.evalExpression(pair.Key)
		value := i.evalExpression(pair.Value)
		hv.Store(href, key, value)
	}
	return href
}

func (i *Interpreter) evalArrayAccess(expr *ast.ArrayAccess) *sv.SV {
	array := i.evalExpression(expr.Array)
	index := i.evalExpression(expr.Index)
	return av.Fetch(array, index)
}

func (i *Interpreter) evalHashAccess(expr *ast.HashAccess) *sv.SV {
	hash := i.evalExpression(expr.Hash)
	key := i.evalExpression(expr.Key)
	return hv.Fetch(hash, key)
}

func (i *Interpreter) evalCallExpr(expr *ast.CallExpr) *sv.SV {
	funcName := ""
	if ident, ok := expr.Function.(*ast.Identifier); ok {
		funcName = ident.Value
	}

	args := make([]*sv.SV, len(expr.Args))
	for idx, arg := range expr.Args {
		args[idx] = i.evalExpression(arg)
	}

	// Built-in functions
	switch funcName {
	case "print":
		return i.builtinPrint(expr)
	case "say":
		return i.builtinSay(expr)
	case "open":
		return i.builtinOpen(expr)
	case "close":
		return i.builtinClose(expr)
	case "length":
		return sv.Length(args[0])
	case "defined":
		return sv.Defined(args[0])
	case "ref":
		return sv.Ref(args[0])
	case "push":
		return i.builtinPush(expr.Args, args)
	case "pop":
		return i.builtinPop(expr.Args)
	case "shift":
		return i.builtinShift(expr.Args)
	case "unshift":
		return i.builtinUnshift(expr.Args, args)
	case "keys":
		return i.builtinKeys(args)
	case "values":
		return i.builtinValues(args)
	case "join":
		return i.builtinJoin(args)
	case "split":
		return i.builtinSplit(args)
	case "substr":
		return i.builtinSubstr(args)
	case "int":
		if len(args) > 0 {
			return sv.NewInt(args[0].AsInt())
		}
		return sv.NewInt(0)
	case "abs":
		return i.builtinAbs(args)
	case "sqrt":
		return i.builtinSqrt(args)
	case "chr":
		return i.builtinChr(args)
	case "ord":
		return i.builtinOrd(args)
	case "lc":
		return sv.Lc(args[0])
	case "uc":
		return sv.Uc(args[0])
	case "chomp":
		return i.builtinChomp(expr.Args)
	case "die":
		return i.builtinDie(args)
	case "warn":
		return i.builtinWarn(args)
	case "exit":
		return i.builtinExit(args)
	case "scalar":
		return i.builtinScalar(args)
	}

	return i.callUserSub(funcName, args)
}

func (i *Interpreter) evalMethodCall(expr *ast.MethodCall) *sv.SV {
	_ = expr // TODO: implement OO method calls
	return sv.NewUndef()
}

func (i *Interpreter) evalRefExpr(expr *ast.RefExpr) *sv.SV {
	val := i.evalExpression(expr.Value)
	return sv.NewRef(val)
}

func (i *Interpreter) evalRangeExpr(expr *ast.RangeExpr) *sv.SV {
	start := i.evalExpression(expr.Start)
	end := i.evalExpression(expr.End)
	elements := sv.Range(start, end)
	return sv.NewArrayRef(elements...)
}

func (i *Interpreter) evalSpecialVar(name string) *sv.SV {
	return i.ctx.GetSpecialVar(name)
}

// ============================================================
// Helper Functions
// ============================================================

func (i *Interpreter) assignBack(expr ast.Expression, value *sv.SV) {
	switch v := expr.(type) {
	case *ast.ScalarVar:
		i.ctx.SetVar(v.Name, value)
	case *ast.ArrayAccess:
		arr := i.evalExpression(v.Array)
		idx := i.evalExpression(v.Index)
		av.Store(arr, idx, value)
	case *ast.HashAccess:
		hash := i.evalExpression(v.Hash)
		key := i.evalExpression(v.Key)
		hv.Store(hash, key, value)
	}
}

func (i *Interpreter) svToList(val *sv.SV) []*sv.SV {
	if val.IsRef() {
		target := val.Deref()
		if target != nil && target.IsArray() {
			return target.ArrayData()
		}
	}
	if val.IsArray() {
		return val.ArrayData()
	}
	return []*sv.SV{val}
}

func (i *Interpreter) interpolateString(s string) string {
	return interpolateRe.ReplaceAllStringFunc(s, func(match string) string {
		var name string
		if match[0] == '@' {
			name = match[1:]
			val := i.ctx.GetVar(name)
			if val != nil && val.IsArray() {
				elements := val.ArrayData()
				parts := make([]string, len(elements))
				for idx, el := range elements {
					parts[idx] = el.AsString()
				}
				return strings.Join(parts, " ")
			}
			return ""
		}
		if strings.HasPrefix(match, "${") {
			name = match[2 : len(match)-1]
		} else {
			name = match[1:]
		}
		val := i.ctx.GetVar(name)
		if val != nil {
			return val.AsString()
		}
		return ""
	})
}

func (i *Interpreter) callUserSub(name string, args []*sv.SV) *sv.SV {
	body := i.ctx.GetSub(name)
	if body == nil {
		return sv.NewUndef()
	}

	i.ctx.PushScope()
	defer i.ctx.PopScope()

	i.ctx.SetArgs(args)
	result := i.evalBlockStmt(body)

	if i.ctx.HasReturn() {
		result = i.ctx.ReturnValue()
		i.ctx.ClearReturn()
	}
	return result
}

func (i *Interpreter) evalArrowAccess(expr *ast.ArrowAccess) *sv.SV {
	left := i.evalExpression(expr.Left)

	// Dereference if needed
	target := left
	if left.IsRef() {
		target = left.Deref()
	}

	// Check what's on the right
	switch right := expr.Right.(type) {
	case *ast.ArrayAccess:
		index := i.evalExpression(right.Index)
		return av.Fetch(target, index)
	case *ast.HashAccess:
		key := i.evalExpression(right.Key)
		return hv.Fetch(target, key)
	default:
		return sv.NewUndef()
	}
}

func (i *Interpreter) evalMatchExpr(expr *ast.MatchExpr) *sv.SV {
	target := i.evalExpression(expr.Target)
	str := target.AsString()

	pattern := expr.Pattern.Pattern
	flags := expr.Pattern.Flags

	// Build regex pattern with flags
	rePattern := pattern
	if strings.Contains(flags, "i") {
		rePattern = "(?i)" + rePattern
	}

	re, err := regexp.Compile(rePattern)
	if err != nil {
		return sv.NewInt(0)
	}

	matched := re.MatchString(str)

	if expr.Negate {
		if matched {
			return sv.NewInt(0)
		}
		return sv.NewInt(1)
	}

	if matched {
		return sv.NewInt(1)
	}
	return sv.NewInt(0)
}

func (i *Interpreter) evalSubstExpr(expr *ast.SubstExpr) *sv.SV {
	target := i.evalExpression(expr.Target)
	str := target.AsString()

	pattern := expr.Pattern
	replacement := expr.Replacement
	flags := expr.Flags

	// Build regex with flags
	rePattern := pattern
	if strings.Contains(flags, "i") {
		rePattern = "(?i)" + rePattern
	}

	re, err := regexp.Compile(rePattern)
	if err != nil {
		return sv.NewInt(0)
	}

	var result string
	if strings.Contains(flags, "g") {
		result = re.ReplaceAllString(str, replacement)
	} else {
		// result = re.ReplaceAllStringFunc(str, func(match string) string {
		// 	// Only replace first occurrence
		// 	return replacement
		// })
		// Actually simpler:
		loc := re.FindStringIndex(str)
		if loc != nil {
			result = str[:loc[0]] + replacement + str[loc[1]:]
		} else {
			result = str
		}
	}

	// Update the variable if it's a scalar
	if v, ok := expr.Target.(*ast.ScalarVar); ok {
		i.ctx.SetVar(v.Name, sv.NewString(result))
	}

	if result != str {
		return sv.NewInt(1)
	}
	return sv.NewInt(0)
}

func (i *Interpreter) evalReadLineExpr(expr *ast.ReadLineExpr) *sv.SV {
	var name string
	if expr.Filehandle != nil {
		switch fh := expr.Filehandle.(type) {
		case *ast.Identifier:
			name = fh.Value
		case *ast.ScalarVar:
			// Get the value which contains the filehandle name
			val := i.ctx.GetVar(fh.Name)
			if val != nil {
				name = val.AsString()
			}
			if name == "" {
				name = fh.Name
			}
		}
	}

	line, ok := i.ctx.ReadLine(name)
	if !ok {
		return sv.NewUndef()
	}
	return sv.NewString(line)
}

func boolToSV(b bool) *sv.SV {
	if b {
		return sv.NewInt(1)
	}
	return sv.NewString("")
}
