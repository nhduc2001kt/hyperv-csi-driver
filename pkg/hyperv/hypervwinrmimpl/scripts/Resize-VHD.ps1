$ErrorActionPreference = 'Stop'

$path = '{{.Path}}'
$size = '{{.Size}}' | ConvertFrom-Json

$vhd = Get-VHD -Path "$path"
if ($vhd.Size -ne $size) {
  Resize-VHD -Path "$path" -SizeBytes $size
}