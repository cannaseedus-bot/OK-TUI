#!/usr/bin/env python3
"""
Download HuggingFace datasets for KHANARY expert training.
Supports: Python, Security, Architecture, Performance, SQL experts.
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


# Dataset configurations for each expert
EXPERT_DATASETS: Dict[str, List[Dict]] = {
    'python': [
        {
            'name': 'openai_humaneval',
            'repo': 'openai_humaneval',
            'split': 'test',
            'config': None,
            'description': 'HumanEval - Python code generation benchmark'
        },
        {
            'name': 'CodeSearchNet',
            'repo': 'code_search_net',
            'split': 'train',
            'config': 'python',
            'description': 'CodeSearchNet Python corpus'
        },
        {
            'name': 'CodeXGLUE',
            'repo': 'glue',
            'split': 'validation',
            'config': 'code_search_net',
            'description': 'CodeXGLUE code search dataset'
        },
        {
            'name': 'The Stack Python',
            'repo': 'bigcode/the-stack-dedup',
            'split': 'train',
            'config': 'data/python',
            'description': 'The Stack - Python code (first 10k samples)'
        }
    ],
    'security': [
        {
            'name': 'SecurityEval',
            'repo': 'khhuang/SecurityEval',
            'split': 'test',
            'config': None,
            'description': 'SecurityEval - vulnerability detection dataset'
        },
        {
            'name': 'OWASP Top 10',
            'repo': 'olivercliff/owasp-top-10-vulnerabilities',
            'split': 'train',
            'config': None,
            'description': 'OWASP Top 10 vulnerability examples'
        },
        {
            'name': 'CVE Descriptions',
            'repo': 'mrwadler/cve-descriptions',
            'split': 'train',
            'config': None,
            'description': 'CVE vulnerability descriptions'
        }
    ],
    'architecture': [
        {
            'name': 'DesignPatterns',
            'repo': 'code-search-net/design-patterns',
            'split': 'train',
            'config': None,
            'description': 'Software design patterns examples'
        },
        {
            'name': 'System Design',
            'repo': 'cassandrapeters/system_design_interview_questions',
            'split': 'train',
            'config': None,
            'description': 'System design interview questions'
        },
        {
            'name': 'The Stack (Architecture)',
            'repo': 'bigcode/the-stack-dedup',
            'split': 'train',
            'config': 'data/typescript',
            'description': 'The Stack - TypeScript architecture patterns'
        }
    ],
    'performance': [
        {
            'name': 'PerformanceBench',
            'repo': 'performancelabs/performancebench',
            'split': 'train',
            'config': None,
            'description': 'Performance optimization benchmarks'
        },
        {
            'name': 'Optimization Tips',
            'repo': 'code-search-net/optimization-tips',
            'split': 'train',
            'config': None,
            'description': 'Code optimization patterns'
        },
        {
            'name': 'The Stack (C++)',
            'repo': 'bigcode/the-stack-dedup',
            'split': 'train',
            'config': 'data/cpp',
            'description': 'The Stack - C++ performance-critical code'
        }
    ],
    'sql': [
        {
            'name': 'Spider',
            'repo': 'spider',
            'split': 'train',
            'config': None,
            'description': 'Spider - Text-to-SQL dataset'
        },
        {
            'name': 'WikiSQL',
            'repo': 'wikisql',
            'split': 'train',
            'config': None,
            'description': 'WikiSQL - Wikipedia SQL queries'
        },
        {
            'name': 'BIRD',
            'repo': 'bird',
            'split': 'train',
            'config': None,
            'description': 'BIRD - Big Real-world Database for SQL'
        }
    ]
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

    def download_expert_datasets(self, expert: str) -> bool:
        """
        Download all datasets for an expert.

        Args:
            expert: Expert name (python, security, etc.)

        Returns:
            True if successful, False otherwise
        """
        if expert not in EXPERT_DATASETS:
            logger.error(f"Unknown expert: {expert}")
            logger.info(f"Available: {', '.join(EXPERT_DATASETS.keys())}")
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
            output_file = expert_dir / f"{name.replace(' ', '_')}.parquet"
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
        results = {}
        for expert in EXPERT_DATASETS.keys():
            results[expert] = self.download_expert_datasets(expert)

        # Summary
        logger.info("\n" + "="*60)
        logger.info("Dataset Download Summary")
        logger.info("="*60)
        for expert, success in results.items():
            status = "✓" if success else "✗"
            logger.info(f"  {status} {expert}")

        logger.info(f"\nTotal statistics:")
        logger.info(f"  Downloaded: {self.stats['downloaded']} datasets")
        logger.info(f"  Failed: {self.stats['failed']} datasets")
        logger.info(f"  Total size: {self.stats['total_size'] / (1024**2):.1f}MB")
        logger.info("="*60)

        return all(results.values())


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Download HuggingFace datasets for KHANARY expert training'
    )
    parser.add_argument(
        '--expert',
        type=str,
        choices=list(EXPERT_DATASETS.keys()) + ['all'],
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

    args = parser.parse_args()

    # Create downloader
    downloader = DatasetDownloader(
        output_dir=args.output,
        max_samples=args.max_samples
    )

    # Download
    if args.expert == 'all':
        success = downloader.download_all_experts()
    else:
        success = downloader.download_expert_datasets(args.expert)

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
