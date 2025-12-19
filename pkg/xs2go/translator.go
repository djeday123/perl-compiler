// pkg/xs2go/translator.go
package xs2go

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Translator struct {
	input     string
	output    strings.Builder
	functions map[string]*XSFunction
	module    string
	package_  string
}

type XSArg struct {
	Name string
	Type string
}

type XSFunction struct {
	Name       string
	ReturnType string
	Args       []XSArg
	Code       string
	OutputVar  string
}

func New() *Translator {
	return &Translator{
		functions: make(map[string]*XSFunction),
	}
}

func (t *Translator) Translate(xsFile string) (string, error) {
	content, err := os.ReadFile(xsFile)
	if err != nil {
		return "", err
	}

	t.input = string(content)

	// Парсим XS
	if err := t.parseXS(); err != nil {
		return "", err
	}

	// Генерируем Go
	t.generateHeader()

	for _, fn := range t.functions {
		t.translateFunction(fn)
	}

	t.generateInit()

	return t.output.String(), nil
}

func (t *Translator) parseXS() error {
	content := t.input

	// Удаляем C комментарии
	commentRe := regexp.MustCompile(`/\*.*?\*/`)
	content = commentRe.ReplaceAllString(content, "")

	// Удаляем #include и другие директивы препроцессора
	preprocessorRe := regexp.MustCompile(`(?m)^#.*$`)
	content = preprocessorRe.ReplaceAllString(content, "")

	// Парсим MODULE = ... PACKAGE = ...
	moduleRe := regexp.MustCompile(`MODULE\s*=\s*(\S+)\s+PACKAGE\s*=\s*(\S+)`)
	if matches := moduleRe.FindStringSubmatch(content); len(matches) >= 3 {
		t.module = matches[1]
		t.package_ = matches[2]
	}

	// Ищем функции
	blocks := regexp.MustCompile(`(?s)(\w+(?:\s*\*)?)\s*\n(\w+)\s*\(([^)]*)\)(.*?)OUTPUT:\s*\n\s*(\w+)`).
		FindAllStringSubmatch(content, -1)

	for _, block := range blocks {
		if len(block) >= 6 {
			fn := &XSFunction{
				ReturnType: strings.TrimSpace(block[1]),
				Name:       strings.TrimSpace(block[2]),
				OutputVar:  strings.TrimSpace(block[5]),
			}

			// Аргументы из сигнатуры (имена)
			argsStr := strings.TrimSpace(block[3])
			argNames := []string{}
			if argsStr != "" {
				for _, arg := range strings.Split(argsStr, ",") {
					argNames = append(argNames, strings.TrimSpace(arg))
				}
			}

			// Парсим тело (между сигнатурой и OUTPUT)
			body := block[4]

			// Разделяем на часть до CODE: и после
			codeParts := regexp.MustCompile(`(?s)(.*)CODE:\s*(.*)`).FindStringSubmatch(body)

			argDeclarations := ""
			codeBody := ""

			if len(codeParts) >= 3 {
				argDeclarations = codeParts[1]
				codeBody = codeParts[2]
			}

			fn.Code = strings.TrimSpace(codeBody)

			// Парсим декларации аргументов (только часть ДО CODE:)
			lines := strings.Split(argDeclarations, "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Формат: "SV *name" или "int a"
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					typeName := strings.Join(parts[:len(parts)-1], " ")
					varName := parts[len(parts)-1]
					varName = strings.TrimPrefix(varName, "*")
					fn.Args = append(fn.Args, XSArg{
						Type: typeName,
						Name: varName,
					})
				} else if len(parts) == 1 && i < len(argNames) {
					// Только тип, имя берём из сигнатуры
					fn.Args = append(fn.Args, XSArg{
						Type: parts[0],
						Name: argNames[i],
					})
				}
			}

			// Если аргументы не найдены в теле, используем имена из сигнатуры
			if len(fn.Args) == 0 && len(argNames) > 0 {
				for _, name := range argNames {
					fn.Args = append(fn.Args, XSArg{
						Type: "SV *",
						Name: name,
					})
				}
			}

			t.functions[fn.Name] = fn
		}
	}

	return nil
}

func (t *Translator) parseArgs(argsStr string) []XSArg {
	var args []XSArg

	parts := strings.Split(argsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Формат: "SV* name" или "int count"
		tokens := strings.Fields(part)
		if len(tokens) >= 2 {
			typeName := strings.Join(tokens[:len(tokens)-1], " ")
			varName := tokens[len(tokens)-1]
			// Убираем * из имени если прилип
			varName = strings.TrimPrefix(varName, "*")
			args = append(args, XSArg{
				Type: typeName,
				Name: varName,
			})
		} else if len(tokens) == 1 {
			// Только тип, имя сгенерируем
			args = append(args, XSArg{
				Type: tokens[0],
				Name: fmt.Sprintf("arg%d", len(args)),
			})
		}
	}

	return args
}

func (t *Translator) generateHeader() {
	t.output.WriteString("// Auto-generated from XS\n")
	t.output.WriteString(fmt.Sprintf("// Module: %s, Package: %s\n\n", t.module, t.package_))
	t.output.WriteString("package main\n\n")
	t.output.WriteString("import (\n")
	t.output.WriteString("\t\"fmt\"\n")
	t.output.WriteString("\t\"strings\"\n")
	t.output.WriteString(")\n\n")
	t.output.WriteString("var _ = fmt.Sprint\n")
	t.output.WriteString("var _ = strings.Cut\n\n")
}

func (t *Translator) translateFunction(fn *XSFunction) {
	// Имя функции: Package::func → perl_Package_func
	goName := "perl_" + strings.ReplaceAll(t.package_, "::", "_") + "_" + fn.Name

	t.output.WriteString(fmt.Sprintf("func %s(args ...*SV) *SV {\n", goName))

	// Извлекаем аргументы
	for i, arg := range fn.Args {
		goType := t.mapType(arg.Type)
		t.output.WriteString(fmt.Sprintf("\t%s := %s\n", arg.Name, t.extractArg(i, arg.Type, goType)))
	}

	if len(fn.Args) > 0 {
		t.output.WriteString("\n")
	}

	// Переменная для результата
	if fn.OutputVar == "RETVAL" {
		t.output.WriteString("\tvar RETVAL *SV\n\n")
	}

	// Транслируем тело
	goCode := t.translateCCode(fn.Code)

	// Добавляем отступы
	lines := strings.Split(goCode, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			t.output.WriteString("\t" + line + "\n")
		}
	}

	t.output.WriteString("\n\treturn RETVAL\n")
	t.output.WriteString("}\n\n")
}

func (t *Translator) mapType(cType string) string {
	cType = strings.TrimSpace(cType)

	switch cType {
	case "SV*", "SV *":
		return "*SV"
	case "AV*", "AV *":
		return "*SV"
	case "HV*", "HV *":
		return "*SV"
	case "int", "I32", "IV":
		return "int64"
	case "double", "NV":
		return "float64"
	case "char*", "char *":
		return "string"
	case "bool":
		return "bool"
	case "void":
		return ""
	default:
		return "*SV"
	}
}

func (t *Translator) extractArg(index int, cType, goType string) string {
	_ = cType
	switch goType {
	case "*SV":
		return fmt.Sprintf("args[%d]", index)
	case "int64":
		return fmt.Sprintf("args[%d].AsInt()", index)
	case "float64":
		return fmt.Sprintf("args[%d].AsFloat()", index)
	case "string":
		return fmt.Sprintf("args[%d].AsString()", index)
	case "bool":
		return fmt.Sprintf("args[%d].IsTrue()", index)
	default:
		return fmt.Sprintf("args[%d]", index)
	}
}

func (t *Translator) translateCCode(cCode string) string {
	code := cCode

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// char *str = SvPV_nolen(name); -> str := name.AsString()
		if matches := regexp.MustCompile(`char\s*\*\s*(\w+)\s*=\s*SvPV_nolen\s*\(\s*(\w+)\s*\)\s*;?`).FindStringSubmatch(line); len(matches) >= 3 {
			line = fmt.Sprintf("%s := %s.AsString()", matches[1], matches[2])
			result = append(result, line)
			continue
		}

		// char *str = SvPV(name, len); -> str := name.AsString()
		if matches := regexp.MustCompile(`char\s*\*\s*(\w+)\s*=\s*SvPV\s*\(\s*(\w+)\s*,\s*\w+\s*\)\s*;?`).FindStringSubmatch(line); len(matches) >= 3 {
			line = fmt.Sprintf("%s := %s.AsString()", matches[1], matches[2])
			result = append(result, line)
			continue
		}

		// newSVpvf("...", args) -> svStr(fmt.Sprintf("...", args))
		line = regexp.MustCompile(`newSVpvf\s*\(\s*"([^"]+)",\s*([^)]+)\)`).ReplaceAllString(line, `svStr(fmt.Sprintf("$1", $2))`)

		// newSVpv(str, 0) -> svStr(str)
		line = regexp.MustCompile(`newSVpv\s*\(\s*([^,]+),\s*\d+\s*\)`).ReplaceAllString(line, "svStr($1)")

		// newSViv(x) -> svInt(x)
		line = regexp.MustCompile(`newSViv\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(line, "svInt(int64($1))")

		// RETVAL = a + b; -> RETVAL = svInt(int64(a + b))
		if strings.HasPrefix(line, "RETVAL = ") && !strings.Contains(line, "svStr") && !strings.Contains(line, "svInt") && !strings.Contains(line, "svFloat") {
			expr := strings.TrimPrefix(line, "RETVAL = ")
			expr = strings.TrimSuffix(expr, ";")
			if !strings.Contains(expr, "\"") {
				line = fmt.Sprintf("RETVAL = svInt(int64(%s))", expr)
			}
		}

		// Убираем лишние точки с запятой
		line = strings.TrimSuffix(line, ";")

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func (t *Translator) generateInit() {
	t.output.WriteString("func init() {\n")

	for _, fn := range t.functions {
		goName := "perl_" + strings.ReplaceAll(t.package_, "::", "_") + "_" + fn.Name
		perlName := t.package_ + "::" + fn.Name
		t.output.WriteString(fmt.Sprintf("\tperl_register_sub(\"%s\", %s)\n", perlName, goName))
	}

	t.output.WriteString("}\n")
}

// Дополнительные методы для SV
const svHelpers = `
// Helper methods for SV
func (sv *SV) IsRef() bool {
	return sv != nil && sv.ref != nil
}

func (sv *SV) IsHash() bool {
	return sv != nil && sv.flags&SVf_HOK != 0
}

func (sv *SV) IsArray() bool {
	return sv != nil && sv.flags&SVf_AOK != 0
}

func (sv *SV) IsInt() bool {
	return sv != nil && sv.flags&SVf_IOK != 0
}

func (sv *SV) IsFloat() bool {
	return sv != nil && sv.flags&SVf_NOK != 0
}

func (sv *SV) Deref() *SV {
	if sv.ref != nil {
		return sv.ref
	}
	return sv
}

type HashIterator struct {
	hv    *SV
	keys  []string
	index int
	key   string
	value *SV
}

func (sv *SV) HashIterator() *HashIterator {
	keys := make([]string, 0, len(sv.hv))
	for k := range sv.hv {
		keys = append(keys, k)
	}
	return &HashIterator{hv: sv, keys: keys, index: -1}
}

func (it *HashIterator) Valid() bool {
	return it.index < len(it.keys)
}

func (it *HashIterator) Next() *HashIterator {
	it.index++
	if it.index < len(it.keys) {
		it.key = it.keys[it.index]
		it.value = it.hv.hv[it.key]
	}
	return it
}

func (it *HashIterator) Key() *SV {
	return svStr(it.key)
}

func (it *HashIterator) Value() *SV {
	return it.value
}
`
