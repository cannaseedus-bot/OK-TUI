#!/usr/bin/env python3
"""
Validate KHANARY expert determinism.
Ensures compiled binaries produce identical outputs across multiple runs.
"""

import os
import sys
import json
import logging
import argparse
import hashlib
from pathlib import Path
from typing import Dict, List, Tuple

try:
    import torch
    from transformers import AutoModel, AutoTokenizer
    import numpy as np
except ImportError:
    print("Error: Required packages not installed.")
    print("Install with: pip install torch transformers numpy")
    sys.exit(1)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# Test prompts for different experts
TEST_PROMPTS = {
    'python': [
        "def fibonacci(n):",
        "class DataProcessor:",
        "import asyncio\nasync def fetch_data():",
    ],
    'security': [
        "sql_query = \"SELECT * FROM users WHERE id=\" + user_id",
        "eval(user_input)",
        "password = os.environ.get('DB_PASSWORD')",
    ],
    'architecture': [
        "class MicroserviceController:",
        "def design_system_architecture():",
        "graph = nx.DiGraph()",
    ],
    'performance': [
        "for i in range(1000000):",
        "def optimize_loop():",
        "matrix = [[0]*1000 for _ in range(1000)]",
    ],
    'sql': [
        "SELECT * FROM orders WHERE date > '2024-01-01'",
        "SELECT COUNT(*) FROM users GROUP BY department",
        "UPDATE products SET price = price * 0.9",
    ]
}


class DeterminismValidator:
    """Validate KHANARY expert determinism."""

    def __init__(self, seed: int = 42):
        """
        Initialize validator.

        Args:
            seed: Random seed for reproducibility
        """
        self.seed = seed
        self._set_determinism_mode(seed)
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        logger.info(f"Using device: {self.device}")

    def _set_determinism_mode(self, seed: int):
        """Set determinism mode for reproducible results."""
        torch.manual_seed(seed)
        torch.cuda.manual_seed_all(seed)
        np.random.seed(seed)

        # Enable deterministic behavior
        torch.backends.cudnn.deterministic = True
        torch.backends.cudnn.benchmark = False

        os.environ['PYTHONHASHSEED'] = str(seed)
        os.environ['CUBLAS_WORKSPACE_CONFIG'] = ':4096:8'  # CUDA determinism

    def load_model(self, checkpoint_path: str) -> bool:
        """
        Load KHANARY expert model.

        Args:
            checkpoint_path: Path to .khμ or checkpoint

        Returns:
            True if successful
        """
        try:
            logger.info(f"Loading model: {checkpoint_path}")

            # Load model and tokenizer
            self.model = AutoModel.from_pretrained(
                checkpoint_path,
                torch_dtype=torch.float16 if self.device == "cuda" else torch.float32,
                device_map="auto" if self.device == "cuda" else None,
                trust_remote_code=True
            )
            self.tokenizer = AutoTokenizer.from_pretrained(checkpoint_path)

            # Set to evaluation mode
            self.model.eval()

            logger.info(f"✓ Model loaded ({self.model.config.model_type})")
            return True

        except Exception as e:
            logger.error(f"Failed to load model: {e}")
            return False

    def get_model_outputs(
        self,
        prompts: List[str],
        max_length: int = 128
    ) -> np.ndarray:
        """
        Get model outputs for test prompts.

        Args:
            prompts: List of input prompts
            max_length: Max generation length

        Returns:
            Output embeddings
        """
        outputs_list = []

        with torch.no_grad():
            for prompt in prompts:
                try:
                    # Tokenize
                    inputs = self.tokenizer(
                        prompt,
                        return_tensors="pt",
                        max_length=512,
                        truncation=True,
                        padding=True
                    ).to(self.device)

                    # Forward pass
                    outputs = self.model(**inputs, output_hidden_states=True)

                    # Extract pooled output (last hidden state mean)
                    hidden_states = outputs.last_hidden_state
                    pooled = hidden_states.mean(dim=1).cpu().numpy()

                    outputs_list.append(pooled)

                except Exception as e:
                    logger.warning(f"Error processing prompt: {e}")
                    outputs_list.append(np.zeros((1, self.model.config.hidden_size)))

        if outputs_list:
            return np.concatenate(outputs_list, axis=0)
        else:
            return np.array([])

    def compute_output_hash(self, outputs: np.ndarray) -> str:
        """
        Compute hash of outputs for verification.

        Args:
            outputs: Model outputs

        Returns:
            SHA256 hash
        """
        output_bytes = outputs.astype(np.float32).tobytes()
        return hashlib.sha256(output_bytes).hexdigest()

    def validate_single_expert(
        self,
        checkpoint_path: str,
        expert_name: str,
        num_runs: int = 10
    ) -> Tuple[bool, Dict]:
        """
        Validate determinism for single expert.

        Args:
            checkpoint_path: Path to checkpoint
            expert_name: Name of expert
            num_runs: Number of validation runs

        Returns:
            Tuple of (success, results_dict)
        """
        logger.info(f"\nValidating {expert_name} expert ({num_runs} runs)...")

        if not self.load_model(checkpoint_path):
            return False, {'error': 'Failed to load model'}

        # Get test prompts
        prompts = TEST_PROMPTS.get(
            expert_name,
            ["test", "validation", "determinism"]
        )

        results = {
            'expert': expert_name,
            'checkpoint': checkpoint_path,
            'num_runs': num_runs,
            'runs': []
        }

        # Run multiple times and collect outputs
        output_hashes = []

        for run in range(num_runs):
            logger.info(f"  Run {run + 1}/{num_runs}...")

            # Reset determinism
            self._set_determinism_mode(self.seed)

            # Get outputs
            outputs = self.get_model_outputs(prompts)
            output_hash = self.compute_output_hash(outputs)

            results['runs'].append({
                'run': run + 1,
                'hash': output_hash,
                'output_shape': outputs.shape
            })

            output_hashes.append(output_hash)

            logger.info(f"    Hash: {output_hash[:16]}...")

        # Check determinism
        unique_hashes = set(output_hashes)
        is_deterministic = len(unique_hashes) == 1

        results['is_deterministic'] = is_deterministic
        results['unique_hashes'] = len(unique_hashes)
        results['success_rate'] = (num_runs - len(unique_hashes) + 1) / num_runs

        if is_deterministic:
            logger.info(f"  ✓ {expert_name} is DETERMINISTIC (all runs identical)")
        else:
            logger.warning(f"  ✗ {expert_name} shows variance ({len(unique_hashes)} unique hashes)")

        return is_deterministic, results

    def validate_parity(
        self,
        checkpoint_path: str,
        reference_path: str,
        expert_name: str
    ) -> Tuple[bool, Dict]:
        """
        Validate parity between two model versions.

        Args:
            checkpoint_path: Current model checkpoint
            reference_path: Reference model checkpoint
            expert_name: Expert name

        Returns:
            Tuple of (parity_ok, results_dict)
        """
        logger.info(f"\nValidating parity for {expert_name} expert...")

        # Load current model
        if not self.load_model(checkpoint_path):
            return False, {'error': 'Failed to load current model'}

        prompts = TEST_PROMPTS.get(expert_name, ["test"])
        current_outputs = self.get_model_outputs(prompts)

        # Load reference model
        if not self.load_model(reference_path):
            return False, {'error': 'Failed to load reference model'}

        reference_outputs = self.get_model_outputs(prompts)

        # Compare
        mse = np.mean((current_outputs - reference_outputs) ** 2)
        cosine_sim = np.mean([
            np.dot(current_outputs[i], reference_outputs[i]) /
            (np.linalg.norm(current_outputs[i]) * np.linalg.norm(reference_outputs[i]) + 1e-8)
            for i in range(len(current_outputs))
        ])

        parity_ok = cosine_sim > 0.99  # 99% similarity threshold

        results = {
            'expert': expert_name,
            'mse': float(mse),
            'cosine_similarity': float(cosine_sim),
            'parity_ok': parity_ok
        }

        if parity_ok:
            logger.info(f"  ✓ Parity OK (cosine_sim={cosine_sim:.4f})")
        else:
            logger.warning(f"  ✗ Parity divergence (cosine_sim={cosine_sim:.4f})")

        return parity_ok, results


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Validate KHANARY expert determinism'
    )
    parser.add_argument(
        '--experts',
        type=str,
        required=True,
        help='Path to expert(s) (glob pattern or directory)'
    )
    parser.add_argument(
        '--expert-names',
        type=str,
        help='Comma-separated expert names (python,security,etc)'
    )
    parser.add_argument(
        '--runs',
        type=int,
        default=10,
        help='Number of validation runs (default: 10)'
    )
    parser.add_argument(
        '--reference',
        type=str,
        help='Reference model path for parity checking'
    )
    parser.add_argument(
        '--seed',
        type=int,
        default=42,
        help='Random seed (default: 42)'
    )
    parser.add_argument(
        '--output',
        type=str,
        help='Output report file (JSON)'
    )
    parser.add_argument(
        '--log-file',
        type=str,
        help='Log file path'
    )

    args = parser.parse_args()

    # Setup logging to file if requested
    if args.log_file:
        file_handler = logging.FileHandler(args.log_file)
        file_handler.setFormatter(logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        ))
        logger.addHandler(file_handler)

    # Create validator
    validator = DeterminismValidator(seed=args.seed)

    # Find experts
    experts_path = Path(args.experts)
    if experts_path.is_file():
        expert_files = [experts_path]
    else:
        # Try as glob pattern
        parent = experts_path.parent if '*' not in str(experts_path) else Path('.')
        pattern = experts_path.name if '*' in str(experts_path) else '*.khμ'
        expert_files = list(parent.glob(pattern))

    if not expert_files:
        logger.error(f"No experts found at {args.experts}")
        return 1

    logger.info(f"Found {len(expert_files)} expert(s)")

    # Validate
    all_results = []
    all_deterministic = True

    for expert_file in expert_files:
        expert_name = expert_file.stem

        # Run determinism validation
        is_det, det_results = validator.validate_single_expert(
            str(expert_file),
            expert_name,
            num_runs=args.runs
        )

        all_results.append({
            'type': 'determinism',
            **det_results
        })

        if not is_det:
            all_deterministic = False

        # Run parity validation if reference provided
        if args.reference:
            _, par_results = validator.validate_parity(
                str(expert_file),
                args.reference,
                expert_name
            )
            all_results.append({
                'type': 'parity',
                **par_results
            })

    # Summary
    logger.info("\n" + "="*60)
    logger.info("Determinism Validation Summary")
    logger.info("="*60)

    for result in all_results:
        if result['type'] == 'determinism':
            expert = result['expert']
            is_det = result['is_deterministic']
            status = "✓" if is_det else "✗"
            logger.info(f"  {status} {expert}: {result['unique_hashes']} unique hashes")

    if all_deterministic:
        logger.info("\n✓ All experts passed determinism validation!")
    else:
        logger.warning("\n✗ Some experts failed determinism validation")

    logger.info("="*60)

    # Save report if requested
    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)

        with open(output_path, 'w') as f:
            json.dump(all_results, f, indent=2)

        logger.info(f"Report saved to {output_path}")

    return 0 if all_deterministic else 1


if __name__ == '__main__':
    sys.exit(main())
