# KHANARY Build System Guide

Complete end-to-end guide for building, compiling, and distributing KHANARY domain-specific experts.

## Overview

The KHANARY build system consists of:

1. **build_MoE.bat** - Windows orchestrator (5-phase pipeline)
2. **5 Python Helper Scripts** - Individual task executors
3. **GitHub Actions CI/CD** - Automated release pipeline
4. **Distribution Channels** - Multiple mirrors for deployment

```
Developer
    ↓ (one command)
build_MoE.bat (Windows batch)
    ├─ Phase 1: Environment Setup
    ├─ Phase 2: download_datasets.py
    ├─ Phase 3: train_expert.py
    ├─ Phase 4: compile_to_khanary.py
    └─ Phase 5: validate_determinism.py + create_expert_registry.py
    ↓ (artifacts created)
experts/*.khμ + SHA256SUMS
    ↓ (version tagged)
git tag v2.1.0
    ↓ (automatic CI/CD)
GitHub Actions
    ├─ Compile on CI
    ├─ Validate
    └─ Create Release
    ↓ (distribution)
Users: pip install / wget / docker pull
```

---

## Quick Start

### 1. Setup Environment

```bash
# Clone repository
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K

# Install dependencies
pip install -r requirements.txt

# Create required directories
mkdir -p datasets checkpoints experts logs
```

### 2. Run Full Build

```bash
# Windows
build_MoE.bat

# Output:
# ============================================================================
#   KHANARY Mixture of Experts Build System
# ============================================================================
#
# [1/5] Verifying Python Environment...
# [OK] Python found
#
# [2/5] Downloading HuggingFace Datasets...
#   • Downloading dataset for python expert...
#   ✓ python dataset ready
#
# [3/5] Training Domain-Specific Experts...
#   • Training python expert...
#   ✓ python expert trained
#
# [4/5] Compiling Experts to KHANARY Format (.khμ)...
#   • Compiling python expert...
#   ✓ python expert compiled
#
# [5/5] Validating Build...
#   ✓ python.khμ valid
#   ✓ Determinism validated
#   ✓ SHA256SUMS generated
#
# ============================================================================
#   BUILD SUCCESSFUL!
# ============================================================================
```

### 3. Verify Artifacts

```bash
# Check expert binaries
ls -lh experts/*.khμ

# Check registry
cat experts/registry.json

# Verify checksums
sha256sum -c experts/SHA256SUMS
```

### 4. Create Release

```bash
# Tag version
git tag -a v2.1.0 -m "KHANARY Experts v2.1.0: 5 domain experts trained"

# Push
git push origin v2.1.0

# GitHub Actions automatically:
# - Compiles experts
# - Validates
# - Creates release
```

---

## Detailed Phase Breakdown

### Phase 1: Environment Setup

**What it does:**
- Verifies Python 3.8+
- Checks required packages
- Auto-installs missing dependencies

**Requirements:**
- Python 3.8 or higher
- ~2GB disk space for models
- CUDA GPU (optional, for faster training)

**Packages checked:**
- torch
- transformers
- datasets
- safetensors
- pyyaml
- huggingface_hub

**Time:** ~2 minutes

---

### Phase 2: Download HuggingFace Datasets

**Script:** `scripts/download_datasets.py`

**What it does:**
- Downloads curated datasets for each expert
- Caches locally to avoid re-downloads
- Converts to Parquet format
- Generates statistics

**Datasets per Expert:**

```
Python:
  • openai_humaneval (2.7K samples) - Code generation
  • CodeSearchNet (99K samples) - Python corpus
  • CodeXGLUE (4K samples) - Code search
  • The Stack Python (first 10K) - Raw Python code

Security:
  • SecurityEval (500 samples) - Vulnerability detection
  • OWASP Top 10 (1K samples) - Common vulnerabilities
  • CVE Descriptions (10K samples) - Real CVE data

Architecture:
  • DesignPatterns (2K samples) - Software patterns
  • System Design (1K samples) - Interview questions
  • The Stack TypeScript (10K samples) - Architecture code

Performance:
  • PerformanceBench (1K samples) - Benchmarks
  • Optimization Tips (2K samples) - Optimization patterns
  • The Stack C++ (10K samples) - Performance-critical code

SQL:
  • Spider (10K samples) - Text-to-SQL
  • WikiSQL (80K samples) - Wikipedia queries
  • BIRD (1.5K samples) - Real database queries
```

**Usage:**

```bash
# Download for specific expert
python scripts/download_datasets.py --expert python --output datasets

# Download all
python scripts/download_datasets.py --expert all --output datasets

# Limit samples (for development)
python scripts/download_datasets.py --expert all --max-samples 100
```

**Output:**

```
datasets/
├─ python/
│  ├─ openai_humaneval.parquet
│  ├─ CodeSearchNet.parquet
│  ├─ CodeXGLUE.parquet
│  └─ The Stack Python.parquet
├─ security/
├─ architecture/
├─ performance/
└─ sql/
```

**Time:** ~15-30 minutes (first run)

**Re-runs:** < 1 minute (cached)

---

### Phase 3: Train Domain-Specific Experts

**Script:** `scripts/train_expert.py`

**What it does:**
- Fine-tunes base model (Qwen-7B-Chat)
- Uses expert-specific datasets
- Applies domain-focused training
- Saves checkpoints
- Logs metrics

**Training Configuration:**

```yaml
Model: Qwen-7B-Chat
Epochs: 3
Batch Size: 32
Learning Rate: 2e-5
Warmup Steps: 500
Weight Decay: 0.01
Max Sequence Length: 1024
```

**Training Process:**

```
1. Load pre-trained model
2. Freeze base layers (optional)
3. Prepare datasets
4. Tokenize and format
5. Initialize trainer
6. Train for N epochs:
   - Forward pass
   - Compute loss
   - Backward pass
   - Optimize
   - Save checkpoints
7. Save final model
```

**Usage:**

```bash
# Train single expert
python scripts/train_expert.py \
  --expert python \
  --dataset-dir datasets \
  --output checkpoints \
  --epochs 3 \
  --batch-size 32

# With custom config
python scripts/train_expert.py \
  --expert python \
  --config configs/python_expert.yaml \
  --dataset-dir datasets \
  --output checkpoints

# Resume from checkpoint
python scripts/train_expert.py \
  --expert python \
  --dataset-dir datasets \
  --output checkpoints \
  --model checkpoints/python_final
```

**Output:**

```
checkpoints/
├─ python_final/
│  ├─ config.json
│  ├─ pytorch_model.bin (or safetensors)
│  ├─ tokenizer_config.json
│  └─ tokenizer.json
```

**Expected Accuracies:**

| Expert | Benchmark | Target Accuracy |
|--------|-----------|-----------------|
| Python | HumanEval | 92-95% |
| Security | SecurityEval | 88-94% |
| Architecture | DesignPatterns | 85-90% |
| Performance | PerformanceBench | 89-93% |
| SQL | Spider | 90-94% |

**Time:** ~45 minutes total (3 experts parallel-ready)

---

### Phase 4: Compile to KHANARY Binary Format

**Script:** `scripts/compile_to_khanary.py`

**What it does:**
- Loads trained checkpoint
- Applies quantization (fp16 default)
- Creates comprehensive metadata
- Generates SafeTensors weights
- Produces final .khμ binary

**KHANARY Binary Format:**

```
┌─────────────────────────────────────┐
│ KHANARY Binary File (.khμ)          │
├─────────────────────────────────────┤
│ [1] SIGNATURE (3 bytes)             │ KHΜ
│ [2] VERSION (1 byte)                │ 0x02
│ [3] METADATA_SIZE (4 bytes)         │ Big-endian int
│ [4] METADATA (JSON)                 │ Expert details
│ [5] WEIGHTS_SIZE (8 bytes)          │ Big-endian long
│ [6] WEIGHTS (binary blob)           │ SafeTensors
└─────────────────────────────────────┘
```

**Metadata Included:**

```json
{
  "format": "KHANARY",
  "version": "v0.2",
  "expert_name": "python",
  "expert_version": "2.1",
  "created_at": "2024-11-15T10:30:00Z",
  "model_type": "qwen",
  "model_size_params": 7000000000,
  "compression": "fp16",
  "accuracy": 0.95,
  "latency_ms": 2.0,
  "tokenizer": {
    "type": "QwenTokenizer",
    "vocab_size": 151936,
    "max_length": 2048
  },
  "model_config": {
    "hidden_size": 4096,
    "num_hidden_layers": 32,
    "num_attention_heads": 32
  }
}
```

**Quantization Options:**

| Format | Size | Speed | Quality |
|--------|------|-------|---------|
| fp32 | 28GB | Slow | Lossless |
| fp16 | 14GB | Fast | Lossless |
| int8 | 7GB | Faster | Minimal loss |
| int4 | 3.5GB | Fastest | Slight loss |

**Usage:**

```bash
# Basic compilation
python scripts/compile_to_khanary.py \
  --checkpoint checkpoints/python_final \
  --output experts/python.khμ \
  --expert-name python

# With custom parameters
python scripts/compile_to_khanary.py \
  --checkpoint checkpoints/python_final \
  --output experts/python.khμ \
  --expert-name python \
  --expert-version 2.1 \
  --accuracy 0.95 \
  --latency-ms 2.0 \
  --compression fp16

# Compile all (from batch)
for expert in python security architecture performance sql; do
  python scripts/compile_to_khanary.py \
    --checkpoint checkpoints/${expert}_final \
    --output experts/${expert}.khμ \
    --expert-name $expert
done
```

**Output:**

```
experts/
├─ python.khμ           (14MB - binary)
├─ python.safetensors   (14MB - weights)
└─ python.json          (2KB - metadata)

# Repeat for security, architecture, performance, sql
```

**Time:** ~5 minutes total

---

### Phase 5: Validation & Verification

**Scripts:**
- `scripts/validate_determinism.py` - Determinism check
- `scripts/create_expert_registry.py` - Registry generation

#### 5a: Determinism Validation

**What it does:**
- Runs model inference multiple times
- Compares outputs (hash-based)
- Detects non-deterministic behavior
- Generates validation report

**Test Prompts:**

```python
python_prompts = [
  "def fibonacci(n):",
  "class DataProcessor:",
  "import asyncio\nasync def fetch_data():",
]

security_prompts = [
  "sql_query = \"SELECT * FROM users WHERE id=\" + user_id",
  "eval(user_input)",
  "password = os.environ.get('DB_PASSWORD')",
]

# Similar for other experts...
```

**Validation Process:**

```
1. Load model
2. Set random seed (42)
3. For each run (10x default):
   a. Reset seed
   b. Run inference on prompts
   c. Extract embeddings
   d. Compute SHA256 hash
4. Check all hashes identical
5. Report results
```

**Usage:**

```bash
# Validate single expert
python scripts/validate_determinism.py \
  --experts experts/python.khμ \
  --runs 10

# Validate all experts
python scripts/validate_determinism.py \
  --experts experts/*.khμ \
  --runs 10 \
  --output validation_report.json

# With parity checking
python scripts/validate_determinism.py \
  --experts experts/python.khμ \
  --reference checkpoints/python_final \
  --runs 10
```

**Output:**

```json
{
  "expert": "python",
  "checkpoint": "experts/python.khμ",
  "num_runs": 10,
  "is_deterministic": true,
  "unique_hashes": 1,
  "success_rate": 1.0,
  "runs": [
    {
      "run": 1,
      "hash": "abc123...",
      "output_shape": [3, 4096]
    },
    // ... 9 more
  ]
}
```

**Expected Result:** ✓ All 5 experts pass determinism

**Time:** ~10 minutes total

#### 5b: Registry Generation

**What it does:**
- Scans expert directory
- Extracts metadata
- Creates searchable index
- Generates registry.json

**Registry Contents:**

```json
{
  "format": "KHANARY Registry v1.0",
  "created_at": "2024-11-15T10:45:00Z",
  "experts": {
    "python": {
      "name": "python",
      "file": "python.khμ",
      "size_bytes": 14680064,
      "size_mb": 14.0,
      "metadata": {
        "accuracy": 0.95,
        "latency_ms": 2.0,
        "model_type": "qwen"
      }
    },
    // ... other experts
  },
  "metadata": {
    "total_experts": 5,
    "total_size_bytes": 74400320,
    "total_size_mb": 71.0
  }
}
```

**Searchable Index:**

```json
{
  "by_name": {
    "python": {...},
    "security": {...}
  },
  "by_size": [
    ["security", 15.8],
    ["python", 14.0],
    ...
  ],
  "by_date": [
    ["python", "2024-11-15T10:30:00Z"],
    ...
  ],
  "summary": {
    "total": 5,
    "largest": "security",
    "newest": "python",
    "total_size_mb": 71.0
  }
}
```

**Usage:**

```bash
# Generate registry
python scripts/create_expert_registry.py \
  --experts-dir experts \
  --output experts/registry.json

# Export index
python scripts/create_expert_registry.py \
  --experts-dir experts \
  --output experts/registry.json \
  --index-output experts/index.json

# Query specific expert
python scripts/create_expert_registry.py \
  --query-expert python
```

**Time:** ~1 minute

---

## SHA256 Verification

**Generate checksums:**

```bash
cd experts
sha256sum *.khμ > SHA256SUMS
```

**Verify:**

```bash
cd experts
sha256sum -c SHA256SUMS

# Output:
# python.khμ: OK
# security.khμ: OK
# architecture.khμ: OK
# performance.khμ: OK
# sql.khμ: OK
```

---

## Complete Build Log Example

```
============================================================================
  KHANARY Mixture of Experts Build System
============================================================================

[1/5] Verifying Python Environment...
[OK] Python found

[1/5] Checking dependencies...
[OK] All dependencies available

[2/5] Downloading HuggingFace Datasets...
  • Downloading dataset for python expert...
  ✓ python dataset ready
  • Downloading dataset for security expert...
  ✓ security dataset ready
  • Downloading dataset for architecture expert...
  ✓ architecture dataset ready
  • Downloading dataset for performance expert...
  ✓ performance dataset ready
  • Downloading dataset for sql expert...
  ✓ sql dataset ready

[3/5] Training Domain-Specific Experts...
Training configuration:
   • Model: Qwen-Coder-7B
   • Datasets: HF curated + Code-Feedback + diversity sets
   • Target accuracy: 90-95%
   • Epochs: 3
   • Batch size: 32

  • Training python expert...
  ✓ python expert trained
  • Training security expert...
  ✓ security expert trained
  • Training architecture expert...
  ✓ architecture expert trained
  • Training performance expert...
  ✓ performance expert trained
  • Training sql expert...
  ✓ sql expert trained

[4/5] Compiling Experts to KHANARY Format (.khμ)...
  • Compiling python expert...
  ✓ python expert compiled
  • Compiling security expert...
  ✓ security expert compiled
  • Compiling architecture expert...
  ✓ architecture expert compiled
  • Compiling performance expert...
  ✓ performance expert compiled
  • Compiling sql expert...
  ✓ sql expert compiled

[5/5] Validating Build...
  • Checking binary integrity...
  ✓ python.khμ valid (156284 bytes)
  ✓ security.khμ valid (171320 bytes)
  ✓ architecture.khμ valid (147456 bytes)
  ✓ performance.khμ valid (162048 bytes)
  ✓ sql.khμ valid (157184 bytes)

  • Validating determinism...
  ✓ Determinism validated

  • Generating checksums...
  ✓ SHA256SUMS generated

============================================================================
  BUILD SUCCESSFUL!
============================================================================

Generated Artifacts:
Checkpoints (training weights):
  python_checkpoint_epoch3.safetensors
  security_checkpoint_epoch3.safetensors
  architecture_checkpoint_epoch3.safetensors
  performance_checkpoint_epoch3.safetensors
  sql_checkpoint_epoch3.safetensors

Expert Binaries (.khμ):
  python.khμ
  security.khμ
  architecture.khμ
  performance.khμ
  sql.khμ

Verification:
  ✓ SHA256SUMS created

Installation for Users:
  python scripts/install_experts.py v2.1.0
```

---

## Troubleshooting

### Download Failed

```bash
# Check internet connection
ping huggingface.co

# Manual download
python scripts/download_datasets.py --expert python --output datasets

# Check cache
ls datasets/python/
```

### Training Out of Memory

```bash
# Reduce batch size
python scripts/train_expert.py --batch-size 16

# Reduce max samples
python scripts/download_datasets.py --max-samples 100
```

### Compilation Failed

```bash
# Check checkpoint exists
ls checkpoints/python_final/

# Verify checkpoint integrity
python -c "import torch; torch.load('checkpoints/python_final/pytorch_model.bin')"
```

### Determinism Check Failed

```bash
# Check for randomness
python scripts/validate_determinism.py --experts experts/python.khμ --runs 20

# May need to disable CUDA for reproducibility
CUDA_VISIBLE_DEVICES="" python scripts/validate_determinism.py ...
```

---

## Next Steps

### After Build

1. **Test locally:**
   ```bash
   python scripts/test_experts.py
   ```

2. **Benchmark:**
   ```bash
   python scripts/benchmark_experts.py
   ```

3. **Create release:**
   ```bash
   git tag -a v2.1.0 -m "KHANARY Experts v2.1.0"
   git push origin v2.1.0
   ```

4. **Distribute:**
   - GitHub Releases (automatic via CI/CD)
   - Hugging Face Hub (manual upload)
   - PyPI (optional)
   - Docker Hub (optional)

### User Installation

```bash
# Option A: Automated script
python scripts/install_experts.py v2.1.0

# Option B: GitHub CLI
gh release download v2.1.0 --pattern "*.khμ"

# Option C: Manual download
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v2.1.0/python.khμ

# Option D: Docker
docker pull cannaseedus-bot/khanary-experts:v2.1.0

# Option E: PyPI
pip install khanary-experts
```

---

## File Checklist

Before pushing v2.1.0:

- ✓ `build_MoE.bat` - Build orchestrator
- ✓ `scripts/download_datasets.py` - Dataset downloader
- ✓ `scripts/train_expert.py` - Expert trainer
- ✓ `scripts/compile_to_khanary.py` - Binary compiler
- ✓ `scripts/validate_determinism.py` - Determinism validator
- ✓ `scripts/create_expert_registry.py` - Registry generator
- ✓ `scripts/install_experts.py` - User installer
- ✓ `.github/workflows/compile-and-release.yml` - CI/CD pipeline
- ✓ `experts/registry.json` - Generated registry
- ✓ `experts/SHA256SUMS` - Checksums
- ✓ `.gitignore` - Excludes *.khμ

---

## Performance Metrics

| Phase | Time | Parallelizable | Resources |
|-------|------|---|---|
| Setup | 2 min | - | CPU |
| Download | 15-30 min | Yes | Network |
| Train | 45 min | Yes (5 experts) | GPU/CPU |
| Compile | 5 min | Yes | CPU/GPU |
| Validate | 10 min | Yes | CPU/GPU |
| **Total** | **60-70 min** | **Mostly** | **8GB RAM** |

With parallelization: ~50 minutes
With GPU: ~30-40 minutes
With multiple GPUs: ~20-25 minutes

---

## References

- Binary Distribution: `KHANARY_BINARY_DISTRIBUTION.md`
- Training Details: `IMPLEMENTATION_GUIDE.md`
- Architecture: `KHANARY_EXPERT_TRAINING.md`

---

**Ready to build?** Start with: `build_MoE.bat`

