package token

import "testing"

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"my", MY},
		{"sub", SUB},
		{"if", IF},
		{"elsif", ELSIF},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
		{"foreach", FOREACH},
		{"return", RETURN},
		{"print", PRINT},
		{"use", USE},
		{"package", PACKAGE},
		{"eq", STREQ},
		{"ne", STRNE},
		{"and", ANDKW},
		{"or", ORKW},
		{"not", NOTKW},
		{"customIdent", IDENT},
		{"variableName", IDENT},
		{"_underscore", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LookupIdent(tt.input)
			if result != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		input    TokenType
		expected bool
	}{
		{MY, true},
		{SUB, true},
		{IF, true},
		{WHILE, true},
		{PRINT, true},
		{IDENT, false},
		{INT, false},
		{STRING, false},
		{PLUS, false},
		{ASSIGN, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := IsKeyword(tt.input)
			if result != tt.expected {
				t.Errorf("IsKeyword(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
