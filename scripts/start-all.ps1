# SoChill Microservices Startup Script - Windows PowerShell
# Start all API and RPC services

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SoChill Microservices Startup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get script directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptPath

Write-Host "Project Root: $projectRoot" -ForegroundColor Green
Write-Host ""

# Define service list
$services = @(
    @{ Name = "User RPC"; Path = "service/user/rpc"; File = "user.go" },
    @{ Name = "User API"; Path = "service/user/api"; File = "user.go" },
    @{ Name = "Relation RPC"; Path = "service/relation/rpc"; File = "relation.go" },
    @{ Name = "Relation API"; Path = "service/relation/api"; File = "relation.go" },
    @{ Name = "Content RPC"; Path = "service/content/rpc"; File = "content.go" },
    @{ Name = "Content API"; Path = "service/content/api"; File = "content.go" },
    @{ Name = "Comment RPC"; Path = "service/comment/rpc"; File = "comment.go" },
    @{ Name = "Comment API"; Path = "service/comment/api"; File = "comment.go" }
)

# Store all processes
$processes = @()

# Start service function
function Start-Service-Internal {
    param(
        [string]$name,
        [string]$path,
        [string]$file
    )
    
    $fullPath = Join-Path $projectRoot $path
    $fullFilePath = Join-Path $fullPath $file
    
    if (-not (Test-Path $fullFilePath)) {
        Write-Host "[ERROR] File not found: $fullFilePath" -ForegroundColor Red
        return $null
    }
    
    Write-Host "[START] $name..." -ForegroundColor Yellow
    
    try {
        $process = Start-Process -FilePath "go" -ArgumentList "run", $file -WorkingDirectory $fullPath -PassThru -WindowStyle Normal
        Write-Host "[OK] $name started (PID: $($process.Id))" -ForegroundColor Green
        return $process
    }
    catch {
        Write-Host "[FAIL] $name failed to start: $_" -ForegroundColor Red
        return $null
    }
}

# Stop all services function
function Stop-AllServices-Internal {
    Write-Host ""
    Write-Host "Stopping all services..." -ForegroundColor Yellow
    
    foreach ($process in $processes) {
        if ($process -and -not $process.HasExited) {
            try {
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
                Write-Host "[STOP] Service PID $($process.Id) stopped" -ForegroundColor Gray
            }
            catch {
                Write-Host "[WARN] Failed to stop service PID $($process.Id)" -ForegroundColor Yellow
            }
        }
    }
}

# Register Ctrl+C event handler
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action {
    Stop-AllServices-Internal
} | Out-Null

# Start all services
Write-Host "Starting services..." -ForegroundColor Cyan
Write-Host ""

foreach ($service in $services) {
    $process = Start-Service-Internal -name $service.Name -path $service.Path -file $service.File
    if ($process) {
        $processes += $process
    }
    Start-Sleep -Milliseconds 500
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  All services started!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Service List:" -ForegroundColor White
foreach ($service in $services) {
    Write-Host "  - $($service.Name)" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Press Ctrl+C to stop all services" -ForegroundColor Yellow
Write-Host ""

# Keep script running
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
}
finally {
    Stop-AllServices-Internal
}
