# KHANARY Quick Start Guide

Get started with KHANARY Mixture of Experts in minutes.

## Installation

### Option 1: Clone Repository (Recommended)

The 5 core experts are included in the repo (~70MB):

```bash
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
ls -lh experts/*.khμ  # See included experts
```

### Option 2: Download Individual Experts

```bash
# Create experts directory
mkdir -p experts
cd experts

# Download from latest release
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/python.khμ
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/javascript.khμ
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/security.khμ
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/react.khμ
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/fastapi.khμ

# Verify checksums
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v3.0.0/SHA256SUMS
sha256sum -c SHA256SUMS
```

### Option 3: Use Installation Script

```bash
# Clone repo (minimal)
git clone --depth 1 https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K

# Run installer
python scripts/install_experts.py --version v3.0.0

# Verify
python scripts/install_experts.py --verify-only
```

---

## Your 5 Core Experts

| Expert | Size | Use Cases | Accuracy |
|--------|------|-----------|----------|
| **Python** | 14MB | Code generation, debugging, optimization | 92-95% |
| **JavaScript** | 14MB | Frontend/backend development, APIs | 91-94% |
| **Security** | 15MB | Vulnerability detection, secure coding | 88-94% |
| **React** | 14MB | Component generation, hooks, patterns | 91-94% |
| **FastAPI** | 14MB | Endpoint generation, async patterns | 90-93% |
| **TOTAL** | **70MB** | Comprehensive web/API development | **90%+** |

---

## Quick Commands

### List Experts

```bash
python scripts/download_datasets.py --list-experts
```

Output:
```
Available KHANARY Experts (40+):
  1. python
  2. javascript
  3. java
  4. cpp
  5. rust
  ...
  40. documentation
```

### Query Expert Details

```bash
python scripts/create_expert_registry.py --query-expert python
```

Output:
```json
{
  "name": "python",
  "file": "python.khμ",
  "size_mb": 14.0,
  "metadata": {
    "accuracy": 0.95,
    "latency_ms": 2.0,
    "model_type": "qwen"
  }
}
```

### Validate Experts

```bash
python scripts/validate_determinism.py --experts experts/*.khμ --runs 5
```

### Create Registry

```bash
python scripts/create_expert_registry.py --experts-dir experts
cat experts/registry.json | jq '.'
```

---

## Expansion: Add More Experts

### Download All 40+ Experts' Datasets

```bash
python scripts/download_datasets.py --expert all

# Or specific group:
python scripts/download_datasets.py --expert programming  # 8 experts
python scripts/download_datasets.py --expert frontend     # 4 experts
python scripts/download_datasets.py --expert backend      # 5 experts
```

### Train Additional Experts

```bash
# Train specific expert
python scripts/train_expert.py --expert java

# Train group
python scripts/train_all_experts.py --group programming --workers 2 --parallel

# Train all 40+
python scripts/train_all_experts.py --group all --workers 4 --parallel
```

### Compile to Binary

```bash
# Compile specific experts
python scripts/compile_all_experts.py \
  --experts java,cpp,rust \
  --compression fp16

# Compile all
python scripts/compile_all_experts.py \
  --experts all \
  --workers 4 \
  --parallel \
  --checksums
```

---

## System Requirements

### Minimum (Use 5 Core Experts)

```
CPU:    Any modern processor
RAM:    4GB
Disk:   500MB (for experts + dependencies)
GPU:    Not required (CPU inference supported)
```

### Recommended (Train New Experts)

```
CPU:    16+ cores
RAM:    32GB
Disk:   500GB (for datasets + training)
GPU:    1× 24GB (RTX 3090, RTX 4090, A5000, etc.)
```

### For Full System (40+ Experts)

```
CPU:    32+ cores
RAM:    64GB
Disk:   1TB (all datasets + checkpoints)
GPU:    4× 24GB (for parallel training)
```

---

## Use Cases

### 1. Python Development

```bash
# Query Python expert
python scripts/create_expert_registry.py --query-expert python

# Use for:
# - Code generation
# - Bug detection
# - Optimization suggestions
# - Best practices
```

### 2. Web Development (React + FastAPI)

```bash
# Query both experts
python scripts/create_expert_registry.py --query-expert react
python scripts/create_expert_registry.py --query-expert fastapi

# Use for:
# - React component generation
# - FastAPI endpoint scaffolding
# - Full-stack development assistance
```

### 3. Security Review

```bash
# Query security expert
python scripts/create_expert_registry.py --query-expert security

# Use for:
# - Vulnerability detection
# - Secure code review
# - OWASP compliance
# - Threat modeling
```

### 4. Expand to 40+ Experts

```bash
# Train all experts
python scripts/train_all_experts.py --group all --workers 4 --parallel

# Creates experts for:
# - All 8 programming languages
# - Data & databases (4 experts)
# - Science & math (6 experts)
# - Frontend/backend frameworks
# - Cloud platforms
# - And 15+ more domains
```

---

## Troubleshooting

### Experts not found

```bash
# Check installation
ls -lh experts/

# Expected output:
# python.khμ       14M
# javascript.khμ   14M
# security.khμ     15M
# react.khμ        14M
# fastapi.khμ      14M
```

### Checksum verification failed

```bash
# Re-download experts
rm experts/*.khμ
python scripts/install_experts.py --version v3.0.0
```

### Training is slow

```bash
# Use parallel workers (requires multiple GPUs)
python scripts/train_all_experts.py \
  --group programming \
  --workers 4 \
  --parallel

# Or reduce batch size
python scripts/train_expert.py --expert python --batch-size 16
```

### Out of memory during training

```bash
# Reduce batch size
python scripts/train_expert.py --expert python --batch-size 8

# Or reduce max samples
python scripts/download_datasets.py --expert python --max-samples 1000
```

---

## Next Steps

### Learn More

1. **EXPERT_CATALOG.md** - All 40+ available experts
2. **TRAINING_GUIDE_40_EXPERTS.md** - Complete training guide
3. **BUILD_SYSTEM_GUIDE.md** - Build orchestration
4. **KHANARY_BINARY_DISTRIBUTION.md** - Distribution strategy

### Contribute

- Submit issues: https://github.com/cannaseedus-bot/Ollama-K/issues
- Discussions: https://github.com/cannaseedus-bot/Ollama-K/discussions
- Pull requests: https://github.com/cannaseedus-bot/Ollama-K/pulls

### Expand System

```bash
# Option 1: Train specific domains
python scripts/train_all_experts.py --group ml_ai
python scripts/train_all_experts.py --group cloud
python scripts/train_all_experts.py --group science

# Option 2: Fine-tune on custom data
python scripts/train_expert.py --expert python --dataset-dir my_data

# Option 3: Add new experts
# Follow TRAINING_GUIDE_40_EXPERTS.md
```

---

## Architecture

```
KHANARY Mixture of Experts
│
├─ 5 Core Experts (.khμ binaries, 70MB)
│  ├─ Python
│  ├─ JavaScript
│  ├─ Security
│  ├─ React
│  └─ FastAPI
│
├─ 40+ Optional Experts
│  ├─ 8 Programming languages
│  ├─ 4 Data & database systems
│  ├─ 6 Science & math domains
│  ├─ 5 Real-world applications
│  └─ 17+ Specialized domains
│
└─ Hybrid Binary Format
   ├─ KHANARY signature
   ├─ Metadata (JSON)
   └─ Weights (SafeTensors)
```

---

## Performance

| Metric | Value |
|--------|-------|
| Model Size | 7B parameters |
| Latency | 2-3ms per request |
| Accuracy | 90%+ |
| Binary Size | 14MB (FP16) per expert |
| Total (5 experts) | 70MB |
| Compression | Lossless (FP16) |

---

## Support & Resources

### Documentation

- Repository: https://github.com/cannaseedus-bot/Ollama-K
- Issues: https://github.com/cannaseedus-bot/Ollama-K/issues
- Discussions: https://github.com/cannaseedus-bot/Ollama-K/discussions

### Related Projects

- Ollama: https://ollama.ai
- HuggingFace Transformers: https://huggingface.co/transformers/
- The Stack: https://huggingface.co/datasets/bigcode/the-stack

---

## License

KHANARY is released under MIT License. See LICENSE file for details.

---

**Ready to get started?** 🚀

```bash
git clone https://github.com/cannaseedus-bot/Ollama-K.git
cd Ollama-K
python scripts/create_expert_registry.py --experts-dir experts
```
