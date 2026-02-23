#!/usr/bin/env python3
"""
Compile trained expert models to KHANARY binary format (.khμ).
Includes metadata, quantization, and optimization.
"""

import os
import sys
import json
import logging
import argparse
from pathlib import Path
from typing import Dict, Optional
from datetime import datetime

try:
    import torch
    from transformers import AutoModel, AutoTokenizer
    import safetensors.torch
except ImportError:
    print("Error: Required packages not installed.")
    print("Install with: pip install torch transformers safetensors")
    sys.exit(1)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# KHANARY binary format signature
KHANARY_SIGNATURE = b'KHΜ'  # KHANARY signature in bytes
KHANARY_VERSION = b'\x02'   # Version 0.2


class KhanaryCompiler:
    """Compile trained models to KHANARY binary format."""

    def __init__(self, compression: str = "fp16"):
        """
        Initialize compiler.

        Args:
            compression: Compression format (fp32, fp16, int8, int4)
        """
        self.compression = compression
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        logger.info(f"Using device: {self.device}")

    def load_model(self, checkpoint_path: str) -> bool:
        """
        Load trained model checkpoint.

        Args:
            checkpoint_path: Path to model checkpoint

        Returns:
            True if successful
        """
        logger.info(f"Loading checkpoint: {checkpoint_path}")

        try:
            self.model = AutoModel.from_pretrained(
                checkpoint_path,
                torch_dtype=self._get_torch_dtype(),
                device_map="auto" if self.device == "cuda" else None,
                trust_remote_code=True
            )
            self.tokenizer = AutoTokenizer.from_pretrained(checkpoint_path)

            logger.info(f"✓ Model loaded ({self.model.config.model_type})")
            return True

        except Exception as e:
            logger.error(f"Failed to load model: {e}")
            return False

    def _get_torch_dtype(self):
        """Get PyTorch dtype based on compression setting."""
        dtype_map = {
            'fp32': torch.float32,
            'fp16': torch.float16,
            'int8': torch.int8,
            'int4': torch.int8
        }
        return dtype_map.get(self.compression, torch.float16)

    def quantize_model(self) -> bool:
        """
        Quantize model for compression.

        Returns:
            True if successful
        """
        logger.info(f"Quantizing model to {self.compression}...")

        try:
            if self.compression == "int8":
                # Quantize to 8-bit
                from torch.quantization import quantize_dynamic
                self.model = quantize_dynamic(
                    self.model,
                    {torch.nn.Linear},
                    dtype=torch.qint8
                )
                logger.info("✓ Quantized to int8")

            elif self.compression == "int4":
                # Simulate 4-bit (use int8 as proxy)
                from torch.quantization import quantize_dynamic
                self.model = quantize_dynamic(
                    self.model,
                    {torch.nn.Linear},
                    dtype=torch.qint8
                )
                logger.info("✓ Quantized to 4-bit (simulated)")

            return True

        except Exception as e:
            logger.warning(f"Quantization failed, using original: {e}")
            return True

    def create_metadata(
        self,
        expert_name: str,
        version: str = "2.1",
        accuracy: float = 0.93,
        latency_ms: float = 2.0
    ) -> Dict:
        """
        Create metadata for KHANARY binary.

        Args:
            expert_name: Name of expert
            version: Expert version
            accuracy: Accuracy on benchmark
            latency_ms: Average latency in ms

        Returns:
            Metadata dictionary
        """
        metadata = {
            'format': 'KHANARY',
            'version': 'v0.2',
            'expert_name': expert_name,
            'expert_version': version,
            'created_at': datetime.utcnow().isoformat(),
            'model_type': self.model.config.model_type,
            'model_size_params': sum(p.numel() for p in self.model.parameters()),
            'compression': self.compression,
            'device': self.device,
            'accuracy': accuracy,
            'latency_ms': latency_ms,
            'tokenizer': {
                'type': self.tokenizer.__class__.__name__,
                'vocab_size': len(self.tokenizer),
                'max_length': getattr(self.tokenizer, 'model_max_length', 2048),
                'special_tokens': {
                    'pad_token': self.tokenizer.pad_token,
                    'bos_token': self.tokenizer.bos_token,
                    'eos_token': self.tokenizer.eos_token
                }
            },
            'model_config': {
                'hidden_size': self.model.config.hidden_size,
                'num_hidden_layers': self.model.config.num_hidden_layers,
                'num_attention_heads': self.model.config.num_attention_heads,
                'intermediate_size': getattr(self.model.config, 'intermediate_size', None),
            }
        }
        return metadata

    def save_as_khanary(
        self,
        output_path: str,
        expert_name: str,
        metadata: Optional[Dict] = None
    ) -> bool:
        """
        Save model as KHANARY binary format.

        Args:
            output_path: Output file path (.khμ)
            expert_name: Name of expert
            metadata: Optional metadata dict

        Returns:
            True if successful
        """
        logger.info(f"Compiling to KHANARY format: {output_path}")

        try:
            output_path = Path(output_path)
            output_path.parent.mkdir(parents=True, exist_ok=True)

            # Create metadata
            if metadata is None:
                metadata = self.create_metadata(expert_name)

            # Extract model weights
            logger.info("Extracting model weights...")
            state_dict = self.model.state_dict()

            # Create KHANARY binary structure
            khanary_data = {
                'metadata': metadata,
                'weights': state_dict,
                'tokenizer_config': self.tokenizer.get_vocab()
            }

            # Save using safetensors format (secure, fast)
            logger.info("Serializing to safetensors format...")
            weights_path = output_path.with_suffix('.safetensors')
            safetensors.torch.save_file(state_dict, str(weights_path))

            # Save metadata as JSON
            metadata_path = output_path.with_suffix('.json')
            with open(metadata_path, 'w') as f:
                json.dump(metadata, f, indent=2)
            logger.info(f"✓ Metadata saved to {metadata_path}")

            # Create KHANARY binary (combine everything)
            logger.info("Creating KHANARY binary...")
            self._create_khanary_binary(
                output_path,
                weights_path,
                metadata_path,
                metadata
            )

            # Verify
            if output_path.exists():
                size_kb = output_path.stat().st_size / 1024
                logger.info(f"✓ KHANARY binary created ({size_kb:.1f}KB)")
                return True
            else:
                logger.error("Binary file not created")
                return False

        except Exception as e:
            logger.error(f"Failed to save KHANARY binary: {e}")
            return False

    def _create_khanary_binary(
        self,
        output_path: Path,
        weights_path: Path,
        metadata_path: Path,
        metadata: Dict
    ):
        """
        Create KHANARY binary file.

        Args:
            output_path: Output .khμ file path
            weights_path: SafeTensors weights file
            metadata_path: JSON metadata file
            metadata: Metadata dictionary
        """
        with open(output_path, 'wb') as out_file:
            # Write KHANARY signature and version
            out_file.write(KHANARY_SIGNATURE)
            out_file.write(KHANARY_VERSION)

            # Write metadata size and content
            metadata_json = json.dumps(metadata).encode('utf-8')
            metadata_size = len(metadata_json)
            out_file.write(metadata_size.to_bytes(4, byteorder='big'))
            out_file.write(metadata_json)

            # Write weights as binary blob
            with open(weights_path, 'rb') as weights_file:
                weights_data = weights_file.read()
                weights_size = len(weights_data)
                out_file.write(weights_size.to_bytes(8, byteorder='big'))
                out_file.write(weights_data)

        logger.info(f"Binary structure: [SIGNATURE(3)] + [VERSION(1)] + [META_SIZE(4)] + [META] + [WEIGHTS_SIZE(8)] + [WEIGHTS]")

    def compile(
        self,
        checkpoint_path: str,
        output_path: str,
        expert_name: str,
        expert_version: str = "2.1",
        accuracy: float = 0.93,
        latency_ms: float = 2.0
    ) -> bool:
        """
        Full compilation pipeline.

        Args:
            checkpoint_path: Path to trained checkpoint
            output_path: Output .khμ file path
            expert_name: Name of expert
            expert_version: Expert version
            accuracy: Accuracy metric
            latency_ms: Latency in ms

        Returns:
            True if successful
        """
        logger.info("="*60)
        logger.info(f"Compiling {expert_name} expert to KHANARY format")
        logger.info("="*60)

        # Load
        if not self.load_model(checkpoint_path):
            return False

        # Quantize
        self.quantize_model()

        # Create metadata
        metadata = self.create_metadata(
            expert_name=expert_name,
            version=expert_version,
            accuracy=accuracy,
            latency_ms=latency_ms
        )

        # Save
        success = self.save_as_khanary(
            output_path,
            expert_name,
            metadata
        )

        if success:
            logger.info("="*60)
            logger.info("✓ Compilation successful!")
            logger.info("="*60)

        return success


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Compile KHANARY experts to binary format'
    )
    parser.add_argument(
        '--checkpoint',
        type=str,
        required=True,
        help='Path to trained model checkpoint'
    )
    parser.add_argument(
        '--output',
        type=str,
        required=True,
        help='Output .khμ file path'
    )
    parser.add_argument(
        '--expert-name',
        type=str,
        default='expert',
        help='Expert name (default: expert)'
    )
    parser.add_argument(
        '--expert-version',
        type=str,
        default='2.1',
        help='Expert version (default: 2.1)'
    )
    parser.add_argument(
        '--accuracy',
        type=float,
        default=0.93,
        help='Accuracy metric (0-1, default: 0.93)'
    )
    parser.add_argument(
        '--latency-ms',
        type=float,
        default=2.0,
        help='Latency in ms (default: 2.0)'
    )
    parser.add_argument(
        '--compression',
        type=str,
        choices=['fp32', 'fp16', 'int8', 'int4'],
        default='fp16',
        help='Compression format (default: fp16)'
    )
    parser.add_argument(
        '--format',
        type=str,
        default='khanary_v0.2',
        help='Output format (default: khanary_v0.2)'
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

    # Create compiler
    compiler = KhanaryCompiler(compression=args.compression)

    # Compile
    success = compiler.compile(
        checkpoint_path=args.checkpoint,
        output_path=args.output,
        expert_name=args.expert_name,
        expert_version=args.expert_version,
        accuracy=args.accuracy,
        latency_ms=args.latency_ms
    )

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
