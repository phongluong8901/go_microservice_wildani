param([string]$target = "linux")
$services = @("api-gateway", "auth-service", "user-service", "wallet-service", "ledger-service", "transaction-service")

foreach ($s in $services) {
    Write-Host "Building $s for $target..."
    $binDir = "$s/bin"
    if (!(Test-Path $binDir)) { New-Item -ItemType Directory -Force -Path $binDir | Out-Null }
    
    $env:CGO_ENABLED = "0"
    if ($target -eq "linux") {
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
        go build -ldflags="-w -s" -o "./$s/bin/$s-linux" "./$s/cmd/main.go"
    } elseif ($target -eq "windows") {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        go build -ldflags="-w -s" -o "./$s/bin/$s.exe" "./$s/cmd/main.go"
    } else {
        go build -o "./$s/bin/$s" "./$s/cmd/main.go"
    }
}
Write-Host "All services built successfully!"