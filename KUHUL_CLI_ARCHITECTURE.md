# KUHUL CLI TUI Architecture

**KUHUL Code-Generation IDE** - Terminal-based intelligent code editor with agent orchestration, file awareness, and crash recovery.

---

## 🎯 Vision

```
Developer → CLI TUI → Project Context → Agent System → Code Generation
                ↓           ↓              ↓              ↓
            Chat/Plan   File Aware    Multi-Agent    KHANARY Experts
            Interface   Validation    Orchestration   (Python, JS, React, etc.)
```

**NOT** a chat interface. **IS** a code IDE where conversations drive edits.

---

## 🏗️ Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│              KUHUL CLI TUI (PowerShell)                 │
├─────────────────────────────────────────────────────────┤
│  UI Layer                                               │
│  ├─ Main viewport (files being edited)                 │
│  ├─ Chat panel (code generation prompts)               │
│  ├─ File tree (project structure)                      │
│  ├─ Status bar (current file, model, agent)            │
│  └─ Debug panel (agent actions, errors)                │
├─────────────────────────────────────────────────────────┤
│  Context Layer                                          │
│  ├─ Project awareness (.kuhul/project.json)            │
│  ├─ File dependency tracking                           │
│  ├─ Environment detection (Node, Python, .NET, etc.)   │
│  ├─ Build/test configuration                          │
│  └─ Constraint enforcement (strict rules)              │
├─────────────────────────────────────────────────────────┤
│  Agent Layer                                            │
│  ├─ Agent manager (spawning, lifecycle)                │
│  ├─ Tool registry (installed tools)                    │
│  ├─ Agent scaffolder (creates agents on demand)        │
│  ├─ KHANARY router (selects expert)                    │
│  └─ Agent debugger (trace, breakpoints)                │
├─────────────────────────────────────────────────────────┤
│  Execution Layer                                        │
│  ├─ File operations (read, write, validate)            │
│  ├─ Code generation pipeline                           │
│  ├─ Testing & validation                               │
│  ├─ Installation/package management                    │
│  └─ Crash recovery & session resume                    │
├─────────────────────────────────────────────────────────┤
│  Storage Layer                                          │
│  ├─ .kuhul/ (config, agents, models, sessions)         │
│  ├─ Session history (resumable)                        │
│  ├─ Agent definitions                                  │
│  ├─ Model registry                                     │
│  └─ Edit history & undo/redo                           │
├─────────────────────────────────────────────────────────┤
│  KHANARY Expert System (40+ Specialists)               │
│  ├─ Python, JavaScript, React, FastAPI, Security...   │
│  ├─ Deterministic outputs                              │
│  └─ Accurate code generation per domain                │
└─────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
.kuhul/
├─ project.json              # Project metadata & constraints
├─ environment.json          # Detected environment (Node, Python, etc.)
├─ rules.json               # Strict policy enforcement
│
├─ agents/
│  ├─ registry.json         # Installed agents
│  ├─ python-coder/
│  │  ├─ manifest.json
│  │  ├─ tools.json        # Available tools
│  │  └─ capabilities.json
│  ├─ react-designer/
│  └─ debugger/
│
├─ models/
│  ├─ registry.json         # Available KHANARY experts
│  └─ sessions/
│
└─ sessions/
   ├─ current.session       # Active session state
   ├─ history.log          # All prompts/edits
   └─ backups/             # Crash recovery
```

---

## 🎮 UI Components (PowerShell TUI)

### Main Screen Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ KUHUL IDE | Project: webapp | Env: Node | Model: javascript    │ ← Status Bar
├──────────────────┬──────────────────────────────────────────────┤
│ File Tree        │  MAIN EDITOR VIEWPORT                        │
│                  │  ─────────────────────────────────────────   │
│ 📁 src/          │  src/components/Button.jsx                   │
│  📄 components   │  ┌─────────────────────────────────────────┐ │
│   🔵 Button.jsx  │  │ 1  import React from 'react'            │ │
│   📄 Modal.jsx   │  │ 2  import './Button.css'                │ │
│  📄 services.js  │  │ 3                                        │ │
│ 📁 api/          │  │ 4  export const Button = ({ onClick })  │ │
│  📄 client.js    │  │ 5    => (                                │ │
│ 📄 index.js      │  │ 6    <button onClick={onClick}>         │ │
│ 📄 package.json  │  │ 7      Click me                         │ │
│                  │  │ 8    </button>                           │ │
│ [New File]       │  │ 9    )                                   │ │
│ [New Folder]     │  └─────────────────────────────────────────┘ │
├──────────────────┼──────────────────────────────────────────────┤
│                  │ CHAT PANEL (Code Generation)                 │
│ CONTEXT:         │ ┌─────────────────────────────────────────┐ │
│ Environment:Node │ │ > Add loading state to Button          │ │
│ Current:Button   │ │                                         │ │
│ Files:1 opened   │ │ Agent[React] planning...                │ │
│ Agents:2 active  │ │ ✓ Will update Button.jsx               │ │
│                  │ │ ✓ Will add loading.css                 │ │
│                  │ │ ? Install react-spinners? (Y/n)        │ │
│                  │ │                                         │ │
│                  │ │ > █ [Type prompt or 'help']             │ │
│                  │ └─────────────────────────────────────────┘ │
├──────────────────┴──────────────────────────────────────────────┤
│ [Debug] python-coder: Generating code | [Agents] 2 | [×] Errors │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🤖 Agent System

### Agent Lifecycle

```
User Prompt
    ↓
[Route to Expert] (KHANARY selector)
    ├─ "Add button click handler" → React Expert
    ├─ "Secure the API" → Security Expert
    └─ "Debug memory leak" → JavaScript Expert
    ↓
[Load/Create Agent]
    ├─ Check if agent exists
    ├─ Load from .kuhul/agents/
    └─ Or create from template
    ↓
[Gather Context]
    ├─ Read affected files
    ├─ Parse dependencies
    ├─ Check build config
    └─ Load constraints
    ↓
[Generate Code]
    ├─ Call KHANARY expert
    ├─ Stream planning
    ├─ Request approval
    └─ Generate edits
    ↓
[Apply Changes]
    ├─ Validate files
    ├─ Check folder structure
    ├─ Run tests
    └─ Update project.json
    ↓
[Resume/Crash Recovery]
    └─ Save state to session history
```

### Agent Types

```
Core Agents (Built-in):
├─ code-generator      # Main code generation agent
├─ debugger            # Debug code, trace execution
├─ file-manager        # Create, move, delete files
├─ installer           # Install packages/tools
├─ model-manager       # Manage KHANARY experts
└─ crash-recovery      # Restore previous sessions

Expert Agents (Per-Domain):
├─ python-coder        # Python development
├─ javascript-coder    # JavaScript/Node.js
├─ react-designer      # React components
├─ fastapi-builder     # FastAPI APIs
├─ security-auditor    # Security review
└─ ... (40+ KHANARY experts available)

On-Demand Agents (Created if needed):
├─ [any KHANARY expert]  # Created from template
├─ [custom tool]         # If requested by LLM
└─ [specialized agent]   # For specific tasks
```

---

## 💬 Chat/Prompt System

### NOT General Chat - Code Generation Prompts

```
User Input Types Allowed:
✅ "Add button to the header component"
✅ "Secure the login endpoint"
✅ "Create a loading spinner"
✅ "Fix the memory leak in this function"
✅ "Deploy to AWS"

Not Allowed:
❌ "What is machine learning?"
❌ "Tell me a joke"
❌ "How do I write a resume?"
```

### Prompt Flow

```
User: "Add loading state to Button component"
        ↓
[Strict Validation]
  ✓ Is this a code generation task?
  ✓ Are we in a project?
  ✓ What files will change?
        ↓
[Environment Detection]
  ✓ Project: React app
  ✓ Framework: Next.js
  ✓ Files: src/components/Button.jsx
        ↓
[Expert Selection]
  ✓ Route to: React Expert (KHANARY)
        ↓
[Planning Phase]
  Agent responds with:
  - What it will do
  - What files it will edit
  - What dependencies it needs
  - If approvals/installations needed
        ↓
[Approval] User: "OK"
        ↓
[Execution]
  - Generate code
  - Apply edits
  - Show diffs
  - Run tests
        ↓
[Verification]
  ✓ Files valid
  ✓ No duplicates
  ✓ Folder structure intact
  ✓ Tests pass
```

---

## 📂 File & Folder Awareness

### Strict Rules

```json
{
  "file_operations": {
    "no_duplicates": true,
    "no_orphans": true,
    "validate_structure": true,
    "require_folder_context": true
  },
  "edit_policy": {
    "single_file_edits": "forbidden",
    "must_be_part_of_cohesive_change": true,
    "related_files_must_be_updated": true
  },
  "validation": {
    "check_imports": true,
    "validate_references": true,
    "enforce_naming": true,
    "prevent_dead_code": true
  }
}
```

### File Awareness Example

```
❌ NOT ALLOWED:
  User: "Update Button.jsx"
  Agent generates: Button.jsx only
  Reason: Button is imported in 5 places
          Must update those files too

✅ ALLOWED:
  User: "Add loading state to Button component"
  Agent generates:
    ├─ src/components/Button.jsx (add prop)
    ├─ src/components/Button.module.css (add styles)
    ├─ src/pages/LoginPage.jsx (pass loading prop)
    ├─ src/hooks/useAuth.js (add loading state)
    └─ package.json (add react-spinners if needed)
  Validation:
    ✓ All imports updated
    ✓ All usages updated
    ✓ Types match
    ✓ No orphaned code
```

---

## 💾 Session Resume & Crash Recovery

### Session State

```
.kuhul/sessions/current.session
├─ timestamp: "2024-11-15T10:30:00Z"
├─ project_path: "/home/user/webapp"
├─ environment: "node"
├─ open_files: ["Button.jsx", "services.js"]
├─ active_agent: "react-designer"
├─ context: {
│   "last_prompt": "Add loading state to Button",
│   "planning": [...],
│   "in_progress": true
│ }
├─ changes: [
│   {
│     "file": "Button.jsx",
│     "before": "...",
│     "after": "...",
│     "applied": true
│   }
│ ]
└─ backup_files: [...]
```

### Crash Recovery Flow

```
User restarts CLI
    ↓
Detect crashed session
    ↓
Load .kuhul/sessions/current.session
    ↓
Show recovery options:
  1. Resume previous session
  2. Review what changed
  3. Continue editing
  4. Start new session
    ↓
User selects: "Resume"
    ↓
Restore state:
  ✓ Open same files
  ✓ Load active agent
  ✓ Show planning from before crash
  ✓ Continue code generation
    ↓
User can:
  - Approve/deny pending changes
  - Continue from where they left off
  - Undo if something was wrong
```

---

## 🔧 Model Manager

### KHANARY Expert Selection

```
Available Models: 40+

By Task:
├─ Code Generation
│  ├─ python (92-95%)
│  ├─ javascript (91-94%)
│  ├─ react (91-94%)
│  └─ fastapi (90-93%)
│
├─ Debugging
│  ├─ debugger (expert)
│  ├─ security (88-94%)
│  └─ performance (86-91%)
│
└─ Infrastructure
   ├─ devops (88-92%)
   ├─ aws (88-92%)
   └─ docker (expert)

User can:
- Select specific expert for task
- Switch models mid-session
- Compare outputs from different models
- Use different experts for different files
```

### Model Status

```
🟢 python.khμ        (14MB, loaded, 92% acc)
🟢 javascript.khμ    (14MB, loaded, 91% acc)
🟡 react.khμ         (14MB, ready, 91% acc)
⚫ fastapi.khμ       (14MB, not loaded)
❌ security.khμ      (15MB, error: corrupted)

Commands:
  load <model>       # Load model into memory
  unload <model>     # Free memory
  check <model>      # Verify integrity
  install <model>    # Download from HF Hub
```

---

## 🐛 Agent Debugger

### Debug Mode

```
> debug on

Now tracking:
├─ Agent decisions
├─ Tool calls
├─ File operations
├─ KHANARY prompts
├─ Generated code
└─ Validation results

Breakpoints:
├─ Before file write
├─ Before package install
├─ Before destructive operation
└─ On validation error
```

### Trace Output

```
[13:42:15] React Expert starting
[13:42:15]   Context: Button.jsx (23 lines)
[13:42:15]   Task: "Add loading state"
[13:42:15]   Files to update: 3
[13:42:16] → Calling KHANARY:react
[13:42:17] ← Generated diff (45 lines)
[13:42:17] Validating...
[13:42:17]   ✓ Import paths correct
[13:42:17]   ✓ No circular dependencies
[13:42:17]   ✓ CSS matches structure
[13:42:17] Ready to apply
[13:42:17]   Files: Button.jsx, Button.module.css
[13:42:17]   New lines: 12
[13:42:17]   Deleted lines: 0
[13:42:17] ✓ Applied successfully
```

---

## 🚀 Environment Detection

### Auto-Detect

```
Detecting environment...
  ✓ Found package.json → Node.js project
  ✓ Framework: Next.js
  ✓ Build tool: webpack
  ✓ Package manager: npm
  ✓ TypeScript: yes
  ✓ Testing: Jest

Project type: React Frontend
Recommended agents:
  - react-designer
  - typescript-checker
  - jest-tester
  - webpack-optimizer
```

### Supported Environments

```
✅ Node.js/JavaScript/TypeScript
✅ Python/FastAPI/Django
✅ React/Vue/Angular
✅ .NET/C#
✅ Go
✅ Rust
✅ Docker/Kubernetes
✅ AWS/GCP/Azure
```

---

## 📋 Strict Policy Enforcement

### Rules Engine

```json
{
  "rules": [
    {
      "name": "no_single_file_changes",
      "description": "All changes must be cohesive",
      "severity": "error",
      "check": "related_files_updated"
    },
    {
      "name": "no_duplicates",
      "description": "No duplicate file names",
      "severity": "error",
      "check": "file_uniqueness"
    },
    {
      "name": "folder_aware",
      "description": "Respect project structure",
      "severity": "error",
      "check": "folder_structure_valid"
    },
    {
      "name": "import_consistency",
      "description": "All imports must be valid",
      "severity": "error",
      "check": "validate_imports"
    },
    {
      "name": "no_orphaned_code",
      "description": "No unused dead code",
      "severity": "warning",
      "check": "code_usage_check"
    }
  ]
}
```

---

## 🛠️ Tool Registry

### Built-in Tools

```
File Operations:
├─ read <path>
├─ write <path> <content>
├─ delete <path>
├─ move <from> <to>
├─ copy <from> <to>
└─ mkdir <path>

Package Management:
├─ npm install <package>
├─ pip install <package>
├─ dotnet add package <package>
└─ verify_installed <package>

Build & Test:
├─ npm run build
├─ npm test
├─ python -m pytest
├─ dotnet build
└─ docker build

Version Control:
├─ git status
├─ git add <files>
├─ git commit <message>
├─ git branch <name>
└─ git push

Debugging:
├─ trace <agent>
├─ breakpoint <location>
├─ inspect <variable>
└─ replay <session_id>
```

### Custom Tools (Agent-Created)

```
Agents can create tools if needed:
  "Agent needs tool 'webpack-optimize' but it doesn't exist"
  → Auto-creates from template
  → Registers in tool registry
  → Uses for current task
  → Available for future tasks
```

---

## 🔄 Implementation Phases

### Phase 1: Core Infrastructure
- [x] Architecture design
- [ ] PowerShell TUI framework
- [ ] Project awareness system
- [ ] File manager
- [ ] Session storage

### Phase 2: Agent System
- [ ] Agent manager
- [ ] Agent scaffold template
- [ ] Tool registry
- [ ] KHANARY router
- [ ] Built-in agents

### Phase 3: Code Generation
- [ ] Prompt validation
- [ ] Context gathering
- [ ] Code generation pipeline
- [ ] Diff visualization
- [ ] Change application

### Phase 4: Debugging & Recovery
- [ ] Agent debugger
- [ ] Crash recovery
- [ ] Session resume
- [ ] Undo/redo system
- [ ] Change history

### Phase 5: Polish & Integration
- [ ] Model manager UI
- [ ] Environment detection
- [ ] Rules enforcement
- [ ] Performance optimization
- [ ] User documentation

---

## 📚 Files to Create

```
Core:
├─ kuhul.ps1               # Main CLI entry point
├─ cli/                    # CLI components
│  ├─ tui.ps1             # TUI framework
│  ├─ renderer.ps1        # Screen rendering
│  ├─ input-handler.ps1   # Keyboard/mouse input
│  └─ layout.ps1          # UI layout logic
│
├─ core/                   # Core systems
│  ├─ project.ps1         # Project management
│  ├─ context.ps1         # Context gathering
│  ├─ environment.ps1     # Environment detection
│  ├─ file-manager.ps1    # File operations
│  └─ validator.ps1       # Strict validation
│
├─ agents/                 # Agent system
│  ├─ agent-manager.ps1   # Agent lifecycle
│  ├─ scaffolder.ps1      # Create agents
│  ├─ router.ps1          # KHANARY routing
│  ├─ built-ins/          # Built-in agents
│  │  ├─ code-generator.ps1
│  │  ├─ debugger.ps1
│  │  ├─ file-manager.ps1
│  │  └─ installer.ps1
│  └─ tools/              # Tool registry
│
├─ models/                 # KHANARY integration
│  ├─ model-manager.ps1   # Load/manage models
│  ├─ experts.ps1         # Expert routing
│  └─ khanary-client.ps1  # Call experts
│
└─ utils/                  # Utilities
   ├─ session.ps1         # Session management
   ├─ recovery.ps1        # Crash recovery
   ├─ logger.ps1          # Logging
   └─ config.ps1          # Configuration
```

---

## 🎯 What Makes This Different

```
Traditional IDEs:
  Editor → Code → File system
  Generic, broad capability

KUHUL CLI TUI:
  Chat Prompt → Agent → Project Context → Validated Edits
  Specialized, strict, purposeful

Key Differences:
✅ Every prompt generates code (not chat)
✅ File-aware (knows dependencies)
✅ Folder-aware (respects structure)
✅ Agent-driven (creates tools as needed)
✅ KHANARY-powered (40+ experts)
✅ Crash-resilient (full recovery)
✅ Strict rules (prevents mistakes)
✅ Debugging built-in (trace everything)
```

---

**Ready to implement? Start with Phase 1 core infrastructure.** 🚀
