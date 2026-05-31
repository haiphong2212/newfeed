$basePath = "d:\newfeed\services\auth-service"
$directories = @(
    "$basePath",
    "$basePath\internal\domain",
    "$basePath\internal\repository",
    "$basePath\internal\infrastructure",
    "$basePath\internal\usecase",
    "$basePath\internal\config",
    "$basePath\internal\delivery\grpc",
    "$basePath\internal\delivery\http",
    "$basePath\proto"
)

foreach ($dir in $directories) {
    if (-not (Test-Path -Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
}
Write-Host "Directories created successfully"
