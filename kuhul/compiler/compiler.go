package compiler

import (
	"fmt"
)

// CompileError represents a compilation error
type CompileError struct {
	Phase    string // Which phase failed
	Message  string
	Line     int
	Column   int
	Severity string // "error" or "warning"
}

// Compiler orchestrates the 7-phase compilation pipeline
type Compiler struct {
	source       string
	filename     string

	// Pipeline stages
	lexer        *Lexer
	parser       *Parser
	analyzer     *SemanticAnalyzer
	codegen      *CodeGenerator

	// Output artifacts
	tokens       []Token
	ast          *Program
	javascript   string

	// Error handling
	errors       []CompileError
	warnings     []CompileError
}

// NewCompiler creates a new compiler
func NewCompiler(source string) *Compiler {
	return NewCompilerWithFilename(source, "unknown.khl")
}

// NewCompilerWithFilename creates a new compiler with filename
func NewCompilerWithFilename(source, filename string) *Compiler {
	return &Compiler{
		source:   source,
		filename: filename,
		errors:   []CompileError{},
		warnings: []CompileError{},
	}
}

// Compile performs the full compilation pipeline
func (c *Compiler) Compile() (string, error) {
	// Phase 1: Lexical Analysis
	if err := c.phase1_Lexing(); err != nil {
		return "", err
	}

	// Phase 2: Syntax Analysis
	if err := c.phase2_Parsing(); err != nil {
		return "", err
	}

	// Phase 3: Semantic Analysis
	if err := c.phase3_SemanticAnalysis(); err != nil {
		return "", err
	}

	// Phase 4: Intermediate Representation (skipped for now - go straight to JS)
	// Phase 5: Optimization (skipped for basic implementation)
	// Phase 6: Code Generation
	if err := c.phase6_CodeGeneration(); err != nil {
		return "", err
	}

	// Phase 7: Assembly/Linking (for now just return JavaScript)
	// In future, could produce bytecode, WASM, etc.

	return c.javascript, nil
}

// phase1_Lexing performs lexical analysis
func (c *Compiler) phase1_Lexing() error {
	c.lexer = NewLexer(c.source)
	tokens, err := c.lexer.Tokenize()
	if err != nil {
		lexErrors := c.lexer.GetErrors()
		for _, e := range lexErrors {
			c.addError(
				"Lexical Analysis",
				e.Message,
				e.Line,
				e.Column,
				"error",
			)
		}
		return err
	}

	c.tokens = tokens
	return nil
}

// phase2_Parsing performs syntax analysis
func (c *Compiler) phase2_Parsing() error {
	c.parser = NewParser(c.tokens)
	ast, err := c.parser.Parse()
	if err != nil {
		return err
	}

	c.ast = ast
	return nil
}

// phase3_SemanticAnalysis performs semantic analysis
func (c *Compiler) phase3_SemanticAnalysis() error {
	if c.ast == nil {
		return fmt.Errorf("no AST for semantic analysis")
	}

	c.analyzer = NewSemanticAnalyzer(c.ast)
	err := c.analyzer.Analyze()
	if err != nil {
		// Add errors but continue (warnings don't stop compilation)
	}

	// Collect errors and warnings
	for _, e := range c.analyzer.GetErrors() {
		c.addError(
			"Semantic Analysis",
			e.Message,
			e.Line,
			e.Column,
			"error",
		)
	}

	for _, w := range c.analyzer.GetWarnings() {
		c.addWarning(
			"Semantic Analysis",
			w.Message,
			w.Line,
			w.Column,
		)
	}

	// Only fail on actual errors, not warnings
	if len(c.analyzer.GetErrors()) > 0 {
		return fmt.Errorf("semantic analysis failed")
	}

	return nil
}

// phase6_CodeGeneration performs code generation
func (c *Compiler) phase6_CodeGeneration() error {
	if c.ast == nil {
		return fmt.Errorf("no AST for code generation")
	}

	c.codegen = NewCodeGenerator(c.ast)
	js, err := c.codegen.Generate()
	if err != nil {
		return err
	}

	c.javascript = js
	return nil
}

// addError adds a compilation error
func (c *Compiler) addError(phase, message string, line, col int, severity string) {
	c.errors = append(c.errors, CompileError{
		Phase:    phase,
		Message:  message,
		Line:     line,
		Column:   col,
		Severity: severity,
	})
}

// addWarning adds a compilation warning
func (c *Compiler) addWarning(phase, message string, line, col int) {
	c.warnings = append(c.warnings, CompileError{
		Phase:    phase,
		Message:  message,
		Line:     line,
		Column:   col,
		Severity: "warning",
	})
}

// GetErrors returns compilation errors
func (c *Compiler) GetErrors() []CompileError {
	return c.errors
}

// GetWarnings returns compilation warnings
func (c *Compiler) GetWarnings() []CompileError {
	return c.warnings
}

// GetTokens returns the tokens from phase 1
func (c *Compiler) GetTokens() []Token {
	return c.tokens
}

// GetAST returns the AST from phase 2
func (c *Compiler) GetAST() *Program {
	return c.ast
}

// GetJavaScript returns the generated JavaScript code
func (c *Compiler) GetJavaScript() string {
	return c.javascript
}

// HasErrors returns true if there are compilation errors
func (c *Compiler) HasErrors() bool {
	return len(c.errors) > 0
}

// HasWarnings returns true if there are compilation warnings
func (c *Compiler) HasWarnings() bool {
	return len(c.warnings) > 0
}

// String returns a human-readable summary of compilation
func (c *Compiler) String() string {
	summary := fmt.Sprintf("Compilation of %s:\n", c.filename)

	if c.HasErrors() {
		summary += fmt.Sprintf("  Errors: %d\n", len(c.errors))
		for _, e := range c.errors {
			summary += fmt.Sprintf("    [%s] %s at %d:%d\n", e.Phase, e.Message, e.Line, e.Column)
		}
	}

	if c.HasWarnings() {
		summary += fmt.Sprintf("  Warnings: %d\n", len(c.warnings))
		for _, w := range c.warnings {
			summary += fmt.Sprintf("    [%s] %s at %d:%d\n", w.Phase, w.Message, w.Line, w.Column)
		}
	}

	if !c.HasErrors() && !c.HasWarnings() {
		summary += "  Status: OK ✓\n"
	}

	return summary
}

// CompileFile compiles a K'UHUL file
func CompileFile(filename string, source string) (string, error) {
	compiler := NewCompilerWithFilename(source, filename)
	return compiler.Compile()
}

// CompileString compiles a K'UHUL string
func CompileString(source string) (string, error) {
	return CompileFile("stdin", source)
}
