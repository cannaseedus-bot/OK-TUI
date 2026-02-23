#!/usr/bin/env python3
"""
Download HuggingFace datasets for KHANARY expert training.
Supports 40+ domain-specific, tool-specific, and application-specific experts.
"""

import os
import sys
import json
import logging
import argparse
from pathlib import Path
from typing import Dict, List

try:
    from datasets import load_dataset, concatenate_datasets
    from huggingface_hub import hf_hub_download
except ImportError:
    print("Error: Required packages not installed.")
    print("Install with: pip install datasets huggingface_hub")
    sys.exit(1)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# Comprehensive dataset configurations for 40+ experts
EXPERT_DATASETS: Dict[str, List[Dict]] = {
    # ========================================================================
    # CORE PROGRAMMING LANGUAGES (8 experts)
    # ========================================================================
    'python': [
        {'name': 'openai_humaneval', 'repo': 'openai_humaneval', 'split': 'test', 'config': None, 'description': 'Python code generation'},
        {'name': 'CodeSearchNet Python', 'repo': 'code_search_net', 'split': 'train', 'config': 'python', 'description': 'Python corpus'},
        {'name': 'CodeXGLUE', 'repo': 'codexglue', 'split': 'train', 'config': None, 'description': 'Code understanding benchmark'},
        {'name': 'The Stack Python', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/python', 'description': 'Raw Python code'},
    ],
    'javascript': [
        {'name': 'JavaScript HumanEval', 'repo': 'openai_humaneval', 'split': 'test', 'config': None, 'description': 'JS code generation'},
        {'name': 'The Stack JavaScript', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/javascript', 'description': 'Raw JavaScript code'},
        {'name': 'Web API Docs', 'repo': 'huggingface-projects/web-api-docs', 'split': 'train', 'config': None, 'description': 'JavaScript Web APIs'},
        {'name': 'Node.js Patterns', 'repo': 'bigcode/nodejs-patterns', 'split': 'train', 'config': None, 'description': 'Node.js patterns'},
    ],
    'java': [
        {'name': 'CodeSearchNet Java', 'repo': 'code_search_net', 'split': 'train', 'config': 'java', 'description': 'Java corpus'},
        {'name': 'The Stack Java', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/java', 'description': 'Raw Java code'},
        {'name': 'Java Design Patterns', 'repo': 'bigcode/java-design-patterns', 'split': 'train', 'config': None, 'description': 'Java patterns'},
    ],
    'cpp': [
        {'name': 'CodeSearchNet C++', 'repo': 'code_search_net', 'split': 'train', 'config': 'cpp', 'description': 'C++ corpus'},
        {'name': 'The Stack C++', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/cpp', 'description': 'Raw C++ code'},
        {'name': 'C++ Performance Tips', 'repo': 'bigcode/cpp-performance-tips', 'split': 'train', 'config': None, 'description': 'C++ optimization'},
    ],
    'rust': [
        {'name': 'The Stack Rust', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/rust', 'description': 'Raw Rust code'},
        {'name': 'Rust by Example', 'repo': 'rust-lang/rust-by-example', 'split': 'train', 'config': None, 'description': 'Rust examples'},
        {'name': 'Rust Patterns', 'repo': 'bigcode/rust-patterns', 'split': 'train', 'config': None, 'description': 'Rust idiomatic patterns'},
    ],
    'go': [
        {'name': 'The Stack Go', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/go', 'description': 'Raw Go code'},
        {'name': 'Go Patterns', 'repo': 'bigcode/go-patterns', 'split': 'train', 'config': None, 'description': 'Go idioms'},
    ],
    'csharp': [
        {'name': 'The Stack C#', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/csharp', 'description': 'Raw C# code'},
        {'name': 'C# Patterns', 'repo': 'bigcode/csharp-patterns', 'split': 'train', 'config': None, 'description': 'C# patterns'},
    ],
    'typescript': [
        {'name': 'The Stack TypeScript', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/typescript', 'description': 'Raw TypeScript'},
        {'name': 'TypeScript Patterns', 'repo': 'bigcode/typescript-patterns', 'split': 'train', 'config': None, 'description': 'TS patterns'},
    ],

    # ========================================================================
    # DATA & DATABASES (4 experts)
    # ========================================================================
    'sql': [
        {'name': 'Spider', 'repo': 'spider', 'split': 'train', 'config': None, 'description': 'Text-to-SQL dataset'},
        {'name': 'WikiSQL', 'repo': 'wikisql', 'split': 'train', 'config': None, 'description': 'Wikipedia SQL queries'},
        {'name': 'BIRD', 'repo': 'bird', 'split': 'train', 'config': None, 'description': 'Big Real-world Database SQL'},
    ],
    'data_engineering': [
        {'name': 'ETL Pipelines', 'repo': 'bigcode/etl-examples', 'split': 'train', 'config': None, 'description': 'ETL pipeline patterns'},
        {'name': 'Data Processing', 'repo': 'bigcode/data-processing', 'split': 'train', 'config': None, 'description': 'Data transformation code'},
        {'name': 'Apache Spark', 'repo': 'bigcode/spark-examples', 'split': 'train', 'config': None, 'description': 'Spark patterns'},
    ],
    'nosql': [
        {'name': 'MongoDB Patterns', 'repo': 'bigcode/mongodb-patterns', 'split': 'train', 'config': None, 'description': 'MongoDB queries'},
        {'name': 'Redis Patterns', 'repo': 'bigcode/redis-patterns', 'split': 'train', 'config': None, 'description': 'Redis operations'},
    ],
    'data_science': [
        {'name': 'Kaggle Datasets', 'repo': 'kaggle/datasets-meta', 'split': 'train', 'config': None, 'description': 'Kaggle data science'},
        {'name': 'Pandas Examples', 'repo': 'bigcode/pandas-examples', 'split': 'train', 'config': None, 'description': 'Data analysis code'},
        {'name': 'Statistics', 'repo': 'bigcode/statistics-tutorials', 'split': 'train', 'config': None, 'description': 'Statistical analysis'},
    ],

    # ========================================================================
    # SCIENCE & MATHEMATICS (6 experts)
    # ========================================================================
    'physics': [
        {'name': 'Physics Papers', 'repo': 'arxiv-physics', 'split': 'train', 'config': None, 'description': 'Physics papers abstracts'},
        {'name': 'Quantum Computing', 'repo': 'bigcode/quantum-computing', 'split': 'train', 'config': None, 'description': 'Quantum algorithms'},
        {'name': 'Physics Formulas', 'repo': 'bigcode/physics-formulas', 'split': 'train', 'config': None, 'description': 'Physics equations'},
    ],
    'chemistry': [
        {'name': 'Chemistry Papers', 'repo': 'arxiv-chemistry', 'split': 'train', 'config': None, 'description': 'Chemistry research'},
        {'name': 'Molecular Properties', 'repo': 'bigcode/molecule-properties', 'split': 'train', 'config': None, 'description': 'Chemical data'},
        {'name': 'Periodic Table', 'repo': 'bigcode/periodic-table', 'split': 'train', 'config': None, 'description': 'Element information'},
    ],
    'mathematics': [
        {'name': 'Math Problems', 'repo': 'openai/mathematics', 'split': 'train', 'config': None, 'description': 'Math problem solving'},
        {'name': 'Calculus', 'repo': 'bigcode/calculus-examples', 'split': 'train', 'config': None, 'description': 'Calculus problems'},
        {'name': 'Linear Algebra', 'repo': 'bigcode/linear-algebra', 'split': 'train', 'config': None, 'description': 'LA concepts'},
    ],
    'biology': [
        {'name': 'Biology Papers', 'repo': 'arxiv-biology', 'split': 'train', 'config': None, 'description': 'Biology research'},
        {'name': 'Bioinformatics', 'repo': 'bigcode/bioinformatics', 'split': 'train', 'config': None, 'description': 'DNA/protein analysis'},
    ],
    'space': [
        {'name': 'Astronomy Papers', 'repo': 'arxiv-astronomy', 'split': 'train', 'config': None, 'description': 'Astronomy research'},
        {'name': 'Space Missions', 'repo': 'bigcode/space-missions', 'split': 'train', 'config': None, 'description': 'Space exploration'},
    ],
    'engineering': [
        {'name': 'Engineering Papers', 'repo': 'arxiv-engineering', 'split': 'train', 'config': None, 'description': 'Engineering research'},
        {'name': 'CAD Models', 'repo': 'bigcode/cad-models', 'split': 'train', 'config': None, 'description': 'CAD specifications'},
    ],

    # ========================================================================
    # DOMAIN-SPECIFIC APPLICATIONS (5 experts)
    # ========================================================================
    'healthcare': [
        {'name': 'Medical Papers', 'repo': 'pubmed-abstracts', 'split': 'train', 'config': None, 'description': 'Medical research'},
        {'name': 'Clinical Notes', 'repo': 'mimic-iii-notes', 'split': 'train', 'config': None, 'description': 'Patient records'},
        {'name': 'Medical Terminology', 'repo': 'bigcode/medical-terms', 'split': 'train', 'config': None, 'description': 'Medical knowledge'},
    ],
    'finance': [
        {'name': 'Financial Papers', 'repo': 'arxiv-finance', 'split': 'train', 'config': None, 'description': 'Finance research'},
        {'name': 'Trading Code', 'repo': 'bigcode/trading-algorithms', 'split': 'train', 'config': None, 'description': 'Trading strategies'},
        {'name': 'Accounting', 'repo': 'bigcode/accounting-concepts', 'split': 'train', 'config': None, 'description': 'Accounting practices'},
    ],
    'legal': [
        {'name': 'Legal Documents', 'repo': 'legal-texts', 'split': 'train', 'config': None, 'description': 'Legal documents'},
        {'name': 'Case Law', 'repo': 'case-law-corpus', 'split': 'train', 'config': None, 'description': 'Court decisions'},
        {'name': 'Contract Analysis', 'repo': 'bigcode/contract-clauses', 'split': 'train', 'config': None, 'description': 'Contract patterns'},
    ],
    'business': [
        {'name': 'Business Plans', 'repo': 'bigcode/business-plans', 'split': 'train', 'config': None, 'description': 'Business strategy'},
        {'name': 'Management', 'repo': 'bigcode/management-concepts', 'split': 'train', 'config': None, 'description': 'Management theory'},
        {'name': 'Marketing', 'repo': 'bigcode/marketing-strategies', 'split': 'train', 'config': None, 'description': 'Marketing techniques'},
    ],
    'education': [
        {'name': 'Educational Content', 'repo': 'bigcode/edu-content', 'split': 'train', 'config': None, 'description': 'Learning materials'},
        {'name': 'Tutorials', 'repo': 'bigcode/tutorials', 'split': 'train', 'config': None, 'description': 'Step-by-step guides'},
        {'name': 'Q&A', 'repo': 'stackoverflow-qa', 'split': 'train', 'config': None, 'description': 'Question-answer pairs'},
    ],

    # ========================================================================
    # FRONTEND & UI/UX (4 experts)
    # ========================================================================
    'react': [
        {'name': 'React Components', 'repo': 'bigcode/react-components', 'split': 'train', 'config': None, 'description': 'React patterns'},
        {'name': 'The Stack JavaScript (React)', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/javascript', 'description': 'React code'},
        {'name': 'React Hooks', 'repo': 'bigcode/react-hooks', 'split': 'train', 'config': None, 'description': 'Hooks patterns'},
    ],
    'vue': [
        {'name': 'Vue Components', 'repo': 'bigcode/vue-components', 'split': 'train', 'config': None, 'description': 'Vue patterns'},
        {'name': 'Vue Examples', 'repo': 'bigcode/vue-examples', 'split': 'train', 'config': None, 'description': 'Vue code'},
    ],
    'angular': [
        {'name': 'Angular Components', 'repo': 'bigcode/angular-components', 'split': 'train', 'config': None, 'description': 'Angular patterns'},
        {'name': 'TypeScript Angular', 'repo': 'bigcode/angular-typescript', 'split': 'train', 'config': None, 'description': 'TS in Angular'},
    ],
    'ux_design': [
        {'name': 'UX Research', 'repo': 'bigcode/ux-research', 'split': 'train', 'config': None, 'description': 'UX methodology'},
        {'name': 'UI Patterns', 'repo': 'bigcode/ui-patterns', 'split': 'train', 'config': None, 'description': 'Design patterns'},
        {'name': 'CSS Guidelines', 'repo': 'bigcode/css-guidelines', 'split': 'train', 'config': None, 'description': 'CSS best practices'},
        {'name': 'Accessibility', 'repo': 'bigcode/a11y-guidelines', 'split': 'train', 'config': None, 'description': 'WCAG guidelines'},
    ],

    # ========================================================================
    # BACKEND & INFRASTRUCTURE (5 experts)
    # ========================================================================
    'nodejs': [
        {'name': 'Node.js Patterns', 'repo': 'bigcode/nodejs-patterns', 'split': 'train', 'config': None, 'description': 'Node patterns'},
        {'name': 'Express.js', 'repo': 'bigcode/expressjs-examples', 'split': 'train', 'config': None, 'description': 'Express patterns'},
        {'name': 'Node.js Best Practices', 'repo': 'bigcode/nodejs-best-practices', 'split': 'train', 'config': None, 'description': 'Node practices'},
    ],
    'django': [
        {'name': 'Django Patterns', 'repo': 'bigcode/django-patterns', 'split': 'train', 'config': None, 'description': 'Django patterns'},
        {'name': 'Django ORM', 'repo': 'bigcode/django-orm', 'split': 'train', 'config': None, 'description': 'Django database'},
        {'name': 'Django Templates', 'repo': 'bigcode/django-templates', 'split': 'train', 'config': None, 'description': 'Template syntax'},
    ],
    'fastapi': [
        {'name': 'FastAPI Patterns', 'repo': 'bigcode/fastapi-patterns', 'split': 'train', 'config': None, 'description': 'FastAPI patterns'},
        {'name': 'AsyncIO', 'repo': 'bigcode/asyncio-patterns', 'split': 'train', 'config': None, 'description': 'Async programming'},
    ],
    'devops': [
        {'name': 'Docker Patterns', 'repo': 'bigcode/docker-patterns', 'split': 'train', 'config': None, 'description': 'Docker files'},
        {'name': 'Kubernetes', 'repo': 'bigcode/kubernetes-manifests', 'split': 'train', 'config': None, 'description': 'K8s YAML'},
        {'name': 'CI/CD Pipelines', 'repo': 'bigcode/cicd-pipelines', 'split': 'train', 'config': None, 'description': 'CI/CD configs'},
        {'name': 'Infrastructure as Code', 'repo': 'bigcode/terraform-patterns', 'split': 'train', 'config': None, 'description': 'Terraform/CloudFormation'},
    ],
    'microservices': [
        {'name': 'Microservice Patterns', 'repo': 'bigcode/microservice-patterns', 'split': 'train', 'config': None, 'description': 'Distributed systems'},
        {'name': 'API Design', 'repo': 'bigcode/api-design-patterns', 'split': 'train', 'config': None, 'description': 'REST/GraphQL'},
        {'name': 'Message Queues', 'repo': 'bigcode/message-queue-patterns', 'split': 'train', 'config': None, 'description': 'Queue implementations'},
    ],

    # ========================================================================
    # CLOUD PLATFORMS (3 experts)
    # ========================================================================
    'aws': [
        {'name': 'AWS CloudFormation', 'repo': 'bigcode/aws-cloudformation', 'split': 'train', 'config': None, 'description': 'AWS IaC'},
        {'name': 'AWS Lambda', 'repo': 'bigcode/aws-lambda', 'split': 'train', 'config': None, 'description': 'Serverless patterns'},
        {'name': 'AWS Best Practices', 'repo': 'bigcode/aws-best-practices', 'split': 'train', 'config': None, 'description': 'AWS optimization'},
    ],
    'gcp': [
        {'name': 'Google Cloud Deployment', 'repo': 'bigcode/gcp-deployment', 'split': 'train', 'config': None, 'description': 'GCP patterns'},
        {'name': 'BigQuery', 'repo': 'bigcode/bigquery-examples', 'split': 'train', 'config': None, 'description': 'SQL analytics'},
    ],
    'azure': [
        {'name': 'Azure ARM Templates', 'repo': 'bigcode/azure-arm', 'split': 'train', 'config': None, 'description': 'Azure IaC'},
        {'name': 'Azure Functions', 'repo': 'bigcode/azure-functions', 'split': 'train', 'config': None, 'description': 'Serverless on Azure'},
    ],

    # ========================================================================
    # TERMINAL & SHELL (3 experts)
    # ========================================================================
    'bash': [
        {'name': 'Bash Scripts', 'repo': 'bigcode/bash-scripts', 'split': 'train', 'config': None, 'description': 'Shell scripting'},
        {'name': 'Unix Tools', 'repo': 'bigcode/unix-tools', 'split': 'train', 'config': None, 'description': 'Command-line tools'},
        {'name': 'System Administration', 'repo': 'bigcode/sysadmin-scripts', 'split': 'train', 'config': None, 'description': 'Admin automation'},
    ],
    'cli_tools': [
        {'name': 'CLI Patterns', 'repo': 'bigcode/cli-patterns', 'split': 'train', 'config': None, 'description': 'Command-line apps'},
        {'name': 'Argparse', 'repo': 'bigcode/argparse-examples', 'split': 'train', 'config': None, 'description': 'CLI argument parsing'},
        {'name': 'Terminal UI', 'repo': 'bigcode/terminal-ui', 'split': 'train', 'config': None, 'description': 'TUI libraries'},
    ],
    'git': [
        {'name': 'Git Workflows', 'repo': 'bigcode/git-workflows', 'split': 'train', 'config': None, 'description': 'Git patterns'},
        {'name': 'Git Hooks', 'repo': 'bigcode/git-hooks', 'split': 'train', 'config': None, 'description': 'Automation scripts'},
        {'name': 'GitHub Actions', 'repo': 'bigcode/github-actions', 'split': 'train', 'config': None, 'description': 'CI/CD workflows'},
    ],

    # ========================================================================
    # TESTING & QUALITY (3 experts)
    # ========================================================================
    'testing': [
        {'name': 'Unit Tests', 'repo': 'bigcode/unit-tests', 'split': 'train', 'config': None, 'description': 'Test patterns'},
        {'name': 'Jest', 'repo': 'bigcode/jest-examples', 'split': 'train', 'config': None, 'description': 'JavaScript testing'},
        {'name': 'Pytest', 'repo': 'bigcode/pytest-examples', 'split': 'train', 'config': None, 'description': 'Python testing'},
    ],
    'debugging': [
        {'name': 'Debugging Techniques', 'repo': 'bigcode/debugging-guide', 'split': 'train', 'config': None, 'description': 'Debug patterns'},
        {'name': 'Logging', 'repo': 'bigcode/logging-patterns', 'split': 'train', 'config': None, 'description': 'Log management'},
        {'name': 'Profiling', 'repo': 'bigcode/profiling-tools', 'split': 'train', 'config': None, 'description': 'Performance profiling'},
    ],
    'monitoring': [
        {'name': 'Observability', 'repo': 'bigcode/observability', 'split': 'train', 'config': None, 'description': 'Monitoring setup'},
        {'name': 'Prometheus', 'repo': 'bigcode/prometheus-config', 'split': 'train', 'config': None, 'description': 'Metrics collection'},
        {'name': 'ELK Stack', 'repo': 'bigcode/elk-setup', 'split': 'train', 'config': None, 'description': 'Log aggregation'},
    ],

    # ========================================================================
    # SECURITY & ADVANCED (4 experts)
    # ========================================================================
    'security': [
        {'name': 'SecurityEval', 'repo': 'khhuang/SecurityEval', 'split': 'test', 'config': None, 'description': 'Vulnerability detection'},
        {'name': 'OWASP Top 10', 'repo': 'olivercliff/owasp-top-10-vulnerabilities', 'split': 'train', 'config': None, 'description': 'Vulnerability examples'},
        {'name': 'CVE Descriptions', 'repo': 'mrwadler/cve-descriptions', 'split': 'train', 'config': None, 'description': 'CVE data'},
    ],
    'cryptography': [
        {'name': 'Cryptography Algorithms', 'repo': 'bigcode/cryptography', 'split': 'train', 'config': None, 'description': 'Crypto patterns'},
        {'name': 'TLS/SSL', 'repo': 'bigcode/tls-ssl', 'split': 'train', 'config': None, 'description': 'Encryption setup'},
    ],
    'blockchain': [
        {'name': 'Smart Contracts', 'repo': 'bigcode/smart-contracts', 'split': 'train', 'config': None, 'description': 'Solidity code'},
        {'name': 'Web3', 'repo': 'bigcode/web3-patterns', 'split': 'train', 'config': None, 'description': 'Blockchain interaction'},
    ],
    'iot': [
        {'name': 'IoT Code', 'repo': 'bigcode/iot-patterns', 'split': 'train', 'config': None, 'description': 'IoT device code'},
        {'name': 'Embedded Systems', 'repo': 'bigcode/embedded-systems', 'split': 'train', 'config': None, 'description': 'Firmware patterns'},
    ],

    # ========================================================================
    # MACHINE LEARNING & AI (3 experts)
    # ========================================================================
    'machine_learning': [
        {'name': 'TensorFlow', 'repo': 'bigcode/tensorflow-examples', 'split': 'train', 'config': None, 'description': 'ML models'},
        {'name': 'PyTorch', 'repo': 'bigcode/pytorch-examples', 'split': 'train', 'config': None, 'description': 'Deep learning'},
        {'name': 'Scikit-learn', 'repo': 'bigcode/sklearn-examples', 'split': 'train', 'config': None, 'description': 'Classical ML'},
    ],
    'nlp': [
        {'name': 'NLP Tasks', 'repo': 'bigcode/nlp-tasks', 'split': 'train', 'config': None, 'description': 'NLP code'},
        {'name': 'Transformers', 'repo': 'bigcode/huggingface-transformers', 'split': 'train', 'config': None, 'description': 'BERT/GPT code'},
        {'name': 'Text Processing', 'repo': 'bigcode/text-processing', 'split': 'train', 'config': None, 'description': 'NLTK/spaCy'},
    ],
    'computer_vision': [
        {'name': 'Computer Vision', 'repo': 'bigcode/computer-vision', 'split': 'train', 'config': None, 'description': 'OpenCV/PyTorch'},
        {'name': 'Image Processing', 'repo': 'bigcode/image-processing', 'split': 'train', 'config': None, 'description': 'Image algorithms'},
    ],

    # ========================================================================
    # SPECIALIZED TOOLS (4 experts)
    # ========================================================================
    'architecture': [
        {'name': 'DesignPatterns', 'repo': 'code-search-net/design-patterns', 'split': 'train', 'config': None, 'description': 'Design patterns'},
        {'name': 'System Design', 'repo': 'cassandrapeters/system_design_interview_questions', 'split': 'train', 'config': None, 'description': 'System design'},
        {'name': 'The Stack TypeScript', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/typescript', 'description': 'Architecture code'},
    ],
    'performance': [
        {'name': 'PerformanceBench', 'repo': 'performancelabs/performancebench', 'split': 'train', 'config': None, 'description': 'Performance benchmarks'},
        {'name': 'Optimization Tips', 'repo': 'code-search-net/optimization-tips', 'split': 'train', 'config': None, 'description': 'Optimization patterns'},
        {'name': 'The Stack C++', 'repo': 'bigcode/the-stack-dedup', 'split': 'train', 'config': 'data/cpp', 'description': 'Performance code'},
    ],
    'graphics': [
        {'name': 'Graphics Programming', 'repo': 'bigcode/graphics-code', 'split': 'train', 'config': None, 'description': 'OpenGL/Vulkan'},
        {'name': 'Game Engines', 'repo': 'bigcode/game-engine-code', 'split': 'train', 'config': None, 'description': 'Unity/Unreal'},
    ],
    'documentation': [
        {'name': 'API Documentation', 'repo': 'bigcode/api-docs', 'split': 'train', 'config': None, 'description': 'API writing'},
        {'name': 'Technical Writing', 'repo': 'bigcode/technical-docs', 'split': 'train', 'config': None, 'description': 'Doc templates'},
        {'name': 'README Files', 'repo': 'bigcode/readme-files', 'split': 'train', 'config': None, 'description': 'README patterns'},
    ],
}


class DatasetDownloader:
    """Download and cache KHANARY training datasets."""

    def __init__(self, output_dir: str, max_samples: int = None):
        """
        Initialize downloader.

        Args:
            output_dir: Directory to save datasets
            max_samples: Max samples per dataset (None = all)
        """
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.max_samples = max_samples
        self.stats = {
            'downloaded': 0,
            'failed': 0,
            'total_size': 0
        }

    def list_all_experts(self) -> List[str]:
        """List all available experts."""
        return sorted(EXPERT_DATASETS.keys())

    def download_expert_datasets(self, expert: str) -> bool:
        """
        Download all datasets for an expert.

        Args:
            expert: Expert name

        Returns:
            True if successful, False otherwise
        """
        if expert not in EXPERT_DATASETS:
            logger.error(f"Unknown expert: {expert}")
            logger.info(f"Available: {', '.join(self.list_all_experts())}")
            return False

        logger.info(f"Downloading datasets for {expert} expert...")

        expert_dir = self.output_dir / expert
        expert_dir.mkdir(parents=True, exist_ok=True)

        datasets_list = EXPERT_DATASETS[expert]
        success_count = 0

        for dataset_config in datasets_list:
            if self._download_dataset(expert, expert_dir, dataset_config):
                success_count += 1
            else:
                self.stats['failed'] += 1

        logger.info(f"Downloaded {success_count}/{len(datasets_list)} datasets for {expert}")
        return success_count > 0

    def _download_dataset(self, expert: str, expert_dir: Path, config: Dict) -> bool:
        """
        Download a single dataset.

        Args:
            expert: Expert name
            expert_dir: Directory for this expert
            config: Dataset configuration

        Returns:
            True if successful
        """
        name = config['name']
        repo = config['repo']
        split = config['split']
        dataset_config = config.get('config')
        description = config['description']

        try:
            logger.info(f"  • {name} ({description})...")

            # Load dataset
            if dataset_config:
                dataset = load_dataset(repo, dataset_config, split=split, trust_remote_code=True)
            else:
                dataset = load_dataset(repo, split=split, trust_remote_code=True)

            # Limit samples if requested
            if self.max_samples and len(dataset) > self.max_samples:
                logger.info(f"    Limiting to {self.max_samples} samples")
                dataset = dataset.select(range(self.max_samples))

            # Save to disk
            output_file = expert_dir / f"{name.replace(' ', '_').replace('/', '_')}.parquet"
            dataset.to_parquet(str(output_file))

            size_mb = output_file.stat().st_size / (1024 * 1024)
            logger.info(f"    ✓ Saved {len(dataset)} samples ({size_mb:.1f}MB)")

            self.stats['downloaded'] += 1
            self.stats['total_size'] += output_file.stat().st_size

            return True

        except Exception as e:
            logger.warning(f"    ✗ Failed to download: {e}")
            return False

    def download_all_experts(self) -> bool:
        """
        Download datasets for all experts.

        Returns:
            True if all successful
        """
        experts = self.list_all_experts()
        results = {}

        logger.info(f"Starting download for {len(experts)} experts...")
        logger.info("=" * 70)

        for expert in experts:
            results[expert] = self.download_expert_datasets(expert)

        # Summary
        logger.info("\n" + "="*70)
        logger.info(f"Dataset Download Summary ({len(experts)} experts)")
        logger.info("="*70)

        successful = sum(1 for v in results.values() if v)
        for expert, success in results.items():
            status = "✓" if success else "✗"
            logger.info(f"  {status} {expert:20}")

        logger.info(f"\nTotal statistics:")
        logger.info(f"  Successfully downloaded: {self.stats['downloaded']} datasets")
        logger.info(f"  Failed: {self.stats['failed']} datasets")
        logger.info(f"  Total size: {self.stats['total_size'] / (1024**2):.1f}MB")
        logger.info(f"  Success rate: {successful}/{len(experts)} experts")
        logger.info("="*70)

        return successful == len(experts)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Download HuggingFace datasets for KHANARY expert training'
    )
    parser.add_argument(
        '--expert',
        type=str,
        choices=['all'] + sorted(EXPERT_DATASETS.keys()),
        default='all',
        help='Expert to download datasets for (default: all)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='datasets',
        help='Output directory for datasets (default: datasets)'
    )
    parser.add_argument(
        '--max-samples',
        type=int,
        default=None,
        help='Maximum samples per dataset (None = all)'
    )
    parser.add_argument(
        '--list-experts',
        action='store_true',
        help='List all available experts'
    )

    args = parser.parse_args()

    # Create downloader
    downloader = DatasetDownloader(
        output_dir=args.output,
        max_samples=args.max_samples
    )

    # List experts if requested
    if args.list_experts:
        experts = downloader.list_all_experts()
        logger.info(f"\nAvailable KHANARY Experts ({len(experts)}):")
        logger.info("=" * 70)
        for i, expert in enumerate(experts, 1):
            dataset_count = len(EXPERT_DATASETS[expert])
            logger.info(f"  {i:2}. {expert:20} ({dataset_count} datasets)")
        logger.info("=" * 70)
        return 0

    # Download
    if args.expert == 'all':
        success = downloader.download_all_experts()
    else:
        success = downloader.download_expert_datasets(args.expert)

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
