# K'UHUL Compiler Architecture

## Overview

The K'UHUL Compiler is a 7-phase pipeline that transforms K'UHUL source code into executable bytecode/JavaScript that runs on the 5-tier execution runtime. It bridges K'UHUL language semantics with Ollama integration.

```
K'UHUL Source → Lexer → Parser → Semantic → IR Gen → Optimizer → Codegen → Assembly
   (.khl)      (Tokens)  (AST)   (Symbols)  (IR)    (Optimized) (JS/WASM) (Binary)
                                                                               ↓
                                                        5-Tier Execution Runtime
                                                    (@mem @call @ipc @db @api)
```

---

## Phase 1: Lexical Analysis

**Input**: K'UHUL source code (.khl files)
**Output**: Token stream with location information
**Responsibility**: Break source into meaningful tokens

### Token Types
```go
type TokenType string

const (
    // Literals
    TOK_NUMBER    = "NUMBER"
    TOK_STRING    = "STRING"
    TOK_SYMBOL    = "SYMBOL"      // Variables/identifiers
    TOK_GLYPHWORD = "GLYPHWORD"   // K'UHUL glyph words (lam, wo, sek, etc.)

    // Keywords (K'UHUL)
    TOK_POP       = "pop"         // Execute/call
    TOK_WO        = "wo"          // Assignment
    TOK_SEK       = "sek"         // Conditional
    TOK_KAYAB     = "k'ayab"      // Loop
    TOK_PACK      = "pack"        // Pack invocation
    TOK_HANDLER   = "handler"     // Handler definition

    // Operators
    TOK_LPAREN    = "("
    TOK_RPAREN    = ")"
    TOK_LBRACE    = "{"
    TOK_RBRACE    = "}"
    TOK_LBRACKET  = "["
    TOK_RBRACKET  = "]"
    TOK_COLON     = ":"
    TOK_COMMA     = ","
    TOK_ARROW     = "=>"
    TOK_PIPE      = "|"

    // Special
    TOK_EOF       = "EOF"
    TOK_NEWLINE   = "NEWLINE"
)
```

### Lexer Implementation
```go
type Token struct {
    Type   TokenType
    Value  string
    Line   int
    Column int
}

type Lexer struct {
    input  string
    pos    int
    line   int
    col    int
    tokens []Token
}
```

---

## Phase 2: Syntax Analysis

**Input**: Token stream
**Output**: Abstract Syntax Tree (AST)
**Responsibility**: Parse tokens into hierarchical structure

### AST Node Types
```go
type ASTNode interface {
    Position() (line, col int)
    Kind() string
}

// Expressions
type NumberLiteral struct {
    Value float64
    Line, Col int
}

type StringLiteral struct {
    Value string
    Line, Col int
}

type Identifier struct {
    Name string
    Line, Col int
}

type BinaryOp struct {
    Left  ASTNode
    Op    TokenType
    Right ASTNode
    Line, Col int
}

// Statements
type Assignment struct {
    Name  string
    Value ASTNode
    Line, Col int
}

type PopStatement struct {  // Execute/call
    Function ASTNode
    Args     []ASTNode
    Line, Col int
}

type IfStatement struct {
    Condition ASTNode
    Then      []ASTNode
    Else      []ASTNode
    Line, Col int
}

type LoopStatement struct {  // K'ayab - loop
    Iterable ASTNode
    Body     []ASTNode
    Line, Col int
}

type PackCall struct {
    Pack    string
    Handler string
    Args    map[string]ASTNode
    Line, Col int
}

type Program struct {
    Statements []ASTNode
}
```

### Parser Implementation
```go
type Parser struct {
    tokens []Token
    pos    int
    current Token
}

func (p *Parser) Parse() (*Program, error)
func (p *Parser) ParseStatement() (ASTNode, error)
func (p *Parser) ParseExpression() (ASTNode, error)
```

---

## Phase 3: Semantic Analysis

**Input**: AST
**Output**: Annotated AST + Symbol Table
**Responsibility**: Type checking, scope resolution, semantic validation

### Symbol Table
```go
type Symbol struct {
    Name     string
    Type     DataType
    Scope    *Scope
    Line     int
    Assigned bool
}

type Scope struct {
    Symbols map[string]*Symbol
    Parent  *Scope
}

type SemanticAnalyzer struct {
    ast          *Program
    globalScope  *Scope
    currentScope *Scope
    errors       []CompileError
}
```

### Type System
```go
type DataType string

const (
    TYPE_NUMBER   = "number"
    TYPE_STRING   = "string"
    TYPE_BOOLEAN  = "bool"
    TYPE_ARRAY    = "array"
    TYPE_OBJECT   = "object"
    TYPE_FUNCTION = "function"
    TYPE_ANY      = "any"
)
```

### Semantic Checks
- Variable redeclaration detection
- Type compatibility verification
- Scope resolution
- Dead code detection
- Unreachable statements

---

## Phase 4: Intermediate Representation

**Input**: Annotated AST
**Output**: Intermediate Representation (IR)
**Responsibility**: Lower-level code representation

### IR Instructions
```go
type IRInstruction string

const (
    IR_LOAD     = "LOAD"      // Load value into register
    IR_STORE    = "STORE"     // Store register to variable
    IR_CALL     = "CALL"      // Call function/pack
    IR_JUMP     = "JUMP"      // Unconditional jump
    IR_JMPC     = "JMPC"      // Jump if condition
    IR_RETURN   = "RETURN"    // Return from function
    IR_PACK     = "PACK"      // Pack invocation
    IR_BINOP    = "BINOP"     // Binary operation
    IR_UNOP     = "UNOP"      // Unary operation
    IR_LABEL    = "LABEL"     // Label for jumps
)

type IRCode struct {
    Instructions []IRInstruction
    SymbolTable  map[string]*Symbol
    Functions    map[string]*IRFunction
}

type IRFunction struct {
    Name         string
    Params       []string
    Instructions []IRInstruction
}
```

---

## Phase 5: Optimization

**Input**: IR
**Output**: Optimized IR
**Responsibility**: Performance optimizations

### Optimizations
1. **Dead Code Elimination**: Remove unused variables/statements
2. **Constant Folding**: Evaluate compile-time constants
3. **Common Subexpression Elimination**: Factor out repeated expressions
4. **Loop Unrolling**: Duplicate loop bodies for small loops
5. **Inlining**: Inline simple function calls

---

## Phase 6: Code Generation

**Input**: Optimized IR
**Output**: JavaScript source code
**Responsibility**: Transform IR to target language

### JavaScript Emission
```javascript
// Example K'UHUL compilation output

// Pack system integration
const packs = {
    lam_o: require('./packs/pack_lam_o'),
    scxq2: require('./packs/pack_scxq2'),
    // ...
};

// Generated code from K'UHUL
function $kuhul_main() {
    const $mem = new Map();  // Memory tier
    const $context = {};      // Execution context

    // Variable declarations
    let model = "llama2";
    let prompt = "What is 2+2?";

    // Pack call: lam_o.infer
    const result = packs.lam_o.handleInfer({
        model: model,
        prompt: prompt,
        temperature: 0.7
    });

    // Memory operations (@mem tier)
    $mem.set('result', result);

    return $mem.get('result');
}

// Execution
$kuhul_main().then(console.log);
```

### WASM Emission (Future)
- Generate WebAssembly text format (.wat)
- Compile to binary (.wasm)
- Support for performance-critical sections

---

## Phase 7: Assembly/Linking

**Input**: JavaScript source
**Output**: Executable bytecode or standalone binary
**Responsibility**: Final assembly and linking

### Options
1. **Direct Execution**: Run JavaScript directly via Node.js
2. **WASM Binding**: Link WASM modules to JavaScript
3. **Binary Compilation**: Generate standalone executable
4. **Docker Image**: Package with runtime dependencies

---

## 5-Tier Execution Runtime

The runtime provides 5 tiers of execution capabilities:

### Tier 1: @mem (Memory Management)
```go
type MemoryTier struct {
    heap      map[string]interface{}
    stack     []interface{}
    registers [16]interface{}
}

func (m *MemoryTier) Allocate(name string, value interface{})
func (m *MemoryTier) Load(name string) interface{}
func (m *MemoryTier) Free(name string)
```

**Operations**:
- Variable allocation/deallocation
- Heap management
- Stack frame management

### Tier 2: @call (Function/Handler Invocation)
```go
type CallTier struct {
    functions map[string]CallableFunc
    packs     map[string]Pack
    stack     []*CallFrame
}

type CallFrame struct {
    Function string
    Args     map[string]interface{}
    Locals   map[string]interface{}
    Return   interface{}
}

func (c *CallTier) Call(function string, args map[string]interface{}) (interface{}, error)
func (c *CallTier) RegisterPack(name string, pack Pack)
```

**Operations**:
- Function calls with argument binding
- Pack handler invocation
- Call stack management
- Return value handling

### Tier 3: @ipc (Inter-Process Communication)
```go
type IPCTier struct {
    channels map[string]*Channel
    workers  map[string]*Worker
}

type Channel struct {
    name    string
    queue   []interface{}
    listeners []func(interface{})
}

func (i *IPCTier) CreateChannel(name string)
func (i *IPCTier) Send(channel string, data interface{})
func (i *IPCTier) Listen(channel string, callback func(interface{}))
func (i *IPCTier) SpawnWorker(id string, fn CallableFunc)
```

**Operations**:
- Channel-based communication
- Worker thread spawning
- Message passing
- Event handling

### Tier 4: @db (Database Operations)
```go
type DBTier struct {
    connections map[string]*DBConnection
    transactions map[string]*Transaction
}

type DBConnection struct {
    driver   string
    url      string
    conn     interface{}
}

func (d *DBTier) Connect(name string, driver string, url string) error
func (d *DBTier) Query(conn string, sql string) ([]map[string]interface{}, error)
func (d *DBTier) Execute(conn string, sql string) (int64, error)
func (d *DBTier) BeginTransaction(conn string) (string, error)
```

**Operations**:
- Database connections (SQL, NoSQL)
- Query execution
- Transaction management
- Connection pooling

### Tier 5: @api (API Integration)
```go
type APITier struct {
    clients  map[string]*APIClient
    handlers map[string]APIHandler
    routes   map[string]Route
}

type APIClient struct {
    baseURL string
    headers map[string]string
    client  *http.Client
}

type Route struct {
    Method   string
    Path     string
    Handler  CallableFunc
}

func (a *APITier) RegisterRoute(method string, path string, handler CallableFunc)
func (a *APITier) MakeRequest(clientID string, method string, path string, body interface{}) (interface{}, error)
func (a *APITier) ServeHTTP()
```

**Operations**:
- REST API requests
- Webhook handling
- HTTP server
- Response transformations

---

## Compiler Pipeline Implementation

### Main Compiler Interface
```go
type Compiler struct {
    source    string
    filename  string

    // Pipeline stages
    lexer      *Lexer
    parser     *Parser
    analyzer   *SemanticAnalyzer
    irGen      *IRGenerator
    optimizer  *Optimizer
    codegen    *CodeGenerator
    assembler  *Assembler

    // Output
    ast       *Program
    ir        *IRCode
    javascript string
    bytecode  []byte

    // Error handling
    errors    []CompileError
}

func (c *Compiler) Compile() (interface{}, error) {
    // Phase 1: Lexical Analysis
    tokens, err := c.lexer.Tokenize(c.source)
    if err != nil {
        return nil, err
    }

    // Phase 2: Syntax Analysis
    ast, err := c.parser.Parse(tokens)
    if err != nil {
        return nil, err
    }

    // Phase 3: Semantic Analysis
    annotatedAST, err := c.analyzer.Analyze(ast)
    if err != nil {
        return nil, err
    }

    // Phase 4: IR Generation
    ir, err := c.irGen.Generate(annotatedAST)
    if err != nil {
        return nil, err
    }

    // Phase 5: Optimization
    optimizedIR, err := c.optimizer.Optimize(ir)
    if err != nil {
        return nil, err
    }

    // Phase 6: Code Generation
    javascript, err := c.codegen.Generate(optimizedIR)
    if err != nil {
        return nil, err
    }

    // Phase 7: Assembly
    bytecode, err := c.assembler.Assemble(javascript)
    if err != nil {
        return nil, err
    }

    return bytecode, nil
}
```

---

## Integration Points

### K'UHUL ↔ XJSON
- K'UHUL source code defines pack handlers and logic
- Compiles down to XJSON-compatible execution
- XJSON results fed back to runtime

### Runtime ↔ Packs
- @call tier invokes pack handlers
- Pack handlers return XJSON responses
- Responses integrated into memory tier

### Execution ↔ Ollama
- pack_lam_o handlers call real Ollama API (from Phase 2)
- Model inference results returned to runtime
- Results stored in @mem tier

### End-to-End Flow
```
K'UHUL Script
    ↓
Compiler (7 phases)
    ↓
JavaScript/Bytecode
    ↓
5-Tier Runtime
    ├─ @mem: Store results
    ├─ @call: Invoke pack_lam_o.infer
    ├─ @ipc: Communicate between components
    ├─ @db: Persist results (optional)
    └─ @api: Expose HTTP endpoints (optional)
    ↓
pack_lam_o
    ↓
Ollama API
    ↓
LLM Response
```

---

## Data Flow Example

### Input: K'UHUL Script
```
wo model "llama2"
wo prompt "What is 2+2?"

pop (lam_o.infer
  :model model
  :prompt prompt
  :temperature 0.7)

// Result stored in implicit variable
```

### After Lexing
```
[WO] [SYMBOL:model] [STRING:"llama2"]
[WO] [SYMBOL:prompt] [STRING:"What is 2+2?"]
[POP] [LPAREN] [PACK:lam_o] [GLYPHWORD:infer] ...
```

### After Parsing (AST)
```
Program {
  Assignment { model = "llama2" }
  Assignment { prompt = "What is 2+2?" }
  PopStatement {
    function: PackCall { lam_o.infer }
    args: { model, prompt, temperature: 0.7 }
  }
}
```

### After IR Generation
```
LABEL start
LOAD "llama2" R0
STORE model R0
LOAD "What is 2+2?" R1
STORE prompt R1
PACK lam_o.infer { model, prompt, 0.7 }
STORE result R0
RETURN
```

### After Code Generation
```javascript
function $kuhul() {
    const $mem = new Map();
    $mem.set('model', 'llama2');
    $mem.set('prompt', 'What is 2+2?');

    const result = packs.lam_o.handleInfer({
        model: $mem.get('model'),
        prompt: $mem.get('prompt'),
        temperature: 0.7
    });

    $mem.set('result', result);
    return result;
}
```

---

## Error Handling

### Compile-Time Errors
- **Lexical**: Invalid tokens
- **Syntax**: Malformed expressions
- **Semantic**: Type mismatches, undefined variables

### Runtime Errors
- **Execution**: Failed pack calls
- **API**: Ollama unavailable
- **Resource**: Out of memory

### Error Recovery
```go
type CompileError struct {
    Phase    string  // Which phase failed
    Message  string
    Line     int
    Column   int
    Severity string  // "error" or "warning"
}
```

---

## Testing Strategy

### Unit Tests
- Lexer: Token generation
- Parser: AST construction
- Analyzer: Type checking
- CodeGen: JavaScript emission

### Integration Tests
- Full pipeline: K'UHUL → JavaScript
- Runtime: Execution of generated code
- Pack Integration: Code → Ollama

### Performance Tests
- Compilation time
- Generated code efficiency
- Runtime memory usage

---

## Future Enhancements

1. **WASM Backend**: Emit WebAssembly for performance
2. **JIT Compilation**: Just-in-time optimization
3. **Parallel Execution**: Multi-threaded pack calls
4. **Caching**: Memoize expensive computations
5. **Debugging**: Breakpoints and step execution
6. **LSP Support**: IDE integration
7. **REPL**: Interactive execution environment

