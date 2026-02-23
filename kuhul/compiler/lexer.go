// Package compiler implements the K'UHUL compiler pipeline.
// The compiler transforms K'UHUL source code through 7 phases:
// 1. Lexical Analysis - Tokenization
// 2. Syntax Analysis - AST construction
// 3. Semantic Analysis - Type checking and scope resolution
// 4. IR Generation - Intermediate representation
// 5. Optimization - Performance improvements
// 6. Code Generation - JavaScript/WASM emission
// 7. Assembly/Linking - Final executable generation
package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType string

const (
	// Literals
	TokenNumber    TokenType = "NUMBER"
	TokenString    TokenType = "STRING"
	TokenSymbol    TokenType = "SYMBOL"
	TokenGlyphWord TokenType = "GLYPHWORD"

	// Keywords (K'UHUL)
	TokenPop     TokenType = "pop"
	TokenWo      TokenType = "wo"
	TokenSek     TokenType = "sek"
	TokenKayab   TokenType = "k'ayab"
	TokenPack    TokenType = "pack"
	TokenHandler TokenType = "handler"

	// Operators
	TokenLParen    TokenType = "("
	TokenRParen    TokenType = ")"
	TokenLBrace    TokenType = "{"
	TokenRBrace    TokenType = "}"
	TokenLBracket  TokenType = "["
	TokenRBracket  TokenType = "]"
	TokenColon     TokenType = ":"
	TokenComma     TokenType = ","
	TokenArrow     TokenType = "=>"
	TokenPipe      TokenType = "|"
	TokenAt        TokenType = "@"

	// Special
	TokenEOF     TokenType = "EOF"
	TokenNewline TokenType = "NEWLINE"
	TokenComment TokenType = "COMMENT"
)

// Token represents a lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
	Literal interface{} // For parsed values (numbers, strings)
}

// Lexer performs lexical analysis on K'UHUL source code
type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	tokens []Token
	errors []LexError

	// Track K'UHUL keywords
	keywords map[string]TokenType
}

// LexError represents a lexical analysis error
type LexError struct {
	Message string
	Line    int
	Column  int
}

// NewLexer creates a new lexer for the given source code
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		col:    1,
		tokens: []Token{},
		errors: []LexError{},
		keywords: map[string]TokenType{
			"pop":      TokenPop,
			"wo":       TokenWo,
			"sek":      TokenSek,
			"k'ayab":   TokenKayab,
			"pack":     TokenPack,
			"handler":  TokenHandler,
			"true":     TokenSymbol, // Boolean literals
			"false":    TokenSymbol,
			"nil":      TokenSymbol,
		},
	}
}

// Tokenize performs lexical analysis and returns tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		l.skipWhitespace()

		if l.pos >= len(l.input) {
			break
		}

		// Try to match different token types
		if l.matchComment() {
			continue
		}
		if l.matchNumber() {
			continue
		}
		if l.matchString() {
			continue
		}
		if l.matchOperator() {
			continue
		}
		if l.matchSymbolOrKeyword() {
			continue
		}

		// Unknown character
		l.addError(fmt.Sprintf("unexpected character: %q", l.current()))
		l.advance()
	}

	l.addToken(TokenEOF, "EOF", nil)

	if len(l.errors) > 0 {
		return nil, fmt.Errorf("lexical errors: %v", l.errors)
	}

	return l.tokens, nil
}

// current returns the current character without advancing
func (l *Lexer) current() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

// peek returns the character at offset without advancing
func (l *Lexer) peek(offset int) rune {
	pos := l.pos + offset
	if pos >= len(l.input) {
		return 0
	}
	return rune(l.input[pos])
}

// advance moves to the next character
func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := rune(l.input[l.pos])
	l.pos++

	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}

	return ch
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.current()) {
		l.advance()
	}
}

// matchComment handles comments
func (l *Lexer) matchComment() bool {
	if l.current() == '/' && l.peek(1) == '/' {
		// Single-line comment
		startCol := l.col
		l.advance()
		l.advance()

		comment := ""
		for l.pos < len(l.input) && l.current() != '\n' {
			comment += string(l.advance())
		}

		l.addTokenAtCol(TokenComment, comment, nil, startCol)
		return true
	}

	return false
}

// matchNumber matches numeric literals
func (l *Lexer) matchNumber() bool {
	if !unicode.IsDigit(l.current()) && l.current() != '.' {
		return false
	}

	startPos := l.pos
	startCol := l.col

	// Collect digits before decimal point
	for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
		l.advance()
	}

	// Handle decimal point
	if l.current() == '.' && unicode.IsDigit(l.peek(1)) {
		l.advance()
		for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
			l.advance()
		}
	}

	numStr := l.input[startPos:l.pos]
	numVal, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		l.addError(fmt.Sprintf("invalid number: %s", numStr))
		return false
	}

	l.addTokenAtCol(TokenNumber, numStr, numVal, startCol)
	return true
}

// matchString matches string literals
func (l *Lexer) matchString() bool {
	if l.current() != '"' && l.current() != '\'' {
		return false
	}

	quote := l.current()
	startCol := l.col
	l.advance() // Skip opening quote

	str := ""
	for l.pos < len(l.input) && l.current() != quote {
		if l.current() == '\\' {
			l.advance()
			switch l.current() {
			case 'n':
				str += "\n"
			case 't':
				str += "\t"
			case 'r':
				str += "\r"
			case '\\':
				str += "\\"
			case '"':
				str += "\""
			case '\'':
				str += "'"
			default:
				str += string(l.current())
			}
			l.advance()
		} else {
			str += string(l.advance())
		}
	}

	if l.current() != quote {
		l.addError("unterminated string")
		return false
	}

	l.advance() // Skip closing quote
	l.addTokenAtCol(TokenString, str, str, startCol)
	return true
}

// matchOperator matches operators and delimiters
func (l *Lexer) matchOperator() bool {
	ch := l.current()
	startCol := l.col

	operators := map[rune]TokenType{
		'(': TokenLParen,
		')': TokenRParen,
		'{': TokenLBrace,
		'}': TokenRBrace,
		'[': TokenLBracket,
		']': TokenRBracket,
		':': TokenColon,
		',': TokenComma,
		'|': TokenPipe,
		'@': TokenAt,
	}

	if tokType, ok := operators[ch]; ok {
		l.advance()
		l.addTokenAtCol(tokType, string(ch), nil, startCol)
		return true
	}

	// Check for =>
	if ch == '=' && l.peek(1) == '>' {
		l.advance()
		l.advance()
		l.addTokenAtCol(TokenArrow, "=>", nil, startCol)
		return true
	}

	return false
}

// matchSymbolOrKeyword matches symbols and keywords
func (l *Lexer) matchSymbolOrKeyword() bool {
	if !isSymbolStart(l.current()) {
		return false
	}

	startPos := l.pos
	startCol := l.col

	// Collect symbol characters (including apostrophes for K'UHUL words)
	for l.pos < len(l.input) && (isSymbolChar(l.current()) || l.current() == '\'') {
		l.advance()
	}

	// Check for pack call pattern (symbol.handler)
	if l.current() == '.' && l.pos < len(l.input)-1 && isSymbolStart(l.peek(1)) {
		// This is a pack call like "lam_o.infer"
		l.advance() // consume the dot

		handlerStart := l.pos
		for l.pos < len(l.input) && isSymbolChar(l.current()) {
			l.advance()
		}

		fullSymbol := l.input[startPos : handlerStart-1] // package name
		handler := l.input[handlerStart:l.pos]           // handler name

		// Emit package symbol first
		l.addTokenAtCol(TokenSymbol, fullSymbol, fullSymbol, startCol)

		// Then emit handler as ".handler" to indicate it's part of a pack call
		l.addTokenAtCol(TokenSymbol, "."+handler, "."+handler, startCol)
		return true
	}

	symbol := l.input[startPos:l.pos]

	// Check if it's a keyword
	if tokType, ok := l.keywords[symbol]; ok {
		l.addTokenAtCol(tokType, symbol, symbol, startCol)
		return true
	}

	// Check if it's a glyph word (K'UHUL-specific)
	if strings.Contains(symbol, "'") {
		l.addTokenAtCol(TokenGlyphWord, symbol, symbol, startCol)
		return true
	}

	// Otherwise it's a regular symbol
	l.addTokenAtCol(TokenSymbol, symbol, symbol, startCol)
	return true
}

// isSymbolStart checks if a rune can start a symbol
func isSymbolStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}

// isSymbolChar checks if a rune can be part of a symbol
func isSymbolChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}

// addToken adds a token to the token stream
func (l *Lexer) addToken(tokType TokenType, value string, literal interface{}) {
	l.tokens = append(l.tokens, Token{
		Type:    tokType,
		Value:   value,
		Line:    l.line,
		Column:  l.col - len(value),
		Literal: literal,
	})
}

// addTokenAtCol adds a token with specific column
func (l *Lexer) addTokenAtCol(tokType TokenType, value string, literal interface{}, col int) {
	l.tokens = append(l.tokens, Token{
		Type:    tokType,
		Value:   value,
		Line:    l.line,
		Column:  col,
		Literal: literal,
	})
}

// addError adds a lexical error
func (l *Lexer) addError(message string) {
	l.errors = append(l.errors, LexError{
		Message: message,
		Line:    l.line,
		Column:  l.col,
	})
}

// GetErrors returns lexical errors
func (l *Lexer) GetErrors() []LexError {
	return l.errors
}
