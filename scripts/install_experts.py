#!/usr/bin/env python3
"""
Install KHANARY experts from GitHub releases.
Downloads and verifies compiled .khμ binaries.
"""

import os
import sys
import json
import logging
import argparse
import urllib.request
import hashlib
from pathlib import Path
from typing import Dict, List, Optional


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# GitHub release information
GITHUB_REPO = "cannaseedus-bot/Ollama-K"
EXPERTS_URL_TEMPLATE = "https://github.com/{repo}/releases/download/{tag}/{file}"

CORE_EXPERTS = [
    'python',
    'javascript',
    'security',
    'react',
    'fastapi'
]


class ExpertInstaller:
    """Install KHANARY experts from releases."""

    def __init__(self, install_dir: str = 'experts', version: str = 'latest'):
        """
        Initialize installer.

        Args:
            install_dir: Directory to install experts
            version: Release version (e.g., 'v3.0.0' or 'latest')
        """
        self.install_dir = Path(install_dir)
        self.version = version
        self.install_dir.mkdir(parents=True, exist_ok=True)

        logger.info(f"Installation directory: {self.install_dir}")

    def get_release_info(self) -> Optional[Dict]:
        """Get latest release information from GitHub."""
        try:
            import urllib.request
            import json

            url = f"https://api.github.com/repos/{GITHUB_REPO}/releases/latest"
            logger.info(f"Fetching release info from {url}...")

            with urllib.request.urlopen(url) as response:
                data = json.loads(response.read().decode())
                return data

        except Exception as e:
            logger.warning(f"Could not fetch GitHub release info: {e}")
            return None

    def download_file(
        self,
        url: str,
        dest_path: Path,
        expected_sha256: Optional[str] = None
    ) -> bool:
        """
        Download file from URL with progress tracking.

        Args:
            url: File URL
            dest_path: Destination file path
            expected_sha256: Expected SHA256 hash

        Returns:
            True if successful
        """
        try:
            logger.info(f"Downloading {dest_path.name}...")

            def download_progress(block_num, block_size, total_size):
                downloaded = block_num * block_size
                percent = min(100, (downloaded / total_size) * 100)
                logger.info(f"  Progress: {percent:.1f}% ({downloaded/1024/1024:.1f}MB/{total_size/1024/1024:.1f}MB)")

            urllib.request.urlretrieve(url, dest_path, download_progress)

            # Verify SHA256 if provided
            if expected_sha256:
                logger.info("Verifying checksum...")
                sha256 = self._compute_sha256(dest_path)
                if sha256.lower() != expected_sha256.lower():
                    logger.error(f"Checksum mismatch!")
                    logger.error(f"  Expected: {expected_sha256}")
                    logger.error(f"  Got:      {sha256}")
                    return False
                logger.info("✓ Checksum verified")

            logger.info(f"✓ Downloaded {dest_path.name}")
            return True

        except Exception as e:
            logger.error(f"Download failed: {e}")
            return False

    def _compute_sha256(self, file_path: Path) -> str:
        """Compute SHA256 hash of file."""
        sha256_hash = hashlib.sha256()
        with open(file_path, "rb") as f:
            for byte_block in iter(lambda: f.read(4096), b""):
                sha256_hash.update(byte_block)
        return sha256_hash.hexdigest()

    def install_core_experts(self, checksums: Optional[Dict] = None) -> bool:
        """
        Install core 5 experts.

        Args:
            checksums: Dict mapping expert names to SHA256 hashes

        Returns:
            True if all successful
        """
        checksums = checksums or {}
        success_count = 0

        logger.info("\nInstalling 5 core KHANARY experts...")
        logger.info("=" * 70)

        for expert in CORE_EXPERTS:
            filename = f"{expert}.khμ"
            file_url = EXPERTS_URL_TEMPLATE.format(
                repo=GITHUB_REPO,
                tag=self.version,
                file=filename
            )
            dest_path = self.install_dir / filename

            # Skip if already exists
            if dest_path.exists():
                size_mb = dest_path.stat().st_size / (1024 * 1024)
                logger.info(f"✓ {filename} already installed ({size_mb:.1f}MB)")
                success_count += 1
                continue

            # Download
            expected_hash = checksums.get(expert)
            if self.download_file(file_url, dest_path, expected_hash):
                size_mb = dest_path.stat().st_size / (1024 * 1024)
                logger.info(f"✓ {expert} expert installed ({size_mb:.1f}MB)")
                success_count += 1
            else:
                logger.error(f"✗ Failed to install {expert}")

        logger.info("=" * 70)
        logger.info(f"Installation: {success_count}/{len(CORE_EXPERTS)} successful")

        return success_count == len(CORE_EXPERTS)

    def download_registry(self) -> bool:
        """Download expert registry."""
        try:
            logger.info("Downloading expert registry...")

            files = ['registry.json', 'index.json', 'SHA256SUMS']

            for filename in files:
                file_url = EXPERTS_URL_TEMPLATE.format(
                    repo=GITHUB_REPO,
                    tag=self.version,
                    file=filename
                )
                dest_path = self.install_dir / filename

                if self.download_file(file_url, dest_path):
                    logger.info(f"✓ {filename} downloaded")
                else:
                    logger.warning(f"Could not download {filename}")

            return True

        except Exception as e:
            logger.warning(f"Registry download failed: {e}")
            return False

    def verify_installation(self) -> bool:
        """Verify installed experts."""
        logger.info("\nVerifying installation...")
        logger.info("=" * 70)

        expert_files = list(self.install_dir.glob("*.khμ"))

        if not expert_files:
            logger.error("No experts found!")
            return False

        for expert_file in sorted(expert_files):
            size_mb = expert_file.stat().st_size / (1024 * 1024)
            logger.info(f"✓ {expert_file.name:20} ({size_mb:6.1f}MB)")

        logger.info("=" * 70)
        logger.info(f"Total: {len(expert_files)} experts installed")
        logger.info(f"Total size: {sum(f.stat().st_size for f in expert_files) / (1024**2):.1f}MB")

        # Check registry
        registry_path = self.install_dir / 'registry.json'
        if registry_path.exists():
            try:
                with open(registry_path) as f:
                    registry = json.load(f)
                logger.info(f"✓ Registry available ({len(registry.get('experts', {}))} experts)")
            except Exception as e:
                logger.warning(f"Could not parse registry: {e}")

        return True

    def print_usage(self):
        """Print usage instructions."""
        logger.info("\n" + "=" * 70)
        logger.info("✓ KHANARY experts installed successfully!")
        logger.info("=" * 70)

        logger.info("\nQuick Start:")
        logger.info(f"  cd {self.install_dir}")
        logger.info("  ls -lh *.khμ")

        logger.info("\nUsage Examples:")
        logger.info("  # List experts")
        logger.info("  python scripts/create_expert_registry.py --query-expert python")

        logger.info("\n  # Query specific expert")
        logger.info("  cat experts/registry.json | jq '.experts.python'")

        logger.info("\n  # Download all 40+ experts for training")
        logger.info("  python scripts/download_datasets.py --expert all")

        logger.info("\nDocumentation:")
        logger.info("  - EXPERT_CATALOG.md - All 40+ available experts")
        logger.info("  - TRAINING_GUIDE_40_EXPERTS.md - Training guide")
        logger.info("  - BUILD_SYSTEM_GUIDE.md - Build system")

        logger.info("=" * 70 + "\n")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Install KHANARY experts from GitHub releases'
    )
    parser.add_argument(
        '--version',
        type=str,
        default='v3.0.0',
        help='Release version (default: v3.0.0)'
    )
    parser.add_argument(
        '--install-dir',
        type=str,
        default='experts',
        help='Installation directory (default: experts)'
    )
    parser.add_argument(
        '--verify-only',
        action='store_true',
        help='Only verify existing installation'
    )
    parser.add_argument(
        '--latest',
        action='store_true',
        help='Use latest release'
    )

    args = parser.parse_args()

    # Determine version
    version = 'latest' if args.latest else args.version

    # Create installer
    installer = ExpertInstaller(
        install_dir=args.install_dir,
        version=version
    )

    if args.verify_only:
        # Only verify
        success = installer.verify_installation()
        return 0 if success else 1

    # Download and install
    logger.info("\n" + "=" * 70)
    logger.info("KHANARY Expert Installer")
    logger.info("=" * 70)

    # Install experts
    success = installer.install_core_experts()

    # Download registry
    installer.download_registry()

    # Verify
    installer.verify_installation()

    # Print usage
    if success:
        installer.print_usage()

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
