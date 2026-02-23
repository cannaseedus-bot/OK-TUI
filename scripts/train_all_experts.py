#!/usr/bin/env python3
"""
Train all KHANARY experts in parallel or sequential mode.
Orchestrates training for 40+ domain-specific experts.
"""

import os
import sys
import json
import logging
import argparse
import subprocess
import time
from pathlib import Path
from typing import List, Dict
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# Expert groups for training
EXPERT_GROUPS = {
    'core': [
        'python', 'javascript', 'security', 'react', 'fastapi'
    ],
    'programming': [
        'python', 'javascript', 'java', 'cpp', 'rust', 'go', 'csharp', 'typescript'
    ],
    'data': [
        'sql', 'data_engineering', 'nosql', 'data_science'
    ],
    'science': [
        'physics', 'chemistry', 'mathematics', 'biology', 'space', 'engineering'
    ],
    'frontend': [
        'react', 'vue', 'angular', 'ux_design'
    ],
    'backend': [
        'nodejs', 'django', 'fastapi', 'devops', 'microservices'
    ],
    'cloud': [
        'aws', 'gcp', 'azure'
    ],
    'terminal': [
        'bash', 'cli_tools', 'git'
    ],
    'testing': [
        'testing', 'debugging', 'monitoring'
    ],
    'security_advanced': [
        'security', 'cryptography', 'blockchain', 'iot'
    ],
    'ml_ai': [
        'machine_learning', 'nlp', 'computer_vision'
    ],
    'specialized': [
        'architecture', 'performance', 'graphics', 'documentation'
    ],
    'domain_apps': [
        'healthcare', 'finance', 'legal', 'business', 'education'
    ],
    'all': []  # Will be populated with all experts
}

# Populate 'all' group
for group, experts in list(EXPERT_GROUPS.items()):
    if group != 'all' and group != 'core':
        EXPERT_GROUPS['all'].extend(experts)


class ExpertTrainer:
    """Orchestrate training for multiple KHANARY experts."""

    def __init__(
        self,
        dataset_dir: str = 'datasets',
        output_dir: str = 'checkpoints',
        logs_dir: str = 'logs',
        max_workers: int = 1,
        epochs: int = 3,
        batch_size: int = 32,
        learning_rate: float = 2e-5
    ):
        """
        Initialize trainer.

        Args:
            dataset_dir: Directory with datasets
            output_dir: Checkpoint output directory
            logs_dir: Logs directory
            max_workers: Max parallel training jobs
            epochs: Training epochs
            batch_size: Batch size
            learning_rate: Learning rate
        """
        self.dataset_dir = Path(dataset_dir)
        self.output_dir = Path(output_dir)
        self.logs_dir = Path(logs_dir)
        self.max_workers = max_workers
        self.epochs = epochs
        self.batch_size = batch_size
        self.learning_rate = learning_rate

        # Create directories
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.logs_dir.mkdir(parents=True, exist_ok=True)

        self.results = {
            'successful': [],
            'failed': [],
            'skipped': [],
            'total_time': 0
        }

    def check_dataset_exists(self, expert: str) -> bool:
        """Check if dataset exists for expert."""
        expert_dir = self.dataset_dir / expert
        if not expert_dir.exists():
            logger.warning(f"Dataset not found for {expert}: {expert_dir}")
            return False
        return True

    def train_expert(self, expert: str) -> Dict:
        """
        Train single expert.

        Args:
            expert: Expert name

        Returns:
            Training result dictionary
        """
        logger.info(f"Starting training for {expert}...")

        # Check dataset
        if not self.check_dataset_exists(expert):
            logger.warning(f"Skipping {expert} - no dataset")
            return {'expert': expert, 'status': 'skipped', 'reason': 'no_dataset'}

        # Prepare command
        log_file = self.logs_dir / f"train_{expert}.log"
        cmd = [
            sys.executable,
            'scripts/train_expert.py',
            '--expert', expert,
            '--dataset-dir', str(self.dataset_dir),
            '--output', str(self.output_dir),
            '--epochs', str(self.epochs),
            '--batch-size', str(self.batch_size),
            '--learning-rate', str(self.learning_rate),
            '--log-file', str(log_file)
        ]

        try:
            start_time = time.time()

            # Run training
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=3600  # 1 hour timeout
            )

            elapsed = time.time() - start_time

            if result.returncode == 0:
                logger.info(f"✓ {expert} trained successfully ({elapsed:.1f}s)")
                return {
                    'expert': expert,
                    'status': 'success',
                    'time': elapsed,
                    'log': str(log_file)
                }
            else:
                logger.error(f"✗ {expert} training failed")
                logger.error(f"  Error: {result.stderr}")
                return {
                    'expert': expert,
                    'status': 'failed',
                    'error': result.stderr[:500],
                    'log': str(log_file)
                }

        except subprocess.TimeoutExpired:
            logger.error(f"✗ {expert} training timeout (1 hour)")
            return {
                'expert': expert,
                'status': 'failed',
                'error': 'timeout'
            }
        except Exception as e:
            logger.error(f"✗ {expert} training error: {e}")
            return {
                'expert': expert,
                'status': 'failed',
                'error': str(e)
            }

    def train_experts(self, experts: List[str], parallel: bool = False) -> bool:
        """
        Train multiple experts.

        Args:
            experts: List of expert names
            parallel: Train in parallel (requires multiple workers)

        Returns:
            True if all successful
        """
        if not experts:
            logger.error("No experts to train")
            return False

        logger.info(f"\n{'='*70}")
        logger.info(f"Training {len(experts)} experts ({self.max_workers} workers)")
        logger.info(f"Configuration:")
        logger.info(f"  Epochs: {self.epochs}")
        logger.info(f"  Batch Size: {self.batch_size}")
        logger.info(f"  Learning Rate: {self.learning_rate}")
        logger.info(f"  Mode: {'Parallel' if parallel else 'Sequential'}")
        logger.info(f"{'='*70}\n")

        start_time = time.time()

        if parallel and self.max_workers > 1:
            # Parallel training
            with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
                futures = {
                    executor.submit(self.train_expert, expert): expert
                    for expert in experts
                }

                for future in as_completed(futures):
                    result = future.result()
                    self._record_result(result)

        else:
            # Sequential training
            for expert in experts:
                result = self.train_expert(expert)
                self._record_result(result)

        total_time = time.time() - start_time
        self.results['total_time'] = total_time

        return self._print_summary(total_time)

    def _record_result(self, result: Dict):
        """Record training result."""
        status = result.get('status', 'unknown')

        if status == 'success':
            self.results['successful'].append(result)
        elif status == 'failed':
            self.results['failed'].append(result)
        else:
            self.results['skipped'].append(result)

    def _print_summary(self, total_time: float) -> bool:
        """Print training summary."""
        logger.info("\n" + "="*70)
        logger.info("Training Summary")
        logger.info("="*70)

        successful = len(self.results['successful'])
        failed = len(self.results['failed'])
        skipped = len(self.results['skipped'])
        total = successful + failed + skipped

        logger.info(f"\nResults:")
        logger.info(f"  ✓ Successful: {successful}/{total}")
        logger.info(f"  ✗ Failed:     {failed}/{total}")
        logger.info(f"  ⊘ Skipped:    {skipped}/{total}")
        logger.info(f"  Total time:   {total_time/60:.1f} minutes")

        if successful > 0:
            logger.info(f"\nSuccessfully trained:")
            for result in self.results['successful']:
                time_str = f"{result.get('time', 0):.1f}s" if 'time' in result else "N/A"
                logger.info(f"  ✓ {result['expert']:20} ({time_str})")

        if failed > 0:
            logger.info(f"\nFailed:")
            for result in self.results['failed']:
                error = result.get('error', 'unknown')
                logger.info(f"  ✗ {result['expert']:20} ({error})")

        if skipped > 0:
            logger.info(f"\nSkipped:")
            for result in self.results['skipped']:
                reason = result.get('reason', 'unknown')
                logger.info(f"  ⊘ {result['expert']:20} ({reason})")

        logger.info("="*70)

        return failed == 0

    def save_results(self, output_file: str):
        """Save training results to JSON."""
        output_path = Path(output_file)
        output_path.parent.mkdir(parents=True, exist_ok=True)

        with open(output_path, 'w') as f:
            json.dump(self.results, f, indent=2)

        logger.info(f"Results saved to {output_file}")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Train all KHANARY experts'
    )
    parser.add_argument(
        '--group',
        type=str,
        choices=list(EXPERT_GROUPS.keys()),
        default='core',
        help='Expert group to train (default: core)'
    )
    parser.add_argument(
        '--experts',
        type=str,
        help='Comma-separated list of experts (overrides --group)'
    )
    parser.add_argument(
        '--dataset-dir',
        type=str,
        default='datasets',
        help='Dataset directory (default: datasets)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='checkpoints',
        help='Checkpoint output directory (default: checkpoints)'
    )
    parser.add_argument(
        '--logs-dir',
        type=str,
        default='logs',
        help='Logs directory (default: logs)'
    )
    parser.add_argument(
        '--workers',
        type=int,
        default=1,
        help='Number of parallel workers (default: 1)'
    )
    parser.add_argument(
        '--parallel',
        action='store_true',
        help='Enable parallel training'
    )
    parser.add_argument(
        '--epochs',
        type=int,
        default=3,
        help='Training epochs (default: 3)'
    )
    parser.add_argument(
        '--batch-size',
        type=int,
        default=32,
        help='Batch size (default: 32)'
    )
    parser.add_argument(
        '--learning-rate',
        type=float,
        default=2e-5,
        help='Learning rate (default: 2e-5)'
    )
    parser.add_argument(
        '--results',
        type=str,
        help='Save results to JSON file'
    )
    parser.add_argument(
        '--list-groups',
        action='store_true',
        help='List all expert groups'
    )

    args = parser.parse_args()

    # List groups if requested
    if args.list_groups:
        logger.info("\nAvailable expert groups:")
        logger.info("="*70)
        for group, experts in EXPERT_GROUPS.items():
            logger.info(f"\n{group.upper()} ({len(experts)} experts):")
            for expert in experts:
                logger.info(f"  • {expert}")
        logger.info("\n" + "="*70)
        return 0

    # Determine experts to train
    if args.experts:
        experts = [e.strip() for e in args.experts.split(',')]
    else:
        experts = EXPERT_GROUPS.get(args.group, [])

    if not experts:
        logger.error(f"Unknown group: {args.group}")
        return 1

    # Create trainer
    trainer = ExpertTrainer(
        dataset_dir=args.dataset_dir,
        output_dir=args.output,
        logs_dir=args.logs_dir,
        max_workers=args.workers if args.parallel else 1,
        epochs=args.epochs,
        batch_size=args.batch_size,
        learning_rate=args.learning_rate
    )

    # Train
    success = trainer.train_experts(
        experts,
        parallel=args.parallel
    )

    # Save results
    if args.results:
        trainer.save_results(args.results)

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
