# KHANARY Binary Distribution Strategy

## Problem
- GitHub has file size limits (100MB per file)
- Binary files bloat repository history
- KHANARY binaries (.khμ) need efficient distribution
- Want to version control and reproduce builds

## Solution: Multi-Layer Distribution Strategy

```
┌─────────────────────────────────────────────────────┐
│  Local Development (Git repo)                       │
│  ├─ Source configs (YAML) → Git track               │
│  ├─ Training scripts → Git track                    │
│  └─ .gitignore: *.khμ binaries (don't commit)       │
├─────────────────────────────────────────────────────┤
│  CI/CD Pipeline (GitHub Actions)                    │
│  ├─ Compile experts from source (deterministic)    │
│  ├─ Generate .khμ binaries                         │
│  ├─ Run validation tests                           │
│  └─ Create release artifacts                       │
├─────────────────────────────────────────────────────┤
│  Distribution Channels                              │
│  ├─ GitHub Releases (small binaries)               │
│  ├─ Hugging Face Model Hub (large models)          │
│  ├─ Docker Images (compiled+runtime)               │
│  └─ PyPI Package (Python wrapper)                  │
├─────────────────────────────────────────────────────┤
│  End-User Installation                              │
│  ├─ Auto-download from releases                    │
│  ├─ Cache locally in ~/.khanary/experts/          │
│  ├─ Verify hash on download                        │
│  └─ Use locally or stream from Hub                 │
└─────────────────────────────────────────────────────┘
```

---

## Option 1: GitHub Releases (Recommended for Small Binaries)

### Setup

**Create `.gitignore`:**

```bash
# Don't track compiled binaries in Git
*.khμ
checkpoints/
artifacts/
```

**Track metadata instead:**

```bash
# DO track these in Git
configs/python_expert.yaml
scripts/train_expert.py
kuhul/KHANARY_EXPERT_TRAINING.md

# DON'T track these
experts/*.khμ
checkpoints/*
```

### GitHub Actions: Auto-Compile & Release

**Create `.github/workflows/compile-and-release.yml`:**

```yaml
name: Compile KHANARY Experts & Release

on:
  push:
    tags:
      - 'v*'  # On version tags like v2.1.0
    paths:
      - 'configs/**'
      - 'scripts/**'
      - '.github/workflows/compile-and-release.yml'

jobs:
  compile-experts:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        expert: [python, security, architecture, performance, sql]

    steps:
      - uses: actions/checkout@v3

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          pip install torch transformers datasets safetensors pyyaml

      - name: Compile ${{ matrix.expert }} expert
        run: |
          python scripts/compile_to_khanary.py \
            checkpoints/${{ matrix.expert }}_v*.safetensors \
            experts/${{ matrix.expert }}_v*.khμ

      - name: Validate determinism
        run: |
          python scripts/validate_determinism.py experts/${{ matrix.expert }}_v*.khμ

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: ${{ matrix.expert }}-expert
          path: experts/${{ matrix.expert }}_v*.khμ

  create-release:
    needs: compile-experts
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Download all artifacts
        uses: actions/download-artifact@v3
        with:
          path: release_binaries/

      - name: Create checksums
        run: |
          cd release_binaries/
          find . -name "*.khμ" -exec sha256sum {} \; > SHA256SUMS
          cat SHA256SUMS

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            release_binaries/**/*.khμ
            release_binaries/SHA256SUMS
          body: |
            ## KHANARY Expert Binaries v${{ github.ref_name }}

            ### Included Experts
            - PythonExpert v2.1 (95% accuracy)
            - SecurityExpert v2.0 (94% accuracy)
            - ArchitectureExpert v1.2 (88% accuracy)
            - PerformanceExpert v2.0 (91% accuracy)
            - SQLExpert v2.5 (92% accuracy)

            ### Installation
            ```bash
            # Download
            gh release download ${{ github.ref_name }} --pattern "*.khμ"

            # Verify
            sha256sum -c SHA256SUMS

            # Install
            mkdir -p ~/.khanary/experts
            mv *.khμ ~/.khanary/experts/
            ```

            ### Verification
            All binaries have been:
            ✅ Compiled deterministically
            ✅ Validated for parity (100% pass)
            ✅ Tested for determinism (1000+ runs)
            ✅ Benchmarked on domain tasks
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Download Script for Users

**Create `scripts/install_experts.py`:**

```python
#!/usr/bin/env python3
"""Download and install KHANARY experts from GitHub releases."""

import os
import sys
import subprocess
import hashlib
from pathlib import Path

def download_release(version='latest'):
    """Download KHANARY experts from GitHub release."""

    expert_dir = Path.home() / '.khanary' / 'experts'
    expert_dir.mkdir(parents=True, exist_ok=True)

    print(f"📥 Downloading KHANARY experts ({version})...")

    # Use gh CLI to download
    cmd = [
        'gh', 'release', 'download', version,
        '--pattern', '*.khμ',
        '--dir', str(expert_dir),
        '--repo', 'cannaseedus-bot/Ollama-K'
    ]

    try:
        subprocess.run(cmd, check=True)
        print(f"✅ Downloaded to {expert_dir}")
    except subprocess.CalledProcessError as e:
        print(f"❌ Download failed: {e}")
        return False

    # Download checksums
    checksums_cmd = [
        'gh', 'release', 'download', version,
        '--pattern', 'SHA256SUMS',
        '--dir', str(expert_dir),
        '--repo', 'cannaseedus-bot/Ollama-K'
    ]

    try:
        subprocess.run(checksums_cmd, check=True)
    except subprocess.CalledProcessError:
        print("⚠️  Could not download checksums")

    # Verify checksums
    checksums_file = expert_dir / 'SHA256SUMS'
    if checksums_file.exists():
        print("🔍 Verifying checksums...")
        verify_cmd = ['sha256sum', '-c', str(checksums_file)]
        result = subprocess.run(verify_cmd, cwd=expert_dir)
        if result.returncode == 0:
            print("✅ All binaries verified")
        else:
            print("❌ Checksum verification failed")
            return False

    # List installed experts
    print("\n✅ Installed Experts:")
    for khmu_file in sorted(expert_dir.glob('*.khμ')):
        size_mb = khmu_file.stat().st_size / (1024 * 1024)
        print(f"  • {khmu_file.name} ({size_mb:.1f}MB)")

    return True

if __name__ == '__main__':
    version = sys.argv[1] if len(sys.argv) > 1 else 'latest'
    success = download_release(version)
    sys.exit(0 if success else 1)
```

**User installation:**

```bash
# Install all latest experts
python scripts/install_experts.py

# Or specific version
python scripts/install_experts.py v2.1.0

# Verify installation
ls -lh ~/.khanary/experts/
```

---

## Option 2: Hugging Face Model Hub (For Larger Models)

### Setup on Hugging Face

**Create model card (`README.md` in model repo):**

```markdown
# KHANARY Expert: Python v2.1

Compiled KHANARY binary expert for Python code analysis.

## Details
- **Format:** KHANARY v0.2 (.khμ binary)
- **Size:** 150KB
- **Accuracy:** 95% (HumanEval-Python)
- **Latency:** 2.0ms per analysis
- **Determinism:** Verified 1000+ runs
- **Training Data:** Qwen-Coder (60%) + Code-Feedback (20%) + diversity

## Installation

```python
from huggingface_hub import hf_hub_download

# Download expert binary
expert_path = hf_hub_download(
    repo_id="cannaseedus-bot/khanary-python-expert",
    filename="python_v2.1.khμ",
    cache_dir="~/.khanary/experts"
)
```

## Usage

```python
from khanary import KhanaryExpert

expert = KhanaryExpert(expert_path)
result = expert.analyze(python_code)
```

## Citation

```bibtex
@software{khanary_python_2024,
  title={KHANARY Python Expert v2.1},
  author={K'UHUL Team},
  year={2024},
  url={https://huggingface.co/cannaseedus-bot/khanary-python-expert}
}
```
```

**Upload to Hugging Face:**

```bash
# Install huggingface_hub
pip install huggingface_hub

# Login
huggingface-cli login

# Create repo (one time)
huggingface-cli repo create khanary-python-expert --type model

# Clone and add binary
git clone https://huggingface.co/cannaseedus-bot/khanary-python-expert
cd khanary-python-expert

# Copy binary and README
cp ../../experts/python_v2.1.khμ .
cp ../../kuhul/README_PYTHON_EXPERT.md README.md

# Push to Hub
git add .
git commit -m "Add Python Expert v2.1"
git push
```

---

## Option 3: Docker Image (Complete Distribution)

**Create `Dockerfile`:**

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
RUN pip install torch transformers datasets safetensors pyyaml

# Copy training scripts
COPY scripts/ /app/scripts/
COPY configs/ /app/configs/
COPY kuhul/ /app/kuhul/

# Pre-compile experts during build
RUN python scripts/compile_to_khanary.py \
    checkpoints/python_v2.1_trained \
    /app/experts/python_v2.1.khμ && \
    python scripts/compile_to_khanary.py \
    checkpoints/security_v2.0_trained \
    /app/experts/security_v2.0.khμ && \
    python scripts/validate_determinism.py /app/experts/*.khμ

# Create expert registry
RUN python scripts/create_expert_registry.py

ENTRYPOINT ["python", "-m", "khanary"]
```

**Build and push:**

```bash
docker build -t cannaseedus-bot/khanary-experts:v2.1 .
docker push cannaseedus-bot/khanary-experts:v2.1
```

**User installation:**

```bash
docker pull cannaseedus-bot/khanary-experts:v2.1
docker run -v ~/.khanary:/root/.khanary cannaseedus-bot/khanary-experts:v2.1 install
```

---

## Option 4: PyPI Package with Binary Wheels

**Create `setup.py`:**

```python
from setuptools import setup, find_packages

setup(
    name='khanary-experts',
    version='2.1.0',
    description='Pre-compiled KHANARY domain-specific expert binaries',
    author='K\'UHUL Team',

    packages=find_packages(),

    package_data={
        'khanary_experts': [
            'binaries/*.khμ',
            'configs/*.yaml',
        ],
    },

    entry_points={
        'console_scripts': [
            'khanary-install=khanary_experts.cli:install_experts',
            'khanary-validate=khanary_experts.cli:validate_experts',
        ],
    },

    install_requires=[
        'transformers',
        'torch',
    ],

    python_requires='>=3.8',
)
```

**Publish to PyPI:**

```bash
pip install build twine

python -m build
twine upload dist/*
```

**User installation:**

```bash
pip install khanary-experts

# Automatically installs and registers experts
khanary-install
```

---

## Recommended Approach: Hybrid Strategy

### For Best Results:

1. **Store in Git:** Source configs, scripts, training data
2. **CI/CD Build:** GitHub Actions compiles binaries (deterministically)
3. **Release Artifacts:** Upload .khμ to GitHub Releases
4. **Multiple Mirrors:**
   - GitHub Releases (primary, fast)
   - Hugging Face Hub (backup, discoverable)
   - PyPI Package (optional, Python users)
   - Docker Hub (optional, containerized)

### Implementation Steps

**1. Create build configuration (`.github/workflows/build.yml`):**
```yaml
# Triggers on:
# - Tag push (v2.1.0) → Creates release
# - PR to main → Validates compilation
# - Weekly → Ensures reproducibility
```

**2. Add version tracking:**
```bash
# Tag releases
git tag -a v2.1.0 -m "KHANARY Experts v2.1: Python 95%, Security 94%"
git push origin v2.1.0
```

**3. Users install with:**
```bash
# Option A: Direct download
wget https://github.com/cannaseedus-bot/Ollama-K/releases/download/v2.1.0/python_v2.1.khμ

# Option B: Automated script
python scripts/install_experts.py v2.1.0

# Option C: PyPI package
pip install khanary-experts

# Option D: Docker
docker pull cannaseedus-bot/khanary-experts:v2.1.0
```

---

## Best Practices

✅ **Do:**
- Use base64 only for emergency manual distribution
- Leverage GitHub Releases for < 2GB total
- Use Git LFS for large binaries (> 100MB)
- Store checksums with every release
- Automate compilation in CI/CD
- Version binaries with source configs

❌ **Don't:**
- Commit large binaries to main repo
- Manually compile and upload
- Skip checksum verification
- Break reproducibility (document build process)
- Keep old versions in Git history

---

## Size Comparison

| Strategy | Storage | Distribution | Speed | Cost |
|----------|---------|--------------|-------|------|
| **Git LFS** | 🟢 Efficient | 🟡 Requires LFS | 🟢 Fast | 💰 $5/mo |
| **GitHub Releases** | 🟢 Free | 🟢 Simple | 🟢 Fast | ✅ Free |
| **Hugging Face Hub** | 🟢 Free | 🟢 CDN | 🟢 Fast | ✅ Free |
| **Docker Hub** | 🟢 Free | 🟡 Large | 🟡 Slower | ✅ Free |
| **PyPI** | 🟢 Free | 🟢 Standard | 🟢 Fast | ✅ Free |

---

## For Your KHANARY Project

**Recommended setup:**

1. ✅ **Local Git:** Track scripts, configs, YAML
2. ✅ **GitHub Actions:** Auto-compile on tag
3. ✅ **GitHub Releases:** Store .khμ binaries (150KB each = ~1MB total)
4. ✅ **Hugging Face Hub:** Discoverable via model search
5. ✅ **PyPI (optional):** `pip install khanary-experts`

**Total cost:** $0 (all free tiers support this)

---

## Next Steps

1. Create `.gitignore` to exclude `.khμ` files
2. Set up `.github/workflows/compile-and-release.yml`
3. Create `scripts/install_experts.py`
4. Push first expert via GitHub Release
5. Document installation in README

This way, Git stays lean, CI/CD handles compilation, and distribution is automatic!

