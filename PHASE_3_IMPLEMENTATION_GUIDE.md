# Phase 3: PWA Unification - Implementation Guide

**Phase 3 Goal**: Merge K'UHUL capabilities into the PWA as a single universal app

**Phase 3 Status**: ✅ Ready to start (Phase 1 complete)
**Expected Duration**: 2-3 weeks effort
**Dependencies**: Phase 1 ✅ Complete

---

## Phase 3 Overview

Phase 3 focuses on bringing K'UHUL code execution and intelligent model orchestration into the Progressive Web App. This enables users to write, execute, and debug K'UHUL programs directly in the browser, with full offline support.

### Key Capabilities to Implement:
1. **Offline K'UHUL Runtime**: Execute K'UHUL code in Service Worker
2. **K'UHUL IDE**: Code editor with syntax highlighting and real-time validation
3. **Model Orchestration UI**: Visual pipeline builder for models
4. **Offline Llama Support**: Progressive model loading and caching

---

## 3.1 Enhanced Service Worker (`pwa/sw.js`)

### Current Status
- ✅ Basic Service Worker with XJSON support
- ✅ Caching strategy (cache-first for static, network-first for API)
- ✅ Fallback routing (orchestrator → direct Ollama)
- ⏳ K'UHUL code execution in SW (needed)
- ⏳ Pack caching/lazy loading (needed)
- ⏳ SCXQ2 cache verification (needed)

### 3.1.1 K'UHUL Code Execution in Service Worker

**Task**: Enable the Service Worker to execute K'UHUL programs offline

**Implementation Steps**:
1. Import K'UHUL runtime from `pwa/lib/khl-runtime.js`
2. Create handler for `/api/kuhul/execute` requests in SW
3. Parse K'UHUL code and execute in SW context
4. Send results back to main thread
5. Cache successful executions for offline replay

**New Handler Pattern**:
```javascript
// In sw.js fetch event
if (isKuhulExecuteRequest(event.request)) {
  event.respondWith(handleKuhulExecuteOffline(event.request));
}

async function handleKuhulExecuteOffline(request) {
  const { source, mode } = await request.json();
  const result = KhlRuntime.execute(source, mode);
  return new Response(JSON.stringify({ ok: true, result }), ...);
}
```

**Files to Modify**:
- `pwa/sw.js`: Add Kuhul execution handler
- `pwa/index.html`: Link K'UHUL libraries in SW scope

### 3.1.2 Pack Caching and Lazy Loading

**Task**: Cache K'UHUL pack code for offline access

**Implementation Steps**:
1. Create pack index structure
2. Implement lazy-load mechanism for pack code
3. Add to STATIC_ASSETS for initial caching
4. Implement on-demand pack fetching

**Pack Cache Structure**:
```javascript
const PACK_CACHE_NAME = 'kuhul-packs-v1';
const PACK_INDEX = {
  'pack_lam_o': { url: '/packs/pack_lam_o.khl', version: '1.0.0' },
  'pack_scxq2': { url: '/packs/pack_scxq2.khl', version: '1.0.0' },
  // ... more packs
};
```

### 3.1.3 SCXQ2 Cache Verification

**Task**: Verify cached responses match expected SCXQ2 fingerprints

**Implementation Steps**:
1. Read SCXQ2 fingerprint from cache metadata
2. Re-compute fingerprint for cached response
3. Verify match before serving from cache
4. Update cache if fingerprint mismatch

---

## 3.2 K'UHUL IDE in PWA (`pwa/lib/kuhul-ide.js`)

### Current Status
- ❌ K'UHUL IDE not yet created (needed)
- ✅ Syntax highlighting libraries available (`pwa/lib/`)
- ✅ Parser available (`pwa/lib/khl-parser.js`)
- ✅ Runtime available (`pwa/lib/khl-runtime.js`)

### 3.2.1 Create K'UHUL IDE Module

**New File**: `pwa/lib/kuhul-ide.js`

**Core Components**:

#### Editor Setup
```javascript
class KuhulIDE {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
    this.editor = this.initializeEditor();
    this.parser = KhlParser;
    this.runtime = KhlRuntime;
  }

  initializeEditor() {
    // Initialize CodeMirror or similar
    // Set up K'UHUL mode with glyphs
  }
}
```

#### Syntax Highlighting
- Custom mode for K'UHUL glyphs: `⟁Pop⟁`, `⟁Wo⟁`, etc.
- Keyword highlighting: `function`, `variable`, `control`, etc.
- Comment highlighting
- String/number/JSON highlighting

#### Real-Time Parsing
```javascript
onCodeChange(source) {
  const { ast, errors } = KhlParser.parse(source);
  this.displayErrors(errors);
  this.updateOutline(ast);
}
```

#### Pack Explorer
```javascript
showPackExplorer() {
  const packs = PackRegistry.list();
  // Display available packs
  // Show pack details (name, version, description)
  // Allow pack selection for composition
}
```

#### Execution Visualization
```javascript
visualizeExecution(result) {
  // Show ABR phases (Atomic Block Representation)
  // Timeline view of execution
  // Variable state inspection
}
```

#### SCXQ2 Inspector
```javascript
showFingerprintInspector() {
  const fp = SCXQ2.fingerprint(result);
  // Display fingerprint hash
  // Show what data went into fingerprint
  // Compare with cached fingerprints
}
```

### 3.2.2 IDE Integration into PWA

**File**: `pwa/index.html`

**Changes**:
1. Add IDE container div
2. Load `kuhul-ide.js` library
3. Initialize IDE on page load
4. Add IDE controls (run, clear, save, load)
5. Connect IDE to existing UI components

**UI Layout**:
```
┌─────────────────────────────────────────────────┐
│          K'UHUL IDE - Universal App            │
├─────────────────┬───────────────────────────────┤
│  Editor (Left)  │  Output Panel (Right)         │
│                 │                               │
│  ⟁Wo⟁ x = 10   │  Execution Results:           │
│  ⟁Wo⟁ y = 20   │  x: 10                       │
│  ⟁Ch'en⟁        │  y: 20                       │
│  {x, y}         │                               │
│                 │  Pack Explorer |SCXQ2 Insp   │
└─────────────────┴───────────────────────────────┘
```

---

## 3.3 Model Orchestration UI (`pwa/lib/orchestrator-ui.js`)

### Current Status
- ❌ Model Orchestration UI not yet created (needed)
- ✅ Model management infrastructure exists
- ✅ Pack system foundation complete

### 3.3.1 Visual Pipeline Builder

**New File**: `pwa/lib/orchestrator-ui.js`

**Components**:

#### Model Nodes
```javascript
class ModelNode {
  constructor(modelName, position) {
    this.model = modelName;
    this.position = position;
    this.inputs = [];
    this.outputs = [];
  }

  render(canvas) {
    // Draw rectangle representing model
    // Show input/output ports
    // Handle drag for repositioning
  }
}
```

#### Connection System
```javascript
connectModels(sourceNode, targetNode) {
  // Create data flow connection
  // Validate model compatibility
  // Store connection metadata
}
```

#### Drag-and-Drop
```javascript
enableDragDrop() {
  // Drag models from palette onto canvas
  // Drop creates new model node
  // Connect existing nodes
}
```

#### Pipeline Execution
```javascript
executePipeline() {
  // Traverse nodes in topological order
  // Execute models in sequence/parallel
  // Stream results between stages
  // Update UI with execution progress
}
```

### 3.3.2 Real-Time Monitoring

**Features**:
- Live execution progress indicator
- Token count tracking
- Latency visualization
- Model load percentage
- Cache hit rate display

### 3.3.3 Multi-Model Routing

**Advanced Feature** (Low Priority):
- Route inputs to multiple models
- Compare model outputs
- Load balancing visualization
- Fallback model selection

---

## 3.4 Offline Llama Support (`pwa/lib/offline-llama.js`)

### Current Status
- ❌ Offline Llama module not yet created (needed)
- ⏳ WebAssembly llama.cpp port (feasibility study needed)
- ⏳ IndexedDB storage (infrastructure ready)
- ⏳ Progressive loading (pattern to implement)

### 3.4.1 IndexedDB Model Weight Storage

**New File**: `pwa/lib/offline-llama.js`

**Database Schema**:
```javascript
const DB_NAME = 'kuhul-models';
const DB_VERSION = 1;

const OBJECT_STORES = {
  models: { keyPath: 'id' }, // Model metadata
  weights: { keyPath: 'id' }, // Model weight tensors
  cache: { keyPath: 'hash' }, // Inference cache
  config: { keyPath: 'key' }  // Configuration
};
```

**Storage Implementation**:
```javascript
class OfflineLlamaDB {
  async storeModelWeights(modelName, weights) {
    const db = await this.openDB();
    const tx = db.transaction('weights', 'readwrite');
    return tx.objectStore('weights').add({
      modelName,
      data: weights,
      timestamp: Date.now()
    });
  }

  async retrieveModelWeights(modelName) {
    const db = await this.openDB();
    const tx = db.transaction('weights', 'readonly');
    return tx.objectStore('weights').get(modelName);
  }
}
```

### 3.4.2 Progressive Model Loading

**Implementation Steps**:
1. Split model into chunks (layers)
2. Load chunks on-demand
3. Cache frequently used chunks
4. Stream chunks to Service Worker
5. Progressively assemble model

**UI Progress Display**:
```javascript
showLoadingProgress(modelName, percent) {
  // Progress bar showing model load
  // ETA calculation
  // Pause/resume capability
  // Network status indicator
}
```

### 3.4.3 Ollama Fallback

**Graceful Degradation**:
```javascript
async inference(prompt, model) {
  try {
    // Try offline Llama
    return await offlineLlama.infer(prompt, model);
  } catch (e) {
    // Fallback to Ollama
    return await ollamaFallback.infer(prompt, model);
  }
}
```

**Detection Logic**:
```javascript
shouldUseOfflineMode() {
  return (
    navigator.onLine === false ||
    model.isDownloadedOffline === true ||
    userPreference === 'offline'
  );
}
```

---

## Implementation Priority

### High Priority (Start First):
1. **3.1.1** - K'UHUL execution in Service Worker
2. **3.2.1** - K'UHUL IDE module creation
3. **3.3.1** - Visual pipeline builder

### Medium Priority (After High):
4. **3.1.2** - Pack caching and lazy loading
5. **3.4.1** - IndexedDB model storage
6. **3.4.2** - Progressive model loading

### Low Priority (Polish/Polish):
7. **3.1.3** - SCXQ2 cache verification
8. **3.3.3** - Multi-model routing
9. **3.4.3** - Advanced fallback strategies

---

## Testing Strategy

### Unit Tests
- K'UHUL IDE parsing and validation
- Pack caching and retrieval
- Model orchestration graph operations
- Offline storage read/write operations

### Integration Tests
- End-to-end K'UHUL code execution in IDE
- Service Worker handling of K'UHUL requests
- Model pipeline execution
- Offline vs. online switching

### E2E Tests
- User writes K'UHUL code → executes in IDE
- User builds model pipeline → executes → shows results
- Model weights download → offline execution → cache hit
- Service Worker fallback on network failure

---

## Success Criteria for Phase 3

### 3.1 Complete When:
- ✅ K'UHUL code execution works offline in Service Worker
- ✅ Pack caching reduces load time by 50%+
- ✅ SCXQ2 verification prevents stale cache hits

### 3.2 Complete When:
- ✅ Users can write K'UHUL in IDE with syntax highlighting
- ✅ Real-time parsing provides error feedback
- ✅ Pack explorer shows available packs
- ✅ SCXQ2 inspector works correctly

### 3.3 Complete When:
- ✅ Visual pipeline builder creates valid model graphs
- ✅ Drag-and-drop creates model nodes
- ✅ Real-time monitoring shows execution progress

### 3.4 Complete When:
- ✅ Model weights load and cache in IndexedDB
- ✅ Progressive loading shows user progress
- ✅ Fallback to Ollama works seamlessly

---

## File Checklist

**New Files to Create**:
- [ ] `pwa/lib/kuhul-ide.js` - K'UHUL IDE module
- [ ] `pwa/lib/orchestrator-ui.js` - Model orchestration UI
- [ ] `pwa/lib/offline-llama.js` - Offline Llama support

**Files to Modify**:
- [ ] `pwa/sw.js` - Add K'UHUL execution handlers
- [ ] `pwa/index.html` - Add IDE container and load new libraries

**New Tests**:
- [ ] `pwa/tests/kuhul-ide.test.js` - IDE tests
- [ ] `pwa/tests/orchestrator-ui.test.js` - Orchestration tests
- [ ] `pwa/tests/offline-llama.test.js` - Offline storage tests

---

## Next Steps

1. **Review** this guide and the Phase 1 completion report
2. **Choose** which 3.x task to start first (recommend 3.1.1)
3. **Create** feature branch: `git checkout -b phase-3/<feature-name>`
4. **Implement** incrementally with tests
5. **Commit** frequently with clear messages
6. **Push** to remote for review

---

## Additional Resources

- Phase 1 Completion Report: `PHASE_1_COMPLETION_REPORT.md`
- PROGRESS_PHASES.md: Overall project roadmap
- TODO.md: Task tracking
- PWA README: `pwa/README.md`
- K'UHUL Docs: Check documentation files in `docs/`

---

**Ready to begin Phase 3? Start with 3.1.1 or choose another priority!**
