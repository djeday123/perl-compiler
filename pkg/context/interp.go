// Interpreter context for eval package
package context

import (
	"perlc/pkg/ast"
	"perlc/pkg/sv"
)

// Context holds interpreter state for a single execution.
type Context struct {
	runtime *Runtime

	// Variable scopes (lexical)
	scopes []map[string]*sv.SV

	// Subroutines
	subs map[string]*ast.BlockStmt

	// Arguments @_
	args *sv.SV

	// Control flow
	returnValue *sv.SV
	hasReturn   bool
	lastLabel   string
	hasLast     bool
	nextLabel   string
	hasNext     bool
}

// New creates a new interpreter context.
func New() *Context {
	return &Context{
		runtime: GetRuntime(),
		scopes:  []map[string]*sv.SV{make(map[string]*sv.SV)},
		subs:    make(map[string]*ast.BlockStmt),
	}
}

// ============================================================
// Variable Management
// ============================================================

// DeclareVar declares a variable in current scope.
func (c *Context) DeclareVar(name string, value *sv.SV, kind string) {
	if len(c.scopes) == 0 {
		c.scopes = append(c.scopes, make(map[string]*sv.SV))
	}
	c.scopes[len(c.scopes)-1][name] = value
}

// SetVar sets a variable value (searches scopes).
func (c *Context) SetVar(name string, value *sv.SV) {
	// Search from innermost to outermost
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if _, ok := c.scopes[i][name]; ok {
			c.scopes[i][name] = value
			return
		}
	}
	// Not found - create in current scope
	if len(c.scopes) == 0 {
		c.scopes = append(c.scopes, make(map[string]*sv.SV))
	}
	c.scopes[len(c.scopes)-1][name] = value
}

// GetVar gets a variable value.
func (c *Context) GetVar(name string) *sv.SV {
	// Search from innermost to outermost
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if v, ok := c.scopes[i][name]; ok {
			return v
		}
	}
	return sv.NewUndef()
}

// PushScope creates a new scope.
func (c *Context) PushScope() {
	c.scopes = append(c.scopes, make(map[string]*sv.SV))
}

// PopScope removes the current scope.
func (c *Context) PopScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

// ============================================================
// Subroutine Management
// ============================================================

// DeclareSub declares a subroutine.
func (c *Context) DeclareSub(name string, body *ast.BlockStmt) {
	c.subs[name] = body
}

// GetSub gets a subroutine body.
func (c *Context) GetSub(name string) *ast.BlockStmt {
	return c.subs[name]
}

// ============================================================
// Arguments @_
// ============================================================

// SetArgs sets @_ for current call.
func (c *Context) SetArgs(args []*sv.SV) {
	c.args = sv.NewArrayRef(args...)
}

// GetArgs returns @_ array.
func (c *Context) GetArgs() *sv.SV {
	if c.args == nil {
		c.args = sv.NewArrayRef()
	}
	return c.args
}

// ============================================================
// Return Control
// ============================================================

// SetReturn sets return value and flag.
func (c *Context) SetReturn(value *sv.SV) {
	c.returnValue = value
	c.hasReturn = true
}

// HasReturn checks if return was called.
func (c *Context) HasReturn() bool {
	return c.hasReturn
}

// ReturnValue gets the return value.
func (c *Context) ReturnValue() *sv.SV {
	if c.returnValue == nil {
		return sv.NewUndef()
	}
	return c.returnValue
}

// ClearReturn clears return flag.
func (c *Context) ClearReturn() {
	c.hasReturn = false
	c.returnValue = nil
}

// ============================================================
// Last Control
// ============================================================

// SetLast sets last flag.
func (c *Context) SetLast(label string) {
	c.lastLabel = label
	c.hasLast = true
}

// HasLast checks if last was called.
func (c *Context) HasLast() bool {
	return c.hasLast
}

// ClearLast clears last flag.
func (c *Context) ClearLast() {
	c.hasLast = false
	c.lastLabel = ""
}

// ============================================================
// Next Control
// ============================================================

// SetNext sets next flag.
func (c *Context) SetNext(label string) {
	c.nextLabel = label
	c.hasNext = true
}

// HasNext checks if next was called.
func (c *Context) HasNext() bool {
	return c.hasNext
}

// ClearNext clears next flag.
func (c *Context) ClearNext() {
	c.hasNext = false
	c.nextLabel = ""
}

// ============================================================
// Special Variables
// ============================================================

// GetSpecialVar gets a special variable by name.
func (c *Context) GetSpecialVar(name string) *sv.SV {
	switch name {
	case "$_":
		return c.runtime.Underscore()
	case "$/":
		return c.runtime.InputRS()
	case "$\\":
		return c.runtime.OutputRS()
	case "$,":
		return c.runtime.OutputFS()
	case "$\"":
		return c.runtime.ListSep()
	case "$$":
		return c.runtime.PID()
	case "$0":
		return c.runtime.ProgName()
	case "$@":
		return c.runtime.EvalError()
	case "$!":
		return c.runtime.OSError()
	case "$?":
		return c.runtime.ChildError()
	case "$&":
		return c.runtime.Match()
	case "$`":
		return c.runtime.PreMatch()
	case "$'":
		return c.runtime.PostMatch()
	case "$+":
		return c.runtime.LastParen()
	case "$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9":
		n := int(name[1] - '0')
		return c.runtime.Capture(n)
	default:
		return sv.NewUndef()
	}
}
