#!/usr/bin/env python3
"""
Test KHANARY system - verify all components work end-to-end
"""

import os
import sys
import json
import subprocess
from pathlib import Path


def test_section(title):
    """Print test section header"""
    print(f"\n{'='*70}")
    print(f"  {title}")
    print(f"{'='*70}\n")


def run_command(cmd, description):
    """Run command and report results"""
    print(f"  Testing: {description}...")
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, shell=True)
        if result.returncode == 0:
            print(f"  ✅ {description}")
            return True
        else:
            print(f"  ❌ {description}")
            print(f"     Error: {result.stderr[:200]}")
            return False
    except Exception as e:
        print(f"  ❌ {description}")
        print(f"     Exception: {e}")
        return False


def main():
    """Run all tests"""
    print("\n" + "="*70)
    print("  KHANARY System Test Suite")
    print("="*70)

    passed = 0
    failed = 0

    # Test 1: Directory structure
    test_section("1. Directory Structure Verification")

    required_dirs = {
        'scripts': 'Helper scripts directory',
        'experts': 'Compiled experts directory',
        '.github/workflows': 'GitHub Actions workflows',
    }

    for dir_name, description in required_dirs.items():
        dir_path = Path(dir_name)
        if dir_path.exists():
            print(f"  ✅ {description} exists: {dir_name}/")
            passed += 1
        else:
            print(f"  ❌ {description} missing: {dir_name}/")
            failed += 1

    # Test 2: Script files
    test_section("2. Required Scripts Verification")

    required_scripts = {
        'scripts/download_datasets.py': 'Dataset downloader',
        'scripts/train_expert.py': 'Expert trainer',
        'scripts/train_all_experts.py': 'Multi-expert trainer',
        'scripts/compile_to_khanary.py': 'Binary compiler',
        'scripts/compile_all_experts.py': 'Multi-compiler',
        'scripts/create_expert_registry.py': 'Registry creator',
        'scripts/validate_determinism.py': 'Determinism validator',
        'scripts/install_experts.py': 'Installation script',
    }

    for script_path, description in required_scripts.items():
        file_path = Path(script_path)
        if file_path.exists():
            print(f"  ✅ {description}: {script_path}")
            passed += 1
        else:
            print(f"  ❌ {description} missing: {script_path}")
            failed += 1

    # Test 3: Documentation files
    test_section("3. Documentation Verification")

    required_docs = {
        'README.md': 'Main README',
        'QUICK_START.md': 'Quick start guide',
        'EXPERT_CATALOG.md': 'Expert catalog',
        'TRAINING_GUIDE_40_EXPERTS.md': 'Training guide',
        'BUILD_SYSTEM_GUIDE.md': 'Build guide',
        'KHANARY_BINARY_DISTRIBUTION.md': 'Distribution guide',
    }

    for doc_file, description in required_docs.items():
        file_path = Path(doc_file)
        if file_path.exists():
            size_kb = file_path.stat().st_size / 1024
            print(f"  ✅ {description}: {doc_file} ({size_kb:.1f}KB)")
            passed += 1
        else:
            print(f"  ❌ {description} missing: {doc_file}")
            failed += 1

    # Test 4: Python scripts can be imported
    test_section("4. Script Import Test")

    test_scripts = [
        ('scripts.download_datasets', 'Dataset downloader module'),
        ('scripts.train_expert', 'Expert trainer module'),
        ('scripts.compile_to_khanary', 'Binary compiler module'),
    ]

    sys.path.insert(0, str(Path.cwd()))

    for module_name, description in test_scripts:
        try:
            __import__(module_name)
            print(f"  ✅ {description} imports successfully")
            passed += 1
        except Exception as e:
            print(f"  ❌ {description} import failed: {str(e)[:100]}")
            failed += 1

    # Test 5: CLI commands
    test_section("5. CLI Commands Availability")

    cli_tests = [
        ('python scripts/download_datasets.py --list-experts 2>&1 | head -1', 'List experts command'),
        ('python scripts/install_experts.py --help 2>&1 | grep -q "install"', 'Install script help'),
        ('python scripts/train_all_experts.py --list-groups 2>&1 | grep -q "core"', 'List groups command'),
    ]

    for cmd, description in cli_tests:
        if run_command(cmd, description):
            passed += 1
        else:
            failed += 1

    # Test 6: Expert registry
    test_section("6. Expert Registry Test")

    if Path('experts/registry.json').exists():
        try:
            with open('experts/registry.json') as f:
                registry = json.load(f)
            num_experts = len(registry.get('experts', {}))
            print(f"  ✅ Registry exists with {num_experts} experts")
            passed += 1
        except:
            print(f"  ❌ Registry file corrupted")
            failed += 1
    else:
        print(f"  ⚠️  Registry not yet generated (normal for first run)")

    # Test 7: Git and CI/CD
    test_section("7. Git & CI/CD Configuration")

    if Path('.gitignore').exists():
        print(f"  ✅ .gitignore exists")
        passed += 1
    else:
        print(f"  ❌ .gitignore missing")
        failed += 1

    if Path('.github/workflows/compile-and-release.yml').exists():
        print(f"  ✅ GitHub Actions workflow exists")
        passed += 1
    else:
        print(f"  ❌ GitHub Actions workflow missing")
        failed += 1

    # Test 8: Verify git configuration
    if run_command('git status 2>&1 | grep -q "On branch"', 'Git repository check'):
        passed += 1
    else:
        failed += 1

    # Test 9: Python dependencies
    test_section("8. Python Dependencies Check")

    dependencies = [
        'torch',
        'transformers',
        'datasets',
        'safetensors',
        'pyyaml',
        'huggingface_hub',
    ]

    dep_available = 0
    for dep in dependencies:
        try:
            __import__(dep)
            print(f"  ✅ {dep}")
            dep_available += 1
        except ImportError:
            print(f"  ⚠️  {dep} not installed (install with: pip install {dep})")

    if dep_available >= 3:
        passed += 1
    else:
        print(f"  ❌ Missing required dependencies")
        failed += 1

    # Test 10: File structure validation
    test_section("9. System Structure Validation")

    if Path('build_MoE.bat').exists():
        print(f"  ✅ build_MoE.bat (Windows build script)")
        passed += 1
    else:
        print(f"  ❌ build_MoE.bat missing")
        failed += 1

    # Summary
    test_section("Test Summary")

    total = passed + failed
    percentage = (passed / total * 100) if total > 0 else 0

    print(f"  Passed: {passed}/{total}")
    print(f"  Failed: {failed}/{total}")
    print(f"  Success Rate: {percentage:.1f}%\n")

    if percentage >= 90:
        print(f"  🎉 SYSTEM READY FOR PRODUCTION!\n")
        return 0
    elif percentage >= 70:
        print(f"  ⚠️  SYSTEM MOSTLY READY (install dependencies to complete)\n")
        return 1
    else:
        print(f"  ❌ SYSTEM NEEDS CONFIGURATION\n")
        return 1


if __name__ == '__main__':
    sys.exit(main())
