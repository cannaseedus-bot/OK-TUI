package compiler

// ASTNode is the interface for all AST nodes
type ASTNode interface {
	NodeType() string
	Position() (line, col int)
}

// Program represents the root AST node
type Program struct {
	Statements []ASTNode
	Line       int
	Column     int
}

func (p *Program) NodeType() string              { return "Program" }
func (p *Program) Position() (line, col int)     { return p.Line, p.Column }

// Expression nodes

// NumberLiteral represents a numeric literal
type NumberLiteral struct {
	Value  float64
	Line   int
	Column int
}

func (n *NumberLiteral) NodeType() string            { return "NumberLiteral" }
func (n *NumberLiteral) Position() (line, col int)   { return n.Line, n.Column }

// StringLiteral represents a string literal
type StringLiteral struct {
	Value  string
	Line   int
	Column int
}

func (s *StringLiteral) NodeType() string            { return "StringLiteral" }
func (s *StringLiteral) Position() (line, col int)   { return s.Line, s.Column }

// Identifier represents a variable or symbol reference
type Identifier struct {
	Name   string
	Line   int
	Column int
}

func (i *Identifier) NodeType() string            { return "Identifier" }
func (i *Identifier) Position() (line, col int)   { return i.Line, i.Column }

// BinaryOp represents a binary operation
type BinaryOp struct {
	Left   ASTNode
	Op     TokenType
	Right  ASTNode
	Line   int
	Column int
}

func (b *BinaryOp) NodeType() string            { return "BinaryOp" }
func (b *BinaryOp) Position() (line, col int)   { return b.Line, b.Column }

// UnaryOp represents a unary operation
type UnaryOp struct {
	Op     TokenType
	Right  ASTNode
	Line   int
	Column int
}

func (u *UnaryOp) NodeType() string            { return "UnaryOp" }
func (u *UnaryOp) Position() (line, col int)   { return u.Line, u.Column }

// ArrayLiteral represents an array literal [...]
type ArrayLiteral struct {
	Elements []ASTNode
	Line     int
	Column   int
}

func (a *ArrayLiteral) NodeType() string            { return "ArrayLiteral" }
func (a *ArrayLiteral) Position() (line, col int)   { return a.Line, a.Column }

// MapLiteral represents a map/object literal {...}
type MapLiteral struct {
	Pairs map[string]ASTNode // key => value
	Line  int
	Column int
}

func (m *MapLiteral) NodeType() string            { return "MapLiteral" }
func (m *MapLiteral) Position() (line, col int)   { return m.Line, m.Column }

// Statement nodes

// Assignment represents a "wo" assignment (assignment)
type Assignment struct {
	Name   string
	Value  ASTNode
	Line   int
	Column int
}

func (a *Assignment) NodeType() string            { return "Assignment" }
func (a *Assignment) Position() (line, col int)   { return a.Line, a.Column }

// PopStatement represents a "pop" statement (execute/call)
type PopStatement struct {
	Function ASTNode                 // Can be PackCall or Identifier
	Args     map[string]ASTNode      // Keyword arguments
	Line     int
	Column   int
}

func (p *PopStatement) NodeType() string            { return "PopStatement" }
func (p *PopStatement) Position() (line, col int)   { return p.Line, p.Column }

// IfStatement represents a "sek" conditional
type IfStatement struct {
	Condition ASTNode
	Then      []ASTNode
	Else      []ASTNode
	Line      int
	Column    int
}

func (i *IfStatement) NodeType() string            { return "IfStatement" }
func (i *IfStatement) Position() (line, col int)   { return i.Line, i.Column }

// LoopStatement represents a "k'ayab" loop
type LoopStatement struct {
	Iterable ASTNode   // What to iterate over
	Body     []ASTNode // Loop body
	Line     int
	Column   int
}

func (l *LoopStatement) NodeType() string            { return "LoopStatement" }
func (l *LoopStatement) Position() (line, col int)   { return l.Line, l.Column }

// PackCall represents a pack invocation
// Example: lam_o.infer :model "llama2" :prompt "hello"
type PackCall struct {
	PackName  string
	Handler   string
	Args      map[string]ASTNode
	Line      int
	Column    int
}

func (p *PackCall) NodeType() string            { return "PackCall" }
func (p *PackCall) Position() (line, col int)   { return p.Line, p.Column }

// HandlerDef represents a handler definition
type HandlerDef struct {
	Name       string
	Params     []string
	Body       []ASTNode
	Line       int
	Column     int
}

func (h *HandlerDef) NodeType() string            { return "HandlerDef" }
func (h *HandlerDef) Position() (line, col int)   { return h.Line, h.Column }

// ReturnStatement represents a return from handler
type ReturnStatement struct {
	Value  ASTNode
	Line   int
	Column int
}

func (r *ReturnStatement) NodeType() string            { return "ReturnStatement" }
func (r *ReturnStatement) Position() (line, col int)   { return r.Line, r.Column }

// RuntimeTierAccess represents access to runtime tiers (@mem, @call, @ipc, @db, @api)
type RuntimeTierAccess struct {
	Tier   string // "mem", "call", "ipc", "db", "api"
	Method string // Operation on the tier
	Args   []ASTNode
	Line   int
	Column int
}

func (r *RuntimeTierAccess) NodeType() string            { return "RuntimeTierAccess" }
func (r *RuntimeTierAccess) Position() (line, col int)   { return r.Line, r.Column }

// Block represents a block of statements
type Block struct {
	Statements []ASTNode
	Line       int
	Column     int
}

func (b *Block) NodeType() string            { return "Block" }
func (b *Block) Position() (line, col int)   { return b.Line, b.Column }

// FunctionCall represents a generic function call
type FunctionCall struct {
	Function ASTNode
	Args     []ASTNode
	Line     int
	Column   int
}

func (f *FunctionCall) NodeType() string            { return "FunctionCall" }
func (f *FunctionCall) Position() (line, col int)   { return f.Line, f.Column }

// IndexAccess represents array/map indexing
type IndexAccess struct {
	Object ASTNode
	Index  ASTNode
	Line   int
	Column int
}

func (i *IndexAccess) NodeType() string            { return "IndexAccess" }
func (i *IndexAccess) Position() (line, col int)   { return i.Line, i.Column }

// MemberAccess represents object property access
type MemberAccess struct {
	Object   ASTNode
	Member   string
	Line     int
	Column   int
}

func (m *MemberAccess) NodeType() string            { return "MemberAccess" }
func (m *MemberAccess) Position() (line, col int)   { return m.Line, m.Column }
