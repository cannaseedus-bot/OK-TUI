#!/usr/bin/env python3
"""
Train KHANARY domain-specific experts using HuggingFace Transformers.
Supports: Python, Security, Architecture, Performance, SQL experts.
"""

import os
import sys
import json
import logging
import argparse
from pathlib import Path
from typing import Dict, Optional

try:
    import torch
    import yaml
    from datasets import load_dataset, concatenate_datasets
    from transformers import (
        AutoTokenizer, AutoModelForCausalLM,
        TrainingArguments, Trainer,
        DataCollatorForLanguageModeling
    )
except ImportError:
    print("Error: Required packages not installed.")
    print("Install with: pip install torch transformers datasets pyyaml")
    sys.exit(1)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class ExpertTrainer:
    """Train domain-specific KHANARY experts."""

    def __init__(
        self,
        model_name: str = "Qwen/Qwen-7B-Chat",
        device: str = "cuda" if torch.cuda.is_available() else "cpu"
    ):
        """
        Initialize expert trainer.

        Args:
            model_name: HuggingFace model to fine-tune
            device: Device to train on (cuda/cpu)
        """
        self.model_name = model_name
        self.device = device
        self.model = None
        self.tokenizer = None

        logger.info(f"Using device: {device}")
        logger.info(f"Model: {model_name}")

    def load_model_and_tokenizer(self):
        """Load model and tokenizer from HuggingFace."""
        logger.info("Loading model and tokenizer...")

        try:
            self.tokenizer = AutoTokenizer.from_pretrained(
                self.model_name,
                trust_remote_code=True
            )
            self.model = AutoModelForCausalLM.from_pretrained(
                self.model_name,
                torch_dtype=torch.float16 if self.device == "cuda" else torch.float32,
                device_map="auto" if self.device == "cuda" else None,
                trust_remote_code=True
            )

            # Add pad token if missing
            if self.tokenizer.pad_token is None:
                self.tokenizer.pad_token = self.tokenizer.eos_token

            logger.info(f"✓ Model loaded: {self.model.config.hidden_size} dims")
            return True

        except Exception as e:
            logger.error(f"Failed to load model: {e}")
            return False

    def load_datasets(
        self,
        dataset_dir: str,
        expert: str,
        max_samples: Optional[int] = None
    ) -> Optional[object]:
        """
        Load training datasets for expert.

        Args:
            dataset_dir: Directory containing datasets
            expert: Expert name
            max_samples: Limit samples (None = all)

        Returns:
            Combined dataset or None if failed
        """
        logger.info(f"Loading datasets for {expert} expert...")

        dataset_dir = Path(dataset_dir) / expert
        if not dataset_dir.exists():
            logger.error(f"Dataset directory not found: {dataset_dir}")
            return None

        datasets = []
        parquet_files = list(dataset_dir.glob("*.parquet"))

        if not parquet_files:
            logger.error(f"No parquet files in {dataset_dir}")
            return None

        for parquet_file in parquet_files:
            try:
                logger.info(f"  • Loading {parquet_file.name}...")
                ds = load_dataset("parquet", data_files=str(parquet_file))["train"]

                if max_samples and len(ds) > max_samples:
                    ds = ds.select(range(max_samples))

                datasets.append(ds)
                logger.info(f"    ✓ {len(ds)} samples")

            except Exception as e:
                logger.warning(f"    ✗ Failed: {e}")

        if not datasets:
            logger.error("No datasets loaded successfully")
            return None

        # Combine datasets
        if len(datasets) == 1:
            combined = datasets[0]
        else:
            combined = concatenate_datasets(datasets)

        logger.info(f"✓ Total samples: {len(combined)}")
        return combined

    def preprocess_data(self, dataset, max_length: int = 1024):
        """
        Tokenize and format data for training.

        Args:
            dataset: Raw dataset
            max_length: Max sequence length

        Returns:
            Preprocessed dataset
        """
        logger.info(f"Preprocessing data (max_length={max_length})...")

        def tokenize_function(examples):
            """Tokenize text examples."""
            # Handle different possible text column names
            text_column = None
            for col in ['text', 'code', 'content', 'input', 'problem_statement', 'query']:
                if col in examples and examples[col]:
                    text_column = col
                    break

            if text_column is None:
                # Use first string column
                for col in examples:
                    if isinstance(examples[col][0], str):
                        text_column = col
                        break

            if text_column is None:
                logger.warning("No text column found")
                return {"input_ids": [], "attention_mask": []}

            outputs = self.tokenizer(
                examples[text_column],
                max_length=max_length,
                truncation=True,
                padding="max_length",
                return_overflowing_tokens=True
            )

            return outputs

        # Apply tokenization
        tokenized = dataset.map(
            tokenize_function,
            batched=True,
            remove_columns=dataset.column_names,
            num_proc=4
        )

        logger.info(f"✓ Tokenized {len(tokenized)} samples")
        return tokenized

    def train(
        self,
        dataset,
        output_dir: str,
        epochs: int = 3,
        batch_size: int = 32,
        learning_rate: float = 2e-5,
        warmup_steps: int = 500,
        weight_decay: float = 0.01,
        expert_name: str = "expert"
    ) -> bool:
        """
        Train expert on dataset.

        Args:
            dataset: Training dataset
            output_dir: Directory to save checkpoints
            epochs: Number of training epochs
            batch_size: Training batch size
            learning_rate: Learning rate
            warmup_steps: Warmup steps
            weight_decay: Weight decay for regularization
            expert_name: Name of expert (for logging)

        Returns:
            True if successful
        """
        if self.model is None or self.tokenizer is None:
            logger.error("Model and tokenizer not loaded")
            return False

        logger.info(f"\nStarting training for {expert_name} expert...")
        logger.info(f"  Epochs: {epochs}")
        logger.info(f"  Batch size: {batch_size}")
        logger.info(f"  Learning rate: {learning_rate}")
        logger.info(f"  Warmup steps: {warmup_steps}")
        logger.info(f"  Dataset size: {len(dataset)}")

        # Create output directory
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)

        # Training arguments
        training_args = TrainingArguments(
            output_dir=str(output_path),
            overwrite_output_dir=True,
            do_train=True,
            do_eval=False,
            num_train_epochs=epochs,
            per_device_train_batch_size=batch_size,
            save_steps=len(dataset) // batch_size // 2,  # Save twice per epoch
            save_total_limit=3,
            logging_steps=100,
            learning_rate=learning_rate,
            warmup_steps=warmup_steps,
            weight_decay=weight_decay,
            fp16=self.device == "cuda",
            gradient_accumulation_steps=1,
            dataloader_num_workers=4
        )

        # Data collator for language modeling
        data_collator = DataCollatorForLanguageModeling(
            tokenizer=self.tokenizer,
            mlm=False
        )

        # Trainer
        trainer = Trainer(
            model=self.model,
            args=training_args,
            train_dataset=dataset,
            data_collator=data_collator
        )

        try:
            # Train
            trainer.train()

            # Save final model
            final_output = output_path / f"{expert_name}_final"
            self.model.save_pretrained(str(final_output))
            self.tokenizer.save_pretrained(str(final_output))

            logger.info(f"✓ Training complete. Model saved to {final_output}")
            return True

        except Exception as e:
            logger.error(f"Training failed: {e}")
            return False


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Train KHANARY domain-specific experts'
    )
    parser.add_argument(
        '--expert',
        type=str,
        required=True,
        choices=['python', 'security', 'architecture', 'performance', 'sql'],
        help='Expert to train'
    )
    parser.add_argument(
        '--config',
        type=str,
        help='Config file for expert (YAML)'
    )
    parser.add_argument(
        '--dataset-dir',
        type=str,
        default='datasets',
        help='Directory containing datasets (default: datasets)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='checkpoints',
        help='Directory to save checkpoints (default: checkpoints)'
    )
    parser.add_argument(
        '--model',
        type=str,
        default='Qwen/Qwen-7B-Chat',
        help='Base model to fine-tune (default: Qwen/Qwen-7B-Chat)'
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
        help='Training batch size (default: 32)'
    )
    parser.add_argument(
        '--learning-rate',
        type=float,
        default=2e-5,
        help='Learning rate (default: 2e-5)'
    )
    parser.add_argument(
        '--max-samples',
        type=int,
        default=None,
        help='Max samples to use'
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

    # Load config if provided
    training_config = {}
    if args.config and Path(args.config).exists():
        try:
            with open(args.config) as f:
                training_config = yaml.safe_load(f)
            logger.info(f"Loaded config from {args.config}")
        except Exception as e:
            logger.warning(f"Failed to load config: {e}")

    # Create trainer
    trainer = ExpertTrainer(
        model_name=training_config.get('model', args.model)
    )

    # Load model
    if not trainer.load_model_and_tokenizer():
        return 1

    # Load datasets
    dataset = trainer.load_datasets(
        dataset_dir=args.dataset_dir,
        expert=args.expert,
        max_samples=args.max_samples
    )

    if dataset is None:
        logger.error("Failed to load dataset")
        return 1

    # Preprocess
    dataset = trainer.preprocess_data(dataset)

    # Train
    success = trainer.train(
        dataset=dataset,
        output_dir=args.output,
        epochs=training_config.get('epochs', args.epochs),
        batch_size=training_config.get('batch_size', args.batch_size),
        learning_rate=training_config.get('learning_rate', args.learning_rate),
        expert_name=args.expert
    )

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
