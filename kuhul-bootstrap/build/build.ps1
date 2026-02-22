Write-Host "Building KUHUL compiler..."

cargo build --release

New-Item -ItemType Directory -Force -Path output

Write-Host "Done"
