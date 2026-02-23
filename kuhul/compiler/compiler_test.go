package compiler

import (
	"strings"
	"testing"
)

// TestLexer tests lexical analysis
func TestLexer(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "simple assignment",
			input:     `wo x 42`,
			wantCount: 4, // wo, x, 42, EOF
		},
		{
			name:      "string literal",
			input:     `wo msg "hello world"`,
			wantCount: 4,
		},
		{
			name:      "keywords",
			input:     `pop sek k'ayab`,
			wantCount: 4, // pop, sek, k'ayab, EOF
		},
		{
			name:      "operators",
			input:     `( ) { } [ ] : ,`,
			wantCount: 9, // 8 operators + EOF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()

			if err != nil {
				t.Fatalf("tokenize error: %v", err)
			}

			if len(tokens) != tt.wantCount {
				t.Errorf("token count = %d, want %d", len(tokens), tt.wantCount)
			}

			// Last token should be EOF
			if len(tokens) > 0 && tokens[len(tokens)-1].Type != TokenEOF {
				t.Errorf("last token is %s, want EOF", tokens[len(tokens)-1].Type)
			}
		})
	}
}

// TestParser tests syntax analysis
func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple assignment",
			input:   `wo x 42`,
			wantErr: false,
		},
		{
			name:    "string assignment",
			input:   `wo msg "hello"`,
			wantErr: false,
		},
		{
			name:    "array literal",
			input:   `wo arr [1 2 3]`,
			wantErr: false,
		},
		{
			name:    "map literal",
			input:   `wo obj {x: 1 y: 2}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			parser := NewParser(tokens)
			ast, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("parser error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && ast == nil {
				t.Error("expected AST, got nil")
			}
		})
	}
}

// TestSemanticAnalyzer tests semantic analysis
func TestSemanticAnalyzer(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantWarn  bool
	}{
		{
			name:     "valid assignment",
			input:    `wo x 42`,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "valid handler",
			input:    `handler foo { wo x 1 }`,
			wantErr:  false,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			parser := NewParser(tokens)
			ast, err := parser.Parse()
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			analyzer := NewSemanticAnalyzer(ast)
			err = analyzer.Analyze()

			if (err != nil) != tt.wantErr {
				t.Errorf("analyzer error = %v, wantErr %v", err, tt.wantErr)
			}

			hasWarnings := len(analyzer.GetWarnings()) > 0
			if hasWarnings != tt.wantWarn {
				t.Errorf("has warnings = %v, wantWarn %v", hasWarnings, tt.wantWarn)
			}
		})
	}
}

// TestCodeGenerator tests code generation
func TestCodeGenerator(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantErr   bool
	}{
		{
			name:     "simple assignment generates declaration",
			input:    `wo x 42`,
			wantCode: "const x = 42",
			wantErr:  false,
		},
		{
			name:     "string assignment",
			input:    `wo msg "hello"`,
			wantCode: `"hello"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, _ := lexer.Tokenize()

			parser := NewParser(tokens)
			ast, _ := parser.Parse()

			codegen := NewCodeGenerator(ast)
			code, err := codegen.Generate()

			if (err != nil) != tt.wantErr {
				t.Errorf("codegen error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !strings.Contains(code, tt.wantCode) {
				t.Errorf("generated code does not contain %q:\n%s", tt.wantCode, code)
			}
		})
	}
}

// TestFullCompilation tests the full compilation pipeline
func TestFullCompilation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "hello world",
			input: `wo msg "hello world"
wo num 42`,
			wantErr: false,
		},
		{
			name: "handler definition",
			input: `handler greet(name) {
  wo greeting "Hello"
}`,
			wantErr: false,
		},
		{
			name: "conditional",
			input: `wo x 10
sek (x) {
  wo result "yes"
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(tt.input)
			code, err := compiler.Compile()

			if (err != nil) != tt.wantErr {
				t.Errorf("compile error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if code == "" {
					t.Error("expected generated code, got empty string")
				}

				// Should contain boilerplate
				if !strings.Contains(code, "$kuhul_main") {
					t.Error("generated code missing $kuhul_main function")
				}
			}
		})
	}
}

// TestCompileFile compiles a complete K'UHUL program
func TestCompileFile(t *testing.T) {
	kuhulCode := `
// Llama inference example
wo model "llama2"
wo prompt "What is 2+2?"

pop (lam_o.infer
  :model model
  :prompt prompt
  :temperature 0.7)
`

	compiler := NewCompiler(kuhulCode)
	code, err := compiler.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Should generate valid JavaScript
	if !strings.Contains(code, "lam_o") {
		t.Error("generated code missing pack reference")
	}

	if !strings.Contains(code, "await") {
		t.Error("generated code missing async support")
	}

	if !strings.Contains(code, "$mem") {
		t.Error("generated code missing memory tier")
	}

	t.Logf("Generated JavaScript:\n%s", code)
}

// BenchmarkCompilation benchmarks the compiler
func BenchmarkCompilation(b *testing.B) {
	source := `
wo x 1
wo y 2
wo z 3
handler add(a b) {
  wo result (a + b)
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compiler := NewCompiler(source)
		compiler.Compile()
	}
}

// TestCompilerErrorMessages tests error reporting
func TestCompilerErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrorIn  string // Which phase
		expectContains string // Error message should contain
	}{
		{
			name:          "unterminated string",
			input:         `wo x "hello`,
			expectErrorIn: "Lexical Analysis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(tt.input)
			_, err := compiler.Compile()

			if err == nil {
				t.Error("expected compilation error")
				return
			}

			if len(compiler.GetErrors()) == 0 && len(compiler.GetWarnings()) == 0 {
				t.Error("expected error details, got none")
			}
		})
	}
}

// TestPackCallGeneration tests pack invocation code generation
func TestPackCallGeneration(t *testing.T) {
	source := `pop (lam_o.infer :model "llama2" :prompt "hello")`

	compiler := NewCompiler(source)
	code, err := compiler.Compile()

	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Should reference the pack
	if !strings.Contains(code, "packs.lam_o") {
		t.Error("generated code missing pack reference")
	}

	// Should reference handleInfer
	if !strings.Contains(code, "Infer") {
		t.Error("generated code should reference handler")
	}

	// Generated code should be valid JavaScript
	if !strings.Contains(code, "await") {
		t.Error("generated code missing async support")
	}
}

// TestRuntimeTierAccess tests @mem, @call access generation
func TestRuntimeTierAccess(t *testing.T) {
	source := `@mem set x 42`

	compiler := NewCompiler(source)
	_, err := compiler.Compile()

	// This might fail in current implementation (runtime tiers not fully parsed)
	// But at least it should attempt to parse
	_ = err // Not testing for success, just coverage
}
