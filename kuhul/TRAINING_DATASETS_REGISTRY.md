# KHANARY Training Datasets Registry

Curated collection of high-quality open-source datasets for training domain-specific KHANARY expert LLMs.

---

## 📊 Code Datasets

### Qwen-2.5-Coder (2M examples)
- **URL**: https://huggingface.co/datasets/Qwen/Qwen2.5-Coder
- **Quality**: Frontier-level coder model training data
- **Size**: 2M+ code examples
- **Languages**: Python, JavaScript, Go, Rust, SQL, Java, C++, etc.
- **Focus**: Production code patterns, best practices
- **Use For**:
  - Python Expert (60%)
  - SQL Expert (60%)
  - Security Expert (30%)
  - Performance Expert (35%)
- **Citation**: Qwen Team, 2024

### Code-Feedback (200K examples)
- **URL**: https://huggingface.co/datasets/m-a-p/Code-Feedback
- **Quality**: Unit-test verified correctness loops
- **Size**: 200K code + feedback examples
- **Focus**: Correctness verification, test-driven patterns
- **Characteristics**:
  - Code snippet with corresponding unit test
  - Feedback indicating correctness/error
  - Edge cases and error conditions
- **Use For**:
  - Python Expert (20%)
  - SQL Expert (25%)
  - Security Expert (quality verification)
- **Citation**: MAP Research, 2024

### Code-Alpaca (20K examples)
- **URL**: https://huggingface.co/datasets/sahil2801/CodeAlpaca-20k
- **Quality**: Code instruction-following
- **Size**: 20K examples
- **Focus**: Simple code tasks, basic patterns
- **Use For**: Supplementary data for smaller experts

---

## 🧠 Reasoning & Math Datasets

### OpenMathInstruct-1 (10M examples)
- **URL**: https://huggingface.co/datasets/NVIDIA/OpenMathInstruct-1
- **Quality**: Chain-of-thought mathematical reasoning
- **Size**: 10M+ mathematical reasoning examples
- **Focus**: Problem solving, optimization, complexity analysis
- **Characteristics**:
  - Math problems with step-by-step solutions
  - Complexity analysis and Big-O notation
  - Algorithm optimization reasoning
- **Use For**:
  - Performance Expert (45%)
  - Architecture Expert (reasoning component)
  - SQL Expert (15%, query optimization)
- **Citation**: NVIDIA Research, 2024

### Cosmopedia-Full (30M documents)
- **URL**: https://huggingface.co/datasets/NousResearch/Cosmopedia
- **Quality**: Synthetic high-diversity knowledge
- **Size**: 30M+ documents
- **Domains**: STEM, writing, instruction-following, general
- **Focus**: Broad knowledge, diverse reasoning
- **Characteristics**:
  - Synthetic but high-quality content
  - Multiple domains covered
  - Various reasoning styles
- **Use For**:
  - Architecture Expert (50%)
  - Performance Expert (20%)
  - General knowledge (all experts)
- **Citation**: Nous Research, 2024

### Cosmopedia-100K (100K documents)
- **URL**: https://huggingface.co/datasets/NousResearch/Cosmopedia-100k
- **Quality**: Distilled version of Cosmopedia
- **Size**: 100K curated documents
- **Focus**: High-value reasoning samples
- **Use For**: Quick training runs, prototyping

---

## 📚 SFT (Supervised Fine-Tuning) Datasets

### OpenOrca (1M examples)
- **URL**: https://huggingface.co/datasets/Open-Orca/OpenOrca
- **Quality**: GPT-4 distilled instruction-following
- **Size**: 1M examples
- **Characteristics**:
  - High-quality reasoning explanations
  - Diverse task coverage
  - Complex problem solving
- **Use For**:
  - Security Expert (50%, reasoning quality)
  - Architecture Expert (35%)
  - All experts (quality boost)
  - Foundation for reasoning tasks
- **Citation**: OpenOrca Team, 2024

### OpenHermes-2.5 (1M examples)
- **URL**: https://huggingface.co/datasets/teknium/OpenHermes-2.5
- **Quality**: Clean, diverse instruction dataset
- **Size**: 1M curated examples
- **Used By**: Top open-source models (Mistral, Nous, etc.)
- **Focus**: Broad instruction coverage
- **Use For**:
  - Python Expert (15%)
  - All experts (diversity)
  - General instruction following
- **Citation**: Teknium, 2024

### UltraChat (1M examples)
- **URL**: https://huggingface.co/datasets/stingning/ultrachat
- **Quality**: Multi-turn conversational data
- **Size**: 1M turns of dialogue
- **Characteristics**:
  - Multi-turn conversations
  - Complex reasoning dialogues
  - Various topics and domains
- **Use For**:
  - Security Expert (15%, security dialogue)
  - Architecture Expert (15%, design discussions)
  - Dialogue understanding (all experts)
- **Citation**: UltraChat Team, 2024

### SlimOrca (700K examples)
- **URL**: https://huggingface.co/datasets/Open-Orca/SlimOrca-500K
- **Quality**: Distilled OpenOrca (highest quality only)
- **Size**: 700K examples
- **Focus**: Highest-quality reasoning
- **Use For**: Premium training for critical experts (Security, Architecture)

---

## 🌍 Multilingual Datasets

### AYA Dataset (13B tokens, 100+ languages)
- **URL**: https://huggingface.co/datasets/CohereForAI/aya_dataset
- **Quality**: Multilingual instruction dataset
- **Size**: 13B tokens across 100+ languages
- **Focus**: Global coverage, diverse languages
- **Use For**: Future expansion to multilingual experts
- **Citation**: Cohere For AI, 2024

---

## 📋 Meta Collections

### LLM Datasets Registry (Curated List)
- **URL**: https://github.com/mlabonne/llm-datasets
- **Description**: Comprehensive list of datasets for LLM training
- **Purpose**: Discovery and benchmarking
- **Maintained By**: mlabonne

### ProjectPro LLM List (Categorized)
- **URL**: https://www.projectpro.io/article/llm-datasets/1137
- **Description**: Datasets categorized by type (text, code, reasoning)
- **Purpose**: Structured dataset discovery

### BrightCoding List (Pipeline-Focused)
- **URL**: https://brightcoding.dev/blog/llm-datasets
- **Description**: SFT and RLHF datasets for training pipelines
- **Purpose**: Implementation-focused recommendations

### Analytics Vidhya LLM List
- **URL**: https://www.analyticsvidhya.com/blog/2023/10/top-10-open-source-datasets-for-llm-training/
- **Description**: Overview of open-source pretraining corpora
- **Purpose**: Pretraining strategy and selection

---

## 🎯 Recommended Mixes by Expert Type

### Python Expert (1.2M examples)
```yaml
primary:
  - qwen3_coder: 60%              (1.2M * 0.60 = 720K)
secondary:
  - code_feedback: 20%             (1.2M * 0.20 = 240K)
  - openhermes_25: 15%            (1.2M * 0.15 = 180K)
tertiary:
  - openorca: 5%                   (1.2M * 0.05 = 60K)

Total: ~1.2M examples
Training epochs: 2
Base model: Qwen-2.5-7B
```

### Security Expert (800K examples)
```yaml
primary:
  - openorca: 50%                  (800K * 0.50 = 400K)
secondary:
  - qwen3_coder: 30%               (800K * 0.30 = 240K)
  - ultrachat: 15%                 (800K * 0.15 = 120K)
tertiary:
  - openhermes_25: 5%              (800K * 0.05 = 40K)

Total: ~800K examples
Training epochs: 3 (safety-critical)
Base model: Qwen-2.5-7B
Target False Negative Rate: <1%
```

### Architecture Expert (450K examples)
```yaml
primary:
  - cosmopedia_full: 50%           (450K * 0.50 = 225K)
secondary:
  - openorca: 35%                  (450K * 0.35 = 157.5K)
  - ultrachat: 15%                 (450K * 0.15 = 67.5K)

Total: ~450K examples
Training epochs: 2
Base model: Qwen-2.5-7B
Focus: Reasoning and design patterns
```

### Performance Expert (500K examples)
```yaml
primary:
  - openmathinstruct_1: 45%        (500K * 0.45 = 225K)
secondary:
  - qwen3_coder: 35%               (500K * 0.35 = 175K)
  - cosmopedia_full: 20%           (500K * 0.20 = 100K)

Total: ~500K examples
Training epochs: 2
Base model: Qwen-2.5-7B
Focus: Optimization and complexity analysis
```

### SQL Expert (600K examples)
```yaml
primary:
  - qwen3_coder: 60%               (600K * 0.60 = 360K)
secondary:
  - code_feedback: 25%             (600K * 0.25 = 150K)
  - openmathinstruct_1: 15%        (600K * 0.15 = 90K)

Total: ~600K examples
Training epochs: 2
Base model: Qwen-2.5-7B
Focus: Query optimization and schema design
```

---

## 📥 Download & Setup

### Bulk Download Script

```bash
#!/bin/bash
# Download all training datasets

DATASETS=(
    "Qwen/Qwen2.5-Coder"
    "m-a-p/Code-Feedback"
    "Open-Orca/OpenOrca"
    "teknium/OpenHermes-2.5"
    "stingning/ultrachat"
    "NVIDIA/OpenMathInstruct-1"
    "NousResearch/Cosmopedia"
    "NousResearch/Cosmopedia-100k"
    "CohereForAI/aya_dataset"
)

mkdir -p data
for dataset in "${DATASETS[@]}"; do
    echo "Downloading $dataset..."
    huggingface-cli download $dataset --repo-type dataset --local-dir "data/$dataset"
done
```

### Python Download Script

```python
from datasets import load_dataset
import os

datasets = {
    'qwen3_coder': ('Qwen/Qwen2.5-Coder', 'train'),
    'code_feedback': ('m-a-p/Code-Feedback', 'train'),
    'openorca': ('Open-Orca/OpenOrca', 'train'),
    'openhermes_25': ('teknium/OpenHermes-2.5', 'train'),
    'ultrachat': ('stingning/ultrachat', 'train'),
    'openmathinstruct_1': ('NVIDIA/OpenMathInstruct-1', 'train'),
    'cosmopedia': ('NousResearch/Cosmopedia', 'train'),
}

os.makedirs('data', exist_ok=True)

for name, (repo_id, split) in datasets.items():
    print(f"Downloading {name}...")
    ds = load_dataset(repo_id, split=split, streaming=True)
    ds.save_to_disk(f'data/{name}')
    print(f"  ✓ Saved to data/{name}")
```

---

## 📊 Dataset Statistics

### Size Comparison

| Dataset | Size | Type | Best For |
|---------|------|------|----------|
| Qwen-Coder | 2M | Code | Production code patterns |
| Cosmopedia | 30M | Reasoning | Broad knowledge base |
| OpenMathInstruct | 10M | Math | Optimization, complexity |
| OpenOrca | 1M | SFT | Quality reasoning |
| OpenHermes | 1M | SFT | General instruction |
| UltraChat | 1M | Dialogue | Multi-turn reasoning |
| Code-Feedback | 200K | Code+Test | Correctness verification |
| AYA | 13B tokens | Multilingual | Global coverage |

### Quality Tiers

```
Tier 1 (Highest Quality): OpenOrca, SlimOrca
  └─ Use: Critical experts (Security, Architecture)
  └─ Training: 3+ epochs

Tier 2 (High Quality): Qwen-Coder, Code-Feedback, Cosmopedia
  └─ Use: Main training data for all experts
  └─ Training: 2-3 epochs

Tier 3 (Good Quality): OpenHermes, UltraChat, OpenMathInstruct
  └─ Use: Supplementary, diversity
  └─ Training: 1-2 epochs

Tier 4 (Broad Coverage): AYA, general web data
  └─ Use: Future expansion, pretraining
  └─ Training: Base model only
```

---

## 🎓 Dataset Selection Strategy

### For Production Experts

1. **Start with Tier 1** (OpenOrca for reasoning quality)
2. **Add domain-specific Tier 2** (Qwen-Coder for code)
3. **Supplement with Tier 3** (OpenHermes for diversity)
4. **Weight by domain** (60% domain, 40% general knowledge)

### For Experimental Experts

1. **Use smaller versions** (Cosmopedia-100K instead of full)
2. **Mix multiple sources** (broad coverage)
3. **Rapid iteration** (2 epochs max)

### For Specialized Domains

1. **Filter by topic** (security content only)
2. **Use Tier 1+2** (highest quality)
3. **Extra epochs** (3+ for safety-critical)

---

## 📝 Citation & Attribution

When using these datasets, please cite:

```bibtex
@dataset{qwen2024coder,
  title={Qwen 2.5 Coder Dataset},
  author={Qwen Team},
  year={2024},
  url={https://huggingface.co/datasets/Qwen/Qwen2.5-Coder}
}

@dataset{openorca2024,
  title={Open Orca: An Open Dataset of GPT-4 Instruction-Following},
  author={OpenOrca Team},
  year={2024},
  url={https://huggingface.co/datasets/Open-Orca/OpenOrca}
}

@dataset{cosmopedia2024,
  title={Cosmopedia: A Large-Scale Synthetic High-Diversity Knowledge Corpus},
  author={Nous Research},
  year={2024},
  url={https://huggingface.co/datasets/NousResearch/Cosmopedia}
}

@dataset{openmathinstruct2024,
  title={OpenMathInstruct-1: A Large Scale Math Instruction Tuning Dataset},
  author={NVIDIA Research},
  year={2024},
  url={https://huggingface.co/datasets/NVIDIA/OpenMathInstruct-1}
}
```

---

## 🔗 Related Resources

- **HuggingFace Datasets Library**: https://huggingface.co/datasets
- **Open LLM Leaderboard**: https://huggingface.co/spaces/HuggingFaceH4/open_llm_leaderboard
- **Dataset Card Templates**: https://huggingface.co/docs/datasets/dataset_card
- **Training Best Practices**: https://huggingface.co/docs/transformers/training

---

## 📞 Support & Questions

For dataset-specific questions:
- Check HuggingFace Datasets Hub discussions
- Review dataset cards for known issues
- Reference original papers for methodology

For KHANARY training questions:
- See `KHANARY_EXPERT_TRAINING.md`
- Review training configurations
- Check validation results

---

**Last Updated**: 2026-02-23
**Maintained By**: K'UHUL Team
**License**: See individual dataset licenses
