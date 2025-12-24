package codegen

import "perlc/pkg/ast"

func (g *Generator) varName(expr ast.Expression) string {
	switch v := expr.(type) {
	case *ast.ScalarVar:
		return "v_" + v.Name
	case *ast.ArrayVar:
		return "a_" + v.Name
	case *ast.HashVar:
		return "h_" + v.Name
	}
	return "_"
}

func (g *Generator) scalarName(name string) string {
	return "v_" + name
}

func (g *Generator) arrayName(name string) string {
	return "a_" + name
}

func (g *Generator) hashName(name string) string {
	return "h_" + name
}

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
