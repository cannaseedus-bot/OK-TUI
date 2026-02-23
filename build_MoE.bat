@echo off
REM ============================================================================
REM KHANARY Mixture of Experts (MoE) Build System
REM Trains all domain-specific experts and compiles to .khμ binaries
REM ============================================================================

setlocal enabledelayedexpansion

REM Configuration
set PYTHON_CMD=python
set SCRIPT_DIR=%~dp0scripts
set CONFIGS_DIR=%~dp0configs
set CHECKPOINTS_DIR=%~dp0checkpoints
set EXPERTS_DIR=%~dp0experts
set LOGS_DIR=%~dp0logs
set DATASETS_DIR=%~dp0datasets

REM Create directories
if not exist "%CHECKPOINTS_DIR%" mkdir "%CHECKPOINTS_DIR%"
if not exist "%EXPERTS_DIR%" mkdir "%EXPERTS_DIR%"
if not exist "%LOGS_DIR%" mkdir "%LOGS_DIR%"
if not exist "%DATASETS_DIR%" mkdir "%DATASETS_DIR%"

REM Colors and formatting
set RED=[91m
set GREEN=[92m
set YELLOW=[93m
set BLUE=[94m
set RESET=[0m

echo.
echo ============================================================================
echo   KHANARY Mixture of Experts Build System
echo ============================================================================
echo.

REM ============================================================================
REM Phase 1: Environment Setup
REM ============================================================================
echo %BLUE%[1/5] Verifying Python Environment...%RESET%
%PYTHON_CMD% --version >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] Python not found. Please install Python 3.8+%RESET%
    exit /b 1
)
echo %GREEN%[OK] Python found%RESET%

REM Check required packages
echo %BLUE%[1/5] Checking dependencies...%RESET%
%PYTHON_CMD% -c "import torch, transformers, datasets, safetensors, pyyaml" >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%[INFO] Installing required packages...%RESET%
    %PYTHON_CMD% -m pip install -q torch transformers datasets safetensors pyyaml huggingface_hub
    if errorlevel 1 (
        echo %RED%[ERROR] Failed to install dependencies%RESET%
        exit /b 1
    )
)
echo %GREEN%[OK] All dependencies available%RESET%
echo.

REM ============================================================================
REM Phase 2: Download Datasets
REM ============================================================================
echo %BLUE%[2/5] Downloading HuggingFace Datasets...%RESET%

set EXPERTS=python security architecture performance sql

for %%E in (%EXPERTS%) do (
    echo %YELLOW%  • Downloading dataset for %%E expert...%RESET%
    %PYTHON_CMD% "%SCRIPT_DIR%\download_datasets.py" --expert %%E --output "%DATASETS_DIR%" >> "%LOGS_DIR%\download_%%E.log" 2>&1
    if errorlevel 1 (
        echo %RED%[WARNING] Failed to download %%E dataset%RESET%
        echo Check log: %LOGS_DIR%\download_%%E.log
    ) else (
        echo %GREEN%  ✓ %%E dataset ready%RESET%
    )
)
echo.

REM ============================================================================
REM Phase 3: Train Experts
REM ============================================================================
echo %BLUE%[3/5] Training Domain-Specific Experts...%RESET%
echo %YELLOW%  Training configuration:%RESET%
echo    • Model: Qwen-Coder-7B
echo    • Datasets: HF curated + Code-Feedback + diversity sets
echo    • Target accuracy: 90-95%%
echo    • Epochs: 3
echo    • Batch size: 32
echo.

for %%E in (%EXPERTS%) do (
    echo %YELLOW%  • Training %%E expert...%RESET%
    %PYTHON_CMD% "%SCRIPT_DIR%\train_expert.py" ^
        --expert %%E ^
        --config "%CONFIGS_DIR%\%%E_expert.yaml" ^
        --dataset-dir "%DATASETS_DIR%" ^
        --output "%CHECKPOINTS_DIR%" ^
        --epochs 3 ^
        --batch-size 32 ^
        --log-file "%LOGS_DIR%\train_%%E.log"

    if errorlevel 1 (
        echo %RED%[ERROR] Failed to train %%E expert%RESET%
        echo Check log: %LOGS_DIR%\train_%%E.log
        exit /b 1
    ) else (
        echo %GREEN%  ✓ %%E expert trained%RESET%
    )
)
echo.

REM ============================================================================
REM Phase 4: Compile to KHANARY Binary Format
REM ============================================================================
echo %BLUE%[4/5] Compiling Experts to KHANARY Format (.khμ)...%RESET%

for %%E in (%EXPERTS%) do (
    echo %YELLOW%  • Compiling %%E expert...%RESET%

    REM Find the latest checkpoint for this expert
    for /f "delims=" %%F in ('dir /b /od "%CHECKPOINTS_DIR%\%%E_*" ^| findstr /v ".lock" ^| tail -1') do (
        set LATEST_CHECKPOINT=%%F
    )

    %PYTHON_CMD% "%SCRIPT_DIR%\compile_to_khanary.py" ^
        --checkpoint "%CHECKPOINTS_DIR%\!LATEST_CHECKPOINT!" ^
        --output "%EXPERTS_DIR%\%%E.khμ" ^
        --format khanary_v0.2 ^
        --log-file "%LOGS_DIR%\compile_%%E.log"

    if errorlevel 1 (
        echo %RED%[ERROR] Failed to compile %%E expert%RESET%
        echo Check log: %LOGS_DIR%\compile_%%E.log
        exit /b 1
    ) else (
        echo %GREEN%  ✓ %%E expert compiled%RESET%
    )
)
echo.

REM ============================================================================
REM Phase 5: Validation & Verification
REM ============================================================================
echo %BLUE%[5/5] Validating Build...%RESET%

echo %YELLOW%  • Checking binary integrity...%RESET%
for %%E in (%EXPERTS%) do (
    if not exist "%EXPERTS_DIR%\%%E.khμ" (
        echo %RED%[ERROR] Missing %%E.khμ binary%RESET%
        exit /b 1
    )

    for /f %%S in ('powershell -Command "(Get-Item '%EXPERTS_DIR%\%%E.khμ').Length"') do (
        set SIZE=%%S
    )

    if !SIZE! LSS 50000 (
        echo %YELLOW%[WARNING] %%E.khμ seems too small (!SIZE! bytes)%RESET%
    ) else (
        echo %GREEN%  ✓ %%E.khμ valid (!SIZE! bytes)%RESET%
    )
)
echo.

echo %YELLOW%  • Validating determinism...%RESET%
%PYTHON_CMD% "%SCRIPT_DIR%\validate_determinism.py" ^
    --experts "%EXPERTS_DIR%\*.khμ" ^
    --runs 10 ^
    --log-file "%LOGS_DIR%\validate_determinism.log"

if errorlevel 1 (
    echo %YELLOW%[WARNING] Determinism validation incomplete%RESET%
) else (
    echo %GREEN%  ✓ Determinism validated%RESET%
)
echo.

echo %YELLOW%  • Generating checksums...%RESET%
cd /d "%EXPERTS_DIR%"
for %%E in (%EXPERTS%) do (
    for /f %%H in ('powershell -Command "Get-FileHash '%%E.khμ' -Algorithm SHA256 | Select-Object -ExpandProperty Hash"') do (
        echo %%H  %%E.khμ >> SHA256SUMS
    )
)
cd /d "%~dp0"
echo %GREEN%  ✓ SHA256SUMS generated%RESET%
echo.

REM ============================================================================
REM Summary
REM ============================================================================
echo.
echo %GREEN%============================================================================%RESET%
echo %GREEN%  BUILD SUCCESSFUL!%RESET%
echo %GREEN%============================================================================%RESET%
echo.

echo %BLUE%Generated Artifacts:%RESET%
echo %YELLOW%  Checkpoints (training weights):%RESET%
dir /b "%CHECKPOINTS_DIR%" | findstr /v ".lock"
echo.

echo %YELLOW%  Expert Binaries (.khμ):%RESET%
dir /b "%EXPERTS_DIR%\*.khμ"
echo.

echo %YELLOW%  Verification:%RESET%
if exist "%EXPERTS_DIR%\SHA256SUMS" (
    echo  ✓ SHA256SUMS created
) else (
    echo  ✗ SHA256SUMS missing
)
echo.

echo %BLUE%Next Steps:%RESET%
echo  1. Review logs: type "%LOGS_DIR%\*.log"
echo  2. Test experts: python scripts/test_experts.py
echo  3. Benchmark: python scripts/benchmark_experts.py
echo  4. Create release: git tag -a v2.1.0 -m "KHANARY Experts v2.1.0"
echo  5. Push: git push origin v2.1.0
echo.

echo %BLUE%Expert Registration:%RESET%
%PYTHON_CMD% "%SCRIPT_DIR%\create_expert_registry.py" --experts-dir "%EXPERTS_DIR%"
if errorlevel 1 (
    echo %YELLOW%[WARNING] Could not auto-register experts%RESET%
) else (
    echo %GREEN%  ✓ Experts registered%RESET%
)
echo.

echo %BLUE%Installation for Users:%RESET%
echo  python scripts/install_experts.py v2.1.0
echo.

echo %GREEN%Build complete!%RESET%
echo %BLUE%============================================================================%RESET%
echo.

endlocal
exit /b 0

REM ============================================================================
REM Error Handling
REM ============================================================================
:error
echo %RED%[ERROR] Build failed!%RESET%
echo Check logs in: %LOGS_DIR%
exit /b 1
