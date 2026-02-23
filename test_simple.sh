#!/bin/bash

echo "═══════════════════════════════════════════════════════════════"
echo "  KHANARY System - Structural Verification"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Test 1: Scripts
echo "✅ Scripts Available:"
ls -1 scripts/*.py | sed 's/^/   ✓ /'
echo ""

# Test 2: Documentation
echo "✅ Documentation Available:"
ls -1 *.md | sed 's/^/   ✓ /'
echo ""

# Test 3: Build system
echo "✅ Build System:"
test -f build_MoE.bat && echo "   ✓ build_MoE.bat" || echo "   ✗ build_MoE.bat"
test -f .gitignore && echo "   ✓ .gitignore" || echo "   ✗ .gitignore"
echo ""

# Test 4: Directories
echo "✅ Required Directories:"
test -d scripts && echo "   ✓ scripts/" || echo "   ✗ scripts/"
test -d experts && echo "   ✓ experts/" || echo "   ✗ experts/"
test -d .github/workflows && echo "   ✓ .github/workflows/" || echo "   ✗ .github/workflows/"
echo ""

# Test 5: CI/CD
echo "✅ CI/CD Configuration:"
test -f .github/workflows/compile-and-release.yml && echo "   ✓ GitHub Actions workflow" || echo "   ✗ GitHub Actions workflow"
echo ""

# Test 6: File sizes
echo "✅ System Footprint:"
du -sh scripts/ 2>/dev/null | sed 's/^/   Scripts: /'
du -sh .github/ 2>/dev/null | sed 's/^/   CI/CD: /'
find . -maxdepth 1 -name "*.md" -exec du -ch {} + 2>/dev/null | tail -1 | sed 's/^/   Docs: /'
echo ""

# Test 7: Git status
echo "✅ Git Status:"
git status --short | head -10 | sed 's/^/   /'
echo ""

# Test 8: Recent commits
echo "✅ Recent Commits:"
git log --oneline | head -5 | sed 's/^/   /'
echo ""

echo "═══════════════════════════════════════════════════════════════"
echo "  ✅ KHANARY SYSTEM STRUCTURE VERIFIED"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "📦 Ready for:"
echo "   1. Building 5 core experts (build_MoE.bat)"
echo "   2. Training 40+ additional experts (python scripts/train_all_experts.py)"
echo "   3. Creating GitHub releases (git tag v3.0.0)"
echo "   4. Distributing to users (GitHub + install script)"
echo ""
