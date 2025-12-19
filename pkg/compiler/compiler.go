// pkg/compiler/compiler.go
package compiler

import (
	"fmt"
	"os"
	"strings"

	"perlc/pkg/ast"
	"perlc/pkg/codegen"
	"perlc/pkg/deps"
	"perlc/pkg/lexer"
	"perlc/pkg/parser"
	"perlc/pkg/xs2go"
)

// Cache для скомпилированных модулей
var moduleCache = make(map[string]string)

// Compile компилирует Perl файл в Go код
func Compile(perlFile string) (string, error) {
	content, err := os.ReadFile(perlFile)
	if err != nil {
		return "", err
	}

	// Используем наш парсер
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse errors: %v", p.Errors())
	}

	// Собираем все use
	modules := collectModules(program)

	// Обрабатываем каждый модуль
	var moduleCode strings.Builder
	for _, mod := range modules {
		code, err := processModule(mod)
		if err != nil {
			// Предупреждаем, но продолжаем
			fmt.Fprintf(os.Stderr, "Warning: could not process module %s: %v\n", mod, err)
			continue
		}
		moduleCode.WriteString(code)
	}

	// Генерируем основной код
	gen := codegen.New()
	mainCode := gen.Generate(program)

	// Собираем всё вместе
	return combineCode(moduleCode.String(), mainCode), nil
}

// collectModules собирает все use декларации из программы
func collectModules(program *ast.Program) []string {
	var modules []string

	for _, stmt := range program.Statements {
		if use, ok := stmt.(*ast.UseDecl); ok {
			modules = append(modules, use.Module)
		}
	}

	return modules
}

// processModule обрабатывает один модуль
func processModule(moduleName string) (string, error) {
	// Проверяем кэш
	if cached, ok := moduleCache[moduleName]; ok {
		return cached, nil
	}

	// Стандартные модули пропускаем
	if isStandardModule(moduleName) {
		return "", nil
	}

	// Проверяем кэш на диске
	if cached, ok := deps.GetCachedModule(moduleName, "latest"); ok {
		moduleCache[moduleName] = cached
		return cached, nil
	}

	// Ищем и анализируем модуль
	info, err := deps.AnalyzeModule(moduleName)
	if err != nil {
		return "", err
	}

	var code strings.Builder

	// Транслируем XS → Go если есть
	if info.HasXS {
		translator := xs2go.New()
		for _, xsFile := range info.XSFiles {
			part, err := translator.Translate(xsFile)
			if err != nil {
				return "", fmt.Errorf("XS translation failed for %s: %w", xsFile, err)
			}
			code.WriteString(part)
			code.WriteString("\n")
		}
	}

	// Транслируем Pure Perl части
	for _, pmFile := range info.PurePerl {
		part, err := compilePerl(pmFile)
		if err != nil {
			return "", fmt.Errorf("Perl compilation failed for %s: %w", pmFile, err)
		}
		code.WriteString(part)
		code.WriteString("\n")
	}

	// Рекурсивно обрабатываем зависимости
	for _, dep := range info.Dependencies {
		depCode, err := processModule(dep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not process dependency %s: %v\n", dep, err)
			continue
		}
		code.WriteString(depCode)
	}

	result := code.String()

	// Сохраняем в кэш
	moduleCache[moduleName] = result
	deps.CacheModule(moduleName, "latest", result)

	return result, nil
}

// compilePerl компилирует .pm файл в Go код
func compilePerl(pmFile string) (string, error) {
	content, err := os.ReadFile(pmFile)
	if err != nil {
		return "", err
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse errors: %v", p.Errors())
	}

	gen := codegen.New()
	return gen.Generate(program), nil
}

// isStandardModule проверяет, является ли модуль стандартным
func isStandardModule(name string) bool {
	standard := map[string]bool{
		"strict":         true,
		"warnings":       true,
		"feature":        true,
		"utf8":           true,
		"vars":           true,
		"constant":       true,
		"Exporter":       true,
		"Carp":           true,
		"Data::Dumper":   true,
		"File::Spec":     true,
		"File::Path":     true,
		"File::Basename": true,
		"Getopt::Long":   true,
		"Pod::Usage":     true,
	}
	return standard[name]
}

// combineCode объединяет код модулей и основной код
func combineCode(moduleCode, mainCode string) string {
	if moduleCode == "" {
		return mainCode
	}

	var result strings.Builder

	// Модули идут первыми (без package main и import)
	result.WriteString("// === Compiled Modules ===\n")
	result.WriteString(moduleCode)
	result.WriteString("\n// === Main Program ===\n")
	result.WriteString(mainCode)

	return result.String()
}
