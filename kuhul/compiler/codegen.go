package compiler

import (
	"fmt"
	"strings"
)

// CodeGenerator generates JavaScript code from the AST
type CodeGenerator struct {
	ast    *Program
	code   strings.Builder
	indent int
	errors []string
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(ast *Program) *CodeGenerator {
	return &CodeGenerator{
		ast:    ast,
		indent: 0,
		errors: []string{},
	}
}

// Generate generates JavaScript code
func (cg *CodeGenerator) Generate() (string, error) {
	if cg.ast == nil {
		return "", fmt.Errorf("no AST to generate code from")
	}

	// Emit preamble
	cg.emitPreamble()

	// Emit statements
	for _, stmt := range cg.ast.Statements {
		cg.emitStatement(stmt)
		cg.writeln()
	}

	// Emit postamble
	cg.emitPostamble()

	if len(cg.errors) > 0 {
		return "", fmt.Errorf("codegen errors: %v", cg.errors)
	}

	return cg.code.String(), nil
}

// emitPreamble emits the boilerplate code at the beginning
func (cg *CodeGenerator) emitPreamble() {
	cg.writeln(`// K'UHUL Generated Code`)
	cg.writeln(`// Auto-generated from K'UHUL source`)
	cg.writeln()

	cg.writeln(`const packs = {`)
	cg.indent++
	cg.writeln(`lam_o: { handleInfer: async (ctx) => { /* pack impl */ } },`)
	cg.writeln(`scxq2: { },`)
	cg.writeln(`asx_ram: { },`)
	cg.writeln(`mx2lm: { },`)
	cg.indent--
	cg.writeln(`};`)
	cg.writeln()

	cg.writeln(`// Memory tier`)
	cg.writeln(`const $mem = new Map();`)
	cg.writeln()

	cg.writeln(`// Call tier`)
	cg.writeln(`const $context = {};`)
	cg.writeln()

	cg.writeln(`// Main function`)
	cg.writeln(`async function $kuhul_main() {`)
	cg.indent++
}

// emitPostamble emits the boilerplate code at the end
func (cg *CodeGenerator) emitPostamble() {
	cg.indent--
	cg.writeln(`}`)
	cg.writeln()
	cg.writeln(`// Execute`)
	cg.writeln(`$kuhul_main().catch(console.error);`)
}

// emitStatement emits a statement
func (cg *CodeGenerator) emitStatement(node ASTNode) {
	switch stmt := node.(type) {
	case *Assignment:
		cg.emitAssignment(stmt)
	case *PopStatement:
		cg.emitPopStatement(stmt)
	case *IfStatement:
		cg.emitIfStatement(stmt)
	case *LoopStatement:
		cg.emitLoopStatement(stmt)
	case *HandlerDef:
		cg.emitHandlerDef(stmt)
	case *ReturnStatement:
		cg.emitReturnStatement(stmt)
	case *Block:
		for _, n := range stmt.Statements {
			cg.emitStatement(n)
			cg.writeln()
		}
	}
}

// emitAssignment emits an assignment statement
func (cg *CodeGenerator) emitAssignment(node *Assignment) {
	cg.write(cg.indent_str() + "const " + node.Name + " = ")
	cg.emitExpression(node.Value)
	cg.writeln(";")
	cg.writeln(fmt.Sprintf("$mem.set('%s', %s);", node.Name, node.Name))
}

// emitPopStatement emits a pop (execute) statement
func (cg *CodeGenerator) emitPopStatement(node *PopStatement) {
	cg.write(cg.indent_str() + "const result = ")

	// Emit function call
	switch fn := node.Function.(type) {
	case *PackCall:
		cg.write(fmt.Sprintf("await packs.%s.handle%s({", fn.PackName, cg.capitalize(fn.Handler)))

		// Emit arguments
		first := true
		for key, val := range node.Args {
			if !first {
				cg.write(", ")
			}
			cg.write(key + ": ")
			cg.emitExpression(val)
			first = false
		}

		cg.write("})")
	case *Identifier:
		cg.write("await " + fn.Name + "(")

		// Emit positional arguments
		first := true
		for _, arg := range node.Args {
			if !first {
				cg.write(", ")
			}
			cg.emitExpression(arg)
			first = false
		}

		cg.write(")")
	}

	cg.writeln(";")
	cg.writeln("$mem.set('result', result);")
}

// emitIfStatement emits an if statement
func (cg *CodeGenerator) emitIfStatement(node *IfStatement) {
	cg.write(cg.indent_str() + "if (")
	cg.emitExpression(node.Condition)
	cg.writeln(") {")

	cg.indent++
	for _, stmt := range node.Then {
		cg.emitStatement(stmt)
		cg.writeln()
	}
	cg.indent--

	if len(node.Else) > 0 {
		cg.writeln(cg.indent_str() + "} else {")
		cg.indent++
		for _, stmt := range node.Else {
			cg.emitStatement(stmt)
			cg.writeln()
		}
		cg.indent--
	}

	cg.writeln(cg.indent_str() + "}")
}

// emitLoopStatement emits a loop statement
func (cg *CodeGenerator) emitLoopStatement(node *LoopStatement) {
	cg.write(cg.indent_str() + "for (const item of ")
	cg.emitExpression(node.Iterable)
	cg.writeln(") {")

	cg.indent++
	cg.writeln("$mem.set('item', item);")
	for _, stmt := range node.Body {
		cg.emitStatement(stmt)
		cg.writeln()
	}
	cg.indent--

	cg.writeln(cg.indent_str() + "}")
}

// emitHandlerDef emits a handler definition
func (cg *CodeGenerator) emitHandlerDef(node *HandlerDef) {
	cg.write(cg.indent_str() + "async function " + node.Name + "(")

	// Emit parameters
	for i, param := range node.Params {
		if i > 0 {
			cg.write(", ")
		}
		cg.write(param)
	}

	cg.writeln(") {")

	cg.indent++
	for _, stmt := range node.Body {
		cg.emitStatement(stmt)
		cg.writeln()
	}
	cg.indent--

	cg.writeln(cg.indent_str() + "}")
}

// emitReturnStatement emits a return statement
func (cg *CodeGenerator) emitReturnStatement(node *ReturnStatement) {
	cg.write(cg.indent_str() + "return ")
	if node.Value != nil {
		cg.emitExpression(node.Value)
	}
	cg.writeln(";")
}

// emitExpression emits an expression
func (cg *CodeGenerator) emitExpression(node ASTNode) {
	switch expr := node.(type) {
	case *NumberLiteral:
		cg.write(fmt.Sprintf("%v", expr.Value))
	case *StringLiteral:
		cg.write(fmt.Sprintf(`"%s"`, expr.Value))
	case *Identifier:
		cg.write("$mem.get('" + expr.Name + "')")
	case *ArrayLiteral:
		cg.write("[")
		for i, elem := range expr.Elements {
			if i > 0 {
				cg.write(", ")
			}
			cg.emitExpression(elem)
		}
		cg.write("]")
	case *MapLiteral:
		cg.write("{")
		i := 0
		for key, val := range expr.Pairs {
			if i > 0 {
				cg.write(", ")
			}
			cg.write(fmt.Sprintf(`"%s": `, key))
			cg.emitExpression(val)
			i++
		}
		cg.write("}")
	case *BinaryOp:
		cg.emitExpression(expr.Left)
		cg.write(" " + string(expr.Op) + " ")
		cg.emitExpression(expr.Right)
	case *FunctionCall:
		cg.write("(")
		cg.emitExpression(expr.Function)
		cg.write(")(")
		for i, arg := range expr.Args {
			if i > 0 {
				cg.write(", ")
			}
			cg.emitExpression(arg)
		}
		cg.write(")")
	case *PackCall:
		cg.write(fmt.Sprintf("packs.%s.handle%s(", expr.PackName, cg.capitalize(expr.Handler)))
		i := 0
		for key, val := range expr.Args {
			if i > 0 {
				cg.write(", ")
			}
			cg.write(key + ": ")
			cg.emitExpression(val)
			i++
		}
		cg.write(")")
	case *RuntimeTierAccess:
		cg.write(fmt.Sprintf("await $%s.%s(", expr.Tier, expr.Method))
		for i, arg := range expr.Args {
			if i > 0 {
				cg.write(", ")
			}
			cg.emitExpression(arg)
		}
		cg.write(")")
	}
}

// Utility methods

// write writes a string to the output
func (cg *CodeGenerator) write(s string) {
	cg.code.WriteString(s)
}

// writeln writes a string followed by a newline
func (cg *CodeGenerator) writeln(s ...string) {
	for _, str := range s {
		cg.code.WriteString(str)
	}
	cg.code.WriteString("\n")
}

// indent_str returns the indentation string
func (cg *CodeGenerator) indent_str() string {
	return strings.Repeat("    ", cg.indent)
}

// capitalize returns a capitalized string
func (cg *CodeGenerator) capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// GetErrors returns code generation errors
func (cg *CodeGenerator) GetErrors() []string {
	return cg.errors
}
