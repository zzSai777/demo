param(
    [string]$HostName = "10.235.106.18",
    [string]$User = "root",
    [string]$ServiceName = "go-demo",
    [string]$RemoteDir = "/opt/go-demo",
    [string]$Port = "8080"
)

$ErrorActionPreference = "Stop"

$target = $HostName
if ($User -ne "") {
    $target = "$User@$HostName"
}

$goCommand = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCommand -and (Test-Path "C:\Program Files\Go\bin\go.exe")) {
    $goCommand = Get-Item "C:\Program Files\Go\bin\go.exe"
}
if (-not $goCommand) {
    throw "Go is not installed or not available in PATH. Install Go locally before running this deploy script."
}

$env:GOOS = "linux"
$env:GOARCH = "amd64"
& $goCommand.Source test ./...
& $goCommand.Source build -o $ServiceName .

ssh -o StrictHostKeyChecking=accept-new $target "sudo mkdir -p $RemoteDir"
scp ".\$ServiceName" "${target}:/tmp/$ServiceName"
ssh $target "sudo mv /tmp/$ServiceName $RemoteDir/$ServiceName && sudo chmod +x $RemoteDir/$ServiceName"

$unit = @"
[Unit]
Description=Go Demo HTTP Service
After=network.target

[Service]
Type=simple
Environment=PORT=$Port
ExecStart=$RemoteDir/$ServiceName
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
"@

$encodedUnit = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($unit))
ssh $target "echo $encodedUnit | base64 -d | sudo tee /etc/systemd/system/$ServiceName.service >/dev/null && sudo systemctl daemon-reload && sudo systemctl enable --now $ServiceName && sudo systemctl restart $ServiceName"

ssh $target "systemctl is-active $ServiceName"
Invoke-WebRequest -UseBasicParsing "http://${HostName}:${Port}/healthz"
