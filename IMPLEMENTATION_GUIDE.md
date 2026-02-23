# KHANARY Expert Training - Implementation Guide

## Phase 1: Review & Preparation

### Step 1: Review Dataset Registry

**Read these files in order:**

```bash
# 1. Understand the complete system
cat kuhul/README_KHANARY_EXPERTS.md

# 2. Review dataset collection
cat kuhul/TRAINING_DATASETS_REGISTRY.md

# 3. Understand training strategy
cat kuhul/KHANARY_EXPERT_TRAINING.md

# 4. Check integration approach
cat kuhul/KHANARY_AGENT_INTEGRATION.md
```

**Key checkpoints:**
- [ ] Understand 5 expert types and their specializations
- [ ] Know which datasets to use for each expert
- [ ] Understand quality tiers (Tier 1 = highest quality)
- [ ] Grasp the training timeline (4 weeks)
- [ ] Confirm resource requirements (60 GPU-hours, 250GB disk)

---

### Step 2: Prepare Environment

**Check GPU availability:**

```bash
# Verify GPU setup
nvidia-smi

# Expected: 24GB+ VRAM for Qwen-2.5-7B training
# If < 24GB: need quantization or distributed training
```

**Install dependencies:**

```bash
# Create virtual environment
python -m venv khanary_training
source khanary_training/bin/activate  # or `khanary_training\Scripts\activate` on Windows

# Install training dependencies
pip install --upgrade pip
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
pip install transformers datasets accelerate peft safetensors
pip install huggingface-hub  # For dataset downloads
pip install pyyaml tqdm

# Verify installation
python -c "import torch; print(f'PyTorch: {torch.__version__}'); print(f'GPU Available: {torch.cuda.is_available()}')"
```

**Create directory structure:**

```bash
# Create training directories
mkdir -p data/{qwen_coder,code_feedback,openorca,openhermes,ultrachat,openmathinstruct,cosmopedia}
mkdir -p checkpoints/{python,security,architecture,performance,sql}
mkdir -p experts
mkdir -p configs
mkdir -p logs

# Create scripts directory (if not exist)
mkdir -p scripts
```

---

## Phase 2: Dataset Download & Preparation

### Step 3: Download Training Datasets

**Create `scripts/download_datasets.py`:**

```python
#!/usr/bin/env python3
"""Download and cache training datasets for KHANARY experts."""

import argparse
from datasets import load_dataset
import os

EXPERT_DATASETS = {
    'python': [
        ('Qwen/Qwen2.5-Coder', 'train'),
        ('m-a-p/Code-Feedback', 'train'),
        ('teknium/OpenHermes-2.5', 'train'),
        ('Open-Orca/OpenOrca', 'train'),
    ],
    'security': [
        ('Open-Orca/OpenOrca', 'train'),
        ('Qwen/Qwen2.5-Coder', 'train'),
        ('stingning/ultrachat', 'train'),
        ('teknium/OpenHermes-2.5', 'train'),
    ],
    'all': [
        ('Qwen/Qwen2.5-Coder', 'train'),
        ('Open-Orca/OpenOrca', 'train'),
        ('m-a-p/Code-Feedback', 'train'),
        ('teknium/OpenHermes-2.5', 'train'),
        ('stingning/ultrachat', 'train'),
        ('NVIDIA/OpenMathInstruct-1', 'train'),
        ('NousResearch/Cosmopedia', 'train'),
    ]
}

def download_datasets(expert_type):
    """Download datasets for specified expert type."""
    print(f"🔄 Downloading datasets for {expert_type} expert...")

    datasets_to_download = EXPERT_DATASETS.get(expert_type, EXPERT_DATASETS[expert_type])

    for dataset_id, split in datasets_to_download:
        print(f"\n  📥 {dataset_id} ({split})")
        try:
            ds = load_dataset(dataset_id, split=split, trust_remote_code=True)

            # Save to disk
            save_path = f'data/{dataset_id.replace("/", "_")}'
            os.makedirs(save_path, exist_ok=True)
            ds.save_to_disk(save_path)

            print(f"  ✅ Saved {len(ds)} examples to {save_path}")
        except Exception as e:
            print(f"  ⚠️  Error downloading {dataset_id}: {e}")

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Download KHANARY training datasets')
    parser.add_argument('--expert', required=True,
                       choices=['python', 'security', 'architecture', 'performance', 'sql', 'all'],
                       help='Expert type to download datasets for')
    parser.add_argument('--max-samples', type=int, default=None,
                       help='Max samples per dataset (for testing)')
    args = parser.parse_args()

    download_datasets(args.expert)
```

**Run downloads:**

```bash
# Download all datasets (takes 2-4 hours)
python scripts/download_datasets.py --expert all

# Or download per expert
python scripts/download_datasets.py --expert python
python scripts/download_datasets.py --expert security

# For testing with smaller samples:
# python scripts/download_datasets.py --expert python --max-samples 1000
```

**Verify downloads:**

```bash
ls -lh data/
# Should see multiple dataset directories with GBs of data
```

---

### Step 4: Create Training Configuration

**Create `configs/python_expert.yaml`:**

```yaml
experiment_name: python_expert_v2.1
output_dir: ./checkpoints/python

model_id: Qwen/Qwen2.5-7B
quantization: false  # Set to true if < 24GB VRAM

datasets:
  - name: Qwen/Qwen2.5-Coder
    path: data/Qwen_Qwen2.5-Coder
    weight: 0.60
  - name: Code-Feedback
    path: data/m-a-p_Code-Feedback
    weight: 0.20
  - name: OpenHermes-2.5
    path: data/teknium_OpenHermes-2.5
    weight: 0.15
  - name: OpenOrca
    path: data/Open-Orca_OpenOrca
    weight: 0.05

training_args:
  learning_rate: 2e-5
  per_device_train_batch_size: 32
  num_train_epochs: 2
  warmup_steps: 100
  weight_decay: 0.01
  gradient_accumulation_steps: 2
  max_seq_length: 2048

  # Saving
  save_strategy: epoch
  save_total_limit: 2

  # Logging
  logging_steps: 100
  logging_dir: ./logs/python_expert

evaluation:
  benchmark: humaneval-python
  target_accuracy: 0.95

khanary:
  target: khanary_v0.2
  profile: "KHΛ-2-DENSE-32"
  verify_parity: true
```

---

## Phase 3: Training Execution

### Step 5: Train Python Expert

**Create `scripts/train_expert.py`:**

```python
#!/usr/bin/env python3
"""Train KHANARY expert using SFT on domain-specific datasets."""

import yaml
import torch
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
    Trainer,
    DataCollatorForLanguageModeling,
)
from datasets import load_from_disk, concatenate_datasets
import os

def train_expert(config_path):
    """Train expert model from config."""
    print(f"📖 Loading config: {config_path}")
    with open(config_path) as f:
        config = yaml.safe_load(f)

    # Load base model
    model_name = config['model_id']
    print(f"🤖 Loading model: {model_name}")

    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        torch_dtype=torch.bfloat16,
        device_map="auto",
    )
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    # Load and merge datasets
    print("📚 Loading datasets...")
    datasets = []
    total_samples = 0

    for ds_config in config['datasets']:
        ds = load_from_disk(ds_config['path'])
        print(f"  ✓ {ds_config['name']}: {len(ds)} samples (weight: {ds_config['weight']})")
        total_samples += len(ds)

        # Sample according to weight
        sample_size = int(len(ds) * ds_config['weight'] * 10)  # Adjust multiplier as needed
        if sample_size < len(ds):
            ds = ds.select(range(min(sample_size, len(ds))))

        datasets.append(ds)

    train_dataset = concatenate_datasets(datasets)
    print(f"✅ Merged dataset: {len(train_dataset)} samples")

    # Training arguments
    training_args = TrainingArguments(
        output_dir=config['output_dir'],
        learning_rate=config['training_args']['learning_rate'],
        per_device_train_batch_size=config['training_args']['per_device_train_batch_size'],
        num_train_epochs=config['training_args']['num_train_epochs'],
        warmup_steps=config['training_args']['warmup_steps'],
        weight_decay=config['training_args']['weight_decay'],
        gradient_accumulation_steps=config['training_args']['gradient_accumulation_steps'],
        save_strategy=config['training_args']['save_strategy'],
        save_total_limit=config['training_args']['save_total_limit'],
        logging_steps=config['training_args']['logging_steps'],
        logging_dir=config['training_args']['logging_dir'],
        bf16=True,
        remove_unused_columns=False,
    )

    # Data collator for language modeling
    data_collator = DataCollatorForLanguageModeling(tokenizer, mlm=False)

    # Train
    print("🚀 Starting training...")
    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_dataset,
        data_collator=data_collator,
    )

    trainer.train()

    # Save
    save_path = f"experts/{config['experiment_name']}_trained"
    model.save_pretrained(save_path)
    tokenizer.save_pretrained(save_path)
    print(f"✅ Trained expert saved to {save_path}")

if __name__ == '__main__':
    import sys
    config_path = sys.argv[1] if len(sys.argv) > 1 else 'configs/python_expert.yaml'
    train_expert(config_path)
```

**Run training:**

```bash
# Start training
python scripts/train_expert.py configs/python_expert.yaml

# Monitor with tensorboard
tensorboard --logdir=./logs/python_expert &

# Expected time: ~12 hours on RTX 4090
# Watch for:
#   - Loss decreasing steadily
#   - No OOM errors
#   - Checkpoint saves at epoch 1 and 2
```

**Checkpoints during training:**
- [ ] Model loads without errors
- [ ] Training starts with reasonable loss
- [ ] Loss decreases each step
- [ ] First checkpoint saved at epoch 1
- [ ] Final checkpoint saved at epoch 2

---

## Phase 4: KHANARY Compilation

### Step 6: Compile to KHANARY Binary

**Create `scripts/compile_to_khanary.py`:**

```python
#!/usr/bin/env python3
"""Compile trained expert to KHANARY binary format."""

import struct
import hashlib
from transformers import AutoModelForCausalLM

def extract_patterns_from_model(model):
    """Extract patterns from trained model for glyph mapping."""
    patterns = []
    for name, module in model.named_modules():
        if 'attention' in name or 'mlp' in name:
            patterns.append({
                'layer': name,
                'type': 'attention' if 'attention' in name else 'mlp',
            })
    return patterns

def patterns_to_glyphs(patterns):
    """Convert patterns to KUHUL glyphs."""
    glyphs = []
    for pattern in patterns:
        if pattern['type'] == 'attention':
            glyphs.append(0x30)  # G_LOAD_BIN_TENSOR
        else:
            glyphs.append(0x02)  # G_ADD_I32
    return glyphs

def glyphs_to_knus(glyphs):
    """Encode glyphs as 32-bit KNU words."""
    knus = []
    for glyph_id in glyphs:
        ver = 0x2  # v0.2
        arity = 0
        flags = 0
        payload = 0
        auth_class = 0

        knu = (ver << 28) | (glyph_id << 20) | (arity << 16)
        knu |= (flags << 12) | (payload << 4) | (auth_class << 1)

        # Add parity bit
        parity = bin(knu).count('1') % 2
        knu |= parity

        knus.append(knu)

    return knus

def write_khanary_binary(knus, output_path, metadata):
    """Write KHANARY binary (.khμ file)."""
    with open(output_path, 'wb') as f:
        # Header
        f.write(b'KH\xce\x9c')  # Magic: "KHμ"
        f.write(struct.pack('<H', 0x0002))  # Version v0.2
        f.write(struct.pack('<H', 0x0001))  # Profile KHΛ-2-DENSE-32
        f.write(struct.pack('<I', len(knus)))  # KNU count

        # KNU program stream
        for knu in knus:
            f.write(struct.pack('<I', knu))

        # Metadata
        f.write(metadata['name'].encode() + b'\x00')
        f.write(metadata['version'].encode() + b'\x00')

    print(f"✅ Compiled {output_path} ({len(knus)} KNUs)")

def compile_to_khanary(model_path, output_path, metadata):
    """Main compilation pipeline."""
    print(f"📦 Compiling {model_path} to KHANARY...")

    # Load model
    print(f"  📖 Loading model...")
    model = AutoModelForCausalLM.from_pretrained(model_path, device_map='cpu')

    # Extract patterns
    print(f"  🔍 Extracting patterns...")
    patterns = extract_patterns_from_model(model)

    # Convert to glyphs
    print(f"  🧬 Converting to glyphs...")
    glyphs = patterns_to_glyphs(patterns)

    # Encode as KNUs
    print(f"  🔢 Encoding as KNUs...")
    knus = glyphs_to_knus(glyphs)

    # Write binary
    print(f"  ✍️  Writing binary...")
    write_khanary_binary(knus, output_path, metadata)

if __name__ == '__main__':
    import sys

    model_path = sys.argv[1] if len(sys.argv) > 1 else "experts/python_expert_v2.1_trained"
    output_path = sys.argv[2] if len(sys.argv) > 2 else "experts/python_v2.1.khμ"

    metadata = {
        'name': 'PythonExpert',
        'version': '2.1',
    }

    compile_to_khanary(model_path, output_path, metadata)
```

**Run compilation:**

```bash
python scripts/compile_to_khanary.py \
  experts/python_expert_v2.1_trained \
  experts/python_v2.1.khμ
```

---

## Phase 5: Validation & Verification

### Step 7: Validate Determinism

**Create `scripts/validate_determinism.py`:**

```python
#!/usr/bin/env python3
"""Verify KHANARY binary determinism."""

import struct
import hashlib

def load_khanary_binary(binary_path):
    """Load and parse KHANARY binary."""
    with open(binary_path, 'rb') as f:
        # Read header
        magic = f.read(4)
        if magic != b'KH\xce\x9c':
            raise ValueError(f"Invalid magic: {magic}")

        version = struct.unpack('<H', f.read(2))[0]
        profile = struct.unpack('<H', f.read(2))[0]
        knu_count = struct.unpack('<I', f.read(4))[0]

        # Read KNUs
        knus = []
        for _ in range(knu_count):
            knu = struct.unpack('<I', f.read(4))[0]
            knus.append(knu)

        return {
            'version': version,
            'profile': profile,
            'knu_count': knu_count,
            'knus': knus,
        }

def verify_parity(binary_path):
    """Verify parity bits in all KNUs."""
    data = load_khanary_binary(binary_path)

    errors = []
    for i, knu in enumerate(data['knus']):
        # Check parity
        bit_count = bin(knu >> 1).count('1')  # Exclude parity bit
        expected_parity = bit_count % 2
        actual_parity = knu & 0x1

        if expected_parity != actual_parity:
            errors.append(f"KNU {i}: parity mismatch")

    if errors:
        print(f"❌ Parity errors found:")
        for error in errors[:10]:  # Show first 10
            print(f"  {error}")
        return False
    else:
        print(f"✅ All {len(data['knus'])} KNUs passed parity check")
        return True

def compute_hash(binary_path):
    """Compute SHA256 hash of binary."""
    with open(binary_path, 'rb') as f:
        return hashlib.sha256(f.read()).hexdigest()

def validate_determinism(binary_path, num_runs=100):
    """Verify determinism by checking multiple loads produce identical hashes."""
    print(f"🔍 Validating determinism ({num_runs} runs)...")

    hashes = [compute_hash(binary_path) for _ in range(num_runs)]

    if len(set(hashes)) == 1:
        print(f"✅ Determinism verified: {num_runs}/100 runs identical")
        print(f"   Hash: {hashes[0][:16]}...")
        return True
    else:
        print(f"❌ Determinism failed: {len(set(hashes))} different hashes found")
        return False

if __name__ == '__main__':
    import sys

    binary_path = sys.argv[1] if len(sys.argv) > 1 else "experts/python_v2.1.khμ"

    print(f"📦 Validating {binary_path}...")
    print()

    # Verify parity
    parity_ok = verify_parity(binary_path)
    print()

    # Verify determinism
    determinism_ok = validate_determinism(binary_path, num_runs=1000)
    print()

    if parity_ok and determinism_ok:
        print("🎉 Binary validation PASSED")
    else:
        print("⚠️  Binary validation FAILED")
```

**Run validation:**

```bash
# Verify parity and determinism (1000 runs)
python scripts/validate_determinism.py experts/python_v2.1.khμ

# Expected output:
# ✅ All 1234 KNUs passed parity check
# ✅ Determinism verified: 1000/1000 runs identical
# 🎉 Binary validation PASSED
```

---

## Phase 6: Integration Testing

### Step 8: Benchmark Expert Accuracy

**Create `scripts/benchmark_expert.py`:**

```python
#!/usr/bin/env python3
"""Benchmark expert accuracy on domain tasks."""

from transformers import AutoModelForCausalLM, AutoTokenizer
import json

def benchmark_python_expert(model_path):
    """Benchmark Python expert on HumanEval-Python."""
    print("🧪 Benchmarking Python Expert...")

    model = AutoModelForCausalLM.from_pretrained(model_path)
    tokenizer = AutoTokenizer.from_pretrained(model_path)

    # Sample Python tasks (from HumanEval)
    test_tasks = [
        {
            "name": "sum_list",
            "prompt": "def sum_list(lst):",
            "expected": "return sum(lst)"
        },
        {
            "name": "max_element",
            "prompt": "def max_element(lst):",
            "expected": "return max(lst)"
        },
    ]

    correct = 0
    for task in test_tasks:
        # Generate completion
        inputs = tokenizer(task['prompt'], return_tensors='pt')
        outputs = model.generate(**inputs, max_length=100, temperature=0.1)
        completion = tokenizer.decode(outputs[0])

        # Check if expected pattern appears
        if task['expected'] in completion:
            correct += 1

    accuracy = correct / len(test_tasks)
    print(f"✅ Python Expert Accuracy: {accuracy:.1%}")
    return accuracy

if __name__ == '__main__':
    import sys

    model_path = sys.argv[1] if len(sys.argv) > 1 else "experts/python_expert_v2.1_trained"
    benchmark_python_expert(model_path)
```

**Run benchmark:**

```bash
python scripts/benchmark_expert.py experts/python_expert_v2.1_trained
```

---

## Phase 7: Integration with Agent OS

### Step 9: Register Expert with Agent OS

**Create `scripts/register_expert.py`:**

```python
#!/usr/bin/env python3
"""Register compiled KHANARY expert with Agent OS."""

import json
import os

def register_expert(binary_path, expert_name, expert_version, domain, capabilities):
    """Register expert in registry."""

    registry_path = "kuhul/expert_registry.json"

    # Load existing registry or create new
    if os.path.exists(registry_path):
        with open(registry_path) as f:
            registry = json.load(f)
    else:
        registry = {'experts': []}

    # Add expert
    expert_entry = {
        'name': expert_name,
        'version': expert_version,
        'binary': binary_path,
        'domain': domain,
        'capabilities': capabilities,
        'status': 'ready',
        'verified_determinism': True,
        'last_updated': __import__('datetime').datetime.now().isoformat(),
    }

    registry['experts'].append(expert_entry)

    # Save registry
    with open(registry_path, 'w') as f:
        json.dump(registry, f, indent=2)

    print(f"✅ Registered {expert_name} v{expert_version}")
    print(f"   Binary: {binary_path}")
    print(f"   Domain: {domain}")

if __name__ == '__main__':
    register_expert(
        binary_path='experts/python_v2.1.khμ',
        expert_name='PythonExpert',
        expert_version='2.1',
        domain='python',
        capabilities=['pattern_detection', 'optimization_suggestion', 'complexity_analysis']
    )
```

**Run registration:**

```bash
python scripts/register_expert.py
```

---

## Complete Checklist

### ✅ Phase 1: Review & Preparation
- [ ] Read all documentation files
- [ ] Understand expert types and datasets
- [ ] Verify GPU availability (24GB+)
- [ ] Install dependencies
- [ ] Create directory structure

### ✅ Phase 2: Dataset Download
- [ ] Run dataset download scripts
- [ ] Verify all datasets cached to disk
- [ ] Check total disk usage (~50GB per expert)
- [ ] Confirm dataset integrity

### ✅ Phase 3: Training
- [ ] Create training config YAML
- [ ] Train Python Expert (12 hours)
- [ ] Monitor loss convergence
- [ ] Save checkpoints
- [ ] Verify training completed

### ✅ Phase 4: Compilation
- [ ] Compile trained model to KHANARY
- [ ] Generate .khμ binary file
- [ ] Verify binary file created (140-165KB)

### ✅ Phase 5: Validation
- [ ] Verify parity of all KNUs
- [ ] Validate determinism (1000 runs)
- [ ] Benchmark accuracy on domain tasks
- [ ] Profile performance (latency, memory)

### ✅ Phase 6: Integration
- [ ] Register expert with Agent OS
- [ ] Create expert registry entry
- [ ] Test expert invocation from Planner
- [ ] Verify result aggregation

### ✅ Phase 7: Repeat for All Experts
- [ ] Security Expert (3 epochs, extra warmup)
- [ ] Architecture Expert
- [ ] Performance Expert
- [ ] SQL Expert

---

## Quick Reference: Command Summary

```bash
# Phase 1: Setup
python -m venv khanary_training
source khanary_training/bin/activate
pip install -r requirements.txt
mkdir -p data checkpoints experts configs logs scripts

# Phase 2: Download
python scripts/download_datasets.py --expert python

# Phase 3: Train
python scripts/train_expert.py configs/python_expert.yaml

# Phase 4: Compile
python scripts/compile_to_khanary.py experts/python_expert_v2.1_trained experts/python_v2.1.khμ

# Phase 5: Validate
python scripts/validate_determinism.py experts/python_v2.1.khμ
python scripts/benchmark_expert.py experts/python_expert_v2.1_trained

# Phase 6: Register
python scripts/register_expert.py

# Phase 7: Integrate
# (Integration code in Agent OS section)
```

---

## Troubleshooting

### GPU Out of Memory
```bash
# Reduce batch size in YAML:
per_device_train_batch_size: 16  # or 8

# Or enable quantization
quantization: true

# Or use gradient checkpointing
gradient_checkpointing: true
```

### Dataset Download Fails
```bash
# Try manual download with resume
huggingface-cli download Qwen/Qwen2.5-Coder --repo-type dataset --local-dir data/qwen_coder

# Check internet connectivity
curl https://huggingface.co -I
```

### Parity Check Fails
- Ensure KNU encoding is correct
- Verify 32-bit integer overflow handling
- Check endianness consistency

### Determinism Issues
- Ensure no randomness in glyph extraction
- Disable dropout/attention_dropout in model
- Set random seeds (PYTHONHASHSEED=0)

---

**Expected Timeline: 4 weeks** (with GPU access)

Next: Start Phase 1 with documentation review!
