#!/usr/bin/env python3
"""
Create KHANARY expert registry.
Generates metadata and searchable index of available experts.
"""

import os
import sys
import json
import logging
import argparse
from pathlib import Path
from typing import Dict, List, Optional
from datetime import datetime

try:
    import safetensors.torch
except ImportError:
    print("Warning: safetensors not installed. Install with: pip install safetensors")


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class ExpertRegistry:
    """Create and manage KHANARY expert registry."""

    def __init__(self, registry_path: str = "experts/registry.json"):
        """
        Initialize registry.

        Args:
            registry_path: Path to registry file
        """
        self.registry_path = Path(registry_path)
        self.registry = {
            'format': 'KHANARY Registry v1.0',
            'created_at': datetime.utcnow().isoformat(),
            'experts': {},
            'metadata': {
                'total_experts': 0,
                'total_size_bytes': 0,
                'compression': 'fp16'
            }
        }

    def scan_experts_dir(self, experts_dir: str) -> List[Path]:
        """
        Scan directory for expert binaries.

        Args:
            experts_dir: Directory containing experts

        Returns:
            List of expert files
        """
        experts_path = Path(experts_dir)

        if not experts_path.exists():
            logger.error(f"Experts directory not found: {experts_dir}")
            return []

        # Find .khμ files
        expert_files = list(experts_path.glob("*.khμ"))
        logger.info(f"Found {len(expert_files)} expert binaries")

        return sorted(expert_files)

    def parse_expert_file(self, expert_path: Path) -> Optional[Dict]:
        """
        Parse KHANARY expert binary and extract metadata.

        Args:
            expert_path: Path to .khμ file

        Returns:
            Expert metadata dictionary
        """
        try:
            expert_name = expert_path.stem

            # Try to load metadata if it exists
            metadata_path = expert_path.with_suffix('.json')
            if metadata_path.exists():
                with open(metadata_path) as f:
                    metadata = json.load(f)
            else:
                # Create basic metadata
                metadata = {
                    'expert_name': expert_name,
                    'format': 'KHANARY',
                    'version': 'v0.2'
                }

            # Add file information
            file_stat = expert_path.stat()
            expert_info = {
                'name': expert_name,
                'file': expert_path.name,
                'size_bytes': file_stat.st_size,
                'size_mb': file_stat.st_size / (1024 * 1024),
                'created': datetime.fromtimestamp(file_stat.st_ctime).isoformat(),
                'modified': datetime.fromtimestamp(file_stat.st_mtime).isoformat(),
                'metadata': metadata
            }

            return expert_info

        except Exception as e:
            logger.warning(f"Failed to parse {expert_path}: {e}")
            return None

    def generate_registry(self, experts_dir: str) -> bool:
        """
        Generate registry for all experts.

        Args:
            experts_dir: Directory containing experts

        Returns:
            True if successful
        """
        logger.info(f"Scanning experts directory: {experts_dir}")

        expert_files = self.scan_experts_dir(experts_dir)

        if not expert_files:
            logger.warning("No expert files found")
            return False

        # Parse each expert
        total_size = 0
        for expert_file in expert_files:
            logger.info(f"  • Processing {expert_file.name}...")

            expert_info = self.parse_expert_file(expert_file)
            if expert_info:
                expert_name = expert_info['name']
                self.registry['experts'][expert_name] = expert_info
                total_size += expert_info['size_bytes']
                logger.info(f"    ✓ {expert_name} ({expert_info['size_mb']:.1f}MB)")

        # Update metadata
        self.registry['metadata']['total_experts'] = len(self.registry['experts'])
        self.registry['metadata']['total_size_bytes'] = total_size
        self.registry['metadata']['total_size_mb'] = total_size / (1024 * 1024)
        self.registry['updated_at'] = datetime.utcnow().isoformat()

        logger.info(f"\nRegistry generated: {len(self.registry['experts'])} experts")
        return True

    def save_registry(self) -> bool:
        """
        Save registry to file.

        Returns:
            True if successful
        """
        try:
            self.registry_path.parent.mkdir(parents=True, exist_ok=True)

            with open(self.registry_path, 'w') as f:
                json.dump(self.registry, f, indent=2)

            logger.info(f"✓ Registry saved to {self.registry_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to save registry: {e}")
            return False

    def generate_index(self) -> Dict:
        """
        Generate searchable index from registry.

        Returns:
            Indexed registry
        """
        index = {
            'by_name': {},
            'by_size': [],
            'by_date': [],
            'summary': {}
        }

        for name, expert in self.registry['experts'].items():
            # Index by name
            index['by_name'][name] = {
                'file': expert['file'],
                'size_mb': expert['size_mb'],
                'accuracy': expert.get('metadata', {}).get('accuracy', 'N/A'),
                'latency_ms': expert.get('metadata', {}).get('latency_ms', 'N/A')
            }

            # Add to size and date lists
            index['by_size'].append((name, expert['size_mb']))
            index['by_date'].append((name, expert['modified']))

        # Sort
        index['by_size'].sort(key=lambda x: x[1], reverse=True)
        index['by_date'].sort(key=lambda x: x[1], reverse=True)

        # Generate summary
        index['summary'] = {
            'total': len(self.registry['experts']),
            'largest': index['by_size'][0][0] if index['by_size'] else 'N/A',
            'newest': index['by_date'][0][0] if index['by_date'] else 'N/A',
            'total_size_mb': self.registry['metadata'].get('total_size_mb', 0)
        }

        return index

    def print_summary(self):
        """Print registry summary."""
        logger.info("\n" + "="*70)
        logger.info("KHANARY Expert Registry Summary")
        logger.info("="*70)

        if not self.registry['experts']:
            logger.info("No experts registered")
            return

        logger.info(f"\nTotal Experts: {len(self.registry['experts'])}")
        logger.info(f"Total Size: {self.registry['metadata'].get('total_size_mb', 0):.1f}MB\n")

        logger.info("Registered Experts:")
        logger.info("-" * 70)

        for name, info in self.registry['experts'].items():
            size_mb = info['size_mb']
            metadata = info.get('metadata', {})
            accuracy = metadata.get('accuracy', 'N/A')
            latency = metadata.get('latency_ms', 'N/A')

            logger.info(f"  {name:15} | {size_mb:8.1f}MB | Accuracy: {accuracy} | Latency: {latency}ms")

        logger.info("-" * 70)
        logger.info("="*70 + "\n")

    def export_index(self, output_path: str) -> bool:
        """
        Export searchable index.

        Args:
            output_path: Path to save index

        Returns:
            True if successful
        """
        try:
            index = self.generate_index()
            output_file = Path(output_path)
            output_file.parent.mkdir(parents=True, exist_ok=True)

            with open(output_file, 'w') as f:
                json.dump(index, f, indent=2)

            logger.info(f"✓ Index exported to {output_file}")
            return True

        except Exception as e:
            logger.error(f"Failed to export index: {e}")
            return False


class RegistryClient:
    """Client for querying KHANARY expert registry."""

    def __init__(self, registry_path: str):
        """
        Initialize client.

        Args:
            registry_path: Path to registry.json
        """
        self.registry_path = Path(registry_path)
        self.registry = None
        self.load()

    def load(self) -> bool:
        """Load registry from file."""
        try:
            if not self.registry_path.exists():
                logger.error(f"Registry not found: {self.registry_path}")
                return False

            with open(self.registry_path) as f:
                self.registry = json.load(f)

            logger.info(f"✓ Registry loaded ({len(self.registry['experts'])} experts)")
            return True

        except Exception as e:
            logger.error(f"Failed to load registry: {e}")
            return False

    def list_experts(self) -> List[str]:
        """List all available experts."""
        if not self.registry:
            return []
        return list(self.registry['experts'].keys())

    def get_expert_info(self, expert_name: str) -> Optional[Dict]:
        """Get detailed info about expert."""
        if not self.registry:
            return None
        return self.registry['experts'].get(expert_name)

    def filter_by_accuracy(self, min_accuracy: float) -> List[str]:
        """Filter experts by minimum accuracy."""
        if not self.registry:
            return []

        results = []
        for name, info in self.registry['experts'].items():
            accuracy = info.get('metadata', {}).get('accuracy', 0)
            if accuracy >= min_accuracy:
                results.append(name)

        return results

    def filter_by_size(self, max_size_mb: float) -> List[str]:
        """Filter experts by maximum size."""
        if not self.registry:
            return []

        results = []
        for name, info in self.registry['experts'].items():
            if info['size_mb'] <= max_size_mb:
                results.append(name)

        return results


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Create and manage KHANARY expert registry'
    )
    parser.add_argument(
        '--experts-dir',
        type=str,
        default='experts',
        help='Directory containing expert binaries (default: experts)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='experts/registry.json',
        help='Output registry file (default: experts/registry.json)'
    )
    parser.add_argument(
        '--index-output',
        type=str,
        help='Export searchable index to file'
    )
    parser.add_argument(
        '--query-expert',
        type=str,
        help='Query info about specific expert'
    )
    parser.add_argument(
        '--list',
        action='store_true',
        help='List all registered experts'
    )

    args = parser.parse_args()

    # Create or load registry
    registry = ExpertRegistry(registry_path=args.output)

    # Generate registry
    if registry.generate_registry(args.experts_dir):
        # Save
        registry.save_registry()

        # Print summary
        registry.print_summary()

        # Export index if requested
        if args.index_output:
            registry.export_index(args.index_output)

        # Query if requested
        if args.query_expert:
            client = RegistryClient(args.output)
            info = client.get_expert_info(args.query_expert)
            if info:
                logger.info(f"\nExpert: {args.query_expert}")
                logger.info(json.dumps(info, indent=2))
            else:
                logger.warning(f"Expert not found: {args.query_expert}")

        # List if requested
        if args.list:
            client = RegistryClient(args.output)
            experts = client.list_experts()
            if experts:
                logger.info("\nRegistered Experts:")
                for expert in experts:
                    logger.info(f"  • {expert}")

        return 0
    else:
        logger.error("Failed to generate registry")
        return 1


if __name__ == '__main__':
    sys.exit(main())
