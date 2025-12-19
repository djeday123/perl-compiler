// pkg/xs2go/translator.go
package xs2go

import (
	"fmt"
	"os"
	"perlc/pkg/c2go"
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

	// Удаляем C комментарии /* ... */
	commentRe := regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = commentRe.ReplaceAllString(content, "")

	// Удаляем #include, #define и другие директивы
	preprocessorRe := regexp.MustCompile(`(?m)^#\s*(include|define|ifdef|ifndef|endif|else|elif|if|undef).*$`)
	content = preprocessorRe.ReplaceAllString(content, "")

	// Удаляем typedef и struct
	typedefRe := regexp.MustCompile(`(?s)typedef\s+(?:struct|enum)?\s*\{[^}]*\}[^;]*;`)
	content = typedefRe.ReplaceAllString(content, "")

	structRe := regexp.MustCompile(`(?s)struct\s+\w+\s*\{[^}]*\}\s*;`)
	content = structRe.ReplaceAllString(content, "")

	// Парсим MODULE = ... PACKAGE = ...
	moduleRe := regexp.MustCompile(`MODULE\s*=\s*(\S+)\s+PACKAGE\s*=\s*(\S+)`)
	if matches := moduleRe.FindStringSubmatch(content); len(matches) >= 3 {
		t.module = matches[1]
		t.package_ = matches[2]
	}

	// Находим начало XS секции (после MODULE =)
	moduleIdx := strings.Index(content, "MODULE =")
	if moduleIdx == -1 {
		return nil
	}
	xsContent := content[moduleIdx:]

	// Ищем функции в двух форматах:
	// Формат 1: тип и имя на одной строке - "int add(a, b)"
	// Формат 2: тип на отдельной строке - "SV *\nhello(name)"

	// Формат 2: тип на отдельной строке (XS стиль)
	funcRe2 := regexp.MustCompile(`(?m)^(\w+\s*\*?)\s*\n(\w+)\s*\(([^)]*)\)`)
	matches2 := funcRe2.FindAllStringSubmatchIndex(xsContent, -1)

	for i, match := range matches2 {
		if len(match) < 8 {
			continue
		}

		returnType := strings.TrimSpace(xsContent[match[2]:match[3]])
		funcName := strings.TrimSpace(xsContent[match[4]:match[5]])
		argsStr := strings.TrimSpace(xsContent[match[6]:match[7]])

		funcStart := match[0]
		funcEnd := len(xsContent)
		if i+1 < len(matches2) {
			funcEnd = matches2[i+1][0]
		}

		funcBody := xsContent[funcStart:funcEnd]

		fn := t.parseXSFunction(returnType, funcName, argsStr, funcBody)
		if fn != nil && fn.Name != "" {
			t.functions[fn.Name] = fn
		}
	}

	// Формат 1: тип и имя на одной строке (fallback)
	funcRe1 := regexp.MustCompile(`(?m)^(\w+)\s+(\w+)\s*\(([^)]*)\)`)
	matches1 := funcRe1.FindAllStringSubmatchIndex(xsContent, -1)

	for i, match := range matches1 {
		if len(match) < 8 {
			continue
		}

		returnType := xsContent[match[2]:match[3]]
		funcName := xsContent[match[4]:match[5]]
		argsStr := xsContent[match[6]:match[7]]

		// Пропускаем если уже нашли эту функцию
		if _, exists := t.functions[funcName]; exists {
			continue
		}

		funcStart := match[0]
		funcEnd := len(xsContent)
		if i+1 < len(matches1) {
			funcEnd = matches1[i+1][0]
		}

		funcBody := xsContent[funcStart:funcEnd]

		fn := t.parseXSFunction(returnType, funcName, argsStr, funcBody)
		if fn != nil && fn.Name != "" {
			t.functions[fn.Name] = fn
		}
	}

	return nil
}

func (t *Translator) parseXSFunction(returnType, funcName, argsStr, body string) *XSFunction {
	// Пропускаем служебные функции
	if t.isSkippable(funcName, "") {
		return nil
	}

	// Пропускаем если это не XS функция (нет PPCODE: или CODE:)
	hasPPCode := strings.Contains(body, "PPCODE:")
	hasCode := strings.Contains(body, "CODE:")

	if !hasPPCode && !hasCode {
		return nil
	}

	fn := &XSFunction{
		ReturnType: strings.TrimSpace(returnType),
		Name:       strings.TrimSpace(funcName),
	}

	// Получаем имена аргументов из сигнатуры
	argNames := []string{}
	if argsStr != "" {
		for _, arg := range strings.Split(argsStr, ",") {
			arg = strings.TrimSpace(arg)
			if arg != "" {
				argNames = append(argNames, arg)
			}
		}
	}

	if hasPPCode {
		// Извлекаем код после PPCODE:
		idx := strings.Index(body, "PPCODE:")
		if idx != -1 {
			// Парсим типы аргументов из части до PPCODE:
			beforeCode := body[:idx]
			fn.Args = t.parseArgTypes(beforeCode, argNames)

			code := body[idx+len("PPCODE:"):]
			fn.Code = t.cleanPPCode(code)
		}
	} else if hasCode {
		codeIdx := strings.Index(body, "CODE:")
		outputIdx := strings.Index(body, "OUTPUT:")

		if codeIdx != -1 {
			// Парсим типы аргументов из части до CODE:
			beforeCode := body[:codeIdx]
			fn.Args = t.parseArgTypes(beforeCode, argNames)

			codeStart := codeIdx + len("CODE:")
			codeEnd := len(body)
			if outputIdx != -1 && outputIdx > codeIdx {
				codeEnd = outputIdx
			}
			fn.Code = strings.TrimSpace(body[codeStart:codeEnd])
		}

		// Извлекаем OUTPUT переменную
		if outputIdx != -1 {
			outputSection := body[outputIdx+len("OUTPUT:"):]
			lines := strings.Split(outputSection, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, ":") {
					fn.OutputVar = line
					break
				}
			}
		}
	}

	// Если типы не найдены, используем SV* по умолчанию
	if len(fn.Args) == 0 && len(argNames) > 0 {
		for _, name := range argNames {
			fn.Args = append(fn.Args, XSArg{
				Type: "SV *",
				Name: name,
			})
		}
	}

	return fn
}

// parseArgTypes парсит типы аргументов из секции между сигнатурой и CODE:
func (t *Translator) parseArgTypes(section string, argNames []string) []XSArg {
	var args []XSArg

	lines := strings.Split(section, "\n")

	// Создаём map для быстрого поиска
	argSet := make(map[string]bool)
	for _, name := range argNames {
		argSet[name] = true
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Пропускаем если это сигнатура функции
		if strings.Contains(line, "(") && strings.Contains(line, ")") {
			continue
		}

		// Формат: "int a" или "SV *name" или "char *str"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			varName := parts[len(parts)-1]
			varName = strings.TrimPrefix(varName, "*")

			// Проверяем что это один из наших аргументов
			if argSet[varName] {
				typeName := strings.Join(parts[:len(parts)-1], " ")
				if strings.HasPrefix(parts[len(parts)-1], "*") {
					typeName += " *"
				}
				args = append(args, XSArg{
					Type: typeName,
					Name: varName,
				})
			}
		}
	}

	// Сохраняем порядок аргументов как в сигнатуре
	orderedArgs := make([]XSArg, 0, len(argNames))
	argMap := make(map[string]XSArg)
	for _, arg := range args {
		argMap[arg.Name] = arg
	}

	for _, name := range argNames {
		if arg, ok := argMap[name]; ok {
			orderedArgs = append(orderedArgs, arg)
		} else {
			// Тип не найден - используем SV* по умолчанию
			orderedArgs = append(orderedArgs, XSArg{
				Type: "SV *",
				Name: name,
			})
		}
	}

	return orderedArgs
}

func (t *Translator) parseXSArgs(argsStr string) []XSArg {
	var args []XSArg

	if argsStr == "" {
		return args
	}

	parts := strings.Split(argsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Убираем значение по умолчанию: "int enable = 1" -> "int enable"
		if idx := strings.Index(part, "="); idx != -1 {
			part = strings.TrimSpace(part[:idx])
		}

		tokens := strings.Fields(part)
		if len(tokens) >= 2 {
			varName := tokens[len(tokens)-1]
			varName = strings.TrimPrefix(varName, "*")
			typeName := strings.Join(tokens[:len(tokens)-1], " ")

			// Если * было в имени, добавляем к типу
			if strings.HasPrefix(tokens[len(tokens)-1], "*") {
				typeName += " *"
			}

			args = append(args, XSArg{
				Type: typeName,
				Name: varName,
			})
		} else if len(tokens) == 1 {
			// Только имя без типа - предполагаем SV*
			args = append(args, XSArg{
				Type: "SV *",
				Name: tokens[0],
			})
		}
	}

	return args
}

func (t *Translator) cleanPPCode(code string) string {
	lines := strings.Split(code, "\n")
	var cleaned []string

	inAlias := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Пропускаем пустые строки в начале
		if len(cleaned) == 0 && trimmed == "" {
			continue
		}

		// Пропускаем ALIAS: секции
		if strings.HasPrefix(trimmed, "ALIAS:") {
			inAlias = true
			continue
		}
		if inAlias {
			// ALIAS строки имеют формат "name = VALUE"
			if regexp.MustCompile(`^\w+\s*=\s*\w+$`).MatchString(trimmed) {
				continue
			}
			inAlias = false
		}

		// Пропускаем ATTRS:
		if strings.HasPrefix(trimmed, "ATTRS:") {
			continue
		}

		// Останавливаемся на следующей функции
		if regexp.MustCompile(`^\w+\s+\w+\s*\(`).MatchString(trimmed) {
			break
		}

		cleaned = append(cleaned, trimmed)
	}

	return strings.Join(cleaned, "\n")
}

func (t *Translator) parsePPCodeFunction(match []string) *XSFunction {
	returnType := strings.TrimSpace(match[1])
	funcName := strings.TrimSpace(match[2])
	argsStr := strings.TrimSpace(match[3])
	body := match[4]

	// Пропускаем служебные функции
	if t.isSkippable(funcName, "") {
		return nil
	}

	fn := &XSFunction{
		ReturnType: returnType,
		Name:       funcName,
		OutputVar:  "", // PPCODE не использует RETVAL
	}

	// Парсим аргументы
	fn.Args = t.parseXSArgs(argsStr)

	// Очищаем код
	fn.Code = t.cleanPPCode(body)

	return fn
}

func (t *Translator) parseCodeFunction(match []string) *XSFunction {
	returnType := strings.TrimSpace(match[1])
	funcName := strings.TrimSpace(match[2])
	argsStr := strings.TrimSpace(match[3])
	body := match[4]
	outputVar := strings.TrimSpace(match[5])

	if t.isSkippable(funcName, "") {
		return nil
	}

	fn := &XSFunction{
		ReturnType: returnType,
		Name:       funcName,
		OutputVar:  outputVar,
	}

	fn.Args = t.parseXSArgs(argsStr)
	fn.Code = strings.TrimSpace(body)

	return fn
}

func (t *Translator) extractFunctions(section string) {
	// Ищем XS функции
	// Формат:
	// ReturnType
	// func_name(args)
	//     type arg1
	//     type arg2
	//     CODE:
	//         ...
	//     OUTPUT:
	//         RETVAL

	// Паттерн для XS функции
	// Ключевой признак: CODE: и OUTPUT: секции
	funcPattern := regexp.MustCompile(`(?sm)^(\w+(?:\s*\*)?)\s*$\s*^(\w+)\s*\(([^)]*)\)\s*$(.*?)^[ \t]+OUTPUT:\s*$\s*^[ \t]+(\w+)`)

	matches := funcPattern.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		returnType := strings.TrimSpace(match[1])
		funcName := strings.TrimSpace(match[2])
		argsStr := strings.TrimSpace(match[3])
		body := match[4]
		outputVar := strings.TrimSpace(match[5])

		// Пропускаем если это не похоже на XS функцию
		if t.isSkippable(funcName, returnType) {
			continue
		}

		fn := &XSFunction{
			ReturnType: returnType,
			Name:       funcName,
			OutputVar:  outputVar,
		}

		// Парсим имена аргументов из сигнатуры
		argNames := []string{}
		if argsStr != "" {
			for _, arg := range strings.Split(argsStr, ",") {
				arg = strings.TrimSpace(arg)
				if arg != "" {
					argNames = append(argNames, arg)
				}
			}
		}

		// Разделяем тело на декларации аргументов и CODE
		t.parseBody(fn, body, argNames)

		// Добавляем функцию только если она валидна
		if fn.Name != "" && (len(fn.Args) > 0 || len(argNames) == 0) {
			t.functions[fn.Name] = fn
		}
	}
}

func (t *Translator) isSkippable(funcName, returnType string) bool {
	_ = returnType
	// Пропускаем внутренние функции
	skipNames := map[string]bool{
		"DESTROY":   true,
		"CLONE":     true,
		"AUTOLOAD":  true,
		"BEGIN":     true,
		"END":       true,
		"UNITCHECK": true,
		"CHECK":     true,
		"INIT":      true,
	}

	if skipNames[funcName] {
		return true
	}

	// Пропускаем если имя начинается с _
	if strings.HasPrefix(funcName, "_") {
		return true
	}

	return false
}

func (t *Translator) parseBody(fn *XSFunction, body string, argNames []string) {
	// Разделяем на часть до CODE: и после
	codeParts := regexp.MustCompile(`(?si)(.*?)CODE:\s*(.*)`).FindStringSubmatch(body)

	argDeclarations := ""
	codeBody := ""

	if len(codeParts) >= 3 {
		argDeclarations = codeParts[1]
		codeBody = codeParts[2]
	} else {
		// Нет CODE: секции
		return
	}

	// Парсим декларации аргументов
	fn.Args = t.parseArgDeclarations(argDeclarations, argNames)

	// Если аргументы не найдены, используем имена из сигнатуры с типом по умолчанию
	if len(fn.Args) == 0 && len(argNames) > 0 {
		for _, name := range argNames {
			fn.Args = append(fn.Args, XSArg{
				Type: "SV *",
				Name: name,
			})
		}
	}

	// Очищаем код от лишнего
	fn.Code = t.cleanCode(codeBody)
}

func (t *Translator) parseArgDeclarations(declarations string, argNames []string) []XSArg {
	_ = argNames
	var args []XSArg

	lines := strings.Split(declarations, "\n")
	argIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Пропускаем PREINIT:, INIT:, ALIAS: и т.д.
		if strings.HasSuffix(line, ":") {
			continue
		}

		// Пропускаем если это не декларация типа
		if !t.isTypeDeclaration(line) {
			continue
		}

		// Формат: "SV *name" или "int a" или "char *str"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Последнее слово - имя переменной
			varName := parts[len(parts)-1]
			// Всё остальное - тип
			typeName := strings.Join(parts[:len(parts)-1], " ")

			// Убираем * из имени если прилипло
			varName = strings.TrimPrefix(varName, "*")
			// Добавляем * к типу если было в имени
			if strings.HasPrefix(parts[len(parts)-1], "*") {
				typeName += " *"
			}

			args = append(args, XSArg{
				Type: typeName,
				Name: varName,
			})
			argIndex++
		}
	}

	return args
}

func (t *Translator) isTypeDeclaration(line string) bool {
	// Проверяем, похоже ли на декларацию типа
	validTypes := []string{
		"SV", "AV", "HV", "CV", "GV",
		"int", "long", "short", "char",
		"unsigned", "signed",
		"double", "float",
		"I32", "I16", "I8",
		"U32", "U16", "U8",
		"IV", "UV", "NV",
		"STRLEN", "Size_t", "bool",
	}

	for _, vt := range validTypes {
		if strings.HasPrefix(line, vt+" ") || strings.HasPrefix(line, vt+"*") || strings.HasPrefix(line, vt+"\t") {
			return true
		}
	}

	return false
}

func (t *Translator) cleanCode(code string) string {
	lines := strings.Split(code, "\n")
	var cleaned []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Пропускаем пустые строки
		if trimmed == "" {
			continue
		}

		// Останавливаемся на OUTPUT: (если вдруг попало)
		if strings.HasPrefix(trimmed, "OUTPUT:") {
			break
		}

		// Пропускаем метки
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "=") {
			continue
		}

		cleaned = append(cleaned, trimmed)
	}

	return strings.Join(cleaned, "\n")
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

// Замени метод translateCCode на:
func (t *Translator) translateCCode(cCode string) string {
	c2goTranslator := c2go.New()
	return c2goTranslator.Translate(cCode)
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
