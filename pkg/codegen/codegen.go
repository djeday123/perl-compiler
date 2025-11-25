// Package codegen implements code generation for Perl AST to x86-64 machine code.
package codegen

import (
	"bytes"
	"fmt"
	"math"

	"github.com/djeday123/perl-compiler/pkg/ast"
)

// Register represents an x86-64 register.
type Register int

// x86-64 registers
const (
	RAX Register = iota
	RBX
	RCX
	RDX
	RSI
	RDI
	RBP
	RSP
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

// VariableInfo holds information about a variable in the symbol table.
type VariableInfo struct {
	Offset int    // Stack offset from RBP
	Size   int    // Size in bytes
	Type   string // "scalar", "array", "hash"
}

// Generator generates x86-64 machine code from a Perl AST.
type Generator struct {
	code         *bytes.Buffer
	data         *bytes.Buffer
	bss          *bytes.Buffer
	labelCounter int
	variables    map[string]*VariableInfo
	stackOffset  int
	stringTable  map[string]string // string literal -> label
	errors       []string
}

// New creates a new code generator.
func New() *Generator {
	return &Generator{
		code:         new(bytes.Buffer),
		data:         new(bytes.Buffer),
		bss:          new(bytes.Buffer),
		labelCounter: 0,
		variables:    make(map[string]*VariableInfo),
		stackOffset:  0,
		stringTable:  make(map[string]string),
		errors:       []string{},
	}
}

// Errors returns the list of code generation errors.
func (g *Generator) Errors() []string {
	return g.errors
}

// newLabel generates a unique label.
func (g *Generator) newLabel(prefix string) string {
	label := fmt.Sprintf(".%s%d", prefix, g.labelCounter)
	g.labelCounter++
	return label
}

// emit writes assembly code to the code buffer.
func (g *Generator) emit(format string, args ...interface{}) {
	fmt.Fprintf(g.code, format+"\n", args...)
}

// emitData writes to the data section.
func (g *Generator) emitData(format string, args ...interface{}) {
	fmt.Fprintf(g.data, format+"\n", args...)
}

// emitBss writes to the BSS section.
func (g *Generator) emitBss(format string, args ...interface{}) {
	fmt.Fprintf(g.bss, format+"\n", args...)
}

// addStringLiteral adds a string to the data section and returns its label.
func (g *Generator) addStringLiteral(s string) string {
	if label, exists := g.stringTable[s]; exists {
		return label
	}

	label := g.newLabel("str")
	g.stringTable[s] = label
	return label
}

// allocateVariable allocates stack space for a variable.
func (g *Generator) allocateVariable(name string, size int, varType string) *VariableInfo {
	g.stackOffset += size
	info := &VariableInfo{
		Offset: -g.stackOffset,
		Size:   size,
		Type:   varType,
	}
	g.variables[name] = info
	return info
}

// Generate generates x86-64 assembly code from the AST.
func (g *Generator) Generate(program *ast.Program) string {
	var output bytes.Buffer

	// Generate code for all statements
	for _, stmt := range program.Statements {
		g.generateStatement(stmt)
	}

	// Write data section
	output.WriteString("section .data\n")
	// Write format strings for print
	output.WriteString("    format_int: db \"%ld\", 0\n")
	output.WriteString("    format_str: db \"%s\", 0\n")
	output.WriteString("    format_float: db \"%f\", 0\n")
	output.WriteString("    newline: db 10, 0\n")

	// Write string literals
	for str, label := range g.stringTable {
		output.WriteString(fmt.Sprintf("    %s: db \"%s\", 0\n", label, escapeString(str)))
	}

	// Write BSS section
	if g.bss.Len() > 0 {
		output.WriteString("\nsection .bss\n")
		output.WriteString(g.bss.String())
	}

	// Write text section
	output.WriteString("\nsection .text\n")
	output.WriteString("    global _start\n")
	output.WriteString("    extern printf\n")
	output.WriteString("    extern exit\n\n")

	// Entry point
	output.WriteString("_start:\n")
	output.WriteString("    push rbp\n")
	output.WriteString("    mov rbp, rsp\n")

	// Allocate stack space for local variables
	if g.stackOffset > 0 {
		alignedOffset := (g.stackOffset + 15) & ^15 // 16-byte alignment
		output.WriteString(fmt.Sprintf("    sub rsp, %d\n", alignedOffset))
	}

	// Write generated code
	output.WriteString(g.code.String())

	// Clean exit
	output.WriteString("\n    ; Exit program\n")
	output.WriteString("    mov rsp, rbp\n")
	output.WriteString("    pop rbp\n")
	output.WriteString("    mov rax, 60\n")
	output.WriteString("    xor rdi, rdi\n")
	output.WriteString("    syscall\n")

	return output.String()
}

// generateStatement generates code for a statement.
func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.MyStatement:
		g.generateMyStatement(s)
	case *ast.AssignmentStatement:
		g.generateAssignmentStatement(s)
	case *ast.ExpressionStatement:
		g.generateExpressionStatement(s)
	case *ast.PrintStatement:
		g.generatePrintStatement(s)
	case *ast.IfStatement:
		g.generateIfStatement(s)
	case *ast.WhileStatement:
		g.generateWhileStatement(s)
	case *ast.ForStatement:
		g.generateForStatement(s)
	case *ast.SubroutineStatement:
		g.generateSubroutineStatement(s)
	case *ast.ReturnStatement:
		g.generateReturnStatement(s)
	case *ast.BlockStatement:
		g.generateBlockStatement(s)
	}
}

// generateMyStatement generates code for a my declaration.
func (g *Generator) generateMyStatement(stmt *ast.MyStatement) {
	var varName string
	var varType string

	switch v := stmt.Name.(type) {
	case *ast.ScalarVariable:
		varName = v.Name
		varType = "scalar"
	case *ast.ArrayVariable:
		varName = v.Name
		varType = "array"
	case *ast.HashVariable:
		varName = v.Name
		varType = "hash"
	default:
		g.errors = append(g.errors, "invalid variable in my statement")
		return
	}

	// Allocate 8 bytes for the variable (64-bit)
	info := g.allocateVariable(varName, 8, varType)

	// If there's an initial value, generate code to compute and store it
	if stmt.Value != nil {
		g.generateExpression(stmt.Value)
		g.emit("    mov [rbp%+d], rax    ; store %s", info.Offset, varName)
	} else {
		// Initialize to 0
		g.emit("    mov qword [rbp%+d], 0    ; initialize %s", info.Offset, varName)
	}
}

// generateAssignmentStatement generates code for an assignment.
func (g *Generator) generateAssignmentStatement(stmt *ast.AssignmentStatement) {
	// Generate the value first
	g.generateExpression(stmt.Value)

	// Store to the variable
	switch v := stmt.Name.(type) {
	case *ast.ScalarVariable:
		if info, exists := g.variables[v.Name]; exists {
			g.emit("    mov [rbp%+d], rax    ; store to %s", info.Offset, v.Name)
		} else {
			// Variable not declared, allocate it
			info := g.allocateVariable(v.Name, 8, "scalar")
			g.emit("    mov [rbp%+d], rax    ; store to %s", info.Offset, v.Name)
		}
	}
}

// generateExpressionStatement generates code for an expression statement.
func (g *Generator) generateExpressionStatement(stmt *ast.ExpressionStatement) {
	if stmt.Expression != nil {
		// Check if it's an assignment expression
		if infix, ok := stmt.Expression.(*ast.InfixExpression); ok && infix.Operator == "=" {
			// Generate the right side
			g.generateExpression(infix.Right)

			// Store to the left side
			switch v := infix.Left.(type) {
			case *ast.ScalarVariable:
				if info, exists := g.variables[v.Name]; exists {
					g.emit("    mov [rbp%+d], rax    ; store to %s", info.Offset, v.Name)
				} else {
					info := g.allocateVariable(v.Name, 8, "scalar")
					g.emit("    mov [rbp%+d], rax    ; store to %s", info.Offset, v.Name)
				}
			}
		} else {
			g.generateExpression(stmt.Expression)
		}
	}
}

// generatePrintStatement generates code for a print statement.
func (g *Generator) generatePrintStatement(stmt *ast.PrintStatement) {
	for _, arg := range stmt.Arguments {
		g.generateExpression(arg)

		// Determine the type and call appropriate print
		switch arg.(type) {
		case *ast.StringLiteral:
			g.emit("    ; print string")
			g.emit("    mov rdi, rax")
			g.emit("    xor rax, rax")
			g.emit("    call puts wrt ..plt")
		case *ast.IntegerLiteral, *ast.ScalarVariable:
			g.emit("    ; print integer")
			g.emit("    mov rsi, rax")
			g.emit("    lea rdi, [rel format_int]")
			g.emit("    xor rax, rax")
			g.emit("    call printf wrt ..plt")
		case *ast.FloatLiteral:
			g.emit("    ; print float")
			g.emit("    movq xmm0, rax")
			g.emit("    lea rdi, [rel format_float]")
			g.emit("    mov rax, 1")
			g.emit("    call printf wrt ..plt")
		default:
			// Generic print
			g.emit("    ; print value")
			g.emit("    mov rsi, rax")
			g.emit("    lea rdi, [rel format_int]")
			g.emit("    xor rax, rax")
			g.emit("    call printf wrt ..plt")
		}
	}
}

// generateIfStatement generates code for an if statement.
func (g *Generator) generateIfStatement(stmt *ast.IfStatement) {
	elseLabel := g.newLabel("else")
	endLabel := g.newLabel("endif")

	// Evaluate condition
	g.generateExpression(stmt.Condition)
	g.emit("    test rax, rax")
	g.emit("    jz %s", elseLabel)

	// True branch
	g.generateBlockStatement(stmt.Consequence)
	g.emit("    jmp %s", endLabel)

	// False branch
	g.emit("%s:", elseLabel)
	if stmt.Alternative != nil {
		g.generateStatement(stmt.Alternative)
	}

	g.emit("%s:", endLabel)
}

// generateWhileStatement generates code for a while loop.
func (g *Generator) generateWhileStatement(stmt *ast.WhileStatement) {
	startLabel := g.newLabel("while_start")
	endLabel := g.newLabel("while_end")

	g.emit("%s:", startLabel)

	// Evaluate condition
	g.generateExpression(stmt.Condition)
	g.emit("    test rax, rax")
	g.emit("    jz %s", endLabel)

	// Loop body
	g.generateBlockStatement(stmt.Body)
	g.emit("    jmp %s", startLabel)

	g.emit("%s:", endLabel)
}

// generateForStatement generates code for a for/foreach loop.
func (g *Generator) generateForStatement(stmt *ast.ForStatement) {
	// This is a simplified implementation for range-based for loops
	startLabel := g.newLabel("for_start")
	endLabel := g.newLabel("for_end")

	// Handle range expression
	if rangeExpr, ok := stmt.List.(*ast.RangeExpression); ok {
		// Allocate iterator variable
		var iterVar string
		if sv, ok := stmt.Variable.(*ast.ScalarVariable); ok {
			iterVar = sv.Name
		} else {
			iterVar = "$_"
		}

		iterInfo := g.allocateVariable(iterVar, 8, "scalar")
		endInfo := g.allocateVariable(iterVar+"_end", 8, "scalar")

		// Initialize iterator to start value
		g.generateExpression(rangeExpr.Start)
		g.emit("    mov [rbp%+d], rax    ; initialize iterator", iterInfo.Offset)

		// Store end value
		g.generateExpression(rangeExpr.End)
		g.emit("    mov [rbp%+d], rax    ; store end value", endInfo.Offset)

		// Loop start
		g.emit("%s:", startLabel)

		// Check if iterator > end
		g.emit("    mov rax, [rbp%+d]", iterInfo.Offset)
		g.emit("    cmp rax, [rbp%+d]", endInfo.Offset)
		g.emit("    jg %s", endLabel)

		// Loop body
		g.generateBlockStatement(stmt.Body)

		// Increment iterator
		g.emit("    inc qword [rbp%+d]", iterInfo.Offset)
		g.emit("    jmp %s", startLabel)

		g.emit("%s:", endLabel)
	}
}

// generateSubroutineStatement generates code for a subroutine definition.
func (g *Generator) generateSubroutineStatement(stmt *ast.SubroutineStatement) {
	g.emit("\n%s:", stmt.Name.Value)
	g.emit("    push rbp")
	g.emit("    mov rbp, rsp")

	// Generate body
	g.generateBlockStatement(stmt.Body)

	g.emit("    mov rsp, rbp")
	g.emit("    pop rbp")
	g.emit("    ret")
}

// generateReturnStatement generates code for a return statement.
func (g *Generator) generateReturnStatement(stmt *ast.ReturnStatement) {
	if stmt.ReturnValue != nil {
		g.generateExpression(stmt.ReturnValue)
	}
	g.emit("    mov rsp, rbp")
	g.emit("    pop rbp")
	g.emit("    ret")
}

// generateBlockStatement generates code for a block of statements.
func (g *Generator) generateBlockStatement(block *ast.BlockStatement) {
	for _, stmt := range block.Statements {
		g.generateStatement(stmt)
	}
}

// generateExpression generates code for an expression, result in RAX.
func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		g.emit("    mov rax, %d", e.Value)

	case *ast.FloatLiteral:
		// Store float as bits in RAX for now
		bits := floatToBits(e.Value)
		g.emit("    mov rax, %d    ; float %f", bits, e.Value)

	case *ast.StringLiteral:
		label := g.addStringLiteral(e.Value)
		g.emit("    lea rax, [rel %s]", label)

	case *ast.ScalarVariable:
		if info, exists := g.variables[e.Name]; exists {
			g.emit("    mov rax, [rbp%+d]    ; load %s", info.Offset, e.Name)
		} else {
			g.emit("    xor rax, rax    ; undefined variable %s", e.Name)
		}

	case *ast.Identifier:
		// Could be a subroutine call without parentheses or a bareword
		g.emit("    ; identifier: %s", e.Value)

	case *ast.PrefixExpression:
		g.generatePrefixExpression(e)

	case *ast.InfixExpression:
		g.generateInfixExpression(e)

	case *ast.PostfixExpression:
		g.generatePostfixExpression(e)

	case *ast.CallExpression:
		g.generateCallExpression(e)

	case *ast.TernaryExpression:
		g.generateTernaryExpression(e)

	case *ast.RangeExpression:
		// For now, just evaluate start (used in for loops separately)
		g.generateExpression(e.Start)
	}
}

// generatePrefixExpression generates code for a prefix expression.
func (g *Generator) generatePrefixExpression(expr *ast.PrefixExpression) {
	g.generateExpression(expr.Right)

	switch expr.Operator {
	case "-":
		g.emit("    neg rax")
	case "!":
		g.emit("    test rax, rax")
		g.emit("    setz al")
		g.emit("    movzx rax, al")
	case "~":
		g.emit("    not rax")
	case "++":
		// Pre-increment
		if sv, ok := expr.Right.(*ast.ScalarVariable); ok {
			if info, exists := g.variables[sv.Name]; exists {
				g.emit("    inc qword [rbp%+d]", info.Offset)
				g.emit("    mov rax, [rbp%+d]", info.Offset)
			}
		}
	case "--":
		// Pre-decrement
		if sv, ok := expr.Right.(*ast.ScalarVariable); ok {
			if info, exists := g.variables[sv.Name]; exists {
				g.emit("    dec qword [rbp%+d]", info.Offset)
				g.emit("    mov rax, [rbp%+d]", info.Offset)
			}
		}
	}
}

// generateInfixExpression generates code for an infix expression.
func (g *Generator) generateInfixExpression(expr *ast.InfixExpression) {
	// Generate left operand
	g.generateExpression(expr.Left)
	g.emit("    push rax")

	// Generate right operand
	g.generateExpression(expr.Right)
	g.emit("    mov rcx, rax")
	g.emit("    pop rax")

	switch expr.Operator {
	case "+":
		g.emit("    add rax, rcx")
	case "-":
		g.emit("    sub rax, rcx")
	case "*":
		g.emit("    imul rax, rcx")
	case "/":
		g.emit("    cqo")
		g.emit("    idiv rcx")
	case "%":
		g.emit("    cqo")
		g.emit("    idiv rcx")
		g.emit("    mov rax, rdx")
	case "**":
		// Power operation - simplified for integer powers
		g.emit("    ; power operation")
		g.emit("    push rbx")
		g.emit("    mov rbx, rax    ; base")
		g.emit("    mov rax, 1      ; result")
		powerLoop := g.newLabel("pow")
		powerEnd := g.newLabel("pow_end")
		g.emit("%s:", powerLoop)
		g.emit("    test rcx, rcx")
		g.emit("    jz %s", powerEnd)
		g.emit("    imul rax, rbx")
		g.emit("    dec rcx")
		g.emit("    jmp %s", powerLoop)
		g.emit("%s:", powerEnd)
		g.emit("    pop rbx")
	case ".":
		// String concatenation - simplified
		g.emit("    ; string concatenation (not fully implemented)")
	case "==":
		g.emit("    cmp rax, rcx")
		g.emit("    sete al")
		g.emit("    movzx rax, al")
	case "!=":
		g.emit("    cmp rax, rcx")
		g.emit("    setne al")
		g.emit("    movzx rax, al")
	case "<":
		g.emit("    cmp rax, rcx")
		g.emit("    setl al")
		g.emit("    movzx rax, al")
	case ">":
		g.emit("    cmp rax, rcx")
		g.emit("    setg al")
		g.emit("    movzx rax, al")
	case "<=":
		g.emit("    cmp rax, rcx")
		g.emit("    setle al")
		g.emit("    movzx rax, al")
	case ">=":
		g.emit("    cmp rax, rcx")
		g.emit("    setge al")
		g.emit("    movzx rax, al")
	case "<=>":
		// Spaceship operator
		cmpLess := g.newLabel("cmp_less")
		cmpEnd := g.newLabel("cmp_end")
		g.emit("    cmp rax, rcx")
		g.emit("    jl %s", cmpLess)
		g.emit("    setg al")
		g.emit("    movzx rax, al")
		g.emit("    jmp %s", cmpEnd)
		g.emit("%s:", cmpLess)
		g.emit("    mov rax, -1")
		g.emit("%s:", cmpEnd)
	case "&&", "and":
		g.emit("    test rax, rax")
		g.emit("    jz .skip_and_%d", g.labelCounter)
		g.emit("    mov rax, rcx")
		g.emit(".skip_and_%d:", g.labelCounter)
		g.labelCounter++
	case "||", "or":
		g.emit("    test rax, rax")
		g.emit("    jnz .skip_or_%d", g.labelCounter)
		g.emit("    mov rax, rcx")
		g.emit(".skip_or_%d:", g.labelCounter)
		g.labelCounter++
	case "&":
		g.emit("    and rax, rcx")
	case "|":
		g.emit("    or rax, rcx")
	case "^":
		g.emit("    xor rax, rcx")
	case "<<":
		g.emit("    shl rax, cl")
	case ">>":
		g.emit("    sar rax, cl")
	}
}

// generatePostfixExpression generates code for a postfix expression.
func (g *Generator) generatePostfixExpression(expr *ast.PostfixExpression) {
	// Load current value first
	g.generateExpression(expr.Left)

	switch expr.Operator {
	case "++":
		if sv, ok := expr.Left.(*ast.ScalarVariable); ok {
			if info, exists := g.variables[sv.Name]; exists {
				g.emit("    ; post-increment %s", sv.Name)
				g.emit("    inc qword [rbp%+d]", info.Offset)
			}
		}
	case "--":
		if sv, ok := expr.Left.(*ast.ScalarVariable); ok {
			if info, exists := g.variables[sv.Name]; exists {
				g.emit("    ; post-decrement %s", sv.Name)
				g.emit("    dec qword [rbp%+d]", info.Offset)
			}
		}
	}
}

// generateCallExpression generates code for a function call.
func (g *Generator) generateCallExpression(expr *ast.CallExpression) {
	// Push arguments in reverse order (x86-64 calling convention)
	for i := len(expr.Arguments) - 1; i >= 0; i-- {
		g.generateExpression(expr.Arguments[i])
		g.emit("    push rax")
	}

	// Get function name
	var funcName string
	if ident, ok := expr.Function.(*ast.Identifier); ok {
		funcName = ident.Value
	}

	// Handle built-in functions
	switch funcName {
	case "defined":
		if len(expr.Arguments) > 0 {
			g.emit("    pop rax")
			g.emit("    test rax, rax")
			g.emit("    setne al")
			g.emit("    movzx rax, al")
		}
	default:
		// Load arguments into registers (System V AMD64 ABI)
		regs := []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
		for i := 0; i < len(expr.Arguments) && i < len(regs); i++ {
			g.emit("    pop %s", regs[i])
		}

		// Call the function
		g.emit("    xor rax, rax")
		g.emit("    call %s", funcName)
	}
}

// generateTernaryExpression generates code for a ternary expression.
func (g *Generator) generateTernaryExpression(expr *ast.TernaryExpression) {
	falseLabel := g.newLabel("ternary_false")
	endLabel := g.newLabel("ternary_end")

	// Evaluate condition
	g.generateExpression(expr.Condition)
	g.emit("    test rax, rax")
	g.emit("    jz %s", falseLabel)

	// True branch
	g.generateExpression(expr.Consequence)
	g.emit("    jmp %s", endLabel)

	// False branch
	g.emit("%s:", falseLabel)
	g.generateExpression(expr.Alternative)

	g.emit("%s:", endLabel)
}

// Helper functions

// floatToBits converts a float64 to its IEEE 754 bit representation.
func floatToBits(f float64) uint64 {
	return math.Float64bits(f)
}

// escapeString escapes special characters in a string for assembly.
func escapeString(s string) string {
	var result bytes.Buffer
	for _, c := range s {
		switch c {
		case '\n':
			result.WriteString(`", 10, "`)
		case '\r':
			result.WriteString(`", 13, "`)
		case '\t':
			result.WriteString(`", 9, "`)
		case '"':
			result.WriteString(`", 34, "`)
		case '\\':
			result.WriteString(`", 92, "`)
		default:
			result.WriteRune(c)
		}
	}
	return result.String()
}
