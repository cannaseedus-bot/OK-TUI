# MICRONAUT ORCHESTRATOR (SCO/1 projection only)

$Root = Split-Path $MyInvocation.MyCommand.Path
$IO = Join-Path $Root "io"
$Chat = Join-Path $IO "chat.txt"
$Stream = Join-Path $IO "stream.txt"

function Invoke-ObjectOp {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $false)]
        $Input
    )

    throw "Object μ-op '$Name' not bound in host orchestrator."
}

Write-Host "Micronaut online."

$lastSize = 0

while ($true) {
    if (Test-Path $Chat) {
        $size = (Get-Item $Chat).Length
        if ($size -gt $lastSize) {
            $entry = Get-Content $Chat -Raw
            $lastSize = $size

            # ---- CM-1 VERIFY ----
            $isValid = Invoke-ObjectOp -Name "cm1_verify" -Input $entry
            if (-not $isValid) {
                Write-Host "CM-1 violation"
                continue
            }

            # ---- SEMANTIC EXTRACTION ----
            $signal = Invoke-ObjectOp -Name "kuhul_tsg" -Input $entry

            # ---- INFERENCE (SEALED) ----
            $response = Invoke-ObjectOp -Name "scxq2_infer" -Input $signal

            # ---- STREAM OUTPUT ----
            Add-Content $Stream (">> $response")
        }
    }

    Start-Sleep -Milliseconds 200
}
