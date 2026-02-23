# KHANARY Training Guide: 40+ Specialized Experts

Complete guide for training, compiling, and deploying 40+ domain-specific KHANARY experts.

## Quick Start

### Option 1: Train Core 5 Experts (Recommended for first-time)

```bash
# 1. Download datasets
python scripts/download_datasets.py --expert all

# 2. Run build system (trains core 5 experts)
build_MoE.bat  # Windows
# or
./build_MoE.sh  # Linux/Mac

# Expected time: 2-3 hours
# Output: 5 × .khμ files (~70MB total)
```

### Option 2: Train All 40+ Experts (Comprehensive)

```bash
# 1. Download all datasets (120+)
python scripts/download_datasets.py --expert all

# 2. Train all experts in groups
python scripts/train_all_experts.py --group all --workers 4 --parallel

# Expected time: 200+ GPU hours (with 4 parallel workers)
# Output: 40+ × .khμ files (~560MB total)
```

### Option 3: Train Custom Subset

```bash
# Train specific experts
python scripts/train_all_experts.py \
  --experts python,javascript,react,fastapi,security \
  --workers 2 --parallel \
  --epochs 3 --batch-size 32

# Expected time: 30-45 minutes (with 2 parallel workers)
# Output: 5 × .khμ files
```

---

## Available Expert Groups

### 1. Core Group (5 experts, 1-2 hours)
```bash
python scripts/train_all_experts.py --group core

# Includes: python, javascript, security, react, fastapi
# Covers: Popular programming languages & frameworks
```

### 2. Programming Languages (8 experts, 3-4 hours)
```bash
python scripts/train_all_experts.py --group programming

# Includes: python, javascript, java, cpp, rust, go, csharp, typescript
# Covers: 8 major programming languages
```

### 3. Data & Databases (4 experts, 1-2 hours)
```bash
python scripts/train_all_experts.py --group data

# Includes: sql, data_engineering, nosql, data_science
# Covers: Query optimization, ETL, analytics
```

### 4. Science & Math (6 experts, 2-3 hours)
```bash
python scripts/train_all_experts.py --group science

# Includes: physics, chemistry, mathematics, biology, space, engineering
# Covers: Academic & scientific domains
```

### 5. Frontend Development (4 experts, 1-2 hours)
```bash
python scripts/train_all_experts.py --group frontend

# Includes: react, vue, angular, ux_design
# Covers: Popular frontend frameworks & UX
```

### 6. Backend Development (5 experts, 2-3 hours)
```bash
python scripts/train_all_experts.py --group backend

# Includes: nodejs, django, fastapi, devops, microservices
# Covers: Popular backend frameworks & DevOps
```

### 7. Cloud Platforms (3 experts, 1-2 hours)
```bash
python scripts/train_all_experts.py --group cloud

# Includes: aws, gcp, azure
# Covers: All major cloud providers
```

### 8. Terminal & Shell (3 experts, 1 hour)
```bash
python scripts/train_all_experts.py --group terminal

# Includes: bash, cli_tools, git
# Covers: Shell scripting & version control
```

### 9. Testing & Quality (3 experts, 1 hour)
```bash
python scripts/train_all_experts.py --group testing

# Includes: testing, debugging, monitoring
# Covers: QA & observability
```

### 10. Security & Advanced (4 experts, 2 hours)
```bash
python scripts/train_all_experts.py --group security_advanced

# Includes: security, cryptography, blockchain, iot
# Covers: Security & advanced topics
```

### 11. Machine Learning (3 experts, 2 hours)
```bash
python scripts/train_all_experts.py --group ml_ai

# Includes: machine_learning, nlp, computer_vision
# Covers: All AI/ML domains
```

### 12. Specialized Tools (4 experts, 2 hours)
```bash
python scripts/train_all_experts.py --group specialized

# Includes: architecture, performance, graphics, documentation
# Covers: Design, optimization, creative
```

### 13. Domain Applications (5 experts, 2-3 hours)
```bash
python scripts/train_all_experts.py --group domain_apps

# Includes: healthcare, finance, legal, business, education
# Covers: Real-world industry domains
```

---

## Training Details

### System Requirements

**For Single Expert:**
- GPU: 1× 24GB VRAM (RTX 3090, RTX 4090, A5000, etc.)
- RAM: 16GB CPU RAM
- Storage: 50GB (dataset + checkpoint)
- Time: 30-60 minutes

**For 5 Experts (Sequential):**
- GPU: 1× 24GB VRAM
- RAM: 16GB CPU RAM
- Storage: 250GB (datasets + checkpoints)
- Time: 2-3 hours

**For 40+ Experts (Parallel, 4 workers):**
- GPU: 4× 24GB VRAM (recommended)
- RAM: 64GB CPU RAM
- Storage: 1TB (all datasets + checkpoints)
- Time: 50+ hours (with 4 parallel workers)

### Training Configuration

```python
Base Model:      Qwen-7B-Chat
Epochs:          3
Batch Size:      32
Learning Rate:   2e-5
Warmup Steps:    500
Weight Decay:    0.01
Max Sequence:    1024
Precision:       FP16 (mixed)
Optimizer:       AdamW
Gradient Accum:  1
```

### Estimated Timings (per expert)

| Task | Sequential | Parallel (4 workers) |
|------|-----------|------------------|
| Download (1 expert) | 5 min | N/A |
| Download (all) | 30-45 min | 10-15 min |
| Train (1 expert) | 30-60 min | 10-15 min |
| Train (5 experts) | 2-3 hours | 30-45 min |
| Train (40 experts) | 30+ hours | 50+ hours* |
| Compile (1 expert) | 2-3 min | N/A |
| Compile (5 experts) | 10-15 min | 3-5 min |
| Compile (40 experts) | 80-120 min | 20-30 min |

*Per-expert training time ≈ 30-60 min; 40 experts = 1200-2400 min ÷ 4 workers = 300-600 min (5-10 hours actual)

---

## Training Workflows

### Workflow 1: Quick MVP (30 minutes)

Best for: Testing, proof-of-concept, development

```bash
# 1. Download core datasets
python scripts/download_datasets.py --expert core --max-samples 100

# 2. Train core experts quickly
python scripts/train_all_experts.py \
  --group core \
  --epochs 1 \
  --batch-size 16

# 3. Compile
python scripts/compile_all_experts.py \
  --experts python,javascript,react \
  --compression int4

# Total time: ~30 minutes
# Binary size: ~50MB
```

### Workflow 2: Standard Build (2-3 hours)

Best for: Production release, quality assurance

```bash
# 1. Download all datasets
python scripts/download_datasets.py --expert all

# 2. Train core 5 experts
build_MoE.bat

# 3. Compile with fp16
python scripts/compile_all_experts.py \
  --experts python,javascript,security,react,fastapi \
  --compression fp16 \
  --checksums

# Total time: ~2-3 hours
# Binary size: ~70MB (fp16) or ~140MB (fp32)
```

### Workflow 3: Comprehensive (50+ hours)

Best for: Complete product, all domains

```bash
# 1. Download all datasets
python scripts/download_datasets.py --expert all

# 2. Train all 40+ experts in parallel
python scripts/train_all_experts.py \
  --group all \
  --workers 4 \
  --parallel \
  --results training_results.json

# 3. Compile all with parallel
python scripts/compile_all_experts.py \
  --experts all \
  --compression fp16 \
  --workers 4 \
  --parallel \
  --checksums \
  --results compilation_results.json

# 4. Create registry
python scripts/create_expert_registry.py \
  --experts-dir experts \
  --output experts/registry.json \
  --index-output experts/index.json

# Total time: ~50-100 hours (depending on hardware)
# Binary size: ~560MB (fp16)
```

### Workflow 4: Staged Release (Multiple phases)

Best for: Incremental rollout

**Phase 1: Core Languages (Day 1)**
```bash
python scripts/train_all_experts.py --group programming --workers 2 --parallel
# 8 language experts: 3-4 hours
```

**Phase 2: Web Stack (Day 2)**
```bash
python scripts/train_all_experts.py --group frontend --workers 1
python scripts/train_all_experts.py --group backend --workers 1
# 9 experts: 3-4 hours each
```

**Phase 3: Data & AI (Day 3)**
```bash
python scripts/train_all_experts.py --group data --workers 2 --parallel
python scripts/train_all_experts.py --group ml_ai --workers 2 --parallel
# 7 experts: 2-3 hours
```

**Phase 4: Specialized (Day 4)**
```bash
python scripts/train_all_experts.py --group science --workers 2 --parallel
python scripts/train_all_experts.py --group domain_apps --workers 2 --parallel
# 11 experts: 4-5 hours
```

---

## Training Commands Reference

### List Available Groups
```bash
python scripts/train_all_experts.py --list-groups
```

### Train Single Expert
```bash
python scripts/train_expert.py --expert python
python scripts/train_expert.py --expert react
python scripts/train_expert.py --expert healthcare
```

### Train Group (Sequential)
```bash
python scripts/train_all_experts.py --group backend
python scripts/train_all_experts.py --group frontend
python scripts/train_all_experts.py --group science
```

### Train Group (Parallel, 2 workers)
```bash
python scripts/train_all_experts.py \
  --group backend \
  --workers 2 \
  --parallel
```

### Train Group (Parallel, 4 workers)
```bash
python scripts/train_all_experts.py \
  --group all \
  --workers 4 \
  --parallel \
  --epochs 3 \
  --batch-size 32
```

### Train Custom List
```bash
python scripts/train_all_experts.py \
  --experts python,javascript,java,cpp,rust,go \
  --workers 3 \
  --parallel
```

### Train with Custom Parameters
```bash
python scripts/train_all_experts.py \
  --group frontend \
  --workers 2 \
  --parallel \
  --epochs 5 \
  --batch-size 16 \
  --learning-rate 1e-4
```

### Train and Save Results
```bash
python scripts/train_all_experts.py \
  --group all \
  --workers 4 \
  --parallel \
  --results training_results.json
```

---

## Compilation Commands Reference

### Compile Single Expert
```bash
python scripts/compile_to_khanary.py \
  --checkpoint checkpoints/python_final \
  --output experts/python.khμ \
  --expert-name python
```

### Compile Multiple Experts
```bash
python scripts/compile_all_experts.py \
  --experts python,javascript,java,cpp,rust \
  --output experts \
  --compression fp16
```

### Compile All with Options
```bash
python scripts/compile_all_experts.py \
  --experts all \
  --checkpoints-dir checkpoints \
  --output experts \
  --compression fp16 \
  --workers 4 \
  --parallel \
  --checksums \
  --results compilation_results.json
```

### Compile with Different Compression
```bash
# FP32 (lossless, largest)
python scripts/compile_all_experts.py \
  --experts all \
  --compression fp32

# FP16 (lossless, balanced)
python scripts/compile_all_experts.py \
  --experts all \
  --compression fp16

# INT8 (minimal loss, smaller)
python scripts/compile_all_experts.py \
  --experts all \
  --compression int8

# INT4 (more loss, smallest)
python scripts/compile_all_experts.py \
  --experts all \
  --compression int4
```

---

## Parallel Training Best Practices

### Rule of Thumb
- 1 expert per 6-8GB GPU VRAM
- 4 experts × 24GB GPU = 2-3 experts per GPU (recommended)
- 8 workers × 24GB GPUs = ~20-24 parallel experts

### Configuration for Different Hardware

**Single GPU (24GB)**
```bash
python scripts/train_all_experts.py \
  --group core \
  --workers 1  # No parallelization
```

**Dual GPUs (2×24GB)**
```bash
python scripts/train_all_experts.py \
  --group backend \
  --workers 2 \
  --parallel
```

**4-GPU Setup (4×24GB)**
```bash
python scripts/train_all_experts.py \
  --group all \
  --workers 4 \
  --parallel
```

**8-GPU Setup (8×24GB)**
```bash
python scripts/train_all_experts.py \
  --group all \
  --workers 8 \
  --parallel  # Run 8 experts in parallel
```

---

## Monitoring Training

### Watch Training Logs
```bash
# Real-time log monitoring
tail -f logs/train_python.log
tail -f logs/train_javascript.log

# Watch all training logs
ls -lt logs/train_*.log | head -5

# Check training progress
grep "epoch\|loss\|accuracy" logs/train_*.log
```

### Check Checkpoint Progress
```bash
# List checkpoints
ls -lh checkpoints/

# Check specific expert checkpoint
ls -lh checkpoints/python_*

# Monitor disk usage
du -sh checkpoints/
du -sh datasets/
```

### Training Results
```bash
# View training results (if saved)
cat training_results.json | python -m json.tool

# Check compilation results
cat compilation_results.json | python -m json.tool
```

---

## Troubleshooting

### Out of Memory (OOM)

```bash
# Reduce batch size
python scripts/train_expert.py --expert python --batch-size 16

# Reduce max workers
python scripts/train_all_experts.py \
  --group backend \
  --workers 1  # Use fewer parallel jobs

# Reduce sequence length (in train_expert.py)
# Change max_length from 1024 to 512
```

### Dataset Not Found

```bash
# Download missing datasets
python scripts/download_datasets.py --expert all

# Check dataset directory
ls -la datasets/

# Download specific expert
python scripts/download_datasets.py --expert python
```

### Training Too Slow

```bash
# Enable parallel training
python scripts/train_all_experts.py \
  --group backend \
  --workers 4 \
  --parallel

# Use multiple GPUs
CUDA_VISIBLE_DEVICES=0,1,2,3 python scripts/train_all_experts.py \
  --group all \
  --workers 4 \
  --parallel

# Reduce epochs for testing
python scripts/train_expert.py --expert python --epochs 1
```

### Compilation Failed

```bash
# Check checkpoint exists
ls checkpoints/python_final/

# Verify checkpoint is valid
python -c "import torch; print(torch.load('checkpoints/python_final/pytorch_model.bin').keys())"

# Compile with verbose logging
python scripts/compile_to_khanary.py \
  --checkpoint checkpoints/python_final \
  --output experts/python.khμ \
  --log-file logs/compile_verbose.log
```

---

## Production Deployment

### Step 1: Train & Compile

```bash
# Train all 40+ experts
python scripts/train_all_experts.py \
  --group all \
  --workers 8 \
  --parallel \
  --results training_results.json

# Compile all experts
python scripts/compile_all_experts.py \
  --experts all \
  --compression fp16 \
  --workers 8 \
  --parallel \
  --checksums \
  --results compilation_results.json
```

### Step 2: Quality Assurance

```bash
# Validate determinism
python scripts/validate_determinism.py \
  --experts experts/*.khμ \
  --runs 10 \
  --output validation_report.json

# Create registry
python scripts/create_expert_registry.py \
  --experts-dir experts \
  --output experts/registry.json
```

### Step 3: Package & Release

```bash
# Create checksum file
cd experts
sha256sum *.khμ > SHA256SUMS

# Tag version
git tag -a v3.0.0-full -m "KHANARY v3.0.0: All 40+ Experts"

# Push to remote
git push origin v3.0.0-full

# Create release
gh release create v3.0.0-full \
  --title "KHANARY v3.0.0 - 40+ Specialized Experts" \
  --notes "Complete expert system with 40+ domain specialists"

# Upload binaries
gh release upload v3.0.0-full experts/*.khμ experts/registry.json
```

---

## Storage & Cleanup

### Disk Usage Estimates

```
Datasets:        150GB+ (120+ datasets)
Checkpoints:     280GB (40 experts × 7B)
Compiled (.khμ): 560MB (40 experts × 14MB fp16)

Total:           ~430GB during build
Final:           ~560MB (just binaries)
```

### Cleanup After Build

```bash
# Remove intermediate checkpoints (keeps final)
rm -rf checkpoints/*_epoch1 checkpoints/*_epoch2

# Archive old datasets
tar -czf datasets_backup.tar.gz datasets/

# Keep only compiled experts
mkdir -p production
cp experts/*.khμ experts/registry.json production/
```

---

## Advanced Customization

### Custom Model Base

```bash
python scripts/train_expert.py \
  --expert python \
  --model "mistral-7b" \
  --dataset-dir datasets \
  --output checkpoints
```

### Transfer Learning

```bash
# Train from existing checkpoint
python scripts/train_expert.py \
  --expert python \
  --model checkpoints/python_v1 \
  --dataset-dir datasets \
  --output checkpoints \
  --epochs 2
```

### Fine-grained Control

Create `configs/expert_name.yaml`:

```yaml
model: "Qwen-7B-Chat"
epochs: 5
batch_size: 16
learning_rate: 1e-4
warmup_steps: 1000
weight_decay: 0.01
gradient_accumulation_steps: 2
```

Then train:
```bash
python scripts/train_expert.py \
  --expert python \
  --config configs/python.yaml \
  --dataset-dir datasets \
  --output checkpoints
```

---

## Performance Targets

```
Programming Languages: 90-95% accuracy
Data & Databases:      87-92% accuracy
Science & Math:        83-90% accuracy
Frontend:              89-94% accuracy
Backend:               87-93% accuracy
Cloud:                 87-92% accuracy
Terminal:              87-92% accuracy
Testing:               85-91% accuracy
Security:              82-94% accuracy
ML/AI:                 86-91% accuracy
Specialized:           82-91% accuracy
Domain:                82-91% accuracy

Latency: 2-5ms per request
Binary Size: 7-15MB per expert (fp16)
```

---

## Next Steps

1. **Start with core training**: `build_MoE.bat`
2. **Expand as needed**: `python scripts/train_all_experts.py --group <group>`
3. **Deploy when ready**: `python scripts/create_expert_registry.py`
4. **Monitor & optimize**: Check logs & validation reports

**Ready to train 40+ specialists!** 🚀

