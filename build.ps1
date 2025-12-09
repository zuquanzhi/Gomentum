$AppName = "gomentum.exe"
$SourceDir = "./cmd/gomentum"

Write-Host "Building $AppName..."
go build -o $AppName $SourceDir

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!"
} else {
    Write-Host "Build failed."
    exit 1
}
