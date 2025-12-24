// Package eval - builtin functions
package eval

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

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

func (i *Interpreter) builtinIndex(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewInt(-1)
	}
	str := args[0].AsString()
	substr := args[1].AsString()

	// Опциональная начальная позиция
	start := 0
	if len(args) >= 3 {
		start = int(args[2].AsInt())
		if start < 0 {
			start = 0
		}
		if start > len(str) {
			return sv.NewInt(-1)
		}
	}

	pos := strings.Index(str[start:], substr)
	if pos == -1 {
		return sv.NewInt(-1)
	}
	return sv.NewInt(int64(pos + start))
}

// ============================================================
// rindex - найти позицию подстроки (с конца)
// ============================================================

func (i *Interpreter) builtinRindex(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewInt(-1)
	}
	str := args[0].AsString()
	substr := args[1].AsString()

	// Опциональная позиция (до которой искать)
	end := len(str)
	if len(args) >= 3 {
		end = int(args[2].AsInt()) + len(substr)
		if end > len(str) {
			end = len(str)
		}
		if end < 0 {
			return sv.NewInt(-1)
		}
	}

	pos := strings.LastIndex(str[:end], substr)
	return sv.NewInt(int64(pos))
}

// ============================================================
// lcfirst - первая буква в нижний регистр
// ============================================================

func (i *Interpreter) builtinLcfirst(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}
	str := args[0].AsString()
	if len(str) == 0 {
		return sv.NewString("")
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return sv.NewString(string(runes))
}

// ============================================================
// ucfirst - первая буква в верхний регистр
// ============================================================

func (i *Interpreter) builtinUcfirst(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}
	str := args[0].AsString()
	if len(str) == 0 {
		return sv.NewString("")
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return sv.NewString(string(runes))
}

// ============================================================
// chop - удалить последний символ
// ============================================================

func (i *Interpreter) builtinChop(exprs []ast.Expression) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewString("")
	}

	var lastChar string
	for _, expr := range exprs {
		if v, ok := expr.(*ast.ScalarVar); ok {
			val := i.ctx.GetVar(v.Name)
			s := val.AsString()
			if len(s) > 0 {
				runes := []rune(s)
				lastChar = string(runes[len(runes)-1])
				s = string(runes[:len(runes)-1])
				i.ctx.SetVar(v.Name, sv.NewString(s))
			}
		}
	}
	return sv.NewString(lastChar)
}

// ============================================================
// sprintf - форматированная строка
// ============================================================

func (i *Interpreter) builtinSprintf(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}

	format := args[0].AsString()

	// Конвертируем аргументы в interface{} для fmt.Sprintf
	// Используем AsString для всех аргументов, Go сам разберётся с форматом
	// Но для %d/%i/%x нужны числа, для %f/%e/%g нужны float
	fmtArgs := make([]interface{}, len(args)-1)

	// Простой подход: парсим формат и выбираем тип
	fmtIdx := 0
	for idx, arg := range args[1:] {
		// Находим следующий % в формате
		for fmtIdx < len(format) {
			if format[fmtIdx] == '%' {
				fmtIdx++
				if fmtIdx < len(format) && format[fmtIdx] == '%' {
					fmtIdx++
					continue // %%
				}
				// Пропускаем флаги и ширину
				for fmtIdx < len(format) {
					c := format[fmtIdx]
					if c == '-' || c == '+' || c == ' ' || c == '#' || c == '0' ||
						(c >= '0' && c <= '9') || c == '.' || c == '*' {
						fmtIdx++
					} else {
						break
					}
				}
				// Смотрим спецификатор
				if fmtIdx < len(format) {
					spec := format[fmtIdx]
					fmtIdx++
					switch spec {
					case 'd', 'i', 'o', 'x', 'X', 'b', 'c':
						fmtArgs[idx] = arg.AsInt()
					case 'e', 'E', 'f', 'F', 'g', 'G':
						fmtArgs[idx] = arg.AsFloat()
					default: // 's', 'v', etc.
						fmtArgs[idx] = arg.AsString()
					}
					break
				}
			} else {
				fmtIdx++
			}
		}
		// Если формат закончился, используем строку
		if fmtArgs[idx] == nil {
			fmtArgs[idx] = arg.AsString()
		}
	}

	result := fmt.Sprintf(format, fmtArgs...)
	return sv.NewString(result)
}

// ============================================================
// quotemeta - экранирование метасимволов regex
// ============================================================

func (i *Interpreter) builtinQuotemeta(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}
	str := args[0].AsString()
	return sv.NewString(regexp.QuoteMeta(str))
}

// ============================================================
// hex - hex строка в число
// ============================================================

func (i *Interpreter) builtinHex(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewInt(0)
	}
	str := args[0].AsString()
	// Убираем префикс 0x если есть
	str = strings.TrimPrefix(str, "0x")
	str = strings.TrimPrefix(str, "0X")

	val, err := strconv.ParseInt(str, 16, 64)
	if err != nil {
		return sv.NewInt(0)
	}
	return sv.NewInt(val)
}

// ============================================================
// oct - octal/hex/binary строка в число
// ============================================================

func (i *Interpreter) builtinOct(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewInt(0)
	}
	str := strings.TrimSpace(args[0].AsString())

	// Определяем базу по префиксу
	var val int64
	var err error

	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		val, err = strconv.ParseInt(str[2:], 16, 64)
	} else if strings.HasPrefix(str, "0b") || strings.HasPrefix(str, "0B") {
		val, err = strconv.ParseInt(str[2:], 2, 64)
	} else if strings.HasPrefix(str, "0") && len(str) > 1 {
		val, err = strconv.ParseInt(str[1:], 8, 64)
	} else {
		val, err = strconv.ParseInt(str, 8, 64)
	}

	if err != nil {
		return sv.NewInt(0)
	}
	return sv.NewInt(val)
}

// ============================================================
// fc - Unicode fold case (для сравнения без учёта регистра)
// ============================================================

func (i *Interpreter) builtinFc(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewString("")
	}
	str := args[0].AsString()
	// Go не имеет прямого fold case, используем ToLower
	return sv.NewString(strings.ToLower(str))
}

// ============================================================
// pack - упаковка данных в бинарную строку
// ============================================================

func (i *Interpreter) builtinPack(args []*sv.SV) *sv.SV {
	if len(args) < 1 {
		return sv.NewString("")
	}

	template := args[0].AsString()
	values := args[1:]

	var buf bytes.Buffer
	valIdx := 0

	for idx := 0; idx < len(template); idx++ {
		if valIdx >= len(values) {
			break
		}

		ch := template[idx]

		// Проверяем count
		count := 1
		if idx+1 < len(template) {
			if template[idx+1] >= '0' && template[idx+1] <= '9' {
				countStr := ""
				for idx+1 < len(template) && template[idx+1] >= '0' && template[idx+1] <= '9' {
					idx++
					countStr += string(template[idx])
				}
				count, _ = strconv.Atoi(countStr)
			} else if template[idx+1] == '*' {
				idx++
				count = len(values) - valIdx
			}
		}

		for c := 0; c < count && valIdx < len(values); c++ {
			val := values[valIdx]

			switch ch {
			case 'A', 'a': // ASCII строка
				s := val.AsString()
				buf.WriteString(s)
				valIdx++
			case 'Z': // Null-terminated строка
				s := val.AsString()
				buf.WriteString(s)
				buf.WriteByte(0)
				valIdx++
			case 'c', 'C': // char
				buf.WriteByte(byte(val.AsInt()))
				valIdx++
			case 's': // signed short (little-endian)
				binary.Write(&buf, binary.LittleEndian, int16(val.AsInt()))
				valIdx++
			case 'S': // unsigned short
				binary.Write(&buf, binary.LittleEndian, uint16(val.AsInt()))
				valIdx++
			case 'l': // signed long
				binary.Write(&buf, binary.LittleEndian, int32(val.AsInt()))
				valIdx++
			case 'L': // unsigned long
				binary.Write(&buf, binary.LittleEndian, uint32(val.AsInt()))
				valIdx++
			case 'q': // signed quad
				binary.Write(&buf, binary.LittleEndian, val.AsInt())
				valIdx++
			case 'Q': // unsigned quad
				binary.Write(&buf, binary.LittleEndian, uint64(val.AsInt()))
				valIdx++
			case 'n': // unsigned short (big-endian)
				binary.Write(&buf, binary.BigEndian, uint16(val.AsInt()))
				valIdx++
			case 'N': // unsigned long (big-endian)
				binary.Write(&buf, binary.BigEndian, uint32(val.AsInt()))
				valIdx++
			case 'f': // float
				binary.Write(&buf, binary.LittleEndian, float32(val.AsFloat()))
				valIdx++
			case 'd': // double
				binary.Write(&buf, binary.LittleEndian, val.AsFloat())
				valIdx++
			case 'H': // hex string
				s := val.AsString()
				for j := 0; j < len(s); j += 2 {
					end := j + 2
					if end > len(s) {
						end = len(s)
					}
					b, _ := strconv.ParseUint(s[j:end], 16, 8)
					buf.WriteByte(byte(b))
				}
				valIdx++
			case 'x': // null byte
				buf.WriteByte(0)
			}
		}
	}

	return sv.NewString(buf.String())
}

// ============================================================
// unpack - распаковка бинарной строки
// ============================================================

func (i *Interpreter) builtinUnpack(args []*sv.SV) *sv.SV {
	if len(args) < 2 {
		return sv.NewArrayRef()
	}

	template := args[0].AsString()
	data := []byte(args[1].AsString())

	var results []*sv.SV
	offset := 0

	for idx := 0; idx < len(template); idx++ {
		if offset >= len(data) {
			break
		}

		ch := template[idx]

		// Проверяем count
		count := 1
		if idx+1 < len(template) {
			if template[idx+1] >= '0' && template[idx+1] <= '9' {
				countStr := ""
				for idx+1 < len(template) && template[idx+1] >= '0' && template[idx+1] <= '9' {
					idx++
					countStr += string(template[idx])
				}
				count, _ = strconv.Atoi(countStr)
			} else if template[idx+1] == '*' {
				idx++
				count = len(data) - offset
			}
		}

		for c := 0; c < count && offset < len(data); c++ {
			switch ch {
			case 'A', 'a': // ASCII строка
				if count > 1 {
					end := offset + count
					if end > len(data) {
						end = len(data)
					}
					results = append(results, sv.NewString(string(data[offset:end])))
					offset = end
					c = count
				} else {
					results = append(results, sv.NewString(string(data[offset])))
					offset++
				}
			case 'Z': // Null-terminated
				end := offset
				for end < len(data) && data[end] != 0 {
					end++
				}
				results = append(results, sv.NewString(string(data[offset:end])))
				offset = end + 1
			case 'c': // signed char
				results = append(results, sv.NewInt(int64(int8(data[offset]))))
				offset++
			case 'C': // unsigned char
				results = append(results, sv.NewInt(int64(data[offset])))
				offset++
			case 's': // signed short
				if offset+2 <= len(data) {
					val := int16(binary.LittleEndian.Uint16(data[offset:]))
					results = append(results, sv.NewInt(int64(val)))
					offset += 2
				}
			case 'S': // unsigned short
				if offset+2 <= len(data) {
					val := binary.LittleEndian.Uint16(data[offset:])
					results = append(results, sv.NewInt(int64(val)))
					offset += 2
				}
			case 'l': // signed long
				if offset+4 <= len(data) {
					val := int32(binary.LittleEndian.Uint32(data[offset:]))
					results = append(results, sv.NewInt(int64(val)))
					offset += 4
				}
			case 'L': // unsigned long
				if offset+4 <= len(data) {
					val := binary.LittleEndian.Uint32(data[offset:])
					results = append(results, sv.NewInt(int64(val)))
					offset += 4
				}
			case 'n': // unsigned short (big-endian)
				if offset+2 <= len(data) {
					val := binary.BigEndian.Uint16(data[offset:])
					results = append(results, sv.NewInt(int64(val)))
					offset += 2
				}
			case 'N': // unsigned long (big-endian)
				if offset+4 <= len(data) {
					val := binary.BigEndian.Uint32(data[offset:])
					results = append(results, sv.NewInt(int64(val)))
					offset += 4
				}
			case 'H': // hex string
				if count > 1 {
					end := offset + (count+1)/2
					if end > len(data) {
						end = len(data)
					}
					var hex strings.Builder
					for j := offset; j < end; j++ {
						hex.WriteString(fmt.Sprintf("%02X", data[j]))
					}
					s := hex.String()
					if len(s) > count {
						s = s[:count]
					}
					results = append(results, sv.NewString(s))
					offset = end
					c = count
				} else {
					results = append(results, sv.NewString(fmt.Sprintf("%X", data[offset]>>4)))
					offset++
				}
			case 'x': // skip byte
				offset++
			}
		}
	}

	return sv.NewArrayRef(results...)
}
