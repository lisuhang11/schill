# SChill Service Startup Script (Windows PowerShell)
# Start all RPC and API services in order

# Set console encoding to UTF-8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$env:LANG = "zh_CN.UTF-8"

# Change code page to UTF-8
chcp 65001 | Out-Null

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SChill Service Startup Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get project root directory reliably
$scriptPath = $MyInvocation.MyCommand.Path
$scriptsDir = Split-Path -Parent $scriptPath
$PROJECT_ROOT = Split-Path -Parent $scriptsDir
Write-Host "Project Root: $PROJECT_ROOT" -ForegroundColor Yellow
Write-Host ""

# Define service list
$SERVICES = @(
    @{ Name = "user-rpc"; Path = "service/user/rpc"; Cmd = "go run user.go" },
    @{ Name = "content-rpc"; Path = "service/content/rpc"; Cmd = "go run content.go" },
    @{ Name = "relation-rpc"; Path = "service/relation/rpc"; Cmd = "go run relation.go" },
    @{ Name = "user-api"; Path = "service/user/api"; Cmd = "go run user.go" },
    @{ Name = "content-api"; Path = "service/content/api"; Cmd = "go run content.go" },
    @{ Name = "relation-api"; Path = "service/relation/api"; Cmd = "go run relation.go" },
    @{ Name = "comment-api"; Path = "service/comment/api"; Cmd = "go run comment.go" }
)

# Store all processes
$processes = @()

function Start-Service {
    param(
        [hashtable]$service
    )
    
    $servicePath = Join-Path $PROJECT_ROOT $service.Path
    
    if (-not (Test-Path $servicePath)) {
        Write-Host "[ERROR] Service directory not found: $servicePath" -ForegroundColor Red
        return $false
    }
    
    Write-Host "[Starting] $($service.Name)..." -ForegroundColor Green
    
    try {
        $process = Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$servicePath'; $($service.Cmd)" -PassThru -WindowStyle Normal
        $processes += $process
        Write-Host "[SUCCESS] $($service.Name) started (PID: $($process.Id))" -ForegroundColor Green
        Start-Sleep -Seconds 2
        return $true
    } catch {
        Write-Host "[FAILED] Cannot start $($service.Name): $_" -ForegroundColor Red
        return $false
    }
}

Write-Host "Starting services..." -ForegroundColor Cyan
Write-Host ""

$successCount = 0
foreach ($service in $SERVICES) {
    if (Start-Service $service) {
        $successCount++
    }
    Write-Host ""
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Startup Complete!" -ForegroundColor Cyan
Write-Host "Successfully started: $successCount / $($SERVICES.Count) services" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press Ctrl+C to stop all services" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Wait for user interrupt
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Write-Host ""
    Write-Host "Stopping all services..." -ForegroundColor Yellow
    foreach ($process in $processes) {
        if (-not $process.HasExited) {
            try {
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
                Write-Host "Stopped process PID: $($process.Id)" -ForegroundColor Gray
            } catch {
                Write-Host "Failed to stop process PID: $($process.Id)" -ForegroundColor Red
            }
        }
    }
    Write-Host "All services stopped" -ForegroundColor Green
}
