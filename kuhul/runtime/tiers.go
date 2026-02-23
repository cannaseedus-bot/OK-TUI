// Package runtime provides the 5-tier execution environment for K'UHUL compiled code.
//
// The runtime consists of 5 tiers:
// 1. @mem - Memory management (heap, stack, registers)
// 2. @call - Function/handler invocation with call stack
// 3. @ipc - Inter-process communication and channels
// 4. @db - Database operations and connection management
// 5. @api - REST API integration and webhooks
package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExecutionContext holds the complete execution environment
type ExecutionContext struct {
	// Tiers
	MemTier  *MemoryTier
	CallTier *CallTier
	IPCTier  *IPCTier
	DBTier   *DatabaseTier
	APIMTier *APITier

	// Global state
	Variables map[string]interface{}
	Packs     map[string]interface{} // Pack handlers
	Context   context.Context
	Cancel    context.CancelFunc

	// Error handling
	Errors []ExecutionError
	mu     sync.RWMutex
}

// ExecutionError represents an error during execution
type ExecutionError struct {
	Tier      string
	Message   string
	Timestamp time.Time
	Context   string
}

// NewExecutionContext creates a new execution context
func NewExecutionContext() *ExecutionContext {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	return &ExecutionContext{
		MemTier:   NewMemoryTier(),
		CallTier:  NewCallTier(),
		IPCTier:   NewIPCTier(),
		DBTier:    NewDatabaseTier(),
		APIMTier:  NewAPITier(),
		Variables: make(map[string]interface{}),
		Packs:     make(map[string]interface{}),
		Context:   ctx,
		Cancel:    cancel,
		Errors:    []ExecutionError{},
	}
}

// AddError records an execution error
func (ec *ExecutionContext) AddError(tier, message, context string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Errors = append(ec.Errors, ExecutionError{
		Tier:      tier,
		Message:   message,
		Timestamp: time.Now(),
		Context:   context,
	})
}

// GetErrors returns all recorded errors
func (ec *ExecutionContext) GetErrors() []ExecutionError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return append([]ExecutionError{}, ec.Errors...)
}

// Cleanup releases resources
func (ec *ExecutionContext) Cleanup() {
	ec.Cancel()
	if ec.DBTier != nil {
		ec.DBTier.Cleanup()
	}
	if ec.IPCTier != nil {
		ec.IPCTier.Cleanup()
	}
}

// ═══════════════════════════════════════════════════════════════════════════════

// TIER 1: Memory Management (@mem)

// MemoryTier manages heap, stack, and registers
type MemoryTier struct {
	heap      map[string]interface{}
	stack     []interface{}
	registers [16]interface{}
	mu        sync.RWMutex
}

// NewMemoryTier creates a new memory tier
func NewMemoryTier() *MemoryTier {
	return &MemoryTier{
		heap:  make(map[string]interface{}),
		stack: make([]interface{}, 0),
	}
}

// Set allocates/updates a variable
func (m *MemoryTier) Set(name string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.heap[name] = value
	return nil
}

// Get retrieves a variable
func (m *MemoryTier) Get(name string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, exists := m.heap[name]
	if !exists {
		return nil, fmt.Errorf("variable %q not found", name)
	}

	return val, nil
}

// Delete removes a variable
func (m *MemoryTier) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.heap, name)
	return nil
}

// PushStack pushes a value onto the stack
func (m *MemoryTier) PushStack(value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stack = append(m.stack, value)
}

// PopStack pops a value from the stack
func (m *MemoryTier) PopStack() (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.stack) == 0 {
		return nil, fmt.Errorf("stack underflow")
	}

	val := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	return val, nil
}

// SetRegister sets a register value
func (m *MemoryTier) SetRegister(index int, value interface{}) error {
	if index < 0 || index >= 16 {
		return fmt.Errorf("register index out of bounds: %d", index)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.registers[index] = value
	return nil
}

// GetRegister gets a register value
func (m *MemoryTier) GetRegister(index int) (interface{}, error) {
	if index < 0 || index >= 16 {
		return nil, fmt.Errorf("register index out of bounds: %d", index)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.registers[index], nil
}

// GetHeapSize returns the number of variables in heap
func (m *MemoryTier) GetHeapSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.heap)
}

// ═══════════════════════════════════════════════════════════════════════════════

// TIER 2: Function Invocation (@call)

// CallableFunc is a callable function type
type CallableFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// CallFrame represents a function call frame
type CallFrame struct {
	FunctionName string
	Args         map[string]interface{}
	Locals       map[string]interface{}
	ReturnValue  interface{}
	ReturnError  error
	Timestamp    time.Time
}

// CallTier manages function invocation and call stack
type CallTier struct {
	functions map[string]CallableFunc
	stack     []*CallFrame
	maxDepth  int
	mu        sync.RWMutex
}

// NewCallTier creates a new call tier
func NewCallTier() *CallTier {
	return &CallTier{
		functions: make(map[string]CallableFunc),
		stack:     make([]*CallFrame, 0),
		maxDepth:  1000,
	}
}

// RegisterFunction registers a callable function
func (c *CallTier) RegisterFunction(name string, fn CallableFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.functions[name] = fn
}

// Call invokes a function
func (c *CallTier) Call(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	c.mu.RLock()
	fn, exists := c.functions[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("function %q not found", name)
	}

	// Check call stack depth
	c.mu.RLock()
	if len(c.stack) >= c.maxDepth {
		c.mu.RUnlock()
		return nil, fmt.Errorf("call stack overflow: max depth %d exceeded", c.maxDepth)
	}
	c.mu.RUnlock()

	// Create call frame
	frame := &CallFrame{
		FunctionName: name,
		Args:         args,
		Locals:       make(map[string]interface{}),
		Timestamp:    time.Now(),
	}

	// Push frame
	c.mu.Lock()
	c.stack = append(c.stack, frame)
	c.mu.Unlock()

	// Execute function
	result, err := fn(ctx, args)

	// Pop frame
	c.mu.Lock()
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
	c.mu.Unlock()

	return result, err
}

// GetCallStack returns the current call stack
func (c *CallTier) GetCallStack() []*CallFrame {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]*CallFrame{}, c.stack...)
}

// GetCallDepth returns the current call depth
func (c *CallTier) GetCallDepth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.stack)
}

// ═══════════════════════════════════════════════════════════════════════════════

// TIER 3: Inter-Process Communication (@ipc)

// Channel represents a communication channel
type Channel struct {
	Name      string
	Queue     []interface{}
	Listeners []func(interface{})
	mu        sync.RWMutex
	closed    bool
}

// Send sends a message on the channel
func (ch *Channel) Send(data interface{}) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return fmt.Errorf("channel %q is closed", ch.Name)
	}

	ch.Queue = append(ch.Queue, data)

	// Notify listeners
	for _, listener := range ch.Listeners {
		go listener(data)
	}

	return nil
}

// Receive receives a message from the channel
func (ch *Channel) Receive() (interface{}, error) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if len(ch.Queue) == 0 {
		return nil, fmt.Errorf("channel %q is empty", ch.Name)
	}

	msg := ch.Queue[0]
	ch.Queue = ch.Queue[1:]
	return msg, nil
}

// Subscribe adds a listener to the channel
func (ch *Channel) Subscribe(listener func(interface{})) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.Listeners = append(ch.Listeners, listener)
}

// Close closes the channel
func (ch *Channel) Close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.closed = true
	return nil
}

// IPCTier manages inter-process communication
type IPCTier struct {
	channels map[string]*Channel
	workers  map[string]context.CancelFunc
	mu       sync.RWMutex
}

// NewIPCTier creates a new IPC tier
func NewIPCTier() *IPCTier {
	return &IPCTier{
		channels: make(map[string]*Channel),
		workers:  make(map[string]context.CancelFunc),
	}
}

// CreateChannel creates a new channel
func (ipc *IPCTier) CreateChannel(name string) (*Channel, error) {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if _, exists := ipc.channels[name]; exists {
		return nil, fmt.Errorf("channel %q already exists", name)
	}

	ch := &Channel{
		Name:  name,
		Queue: make([]interface{}, 0),
	}

	ipc.channels[name] = ch
	return ch, nil
}

// GetChannel retrieves a channel
func (ipc *IPCTier) GetChannel(name string) (*Channel, error) {
	ipc.mu.RLock()
	defer ipc.mu.RUnlock()

	ch, exists := ipc.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %q not found", name)
	}

	return ch, nil
}

// Send sends a message on a channel
func (ipc *IPCTier) Send(channelName string, data interface{}) error {
	ch, err := ipc.GetChannel(channelName)
	if err != nil {
		return err
	}

	return ch.Send(data)
}

// Receive receives a message from a channel
func (ipc *IPCTier) Receive(channelName string) (interface{}, error) {
	ch, err := ipc.GetChannel(channelName)
	if err != nil {
		return nil, err
	}

	return ch.Receive()
}

// Cleanup closes all channels and cancels workers
func (ipc *IPCTier) Cleanup() {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	// Cancel all workers
	for _, cancel := range ipc.workers {
		cancel()
	}

	// Close all channels
	for _, ch := range ipc.channels {
		_ = ch.Close()
	}

	ipc.channels = make(map[string]*Channel)
	ipc.workers = make(map[string]context.CancelFunc)
}

// ═══════════════════════════════════════════════════════════════════════════════

// TIER 4: Database Operations (@db)

// DBConnection represents a database connection
type DBConnection struct {
	Name    string
	Driver  string
	URL     string
	Connected bool
}

// DatabaseTier manages database connections and operations
type DatabaseTier struct {
	connections map[string]*DBConnection
	mu          sync.RWMutex
}

// NewDatabaseTier creates a new database tier
func NewDatabaseTier() *DatabaseTier {
	return &DatabaseTier{
		connections: make(map[string]*DBConnection),
	}
}

// Connect establishes a database connection
func (db *DatabaseTier) Connect(name, driver, url string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.connections[name]; exists {
		return fmt.Errorf("connection %q already exists", name)
	}

	// For now, just record the connection
	// In production, would actually connect to database
	db.connections[name] = &DBConnection{
		Name:      name,
		Driver:    driver,
		URL:       url,
		Connected: true,
	}

	return nil
}

// Disconnect closes a database connection
func (db *DatabaseTier) Disconnect(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	conn, exists := db.connections[name]
	if !exists {
		return fmt.Errorf("connection %q not found", name)
	}

	conn.Connected = false
	return nil
}

// Query executes a query
func (db *DatabaseTier) Query(connName, sql string) ([]map[string]interface{}, error) {
	db.mu.RLock()
	conn, exists := db.connections[connName]
	db.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection %q not found", connName)
	}

	if !conn.Connected {
		return nil, fmt.Errorf("connection %q is closed", connName)
	}

	// Return mock results for now
	return []map[string]interface{}{}, nil
}

// Execute executes a statement
func (db *DatabaseTier) Execute(connName, sql string) (int64, error) {
	db.mu.RLock()
	conn, exists := db.connections[connName]
	db.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("connection %q not found", connName)
	}

	if !conn.Connected {
		return 0, fmt.Errorf("connection %q is closed", connName)
	}

	// Return mock result for now
	return 1, nil
}

// Cleanup closes all connections
func (db *DatabaseTier) Cleanup() {
	db.mu.Lock()
	defer db.mu.Unlock()

	for name := range db.connections {
		db.connections[name].Connected = false
	}
}

// ═══════════════════════════════════════════════════════════════════════════════

// TIER 5: API Integration (@api)

// Route represents an HTTP route
type Route struct {
	Method  string
	Path    string
	Handler CallableFunc
}

// APITier manages REST API integration
type APITier struct {
	routes  map[string]*Route
	clients map[string]interface{} // API clients
	mu      sync.RWMutex
}

// NewAPITier creates a new API tier
func NewAPITier() *APITier {
	return &APITier{
		routes:  make(map[string]*Route),
		clients: make(map[string]interface{}),
	}
}

// RegisterRoute registers an HTTP route
func (api *APITier) RegisterRoute(method, path string, handler CallableFunc) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	key := fmt.Sprintf("%s %s", method, path)

	if _, exists := api.routes[key]; exists {
		return fmt.Errorf("route %q already exists", key)
	}

	api.routes[key] = &Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}

	return nil
}

// GetRoute retrieves a route
func (api *APITier) GetRoute(method, path string) (*Route, error) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	key := fmt.Sprintf("%s %s", method, path)
	route, exists := api.routes[key]

	if !exists {
		return nil, fmt.Errorf("route %q not found", key)
	}

	return route, nil
}

// ListRoutes returns all registered routes
func (api *APITier) ListRoutes() []*Route {
	api.mu.RLock()
	defer api.mu.RUnlock()

	routes := make([]*Route, 0)
	for _, route := range api.routes {
		routes = append(routes, route)
	}

	return routes
}

// RegisterClient registers an API client
func (api *APITier) RegisterClient(name string, client interface{}) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.clients[name] = client
}

// GetClient retrieves an API client
func (api *APITier) GetClient(name string) (interface{}, error) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	client, exists := api.clients[name]
	if !exists {
		return nil, fmt.Errorf("client %q not found", name)
	}

	return client, nil
}
