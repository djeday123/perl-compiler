// pkg/c2go/translator.go
package c2go

import (
	"fmt"
	"regexp"
	"strings"
)

// Translator транслирует C код в Go
type Translator struct {
	// Можно добавить настройки
}

func New() *Translator {
	return &Translator{}
}

// Translate транслирует C код в Go
func (t *Translator) Translate(cCode string) string {
	lines := strings.Split(cCode, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if t.isSkipLine(line) {
			continue
		}

		translated := t.TranslateLine(line)
		if translated != "" {
			result = append(result, translated)
		}
	}

	return strings.Join(result, "\n")
}

// isSkipLine проверяет нужно ли пропустить строку
func (t *Translator) isSkipLine(line string) bool {
	skipPrefixes := []string{
		// Perl/XS макросы
		"PUTBACK", "SPAGAIN", "EXTEND", "PUSHMARK",
		"XPUSHs", "PUSHs", "POPs", "TOPs",
		"ENTER", "LEAVE", "SAVETMPS", "FREETMPS",
		"dSP", "dXSARGS", "dMARK", "dITEMS",
		"XSRETURN", "RETURN",
		"SvREFCNT_inc", "SvREFCNT_dec",
		"SvREADONLY",
		"ST (",
		// Память
		"Safefree", "Newx",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	// goto и метки
	if strings.HasPrefix(line, "goto ") {
		return true
	}
	if regexp.MustCompile(`^\w+:$`).MatchString(line) && line != "default:" {
		return true
	}

	return false
}

// TranslateLine транслирует одну строку C в Go
func (t *Translator) TranslateLine(line string) string {
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	// === Объявления переменных ===

	// char *str = SvPV_nolen(x)
	if m := regexp.MustCompile(`^char\s*\*\s*(\w+)\s*=\s*SvPV_nolen\s*\(\s*(\w+)\s*\)`).FindStringSubmatch(line); len(m) >= 3 {
		return fmt.Sprintf("%s := %s.AsString()", m[1], m[2])
	}

	// char *str = SvPV(x, len)
	if m := regexp.MustCompile(`^char\s*\*\s*(\w+)\s*=\s*SvPV\s*\(\s*(\w+)\s*,`).FindStringSubmatch(line); len(m) >= 3 {
		return fmt.Sprintf("%s := %s.AsString()", m[1], m[2])
	}

	// char *str;
	if m := regexp.MustCompile(`^char\s*\*\s*(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s string", m[1])
	}

	// SV *var = expr
	if m := regexp.MustCompile(`^SV\s*\*\s*(\w+)\s*=\s*(.+)`).FindStringSubmatch(line); len(m) >= 3 {
		return fmt.Sprintf("%s := %s", m[1], t.TranslateExpr(m[2]))
	}

	// SV *var;
	if m := regexp.MustCompile(`^SV\s*\*\s*(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s *SV", m[1])
	}

	// AV *var;
	if m := regexp.MustCompile(`^AV\s*\*\s*(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s *SV // array", m[1])
	}

	// HV *var;
	if m := regexp.MustCompile(`^HV\s*\*\s*(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s *SV // hash", m[1])
	}

	// int var = expr
	if m := regexp.MustCompile(`^int\s+(\w+)\s*=\s*(.+)`).FindStringSubmatch(line); len(m) >= 3 {
		return fmt.Sprintf("%s := int64(%s)", m[1], t.TranslateExpr(m[2]))
	}

	// int var;
	if m := regexp.MustCompile(`^int\s+(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s int64", m[1])
	}

	// STRLEN var;
	if m := regexp.MustCompile(`^STRLEN\s+(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s int", m[1])
	}

	// UV var;
	if m := regexp.MustCompile(`^UV\s+(\w+)$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("var %s uint64", m[1])
	}

	// === Присваивания ===

	// RETVAL = expr
	if m := regexp.MustCompile(`^RETVAL\s*=\s*(.+)`).FindStringSubmatch(line); len(m) >= 2 {
		expr := t.TranslateExpr(m[1])
		// Если выражение простое (арифметика) и не обёрнуто в sv*, оборачиваем
		if !strings.HasPrefix(expr, "sv") && !strings.HasPrefix(expr, "nil") {
			// Проверяем что это арифметика или переменная
			if regexp.MustCompile(`^[\w\s\+\-\*/%\(\)]+$`).MatchString(expr) {
				return fmt.Sprintf("RETVAL = svInt(int64(%s))", expr)
			}
		}
		return fmt.Sprintf("RETVAL = %s", expr)
	}

	// var = expr
	if m := regexp.MustCompile(`^(\w+)\s*=\s*(.+)`).FindStringSubmatch(line); len(m) >= 3 {
		return fmt.Sprintf("%s = %s", m[1], t.TranslateExpr(m[2]))
	}

	// === Условия ===

	// if (expr) {
	if m := regexp.MustCompile(`^if\s*\(\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("if %s {", t.TranslateCondition(m[1]))
	}

	// } else if (expr) {
	if m := regexp.MustCompile(`^}\s*else\s+if\s*\(\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("} else if %s {", t.TranslateCondition(m[1]))
	}

	// else if (expr) {
	if m := regexp.MustCompile(`^else\s+if\s*\(\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("} else if %s {", t.TranslateCondition(m[1]))
	}

	// } else {
	if regexp.MustCompile(`^}\s*else\s*\{?$`).MatchString(line) {
		return "} else {"
	}

	// else {
	if regexp.MustCompile(`^else\s*\{?$`).MatchString(line) {
		return "} else {"
	}

	// }
	if line == "}" {
		return "}"
	}

	// {
	if line == "{" {
		return "{"
	}

	// === Циклы ===

	// while (expr) {
	if m := regexp.MustCompile(`^while\s*\(\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("for %s {", t.TranslateCondition(m[1]))
	}

	// for (;;) {
	if regexp.MustCompile(`^for\s*\(\s*;;\s*\)\s*\{?$`).MatchString(line) {
		return "for {"
	}

	// for (init; cond; post) {
	if m := regexp.MustCompile(`^for\s*\(\s*(.+);\s*(.+);\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 4 {
		init := t.TranslateExpr(m[1])
		cond := t.TranslateCondition(m[2])
		post := t.TranslateExpr(m[3])
		return fmt.Sprintf("for %s; %s; %s {", init, cond, post)
	}

	// do {
	if line == "do {" || line == "do" {
		return "for {"
	}

	// } while (expr);
	if m := regexp.MustCompile(`^}\s*while\s*\(\s*(.+)\s*\)`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("if !(%s) { break } }", t.TranslateCondition(m[1]))
	}

	// break;
	if line == "break" {
		return "break"
	}

	// continue;
	if line == "continue" {
		return "continue"
	}

	// === Switch ===

	// switch (expr) {
	if m := regexp.MustCompile(`^switch\s*\(\s*(.+)\s*\)\s*\{?$`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("switch %s {", t.TranslateExpr(m[1]))
	}

	// case 'x':
	if m := regexp.MustCompile(`^case\s+'(.)':`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("case '%s':", m[1])
	}

	// case X:
	if m := regexp.MustCompile(`^case\s+(\w+):`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("case %s:", m[1])
	}

	// default:
	if line == "default:" {
		return "default:"
	}

	// === Вызовы функций ===

	// croak(...)
	if m := regexp.MustCompile(`^croak\s*\((.+)\)`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("panic(%s)", m[1])
	}

	// warn(...)
	if m := regexp.MustCompile(`^warn\s*\((.+)\)`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("fmt.Fprintf(os.Stderr, %s)", m[1])
	}

	// return expr
	if m := regexp.MustCompile(`^return\s+(.+)`).FindStringSubmatch(line); len(m) >= 2 {
		return fmt.Sprintf("return %s", t.TranslateExpr(m[1]))
	}

	// return;
	if line == "return" {
		return "return"
	}

	// Общая трансляция
	return t.TranslateExpr(line)
}

// TranslateExpr транслирует C выражение в Go
func (t *Translator) TranslateExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	expr = strings.TrimSuffix(expr, ";")

	// === Создание SV ===

	// newSVpvf("fmt", args)
	if m := regexp.MustCompile(`newSVpvf\s*\(\s*"([^"]+)"\s*,\s*(.+)\)`).FindStringSubmatch(expr); len(m) >= 3 {
		return fmt.Sprintf(`svStr(fmt.Sprintf("%s", %s))`, m[1], t.TranslateExpr(m[2]))
	}

	// newSVpv(str, len)
	expr = regexp.MustCompile(`newSVpv\s*\(\s*([^,]+)\s*,\s*\d+\s*\)`).ReplaceAllString(expr, "svStr($1)")

	// newSVpvn(str, len)
	expr = regexp.MustCompile(`newSVpvn\s*\(\s*([^,]+)\s*,\s*\d+\s*\)`).ReplaceAllString(expr, "svStr($1)")

	// newSVpvn("", 0)
	expr = regexp.MustCompile(`newSVpvn\s*\(\s*""\s*,\s*0\s*\)`).ReplaceAllString(expr, `svStr("")`)

	// newSViv(x)
	expr = regexp.MustCompile(`newSViv\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(expr, "svInt(int64($1))")

	// newSVuv(x)
	expr = regexp.MustCompile(`newSVuv\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(expr, "svInt(int64($1))")

	// newSVnv(x)
	expr = regexp.MustCompile(`newSVnv\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(expr, "svFloat($1)")

	// newSVsv(x)
	expr = regexp.MustCompile(`newSVsv\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(expr, "svCopy($1)")

	// newRV_inc(x) / newRV_noinc(x)
	expr = regexp.MustCompile(`newRV_(?:inc|noinc)\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(expr, "svRef($1)")

	// newHV()
	expr = regexp.MustCompile(`newHV\s*\(\s*\)`).ReplaceAllString(expr, "svHash()")

	// newAV()
	expr = regexp.MustCompile(`newAV\s*\(\s*\)`).ReplaceAllString(expr, "svArray()")

	// === Доступ к SV ===

	// SvPV_nolen(x)
	expr = regexp.MustCompile(`SvPV_nolen\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.AsString()")

	// SvPV(x, len)
	expr = regexp.MustCompile(`SvPV\s*\(\s*(\w+)\s*,\s*\w+\s*\)`).ReplaceAllString(expr, "$1.AsString()")

	// SvIV(x)
	expr = regexp.MustCompile(`SvIV\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.AsInt()")

	// SvUV(x)
	expr = regexp.MustCompile(`SvUV\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "uint64($1.AsInt())")

	// SvNV(x)
	expr = regexp.MustCompile(`SvNV\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.AsFloat()")

	// SvTRUE(x)
	expr = regexp.MustCompile(`SvTRUE\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.IsTrue()")

	// SvOK(x)
	expr = regexp.MustCompile(`SvOK\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "(!$1.IsUndef())")

	// SvROK(x)
	expr = regexp.MustCompile(`SvROK\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.IsRef()")

	// SvRV(x)
	expr = regexp.MustCompile(`SvRV\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "$1.Deref()")

	// SvCUR(x) - длина строки
	expr = regexp.MustCompile(`SvCUR\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "len($1.AsString())")

	// === Хеши ===

	// hv_store(hv, key, len, val, hash)
	expr = regexp.MustCompile(`hv_store\s*\(\s*(\w+)\s*,\s*([^,]+)\s*,\s*\d+\s*,\s*([^,]+)\s*,\s*\d+\s*\)`).
		ReplaceAllString(expr, "svHSet($1, svStr($2), $3)")

	// hv_fetch(hv, key, len, lval)
	expr = regexp.MustCompile(`hv_fetch\s*\(\s*(\w+)\s*,\s*([^,]+)\s*,\s*\d+\s*,\s*\d+\s*\)`).
		ReplaceAllString(expr, "svHGet($1, svStr($2))")

	// hv_exists(hv, key, len)
	expr = regexp.MustCompile(`hv_exists\s*\(\s*(\w+)\s*,\s*([^,]+)\s*,\s*\d+\s*\)`).
		ReplaceAllString(expr, "svHExists($1, svStr($2))")

	// hv_delete(hv, key, len, flags)
	expr = regexp.MustCompile(`hv_delete\s*\(\s*(\w+)\s*,\s*([^,]+)\s*,\s*\d+\s*,\s*\w+\s*\)`).
		ReplaceAllString(expr, "svHDelete($1, svStr($2))")

	// === Массивы ===

	// av_push(av, val)
	expr = regexp.MustCompile(`av_push\s*\(\s*(\w+)\s*,\s*([^)]+)\s*\)`).
		ReplaceAllString(expr, "svPush($1, $2)")

	// av_pop(av)
	expr = regexp.MustCompile(`av_pop\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "svPop($1)")

	// av_shift(av)
	expr = regexp.MustCompile(`av_shift\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "svShift($1)")

	// av_len(av)
	expr = regexp.MustCompile(`av_len\s*\(\s*(\w+)\s*\)`).ReplaceAllString(expr, "(len($1.av)-1)")

	// av_fetch(av, idx, lval)
	expr = regexp.MustCompile(`av_fetch\s*\(\s*(\w+)\s*,\s*([^,]+)\s*,\s*\d+\s*\)`).
		ReplaceAllString(expr, "svAGet($1, $2)")

	// === Приведение типов ===

	// (HV *)x
	expr = regexp.MustCompile(`\(\s*HV\s*\*\s*\)\s*(\w+)`).ReplaceAllString(expr, "$1")

	// (AV *)x
	expr = regexp.MustCompile(`\(\s*AV\s*\*\s*\)\s*(\w+)`).ReplaceAllString(expr, "$1")

	// (SV *)x
	expr = regexp.MustCompile(`\(\s*SV\s*\*\s*\)\s*(\w+)`).ReplaceAllString(expr, "$1")

	// (IV)x
	expr = regexp.MustCompile(`\(\s*IV\s*\)\s*(\w+)`).ReplaceAllString(expr, "int64($1)")

	// (UV)x
	expr = regexp.MustCompile(`\(\s*UV\s*\)\s*(\w+)`).ReplaceAllString(expr, "uint64($1)")

	// === Константы ===
	expr = strings.ReplaceAll(expr, "NULL", "nil")
	expr = strings.ReplaceAll(expr, "&PL_sv_undef", "svUndef()")

	return expr
}

// TranslateCondition транслирует условие C в Go
func (t *Translator) TranslateCondition(cond string) string {
	cond = strings.TrimSpace(cond)

	// Убираем внешние скобки если они парные
	if strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") {
		inner := cond[1 : len(cond)-1]
		depth := 0
		valid := true
		for _, ch := range inner {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				if depth < 0 {
					valid = false
					break
				}
			}
		}
		if valid && depth == 0 {
			cond = inner
		}
	}

	// !x -> !x
	// x == y -> x == y
	// Применяем общую трансляцию
	return t.TranslateExpr(cond)
}
