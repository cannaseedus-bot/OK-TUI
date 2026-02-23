#!/usr/bin/env python3
"""
Compile all trained KHANARY experts to .khμ binary format.
Orchestrates compilation for 40+ domain-specific experts.
"""

import os
import sys
import json
import logging
import argparse
import subprocess
import time
from pathlib import Path
from typing import List, Dict, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class ExpertCompiler:
    """Orchestrate compilation for multiple KHANARY experts."""

    def __init__(
        self,
        checkpoints_dir: str = 'checkpoints',
        output_dir: str = 'experts',
        logs_dir: str = 'logs',
        compression: str = 'fp16',
        max_workers: int = 1
    ):
        """
        Initialize compiler.

        Args:
            checkpoints_dir: Checkpoint directory
            output_dir: Output directory for .khμ files
            logs_dir: Logs directory
            compression: Compression format (fp32, fp16, int8, int4)
            max_workers: Max parallel compilation jobs
        """
        self.checkpoints_dir = Path(checkpoints_dir)
        self.output_dir = Path(output_dir)
        self.logs_dir = Path(logs_dir)
        self.compression = compression
        self.max_workers = max_workers

        # Create directories
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.logs_dir.mkdir(parents=True, exist_ok=True)

        self.results = {
            'successful': [],
            'failed': [],
            'skipped': [],
            'total_size': 0,
            'total_time': 0
        }

    def find_checkpoint(self, expert: str) -> Optional[Path]:
        """Find latest checkpoint for expert."""
        checkpoints = list(self.checkpoints_dir.glob(f"{expert}_final"))
        if not checkpoints:
            checkpoints = list(self.checkpoints_dir.glob(f"{expert}*"))

        if checkpoints:
            # Sort by modification time, get latest
            checkpoints.sort(key=lambda p: p.stat().st_mtime, reverse=True)
            return checkpoints[0]

        return None

    def compile_expert(
        self,
        expert: str,
        accuracy: float = 0.9,
        latency_ms: float = 2.0
    ) -> Dict:
        """
        Compile single expert.

        Args:
            expert: Expert name
            accuracy: Expected accuracy
            latency_ms: Expected latency in ms

        Returns:
            Compilation result dictionary
        """
        logger.info(f"Compiling {expert}...")

        # Find checkpoint
        checkpoint = self.find_checkpoint(expert)
        if not checkpoint:
            logger.warning(f"Checkpoint not found for {expert}")
            return {'expert': expert, 'status': 'skipped', 'reason': 'no_checkpoint'}

        # Prepare paths
        output_file = self.output_dir / f"{expert}.khμ"
        log_file = self.logs_dir / f"compile_{expert}.log"

        # Prepare command
        cmd = [
            sys.executable,
            'scripts/compile_to_khanary.py',
            '--checkpoint', str(checkpoint),
            '--output', str(output_file),
            '--expert-name', expert,
            '--compression', self.compression,
            '--accuracy', str(accuracy),
            '--latency-ms', str(latency_ms),
            '--log-file', str(log_file)
        ]

        try:
            start_time = time.time()

            # Run compilation
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=600  # 10 minute timeout
            )

            elapsed = time.time() - start_time

            if result.returncode == 0 and output_file.exists():
                size_mb = output_file.stat().st_size / (1024 * 1024)
                logger.info(f"✓ {expert} compiled ({size_mb:.1f}MB, {elapsed:.1f}s)")

                return {
                    'expert': expert,
                    'status': 'success',
                    'file': str(output_file),
                    'size_mb': size_mb,
                    'time': elapsed,
                    'log': str(log_file)
                }
            else:
                logger.error(f"✗ {expert} compilation failed")
                logger.error(f"  Error: {result.stderr}")
                return {
                    'expert': expert,
                    'status': 'failed',
                    'error': result.stderr[:500],
                    'log': str(log_file)
                }

        except subprocess.TimeoutExpired:
            logger.error(f"✗ {expert} compilation timeout")
            return {
                'expert': expert,
                'status': 'failed',
                'error': 'timeout'
            }
        except Exception as e:
            logger.error(f"✗ {expert} compilation error: {e}")
            return {
                'expert': expert,
                'status': 'failed',
                'error': str(e)
            }

    def compile_experts(
        self,
        experts: List[str],
        accuracies: Optional[Dict[str, float]] = None,
        latencies: Optional[Dict[str, float]] = None,
        parallel: bool = False
    ) -> bool:
        """
        Compile multiple experts.

        Args:
            experts: List of expert names
            accuracies: Per-expert accuracies (optional)
            latencies: Per-expert latencies (optional)
            parallel: Compile in parallel

        Returns:
            True if all successful
        """
        if not experts:
            logger.error("No experts to compile")
            return False

        accuracies = accuracies or {}
        latencies = latencies or {}

        logger.info(f"\n{'='*70}")
        logger.info(f"Compiling {len(experts)} experts")
        logger.info(f"Configuration:")
        logger.info(f"  Compression: {self.compression}")
        logger.info(f"  Workers: {self.max_workers if parallel else 1}")
        logger.info(f"  Mode: {'Parallel' if parallel else 'Sequential'}")
        logger.info(f"{'='*70}\n")

        start_time = time.time()

        if parallel and self.max_workers > 1:
            # Parallel compilation
            with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
                futures = {
                    executor.submit(
                        self.compile_expert,
                        expert,
                        accuracy=accuracies.get(expert, 0.9),
                        latency_ms=latencies.get(expert, 2.0)
                    ): expert for expert in experts
                }

                for future in as_completed(futures):
                    result = future.result()
                    self._record_result(result)

        else:
            # Sequential compilation
            for expert in experts:
                result = self.compile_expert(
                    expert,
                    accuracy=accuracies.get(expert, 0.9),
                    latency_ms=latencies.get(expert, 2.0)
                )
                self._record_result(result)

        total_time = time.time() - start_time
        self.results['total_time'] = total_time

        return self._print_summary(total_time)

    def _record_result(self, result: Dict):
        """Record compilation result."""
        status = result.get('status', 'unknown')

        if status == 'success':
            self.results['successful'].append(result)
            self.results['total_size'] += result.get('size_mb', 0)
        elif status == 'failed':
            self.results['failed'].append(result)
        else:
            self.results['skipped'].append(result)

    def _print_summary(self, total_time: float) -> bool:
        """Print compilation summary."""
        logger.info("\n" + "="*70)
        logger.info("Compilation Summary")
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
        logger.info(f"  Total size:   {self.results['total_size']:.1f}MB")

        if successful > 0:
            logger.info(f"\nSuccessfully compiled:")
            for result in self.results['successful']:
                size = result.get('size_mb', 0)
                time_str = f"{result.get('time', 0):.1f}s" if 'time' in result else "N/A"
                logger.info(f"  ✓ {result['expert']:20} ({size:6.1f}MB, {time_str})")

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

    def generate_checksums(self) -> bool:
        """Generate SHA256 checksums for all compiled experts."""
        try:
            checksum_file = self.output_dir / 'SHA256SUMS'

            logger.info(f"\nGenerating checksums...")

            cmd = [
                'sha256sum'
            ] + [str(p) for p in self.output_dir.glob('*.khμ')]

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                with open(checksum_file, 'w') as f:
                    f.write(result.stdout)

                logger.info(f"✓ SHA256SUMS generated ({checksum_file})")
                return True
            else:
                logger.warning("Failed to generate checksums")
                return False

        except Exception as e:
            logger.warning(f"Checksum generation error: {e}")
            return False

    def save_results(self, output_file: str):
        """Save compilation results to JSON."""
        output_path = Path(output_file)
        output_path.parent.mkdir(parents=True, exist_ok=True)

        with open(output_path, 'w') as f:
            json.dump(self.results, f, indent=2)

        logger.info(f"Results saved to {output_file}")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Compile all KHANARY experts to binary format'
    )
    parser.add_argument(
        '--experts',
        type=str,
        help='Comma-separated list of experts to compile'
    )
    parser.add_argument(
        '--checkpoints-dir',
        type=str,
        default='checkpoints',
        help='Checkpoints directory (default: checkpoints)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='experts',
        help='Output directory (default: experts)'
    )
    parser.add_argument(
        '--logs-dir',
        type=str,
        default='logs',
        help='Logs directory (default: logs)'
    )
    parser.add_argument(
        '--compression',
        type=str,
        choices=['fp32', 'fp16', 'int8', 'int4'],
        default='fp16',
        help='Compression format (default: fp16)'
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
        help='Enable parallel compilation'
    )
    parser.add_argument(
        '--checksums',
        action='store_true',
        help='Generate SHA256 checksums'
    )
    parser.add_argument(
        '--results',
        type=str,
        help='Save results to JSON file'
    )

    args = parser.parse_args()

    if not args.experts:
        logger.error("No experts specified. Use --experts expert1,expert2,...")
        return 1

    # Parse experts
    experts = [e.strip() for e in args.experts.split(',')]

    # Create compiler
    compiler = ExpertCompiler(
        checkpoints_dir=args.checkpoints_dir,
        output_dir=args.output,
        logs_dir=args.logs_dir,
        compression=args.compression,
        max_workers=args.workers if args.parallel else 1
    )

    # Compile
    success = compiler.compile_experts(
        experts,
        parallel=args.parallel
    )

    # Generate checksums if requested
    if args.checksums and success:
        compiler.generate_checksums()

    # Save results
    if args.results:
        compiler.save_results(args.results)

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
