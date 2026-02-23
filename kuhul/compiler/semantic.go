package compiler

import (
	"fmt"
)

// DataType represents a data type in the K'UHUL type system
type DataType string

const (
	TypeNumber   DataType = "number"
	TypeString   DataType = "string"
	TypeBoolean  DataType = "bool"
	TypeArray    DataType = "array"
	TypeObject   DataType = "object"
	TypeFunction DataType = "function"
	TypeAny      DataType = "any"
	TypeNil      DataType = "nil"
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name     string
	Type     DataType
	Scope    *Scope
	Line     int
	Column   int
	Assigned bool
	Used     bool
}

// Scope represents a scope in the symbol table
type Scope struct {
	Symbols map[string]*Symbol
	Parent  *Scope
	Level   int // 0 = global, 1+ = nested
}

// SemanticError represents a semantic analysis error
type SemanticError struct {
	Message string
	Line    int
	Column  int
	Code    string // Error code for classification
}

// SemanticAnalyzer performs semantic analysis on the AST
type SemanticAnalyzer struct {
	ast           *Program
	globalScope   *Scope
	currentScope  *Scope
	errors        []SemanticError
	warnings      []SemanticError
	annotatedAST  *Program
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer(ast *Program) *SemanticAnalyzer {
	globalScope := &Scope{
		Symbols: make(map[string]*Symbol),
		Parent:  nil,
		Level:   0,
	}

	return &SemanticAnalyzer{
		ast:          ast,
		globalScope:  globalScope,
		currentScope: globalScope,
		errors:       []SemanticError{},
		warnings:     []SemanticError{},
		annotatedAST: ast,
	}
}

// Analyze performs semantic analysis
func (s *SemanticAnalyzer) Analyze() error {
	if s.ast == nil {
		return fmt.Errorf("no AST to analyze")
	}

	// First pass: collect all declarations
	for _, stmt := range s.ast.Statements {
		s.analyzeStatement(stmt)
	}

	if len(s.errors) > 0 {
		return fmt.Errorf("semantic errors: %v", s.errors)
	}

	return nil
}

// analyzeStatement analyzes a statement
func (s *SemanticAnalyzer) analyzeStatement(node ASTNode) {
	switch stmt := node.(type) {
	case *Assignment:
		s.analyzeAssignment(stmt)
	case *PopStatement:
		s.analyzePopStatement(stmt)
	case *IfStatement:
		s.analyzeIfStatement(stmt)
	case *LoopStatement:
		s.analyzeLoopStatement(stmt)
	case *HandlerDef:
		s.analyzeHandlerDef(stmt)
	case *ReturnStatement:
		s.analyzeReturnStatement(stmt)
	case *Block:
		for _, n := range stmt.Statements {
			s.analyzeStatement(n)
		}
	}
}

// analyzeAssignment analyzes an assignment
func (s *SemanticAnalyzer) analyzeAssignment(node *Assignment) {
	// Check if variable is already defined (redeclaration)
	if sym, exists := s.currentScope.Symbols[node.Name]; exists {
		s.addError(
			fmt.Sprintf("variable %q already defined at line %d", node.Name, sym.Line),
			node.Line, node.Column, "REDECLARATION",
		)
		return
	}

	// Analyze the RHS expression
	valueType := s.inferType(node.Value)

	// Add symbol to scope
	s.currentScope.Symbols[node.Name] = &Symbol{
		Name:     node.Name,
		Type:     valueType,
		Scope:    s.currentScope,
		Line:     node.Line,
		Column:   node.Column,
		Assigned: true,
		Used:     false,
	}
}

// analyzePopStatement analyzes a pop (execute) statement
func (s *SemanticAnalyzer) analyzePopStatement(node *PopStatement) {
	// The function should be either an identifier or a pack call
	switch fn := node.Function.(type) {
	case *PackCall:
		// Check that the pack exists (would be runtime check)
		_ = fn
	case *Identifier:
		// Check that the identifier is defined
		if !s.isSymbolDefined(fn.Name) {
			s.addWarning(
				fmt.Sprintf("calling undefined function %q", fn.Name),
				fn.Line, fn.Column, "UNDEFINED_FUNCTION",
			)
		}
	}

	// Analyze arguments
	for _, arg := range node.Args {
		s.inferType(arg)
	}
}

// analyzeIfStatement analyzes an if statement
func (s *SemanticAnalyzer) analyzeIfStatement(node *IfStatement) {
	// Analyze condition
	condType := s.inferType(node.Condition)

	// Condition should be boolean-like
	if condType != TypeBoolean && condType != TypeAny {
		s.addWarning(
			fmt.Sprintf("condition is %s, expected boolean", condType),
			node.Line, node.Column, "TYPE_MISMATCH",
		)
	}

	// Analyze then branch in new scope
	s.pushScope()
	for _, stmt := range node.Then {
		s.analyzeStatement(stmt)
	}
	s.popScope()

	// Analyze else branch in new scope
	if len(node.Else) > 0 {
		s.pushScope()
		for _, stmt := range node.Else {
			s.analyzeStatement(stmt)
		}
		s.popScope()
	}
}

// analyzeLoopStatement analyzes a loop statement
func (s *SemanticAnalyzer) analyzeLoopStatement(node *LoopStatement) {
	// Analyze iterable
	iterType := s.inferType(node.Iterable)

	// Iterable should be array-like
	if iterType != TypeArray && iterType != TypeAny {
		s.addWarning(
			fmt.Sprintf("loop iterable is %s, expected array", iterType),
			node.Line, node.Column, "TYPE_MISMATCH",
		)
	}

	// Analyze body in new scope
	s.pushScope()
	for _, stmt := range node.Body {
		s.analyzeStatement(stmt)
	}
	s.popScope()
}

// analyzeHandlerDef analyzes a handler definition
func (s *SemanticAnalyzer) analyzeHandlerDef(node *HandlerDef) {
	// Check for redeclaration
	if sym, exists := s.currentScope.Symbols[node.Name]; exists {
		s.addError(
			fmt.Sprintf("handler %q already defined at line %d", node.Name, sym.Line),
			node.Line, node.Column, "REDECLARATION",
		)
		return
	}

	// Add handler symbol
	s.currentScope.Symbols[node.Name] = &Symbol{
		Name:     node.Name,
		Type:     TypeFunction,
		Scope:    s.currentScope,
		Line:     node.Line,
		Column:   node.Column,
		Assigned: true,
		Used:     false,
	}

	// Create new scope for handler body
	s.pushScope()

	// Add parameters to scope
	for _, param := range node.Params {
		s.currentScope.Symbols[param] = &Symbol{
			Name:     param,
			Type:     TypeAny,
			Scope:    s.currentScope,
			Line:     node.Line,
			Column:   node.Column,
			Assigned: true,
			Used:     false,
		}
	}

	// Analyze body
	for _, stmt := range node.Body {
		s.analyzeStatement(stmt)
	}

	s.popScope()
}

// analyzeReturnStatement analyzes a return statement
func (s *SemanticAnalyzer) analyzeReturnStatement(node *ReturnStatement) {
	if node.Value != nil {
		s.inferType(node.Value)
	}
}

// inferType infers the type of an expression
func (s *SemanticAnalyzer) inferType(node ASTNode) DataType {
	switch expr := node.(type) {
	case *NumberLiteral:
		return TypeNumber
	case *StringLiteral:
		return TypeString
	case *ArrayLiteral:
		return TypeArray
	case *MapLiteral:
		return TypeObject
	case *Identifier:
		sym := s.lookupSymbol(expr.Name)
		if sym != nil {
			sym.Used = true
			return sym.Type
		}
		// Unknown identifier - might be a constant or built-in
		return TypeAny
	case *BinaryOp:
		return s.inferBinaryOpType(expr)
	case *UnaryOp:
		return s.inferUnaryOpType(expr)
	case *FunctionCall:
		return TypeAny // Assume any type for now
	case *PackCall:
		return TypeAny // Pack calls can return any type
	case *RuntimeTierAccess:
		return TypeAny // Runtime tier access returns any type
	default:
		return TypeAny
	}
}

// inferBinaryOpType infers the type of a binary operation
func (s *SemanticAnalyzer) inferBinaryOpType(node *BinaryOp) DataType {
	leftType := s.inferType(node.Left)
	rightType := s.inferType(node.Right)

	// Type checking for operators
	switch node.Op {
	case TokenColon: // Key-value in map
		return TypeObject
	default:
		// For now, return the left type or any if mixed
		if leftType == rightType {
			return leftType
		}
		return TypeAny
	}
}

// inferUnaryOpType infers the type of a unary operation
func (s *SemanticAnalyzer) inferUnaryOpType(node *UnaryOp) DataType {
	return s.inferType(node.Right)
}

// isSymbolDefined checks if a symbol is defined
func (s *SemanticAnalyzer) isSymbolDefined(name string) bool {
	return s.lookupSymbol(name) != nil
}

// lookupSymbol looks up a symbol in the current scope hierarchy
func (s *SemanticAnalyzer) lookupSymbol(name string) *Symbol {
	scope := s.currentScope
	for scope != nil {
		if sym, exists := scope.Symbols[name]; exists {
			return sym
		}
		scope = scope.Parent
	}
	return nil
}

// pushScope pushes a new scope
func (s *SemanticAnalyzer) pushScope() {
	newScope := &Scope{
		Symbols: make(map[string]*Symbol),
		Parent:  s.currentScope,
		Level:   s.currentScope.Level + 1,
	}
	s.currentScope = newScope
}

// popScope pops the current scope
func (s *SemanticAnalyzer) popScope() {
	if s.currentScope.Parent != nil {
		s.currentScope = s.currentScope.Parent
	}
}

// addError adds a semantic error
func (s *SemanticAnalyzer) addError(message string, line, col int, code string) {
	s.errors = append(s.errors, SemanticError{
		Message: message,
		Line:    line,
		Column:  col,
		Code:    code,
	})
}

// addWarning adds a semantic warning
func (s *SemanticAnalyzer) addWarning(message string, line, col int, code string) {
	s.warnings = append(s.warnings, SemanticError{
		Message: message,
		Line:    line,
		Column:  col,
		Code:    code,
	})
}

// GetErrors returns semantic errors
func (s *SemanticAnalyzer) GetErrors() []SemanticError {
	return s.errors
}

// GetWarnings returns semantic warnings
func (s *SemanticAnalyzer) GetWarnings() []SemanticError {
	return s.warnings
}

// GetAnnotatedAST returns the annotated AST
func (s *SemanticAnalyzer) GetAnnotatedAST() *Program {
	return s.annotatedAST
}

// GetSymbolTable returns the global symbol table
func (s *SemanticAnalyzer) GetSymbolTable() *Scope {
	return s.globalScope
}
