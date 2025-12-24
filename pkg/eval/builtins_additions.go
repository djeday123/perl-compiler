// Дополнения для pkg/eval/builtins.go
// Добавить в switch в evalCallExpr после case "scalar":

/*
Добавь в evalCallExpr (pkg/eval/eval.go) в switch funcName эти case:

	case "reverse":
		return i.builtinReverse(expr.Args, args)
	case "sort":
		return i.builtinSort(expr.Args, args)
	case "exists":
		return i.builtinExists(expr)
	case "delete":
		return i.builtinDelete(expr)
*/

// Добавить эти функции в pkg/eval/builtins.go:

package eval

import (
	"bufio"
	"fmt"
	"perlc/pkg/ast"
	"perlc/pkg/av"
	"perlc/pkg/hv"
	"perlc/pkg/sv"
	"sort"
)

func (i *Interpreter) builtinReverse(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewArrayRef()
	}

	// Проверяем, если аргумент - переменная массива
	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		if arrSV == nil || (!arrSV.IsArray() && !arrSV.IsRef()) {
			return sv.NewArrayRef()
		}

		// Получаем данные
		var elements []*sv.SV
		if arrSV.IsRef() {
			deref := arrSV.Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if arrSV.IsArray() {
			elements = arrSV.ArrayData()
		}

		if elements == nil {
			return sv.NewArrayRef()
		}

		// Создаём новый массив с элементами в обратном порядке
		reversed := make([]*sv.SV, len(elements))
		for i, j := 0, len(elements)-1; j >= 0; i, j = i+1, j-1 {
			reversed[i] = elements[j]
		}
		return sv.NewArrayRef(reversed...)
	}

	// Если передан первый аргумент как значение
	if len(args) > 0 && args[0] != nil {
		var elements []*sv.SV
		if args[0].IsRef() {
			deref := args[0].Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if args[0].IsArray() {
			elements = args[0].ArrayData()
		}

		if elements != nil {
			reversed := make([]*sv.SV, len(elements))
			for i, j := 0, len(elements)-1; j >= 0; i, j = i+1, j-1 {
				reversed[i] = elements[j]
			}
			return sv.NewArrayRef(reversed...)
		}
	}

	return sv.NewArrayRef()
}

func (i *Interpreter) BuiltinSort_vOld(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewArrayRef()
	}

	// Проверяем, если аргумент - переменная массива
	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		if arrSV == nil || (!arrSV.IsArray() && !arrSV.IsRef()) {
			return sv.NewArrayRef()
		}

		// Получаем данные
		var elements []*sv.SV
		if arrSV.IsRef() {
			deref := arrSV.Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if arrSV.IsArray() {
			elements = arrSV.ArrayData()
		}

		if elements == nil {
			return sv.NewArrayRef()
		}

		// Создаём копию для сортировки
		sorted := make([]*sv.SV, len(elements))
		copy(sorted, elements)

		// Сортируем вручную по строковому значению
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].AsString() > sorted[j].AsString() {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		return sv.NewArrayRef(sorted...)
	}

	// Если передан первый аргумент как значение
	if len(args) > 0 && args[0] != nil {
		var elements []*sv.SV
		if args[0].IsRef() {
			deref := args[0].Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if args[0].IsArray() {
			elements = args[0].ArrayData()
		}

		if elements != nil {
			sorted := make([]*sv.SV, len(elements))
			copy(sorted, elements)

			// Сортируем вручную
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].AsString() > sorted[j].AsString() {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}

			return sv.NewArrayRef(sorted...)
		}
	}

	return sv.NewArrayRef()
}

func (i *Interpreter) builtinSort(exprs []ast.Expression, args []*sv.SV) *sv.SV {
	if len(exprs) == 0 {
		return sv.NewArrayRef()
	}

	// Проверяем, если аргумент - переменная массива
	if arrVar, ok := exprs[0].(*ast.ArrayVar); ok {
		arrSV := i.ctx.GetVar(arrVar.Name)
		if arrSV == nil || (!arrSV.IsArray() && !arrSV.IsRef()) {
			return sv.NewArrayRef()
		}

		// Получаем данные
		var elements []*sv.SV
		if arrSV.IsRef() {
			deref := arrSV.Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if arrSV.IsArray() {
			elements = arrSV.ArrayData()
		}

		if elements == nil {
			return sv.NewArrayRef()
		}

		// Создаём копию для сортировки
		sorted := make([]*sv.SV, len(elements))
		copy(sorted, elements)

		// Сортируем
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].AsString() < sorted[j].AsString()
		})

		return sv.NewArrayRef(sorted...)
	}

	// Если передан первый аргумент как значение
	if len(args) > 0 && args[0] != nil {
		var elements []*sv.SV
		if args[0].IsRef() {
			deref := args[0].Deref()
			if deref != nil && deref.IsArray() {
				elements = deref.ArrayData()
			}
		} else if args[0].IsArray() {
			elements = args[0].ArrayData()
		}

		if elements != nil {
			sorted := make([]*sv.SV, len(elements))
			copy(sorted, elements)

			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].AsString() < sorted[j].AsString()
			})

			return sv.NewArrayRef(sorted...)
		}
	}

	return sv.NewArrayRef()
}

func (i *Interpreter) builtinExists(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) == 0 {
		return sv.NewString("")
	}

	// exists $hash{key}
	if hashAccess, ok := expr.Args[0].(*ast.HashAccess); ok {
		hash := i.evalExpression(hashAccess.Hash)
		key := i.evalExpression(hashAccess.Key)
		return hv.Exists(hash, key)
	}

	// exists $array[idx]
	if arrAccess, ok := expr.Args[0].(*ast.ArrayAccess); ok {
		arr := i.evalExpression(arrAccess.Array)
		idx := i.evalExpression(arrAccess.Index)
		return av.Exists(arr, idx)
	}

	return sv.NewString("")
}

func (i *Interpreter) builtinDelete(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) == 0 {
		return sv.NewUndef()
	}

	// delete $hash{key}
	if hashAccess, ok := expr.Args[0].(*ast.HashAccess); ok {
		hash := i.evalExpression(hashAccess.Hash)
		key := i.evalExpression(hashAccess.Key)
		return hv.Delete(hash, key)
	}

	// delete $array[idx]
	if arrAccess, ok := expr.Args[0].(*ast.ArrayAccess); ok {
		arr := i.evalExpression(arrAccess.Array)
		idx := i.evalExpression(arrAccess.Index)
		return av.Delete(arr, idx)
	}

	return sv.NewUndef()
}

// ============================================================
// НОВЫЕ BUILTIN ФУНКЦИИ ДЛЯ pkg/eval/builtins.go
// ============================================================

// grep - фильтрация массива
// grep { $_ > 5 } @arr  или  grep /pattern/, @arr
// grep - фильтрация массива
// grep { $_ > 5 } @arr  или  grep /pattern/, @arr
func (i *Interpreter) builtinGrep(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) < 2 {
		return sv.NewArrayRef()
	}

	// Получаем массив (второй аргумент)
	var elements []*sv.SV
	listArg := i.evalExpression(expr.Args[1])
	if listArg.IsRef() {
		if deref := listArg.Deref(); deref != nil && deref.IsArray() {
			elements = deref.ArrayData()
		}
	} else if listArg.IsArray() {
		elements = listArg.ArrayData()
	}

	var results []*sv.SV

	// Первый аргумент - блок или regex
	switch block := expr.Args[0].(type) {
	case *ast.AnonSubExpr:
		for _, el := range elements {
			i.ctx.SetVar("_", el)
			result := i.evalBlockStmt(block.Body)
			if result.AsBool() {
				results = append(results, el)
			}
		}
	case *ast.MatchExpr:
		// grep /pattern/, @arr
		for _, el := range elements {
			i.ctx.SetVar("_", el)
			result := i.evalMatchExpr(block)
			if result.AsBool() {
				results = append(results, el)
			}
		}
	default:
		// grep EXPR, @arr - вычисляем выражение для каждого элемента
		for _, el := range elements {
			i.ctx.SetVar("_", el)
			result := i.evalExpression(expr.Args[0])
			if result.AsBool() {
				results = append(results, el)
			}
		}
	}

	return sv.NewArraySV(results...)
}

// map - трансформация массива
// map { $_ * 2 } @arr  или  map { expr } @arr
func (i *Interpreter) builtinMap(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) < 2 {
		return sv.NewArrayRef()
	}

	// Получаем массив (второй аргумент)
	var elements []*sv.SV
	listArg := i.evalExpression(expr.Args[1])
	if listArg.IsRef() {
		if deref := listArg.Deref(); deref != nil && deref.IsArray() {
			elements = deref.ArrayData()
		}
	} else if listArg.IsArray() {
		elements = listArg.ArrayData()
	}

	var results []*sv.SV

	// Первый аргумент - блок
	switch block := expr.Args[0].(type) {
	case *ast.AnonSubExpr:
		// map { expr } @arr - анонимный блок
		for _, el := range elements {
			i.ctx.SetVar("_", el)
			result := i.evalBlockStmt(block.Body)
			// map может возвращать несколько значений
			if result.IsArray() {
				results = append(results, result.ArrayData()...)
			} else {
				results = append(results, result)
			}
		}
	default:
		// map EXPR, @arr
		for _, el := range elements {
			i.ctx.SetVar("_", el)
			result := i.evalExpression(expr.Args[0])
			results = append(results, result)
		}
	}

	return sv.NewArraySV(results...)
}

// wantarray - контекст вызова
// Возвращает: undef (void context), 0 (scalar context), 1 (list context)
func (i *Interpreter) builtinWantarray(args []*sv.SV) *sv.SV {
	_ = args
	ctx := i.ctx.Wantarray()
	if ctx == nil {
		return sv.NewUndef() // void context
	}
	if *ctx == 1 {
		return sv.NewInt(0) // scalar context - возвращаем false (но defined)
	}
	return sv.NewInt(1) // list context - возвращаем true
}

// each - итерация по хешу, возвращает (key, value)
func (i *Interpreter) builtinEach(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewArrayRef()
	}

	hash := args[0]
	if hash.IsRef() {
		hash = hash.Deref()
	}
	if hash == nil || !hash.IsHash() {
		return sv.NewArrayRef()
	}

	// Используем внутренний итератор хеша
	pair := hv.Each(hash)
	if len(pair) == 0 {
		return sv.NewArrayRef()
	}
	return sv.NewArrayRef(pair...)
}

// pos - позиция последнего совпадения regex
// pos($var) - получить позицию
// В Perl также можно pos($var) = N для установки, но это lvalue
func (i *Interpreter) builtinPos(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		// pos() без аргументов - позиция для $_
		pos, ok := i.ctx.GetPos("_")
		if !ok {
			return sv.NewUndef()
		}
		return sv.NewInt(int64(pos))
	}

	// pos($var) - нужно получить имя переменной
	// Но args[0] уже вычислен, поэтому мы не знаем имя
	// Упрощённая реализация: ищем по значению строки
	// TODO: для полной реализации нужно передавать expr

	// Пока возвращаем позицию для $_ если есть аргумент
	pos, ok := i.ctx.GetPos("_")
	if !ok {
		return sv.NewUndef()
	}
	return sv.NewInt(int64(pos))
}

// printf - форматированный вывод
func (i *Interpreter) builtinPrintf(args []*sv.SV) *sv.SV {
	if len(args) == 0 {
		return sv.NewInt(0)
	}

	format := args[0].AsString()
	fmtArgs := make([]interface{}, len(args)-1)

	// Парсим формат для определения типов (как в sprintf)
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

	result := fmt.Sprintf(format, fmtArgs...)
	fmt.Fprint(i.stdout, result)
	return sv.NewInt(int64(len(result)))
}

// eof - проверка конца файла
func (i *Interpreter) builtinEof(expr *ast.CallExpr) *sv.SV {
	// Без аргументов - проверяем ARGV или последний прочитанный файл
	if len(expr.Args) == 0 {
		return sv.NewInt(1) // По умолчанию EOF
	}

	// Получаем имя filehandle из AST
	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value
	default:
		fhName = i.evalExpression(expr.Args[0]).AsString()
	}

	fh := i.ctx.GetFileHandle(fhName)
	if fh == nil {
		return sv.NewInt(1) // Нет файла = EOF
	}

	// Проверяем есть ли ещё данные
	if fh.Scanner != nil {
		return sv.NewInt(0)
	}

	// Для файла проверяем позицию
	if fh.File != nil {
		pos, _ := fh.File.Seek(0, 1) // Текущая позиция
		fi, err := fh.File.Stat()
		if err == nil && pos >= fi.Size() {
			return sv.NewInt(1) // EOF
		}
		return sv.NewInt(0)
	}

	return sv.NewInt(1)
}

// tell - позиция в файле
func (i *Interpreter) builtinTell(expr *ast.CallExpr) *sv.SV {
	if len(expr.Args) == 0 {
		return sv.NewInt(-1)
	}
	// Получаем имя filehandle из AST
	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value
	default:
		// Fallback - вычисляем значение
		fhName = i.evalExpression(expr.Args[0]).AsString()
	}

	fh := i.ctx.GetFileHandle(fhName)
	if fh == nil || fh.File == nil {
		return sv.NewInt(-1)
	}

	pos, err := fh.File.Seek(0, 1) // SEEK_CUR = 1
	if err != nil {
		return sv.NewInt(-1)
	}
	return sv.NewInt(pos)
}

// seek - перемещение в файле
func (i *Interpreter) builtinSeek(expr *ast.CallExpr) *sv.SV {
	// seek(FH, POSITION, WHENCE)
	if len(expr.Args) < 3 {
		return sv.NewInt(0)
	}

	// Получаем имя filehandle из AST
	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value
	default:
		fhName = i.evalExpression(expr.Args[0]).AsString()
	}

	position := i.evalExpression(expr.Args[1]).AsInt()
	whence := int(i.evalExpression(expr.Args[2]).AsInt())

	fh := i.ctx.GetFileHandle(fhName)
	if fh == nil || fh.File == nil {
		return sv.NewInt(0)
	}

	_, err := fh.File.Seek(position, whence)
	if err != nil {
		return sv.NewInt(0)
	}

	// После seek нужно пересоздать Scanner если он был
	if fh.Scanner != nil {
		fh.Scanner = bufio.NewScanner(fh.File)
	}

	return sv.NewInt(1)
}

// read - чтение байтов из файла
func (i *Interpreter) builtinRead(expr *ast.CallExpr, args []*sv.SV) *sv.SV {
	// read(FH, SCALAR, LENGTH, [OFFSET])
	if len(args) < 3 {
		return sv.NewUndef()
	}

	fhName := args[0].AsString()
	length := int(args[2].AsInt())
	offset := 0
	if len(args) >= 4 {
		offset = int(args[3].AsInt())
	}

	fh := i.ctx.GetFileHandle(fhName)
	if fh == nil || fh.File == nil {
		return sv.NewUndef()
	}

	// Читаем данные
	buf := make([]byte, length)
	n, err := fh.File.Read(buf)
	if err != nil && n == 0 {
		return sv.NewUndef()
	}

	// Получаем переменную для записи
	if len(expr.Args) >= 2 {
		if scalarVar, ok := expr.Args[1].(*ast.ScalarVar); ok {
			// Получаем текущее значение или создаём новое
			current := i.ctx.GetVar(scalarVar.Name)
			var result string
			if current != nil && offset > 0 {
				result = current.AsString()
				// Расширяем если нужно
				for len(result) < offset {
					result += "\x00"
				}
				result = result[:offset] + string(buf[:n])
			} else {
				result = string(buf[:n])
			}
			i.ctx.SetVar(scalarVar.Name, sv.NewString(result))
		}
	}

	return sv.NewInt(int64(n))
}

// binmode - установка бинарного режима
func (i *Interpreter) builtinBinmode(expr *ast.CallExpr) *sv.SV {
	// binmode(FH) или binmode(FH, LAYER)
	// В Go файлы уже бинарные по умолчанию
	if len(expr.Args) == 0 {
		return sv.NewInt(1)
	}

	// Получаем имя filehandle из AST
	var fhName string
	switch fh := expr.Args[0].(type) {
	case *ast.ScalarVar:
		fhName = fh.Name
	case *ast.Identifier:
		fhName = fh.Value // STDOUT, STDERR, STDIN
	default:
		fhName = i.evalExpression(expr.Args[0]).AsString()
	}

	// Для стандартных потоков всегда успех
	if fhName == "STDOUT" || fhName == "STDERR" || fhName == "STDIN" {
		return sv.NewInt(1)
	}

	fh := i.ctx.GetFileHandle(fhName)
	if fh == nil {
		return sv.NewInt(0)
	}

	// Layer (":utf8", ":raw", etc.) - пока игнорируем
	return sv.NewInt(1)
}
