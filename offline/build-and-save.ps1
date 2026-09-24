# ============================================================================
# LAF offline image build and export
# Host: Windows with Docker Desktop running + Internet
# Output: dist-offline\laf-images.tar
# Live progress is shown on screen and also written to dist-offline\build.log
# ============================================================================
$ErrorActionPreference = "Continue"
Set-Location -Path (Join-Path $PSScriptRoot "..")
$root = (Get-Location).Path
$log  = Join-Path $root "dist-offline\build.log"
New-Item -ItemType Directory -Force -Path "dist-offline" | Out-Null
"==== LAF build started $(Get-Date) ====" | Out-File -FilePath $log -Encoding utf8

function Say($msg, $color) { Write-Host $msg -ForegroundColor $color; "`n$msg" | Out-File -FilePath $log -Append -Encoding utf8 }

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  LAF offline image build and export"
Write-Host "  Project root : $root"
Write-Host "  Log file     : $log"
Write-Host "============================================================" -ForegroundColor Cyan

function RunLive($desc, $exe, $argList) {
    Say ">> $desc" "Yellow"
    & $exe @argList 2>&1 | Tee-Object -FilePath $log -Append
    return $LASTEXITCODE
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Say "[ERROR] docker not found. Install Docker Desktop and start it." "Red"; Read-Host "Press Enter"; exit 1
}
if ((RunLive "check docker engine" docker @("info")) -ne 0) {
    Say "[ERROR] Docker engine not running. Open Docker Desktop until 'Engine running'." "Red"; Read-Host "Press Enter"; exit 1
}
if ((RunLive "check docker compose" docker @("compose","version")) -ne 0) {
    Say "[ERROR] docker compose v2 not available. Update Docker Desktop." "Red"; Read-Host "Press Enter"; exit 1
}

$mirrors = @("docker.m.daocloud.io","docker.1panel.live","dockerproxy.cn")
$images  = @("mysql:8.0","golang:1.26.5-alpine","node:20-alpine","nginx:1.27-alpine","alpine:3.20")

function PullImage($img) {
    foreach ($m in $mirrors) {
        Say "  [$img] try mirror: $m" "Cyan"
        & docker pull "$m/library/$img" 2>&1 | Tee-Object -FilePath $log -Append
        if ($LASTEXITCODE -eq 0) {
            & docker tag "$m/library/$img" $img 2>&1 | Tee-Object -FilePath $log -Append
            if ($LASTEXITCODE -eq 0) { Say "  [$img] OK (from $m)" "Green"; return $true }
        }
    }
    Say "  [$img] try direct Docker Hub: $img" "Cyan"
    & docker pull $img 2>&1 | Tee-Object -FilePath $log -Append
    if ($LASTEXITCODE -eq 0) { Say "  [$img] OK (from Docker Hub)" "Green"; return $true }
    Say "  [$img] NOT pulled; compose build will retry" "Yellow"
    return $false
}

Write-Host "[1/3] Pre-pulling base images via China mirrors ..." -ForegroundColor Green
foreach ($img in $images) { [void](PullImage $img) }

Write-Host "[2/3] docker compose build ..." -ForegroundColor Green
$code = RunLive "docker compose build" docker @("compose","build")
if ($code -ne 0) { Say "[FAILED] docker compose build failed." "Red"; Read-Host "Press Enter"; exit 1 }

Write-Host "[3/3] docker save -o dist-offline\laf-images.tar ..." -ForegroundColor Green
$code = RunLive "docker save" docker @("save","-o","dist-offline\laf-images.tar","laf-backend:latest","laf-frontend:latest","mysql:8.0")
if ($code -ne 0) { Say "[FAILED] docker save failed." "Red"; Read-Host "Press Enter"; exit 1 }

$tar = Join-Path $root "dist-offline\laf-images.tar"
if (-not (Test-Path $tar)) { Say "[FAILED] tar file was NOT created." "Red"; Read-Host "Press Enter"; exit 1 }

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  DONE. Output file: $tar"
Write-Host "============================================================" -ForegroundColor Cyan
Get-Item $tar | Format-List Name, Length, FullName
Write-Host "Next:"
Write-Host "  1) Zip the whole project (include dist-offline; may exclude frontend-demo\node_modules)"
Write-Host "  2) Upload to server via Aliyun Workbench 'Upload File' into /root"
Write-Host "  3) On server run:  cd /opt/LAF && bash offline/load-and-run.sh"
Read-Host "Press Enter to close"
