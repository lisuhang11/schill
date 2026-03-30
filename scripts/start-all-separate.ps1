# SoChill Microservices Startup Script - Windows PowerShell
# Start each service in separate window for easy log viewing

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SoChill Microservices Startup" -ForegroundColor Cyan
Write-Host "  (Separate Windows)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get script directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptPath

Write-Host "Project Root: $projectRoot" -ForegroundColor Green
Write-Host ""

# Define service list (start RPC first, then API)
$services = @(
    @{ Name = "User RPC"; Path = "service/user/rpc"; File = "user.go" },
    @{ Name = "Relation RPC"; Path = "service/relation/rpc"; File = "relation.go" },
    @{ Name = "Content RPC"; Path = "service/content/rpc"; File = "content.go" },
    @{ Name = "Comment RPC"; Path = "service/comment/rpc"; File = "comment.go" },
    @{ Name = "User API"; Path = "service/user/api"; File = "user.go" },
    @{ Name = "Relation API"; Path = "service/relation/api"; File = "relation.go" },
    @{ Name = "Content API"; Path = "service/content/api"; File = "content.go" },
    @{ Name = "Comment API"; Path = "service/comment/api"; File = "comment.go" }
)

# Start service in new window function
function Start-Service-Window {
    param(
        [string]$name,
        [string]$path,
        [string]$file
    )
    
    $fullPath = Join-Path $projectRoot $path
    $fullFilePath = Join-Path $fullPath $file
    
    if (-not (Test-Path $fullFilePath)) {
        Write-Host "[ERROR] File not found: $fullFilePath" -ForegroundColor Red
        return
    }
    
    Write-Host "[START] $name..." -ForegroundColor Yellow
    
    try {
        $command = "cd '$fullPath'; go run '$file'; pause"
        
        Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", $command -WindowStyle Normal
        Write-Host "[OK] $name started" -ForegroundColor Green
    }
    catch {
        Write-Host "[FAIL] $name failed to start: $_" -ForegroundColor Red
    }
}

# Start all services
Write-Host "Starting services..." -ForegroundColor Cyan
Write-Host ""

foreach ($service in $services) {
    Start-Service-Window -name $service.Name -path $service.Path -file $service.File
    Start-Sleep -Milliseconds 800
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  All service start commands sent!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Service List:" -ForegroundColor White
foreach ($service in $services) {
    Write-Host "  - $($service.Name)" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Tip: Each service runs in separate PowerShell window" -ForegroundColor Yellow
Write-Host "     Close window to stop the service" -ForegroundColor Yellow
Write-Host ""
