$ErrorActionPreference = 'Stop'

$path = '{{.Path}}'
$targetDirectory = (Split-Path "$path" -Parent)
$targetName = (Split-Path "$path" -Leaf)
$targetName = $targetName.Substring(0, $targetName.LastIndexOf('.')).split('\')[-1]

Get-ChildItem -Path $targetDirectory | Where-Object { $_.BaseName.StartsWith($targetName) } | ForEach-Object {
  Remove-Item $_.FullName -Force
}