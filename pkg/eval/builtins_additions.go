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
