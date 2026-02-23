package compiler

import (
	"fmt"
	"strings"
)

// Parser performs syntax analysis on tokens
type Parser struct {
	tokens   []Token
	pos      int
	current  Token
	errors   []ParseError
}

// ParseError represents a parse error
type ParseError struct {
	Message string
	Line    int
	Column  int
	Token   string
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
		errors: []ParseError{},
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// Parse performs parsing and returns the AST
func (p *Parser) Parse() (*Program, error) {
	program := &Program{
		Statements: []ASTNode{},
		Line:       1,
		Column:     1,
	}

	for !p.isAtEnd() {
		// Skip comments and newlines
		if p.current.Type == TokenComment || p.current.Type == TokenNewline {
			p.advance()
			continue
		}

		stmt, err := p.parseStatement()
		if err != nil {
			p.addError(err.Error())
			p.synchronize()
			continue
		}

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}

	return program, nil
}

// parseStatement parses a statement
func (p *Parser) parseStatement() (ASTNode, error) {
	switch p.current.Type {
	case TokenWo:
		return p.parseAssignment()
	case TokenPop:
		return p.parsePopStatement()
	case TokenSek:
		return p.parseIfStatement()
	case TokenKayab:
		return p.parseLoopStatement()
	case TokenHandler:
		return p.parseHandlerDef()
	default:
		// Try to parse as expression statement
		expr, err := p.parseExpression()
		return expr, err
	}
}

// parseAssignment parses a "wo" assignment
func (p *Parser) parseAssignment() (*Assignment, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "wo"

	if p.current.Type != TokenSymbol {
		return nil, fmt.Errorf("expected symbol after 'wo', got %s", p.current.Type)
	}

	name := p.current.Value
	p.advance()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &Assignment{
		Name:   name,
		Value:  value,
		Line:   line,
		Column: col,
	}, nil
}

// parsePopStatement parses a "pop" execution statement
func (p *Parser) parsePopStatement() (*PopStatement, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "pop"

	// Parse the function/handler to call
	fn, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Parse keyword arguments
	args := make(map[string]ASTNode)

	for p.current.Type == TokenColon {
		p.advance() // consume ":"

		if p.current.Type != TokenSymbol {
			return nil, fmt.Errorf("expected argument name, got %s", p.current.Type)
		}

		argName := p.current.Value
		p.advance()

		argValue, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		args[argName] = argValue
	}

	return &PopStatement{
		Function: fn,
		Args:     args,
		Line:     line,
		Column:   col,
	}, nil
}

// parseIfStatement parses a "sek" conditional
func (p *Parser) parseIfStatement() (*IfStatement, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "sek"

	// Parse condition
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Parse then branch
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	// Check for else branch
	var elseBlock []ASTNode
	if p.matchKeyword("else") || p.matchKeyword("sek.else") {
		p.advance()
		var err error
		elseBlock, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &IfStatement{
		Condition: condition,
		Then:      then,
		Else:      elseBlock,
		Line:      line,
		Column:    col,
	}, nil
}

// parseLoopStatement parses a "k'ayab" loop
func (p *Parser) parseLoopStatement() (*LoopStatement, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "k'ayab"

	// Parse iterable
	iterable, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Parse loop body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &LoopStatement{
		Iterable: iterable,
		Body:     body,
		Line:     line,
		Column:   col,
	}, nil
}

// parseHandlerDef parses a handler definition
func (p *Parser) parseHandlerDef() (*HandlerDef, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "handler"

	if p.current.Type != TokenSymbol {
		return nil, fmt.Errorf("expected handler name, got %s", p.current.Type)
	}

	name := p.current.Value
	p.advance()

	// Parse parameters (optional)
	var params []string
	if p.current.Type == TokenLParen {
		p.advance()

		for p.current.Type != TokenRParen && !p.isAtEnd() {
			if p.current.Type != TokenSymbol {
				return nil, fmt.Errorf("expected parameter name")
			}

			params = append(params, p.current.Value)
			p.advance()

			if p.current.Type == TokenComma {
				p.advance()
			}
		}

		if p.current.Type != TokenRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		p.advance()
	}

	// Parse body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &HandlerDef{
		Name:   name,
		Params: params,
		Body:   body,
		Line:   line,
		Column: col,
	}, nil
}

// parseBlock parses a block of statements
func (p *Parser) parseBlock() ([]ASTNode, error) {
	var statements []ASTNode

	// If next token is {, consume it
	if p.current.Type == TokenLBrace {
		p.advance()

		for p.current.Type != TokenRBrace && !p.isAtEnd() {
			if p.current.Type == TokenComment || p.current.Type == TokenNewline {
				p.advance()
				continue
			}

			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}

			if stmt != nil {
				statements = append(statements, stmt)
			}
		}

		if p.current.Type != TokenRBrace {
			return nil, fmt.Errorf("expected '}'")
		}
		p.advance()
	}

	return statements, nil
}

// parseExpression parses an expression
func (p *Parser) parseExpression() (ASTNode, error) {
	return p.parsePrimary()
}

// parsePrimary parses a primary expression
func (p *Parser) parsePrimary() (ASTNode, error) {
	switch p.current.Type {
	case TokenNumber:
		val := p.current.Literal.(float64)
		line := p.current.Line
		col := p.current.Column
		p.advance()
		return &NumberLiteral{Value: val, Line: line, Column: col}, nil

	case TokenString:
		val := p.current.Literal.(string)
		line := p.current.Line
		col := p.current.Column
		p.advance()
		return &StringLiteral{Value: val, Line: line, Column: col}, nil

	case TokenSymbol:
		return p.parseSymbolOrPackCall()

	case TokenGlyphWord:
		return p.parseGlyphWord()

	case TokenAt:
		return p.parseRuntimeTierAccess()

	case TokenLBracket:
		return p.parseArrayLiteral()

	case TokenLBrace:
		return p.parseMapLiteral()

	case TokenLParen:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.current.Type != TokenRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		p.advance()
		return expr, nil

	default:
		return nil, fmt.Errorf("unexpected token: %s", p.current.Type)
	}
}

// parseSymbolOrPackCall parses a symbol or pack call
func (p *Parser) parseSymbolOrPackCall() (ASTNode, error) {
	line := p.current.Line
	col := p.current.Column
	name := p.current.Value
	p.advance()

	// Check for pack call (package.handler)
	if p.current.Type == TokenSymbol && strings.HasPrefix(p.current.Value, ".") {
		handler := strings.TrimPrefix(p.current.Value, ".")
		p.advance()

		// Parse keyword arguments
		args := make(map[string]ASTNode)

		for p.current.Type == TokenColon {
			p.advance()

			if p.current.Type != TokenSymbol {
				return nil, fmt.Errorf("expected argument name")
			}

			argName := p.current.Value
			p.advance()

			argValue, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			args[argName] = argValue
		}

		return &PackCall{
			PackName: name,
			Handler:  handler,
			Args:     args,
			Line:     line,
			Column:   col,
		}, nil
	}

	// Otherwise it's just an identifier
	return &Identifier{Name: name, Line: line, Column: col}, nil
}

// parseGlyphWord handles K'UHUL glyph words
func (p *Parser) parseGlyphWord() (ASTNode, error) {
	name := p.current.Value
	line := p.current.Line
	col := p.current.Column
	p.advance()

	// Glyph words are treated as special identifiers
	return &Identifier{Name: name, Line: line, Column: col}, nil
}

// parseRuntimeTierAccess parses @mem, @call, @ipc, @db, @api access
func (p *Parser) parseRuntimeTierAccess() (ASTNode, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "@"

	if p.current.Type != TokenSymbol {
		return nil, fmt.Errorf("expected tier name after @")
	}

	tier := p.current.Value
	p.advance()

	// Parse method/operation
	var method string
	var args []ASTNode

	if p.current.Type == TokenSymbol {
		method = p.current.Value
		p.advance()

		// Parse arguments if present
		if p.current.Type == TokenLParen {
			p.advance()

			for p.current.Type != TokenRParen && !p.isAtEnd() {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}

				args = append(args, arg)

				if p.current.Type == TokenComma {
					p.advance()
				}
			}

			if p.current.Type != TokenRParen {
				return nil, fmt.Errorf("expected ')'")
			}
			p.advance()
		}
	}

	return &RuntimeTierAccess{
		Tier:   tier,
		Method: method,
		Args:   args,
		Line:   line,
		Column: col,
	}, nil
}

// parseArrayLiteral parses an array literal [...]
func (p *Parser) parseArrayLiteral() (*ArrayLiteral, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "["

	var elements []ASTNode

	for p.current.Type != TokenRBracket && !p.isAtEnd() {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		elements = append(elements, elem)

		if p.current.Type == TokenComma {
			p.advance()
		}
	}

	if p.current.Type != TokenRBracket {
		return nil, fmt.Errorf("expected ']'")
	}
	p.advance()

	return &ArrayLiteral{
		Elements: elements,
		Line:     line,
		Column:   col,
	}, nil
}

// parseMapLiteral parses a map literal {...}
func (p *Parser) parseMapLiteral() (*MapLiteral, error) {
	line := p.current.Line
	col := p.current.Column

	p.advance() // consume "{"

	pairs := make(map[string]ASTNode)

	for p.current.Type != TokenRBrace && !p.isAtEnd() {
		// Parse key
		var key string
		if p.current.Type == TokenSymbol {
			key = p.current.Value
			p.advance()
		} else if p.current.Type == TokenString {
			key = p.current.Literal.(string)
			p.advance()
		} else {
			return nil, fmt.Errorf("expected key in map literal")
		}

		// Expect colon
		if p.current.Type != TokenColon {
			return nil, fmt.Errorf("expected ':' after key in map literal")
		}
		p.advance()

		// Parse value
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		pairs[key] = value

		if p.current.Type == TokenComma {
			p.advance()
		}
	}

	if p.current.Type != TokenRBrace {
		return nil, fmt.Errorf("expected '}'")
	}
	p.advance()

	return &MapLiteral{
		Pairs:  pairs,
		Line:   line,
		Column: col,
	}, nil
}

// Utility methods

// advance moves to the next token
func (p *Parser) advance() {
	if p.pos < len(p.tokens)-1 {
		p.pos++
		p.current = p.tokens[p.pos]
	}
}

// isAtEnd checks if at end of tokens
func (p *Parser) isAtEnd() bool {
	return p.current.Type == TokenEOF
}

// matchKeyword checks if current token matches a keyword
func (p *Parser) matchKeyword(keyword string) bool {
	// For multi-word keywords like "sek.else"
	return p.current.Value == keyword
}

// addError adds a parse error
func (p *Parser) addError(message string) {
	p.errors = append(p.errors, ParseError{
		Message: message,
		Line:    p.current.Line,
		Column:  p.current.Column,
		Token:   p.current.Value,
	})
}

// synchronize tries to recover from parse errors
func (p *Parser) synchronize() {
	for !p.isAtEnd() {
		p.advance()

		// Skip to next statement
		if p.current.Type == TokenWo || p.current.Type == TokenPop ||
			p.current.Type == TokenSek || p.current.Type == TokenKayab {
			return
		}
	}
}
