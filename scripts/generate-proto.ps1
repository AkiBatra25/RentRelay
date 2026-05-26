$ErrorActionPreference = "Stop"

$goBin = Join-Path $env:USERPROFILE "go\bin"
$env:PATH = "$goBin;$env:PATH"

buf generate
