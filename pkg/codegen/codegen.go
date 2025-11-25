// Package codegen provides code generation for Perl AST to machine code.
package codegen

import (
	"fmt"
	"strings"

	"github.com/djeday123/perl-compiler/pkg/ast"
)

// VarInfo stores information about a variable.
type VarInfo struct {
	Offset   int
	IsString bool
}

// CodeGenerator generates machine code from Perl AST.
type CodeGenerator struct {
	output       strings.Builder
	labelCounter int
	variables    map[string]VarInfo // variable name -> variable info
	stackOffset  int
	dataSection  strings.Builder
	stringLabels map[string]string
	functions    map[string]bool
}

// New creates a new CodeGenerator.
func New() *CodeGenerator {
	return &CodeGenerator{
		variables:    make(map[string]VarInfo),
		stringLabels: make(map[string]string),
		functions:    make(map[string]bool),
	}
}

// Generate generates x86-64 assembly code from the AST.
func (cg *CodeGenerator) Generate(program *ast.Program) string {
	// Generate data section for strings
	cg.dataSection.WriteString(".section .data\n")
	cg.dataSection.WriteString("    newline: .asciz \"\\n\"\n")
	cg.dataSection.WriteString("    int_format: .asciz \"%ld\"\n")
	cg.dataSection.WriteString("    str_format: .asciz \"%s\"\n")

	// Generate text section
	cg.output.WriteString("\n.section .text\n")
	cg.output.WriteString(".globl main\n")

	// First pass: collect function definitions
	for _, stmt := range program.Statements {
		if sub, ok := stmt.(*ast.SubroutineStatement); ok {
			cg.functions[sub.Name] = true
		}
	}

	// Second pass: generate subroutines
	for _, stmt := range program.Statements {
		if sub, ok := stmt.(*ast.SubroutineStatement); ok {
			cg.generateSubroutine(sub)
		}
	}

	// Generate main function
	cg.output.WriteString("\nmain:\n")
	cg.output.WriteString("    pushq %rbp\n")
	cg.output.WriteString("    movq %rsp, %rbp\n")
	cg.output.WriteString("    subq $256, %rsp\n") // Reserve stack space

	// Generate code for non-subroutine statements
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.SubroutineStatement); !ok {
			cg.generateStatement(stmt)
		}
	}

	// Return 0 from main
	cg.output.WriteString("\n    # Exit program\n")
	cg.output.WriteString("    movq $0, %rax\n")
	cg.output.WriteString("    movq %rbp, %rsp\n")
	cg.output.WriteString("    popq %rbp\n")
	cg.output.WriteString("    ret\n")

	// Add note section to mark stack as non-executable
	cg.output.WriteString("\n.section .note.GNU-stack,\"\",@progbits\n")

	return cg.dataSection.String() + cg.output.String()
}

func (cg *CodeGenerator) generateSubroutine(sub *ast.SubroutineStatement) {
	cg.output.WriteString(fmt.Sprintf("\n%s:\n", sub.Name))
	cg.output.WriteString("    pushq %rbp\n")
	cg.output.WriteString("    movq %rsp, %rbp\n")
	cg.output.WriteString("    subq $128, %rsp\n")

	// Reset variable tracking for this function
	oldVars := cg.variables
	oldOffset := cg.stackOffset
	cg.variables = make(map[string]VarInfo)
	cg.stackOffset = 0

	for _, stmt := range sub.Body.Statements {
		cg.generateStatement(stmt)
	}

	// Default return 0 if no explicit return
	cg.output.WriteString("    movq $0, %rax\n")
	cg.output.WriteString("    movq %rbp, %rsp\n")
	cg.output.WriteString("    popq %rbp\n")
	cg.output.WriteString("    ret\n")

	// Restore parent scope
	cg.variables = oldVars
	cg.stackOffset = oldOffset
}

func (cg *CodeGenerator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.MyStatement:
		cg.generateMyStatement(s)
	case *ast.AssignmentStatement:
		cg.generateAssignmentStatement(s)
	case *ast.ExpressionStatement:
		cg.generateExpressionStatement(s)
	case *ast.PrintStatement:
		cg.generatePrintStatement(s)
	case *ast.ReturnStatement:
		cg.generateReturnStatement(s)
	case *ast.IfStatement:
		cg.generateIfStatement(s)
	case *ast.WhileStatement:
		cg.generateWhileStatement(s)
	case *ast.ForStatement:
		cg.generateForStatement(s)
	case *ast.BlockStatement:
		for _, innerStmt := range s.Statements {
			cg.generateStatement(innerStmt)
		}
	}
}

func (cg *CodeGenerator) generateMyStatement(stmt *ast.MyStatement) {
	// Allocate space on stack for the variable
	cg.stackOffset += 8
	varName := stmt.Name.Sigil + stmt.Name.Value

	// Determine if value is a string
	isString := false
	if stmt.Value != nil {
		_, isString = stmt.Value.(*ast.StringLiteral)
	}

	cg.variables[varName] = VarInfo{Offset: cg.stackOffset, IsString: isString}

	cg.output.WriteString(fmt.Sprintf("    # my %s\n", varName))

	if stmt.Value != nil {
		cg.generateExpression(stmt.Value)
		cg.output.WriteString(fmt.Sprintf("    movq %%rax, -%d(%%rbp)\n", cg.stackOffset))
	} else {
		// Initialize to 0
		cg.output.WriteString(fmt.Sprintf("    movq $0, -%d(%%rbp)\n", cg.stackOffset))
	}
}

func (cg *CodeGenerator) generateAssignmentStatement(stmt *ast.AssignmentStatement) {
	varName := stmt.Name.Sigil + stmt.Name.Value
	varInfo, exists := cg.variables[varName]

	// Determine if value is a string
	isString := false
	_, isString = stmt.Value.(*ast.StringLiteral)

	if !exists {
		// Auto-vivify the variable
		cg.stackOffset += 8
		varInfo = VarInfo{Offset: cg.stackOffset, IsString: isString}
		cg.variables[varName] = varInfo
	} else {
		varInfo.IsString = isString
		cg.variables[varName] = varInfo
	}

	cg.output.WriteString(fmt.Sprintf("    # %s = ...\n", varName))
	cg.generateExpression(stmt.Value)
	cg.output.WriteString(fmt.Sprintf("    movq %%rax, -%d(%%rbp)\n", varInfo.Offset))
}

func (cg *CodeGenerator) generateExpressionStatement(stmt *ast.ExpressionStatement) {
	if stmt.Expression != nil {
		cg.generateExpression(stmt.Expression)
	}
}

func (cg *CodeGenerator) generatePrintStatement(stmt *ast.PrintStatement) {
	cg.output.WriteString("    # print statement\n")

	for _, arg := range stmt.Arguments {
		cg.generateExpression(arg)

		// Determine if this is a string based on expression type or variable info
		isString := false
		switch a := arg.(type) {
		case *ast.StringLiteral:
			isString = true
		case *ast.Identifier:
			varName := a.Sigil + a.Value
			if info, ok := cg.variables[varName]; ok {
				isString = info.IsString
			}
		}

		if isString {
			cg.output.WriteString("    movq %rax, %rsi\n")
			cg.output.WriteString("    leaq str_format(%rip), %rdi\n")
		} else {
			cg.output.WriteString("    movq %rax, %rsi\n")
			cg.output.WriteString("    leaq int_format(%rip), %rdi\n")
		}

		cg.output.WriteString("    xorq %rax, %rax\n")
		cg.output.WriteString("    call printf@PLT\n")
	}

	// Print newline
	cg.output.WriteString("    leaq newline(%rip), %rdi\n")
	cg.output.WriteString("    call puts@PLT\n")
}

func (cg *CodeGenerator) generateReturnStatement(stmt *ast.ReturnStatement) {
	cg.output.WriteString("    # return\n")
	if stmt.ReturnValue != nil {
		cg.generateExpression(stmt.ReturnValue)
	} else {
		cg.output.WriteString("    movq $0, %rax\n")
	}
	cg.output.WriteString("    movq %rbp, %rsp\n")
	cg.output.WriteString("    popq %rbp\n")
	cg.output.WriteString("    ret\n")
}

func (cg *CodeGenerator) generateIfStatement(stmt *ast.IfStatement) {
	elseLabel := cg.newLabel()
	endLabel := cg.newLabel()

	cg.output.WriteString("    # if statement\n")
	cg.generateExpression(stmt.Condition)
	cg.output.WriteString("    cmpq $0, %rax\n")
	cg.output.WriteString(fmt.Sprintf("    je %s\n", elseLabel))

	// Generate consequence
	for _, s := range stmt.Consequence.Statements {
		cg.generateStatement(s)
	}
	cg.output.WriteString(fmt.Sprintf("    jmp %s\n", endLabel))

	// Generate else
	cg.output.WriteString(fmt.Sprintf("%s:\n", elseLabel))
	if stmt.Alternative != nil {
		cg.generateStatement(stmt.Alternative)
	}

	cg.output.WriteString(fmt.Sprintf("%s:\n", endLabel))
}

func (cg *CodeGenerator) generateWhileStatement(stmt *ast.WhileStatement) {
	startLabel := cg.newLabel()
	endLabel := cg.newLabel()

	cg.output.WriteString("    # while loop\n")
	cg.output.WriteString(fmt.Sprintf("%s:\n", startLabel))
	cg.generateExpression(stmt.Condition)
	cg.output.WriteString("    cmpq $0, %rax\n")
	cg.output.WriteString(fmt.Sprintf("    je %s\n", endLabel))

	for _, s := range stmt.Body.Statements {
		cg.generateStatement(s)
	}
	cg.output.WriteString(fmt.Sprintf("    jmp %s\n", startLabel))

	cg.output.WriteString(fmt.Sprintf("%s:\n", endLabel))
}

func (cg *CodeGenerator) generateForStatement(stmt *ast.ForStatement) {
	// For simplicity, treat as a counted loop with a range
	startLabel := cg.newLabel()
	endLabel := cg.newLabel()

	// Allocate loop variable
	varName := stmt.Variable.Sigil + stmt.Variable.Value
	cg.stackOffset += 8
	cg.variables[varName] = VarInfo{Offset: cg.stackOffset, IsString: false}
	loopVarOffset := cg.stackOffset

	cg.output.WriteString("    # for loop\n")

	// Initialize loop variable (assume it's a range, start from 0 or first value)
	if rangeExpr, ok := stmt.List.(*ast.RangeExpression); ok {
		cg.generateExpression(rangeExpr.Start)
	} else {
		cg.output.WriteString("    movq $0, %rax\n")
	}
	cg.output.WriteString(fmt.Sprintf("    movq %%rax, -%d(%%rbp)\n", loopVarOffset))

	cg.output.WriteString(fmt.Sprintf("%s:\n", startLabel))

	// Check loop condition
	if rangeExpr, ok := stmt.List.(*ast.RangeExpression); ok {
		cg.output.WriteString(fmt.Sprintf("    movq -%d(%%rbp), %%rax\n", loopVarOffset))
		cg.output.WriteString("    pushq %rax\n")
		cg.generateExpression(rangeExpr.End)
		cg.output.WriteString("    movq %rax, %rbx\n")
		cg.output.WriteString("    popq %rax\n")
		cg.output.WriteString("    cmpq %rbx, %rax\n")
		cg.output.WriteString(fmt.Sprintf("    jg %s\n", endLabel))
	}

	// Generate loop body
	for _, s := range stmt.Body.Statements {
		cg.generateStatement(s)
	}

	// Increment loop variable
	cg.output.WriteString(fmt.Sprintf("    incq -%d(%%rbp)\n", loopVarOffset))
	cg.output.WriteString(fmt.Sprintf("    jmp %s\n", startLabel))

	cg.output.WriteString(fmt.Sprintf("%s:\n", endLabel))
}

func (cg *CodeGenerator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		cg.output.WriteString(fmt.Sprintf("    movq $%d, %%rax\n", e.Value))

	case *ast.FloatLiteral:
		// For simplicity, treat as integer for now
		cg.output.WriteString(fmt.Sprintf("    movq $%d, %%rax\n", int64(e.Value)))

	case *ast.StringLiteral:
		label := cg.addStringConstant(e.Value)
		cg.output.WriteString(fmt.Sprintf("    leaq %s(%%rip), %%rax\n", label))

	case *ast.Identifier:
		varName := e.Sigil + e.Value
		if info, ok := cg.variables[varName]; ok {
			cg.output.WriteString(fmt.Sprintf("    movq -%d(%%rbp), %%rax\n", info.Offset))
		} else {
			// Undefined variable, return 0
			cg.output.WriteString("    movq $0, %rax\n")
		}

	case *ast.PrefixExpression:
		cg.generateExpression(e.Right)
		switch e.Operator {
		case "-":
			cg.output.WriteString("    negq %rax\n")
		case "!":
			cg.output.WriteString("    cmpq $0, %rax\n")
			cg.output.WriteString("    sete %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		}

	case *ast.InfixExpression:
		cg.generateExpression(e.Left)
		cg.output.WriteString("    pushq %rax\n")
		cg.generateExpression(e.Right)
		cg.output.WriteString("    movq %rax, %rbx\n")
		cg.output.WriteString("    popq %rax\n")

		switch e.Operator {
		case "+":
			cg.output.WriteString("    addq %rbx, %rax\n")
		case "-":
			cg.output.WriteString("    subq %rbx, %rax\n")
		case "*":
			cg.output.WriteString("    imulq %rbx, %rax\n")
		case "/":
			cg.output.WriteString("    cqto\n")
			cg.output.WriteString("    idivq %rbx\n")
		case "%":
			cg.output.WriteString("    cqto\n")
			cg.output.WriteString("    idivq %rbx\n")
			cg.output.WriteString("    movq %rdx, %rax\n")
		case "==":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    sete %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case "!=":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    setne %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case "<":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    setl %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case ">":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    setg %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case "<=":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    setle %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case ">=":
			cg.output.WriteString("    cmpq %rbx, %rax\n")
			cg.output.WriteString("    setge %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case "&&":
			cg.output.WriteString("    testq %rax, %rax\n")
			cg.output.WriteString("    setne %al\n")
			cg.output.WriteString("    testq %rbx, %rbx\n")
			cg.output.WriteString("    setne %bl\n")
			cg.output.WriteString("    andb %bl, %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case "||":
			cg.output.WriteString("    orq %rbx, %rax\n")
			cg.output.WriteString("    testq %rax, %rax\n")
			cg.output.WriteString("    setne %al\n")
			cg.output.WriteString("    movzbq %al, %rax\n")
		case ".":
			// String concatenation - not fully implemented, return left operand
			cg.output.WriteString("    # String concatenation not fully implemented\n")
		}

	case *ast.CallExpression:
		// Push arguments in reverse order
		for i := len(e.Arguments) - 1; i >= 0; i-- {
			cg.generateExpression(e.Arguments[i])
			cg.output.WriteString("    pushq %rax\n")
		}

		// Move first 6 arguments to registers (x86-64 calling convention)
		regs := []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}
		for i := 0; i < len(e.Arguments) && i < 6; i++ {
			cg.output.WriteString(fmt.Sprintf("    popq %s\n", regs[i]))
		}

		cg.output.WriteString(fmt.Sprintf("    call %s\n", e.Function))
	}
}

func (cg *CodeGenerator) newLabel() string {
	label := fmt.Sprintf(".L%d", cg.labelCounter)
	cg.labelCounter++
	return label
}

func (cg *CodeGenerator) addStringConstant(s string) string {
	if label, ok := cg.stringLabels[s]; ok {
		return label
	}

	label := fmt.Sprintf("str_%d", len(cg.stringLabels))
	cg.stringLabels[s] = label

	// Escape the string for assembly
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")

	cg.dataSection.WriteString(fmt.Sprintf("    %s: .asciz \"%s\"\n", label, escaped))

	return label
}
