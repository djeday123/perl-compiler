// Package errors provides error types and handling for the Perl compiler.
package errors

import (
	"fmt"
	"strings"
)

// CompileError represents a compilation error.
type CompileError struct {
	Line    int
	Column  int
	Message string
	Phase   string // "lexer", "parser", "codegen"
}

func (e *CompileError) Error() string {
	return fmt.Sprintf("%s error at line %d, column %d: %s", e.Phase, e.Line, e.Column, e.Message)
}

// ErrorList is a collection of compile errors.
type ErrorList struct {
	Errors []*CompileError
}

// Add adds a new error to the list.
func (el *ErrorList) Add(line, column int, message, phase string) {
	el.Errors = append(el.Errors, &CompileError{
		Line:    line,
		Column:  column,
		Message: message,
		Phase:   phase,
	})
}

// HasErrors returns true if there are any errors.
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// Error returns a string representation of all errors.
func (el *ErrorList) Error() string {
	if len(el.Errors) == 0 {
		return ""
	}

	var msgs []string
	for _, e := range el.Errors {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// New creates a new error.
func New(line, column int, message, phase string) *CompileError {
	return &CompileError{
		Line:    line,
		Column:  column,
		Message: message,
		Phase:   phase,
	}
}

// NewList creates a new error list.
func NewList() *ErrorList {
	return &ErrorList{Errors: []*CompileError{}}
}
