$ErrorActionPreference = 'Stop'

$path = '{{.Path}}'

if (Test-Path $path) {
  $exists = ConvertTo-Json -InputObject @{ Exists = $true }
  $exists
}
else {
  $exists = ConvertTo-Json -InputObject @{ Exists = $false }
  $exists
}