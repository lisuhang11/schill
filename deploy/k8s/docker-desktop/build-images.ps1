param(
    [string]$Tag = "latest"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\\..\\..")
Set-Location $repoRoot
$images = @(
    @{ Name = "schill-gateway"; Dockerfile = "service/gateway/Dockerfile"; Context = "." },
    @{ Name = "schill-user-rpc"; Dockerfile = "service/user/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-content-rpc"; Dockerfile = "service/content/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-feed-rpc"; Dockerfile = "service/feed/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-comment-rpc"; Dockerfile = "service/comment/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-interaction-rpc"; Dockerfile = "service/interaction/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-relation-rpc"; Dockerfile = "service/relation/rpc/Dockerfile"; Context = "." },
    @{ Name = "schill-search-api"; Dockerfile = "service/search/api/Dockerfile"; Context = "." },
    @{ Name = "schill-canal"; Dockerfile = "service/canal/Dockerfile"; Context = "." }
)

foreach ($image in $images) {
    $tagged = "$($image.Name):$Tag"
    Write-Host "Building $tagged"
    docker build -f $image.Dockerfile -t $tagged $image.Context
}

$esImage = "schill-elasticsearch:8.6.1-plugins"
Write-Host "Building $esImage"
docker build -f deploy/k8s/docker-desktop/elasticsearch/Dockerfile -t $esImage deploy/k8s/docker-desktop/elasticsearch

Write-Host ""
Write-Host "Build complete."
Write-Host "Suggested deploy command:"
Write-Host "kubectl apply -k deploy/k8s/docker-desktop"
