$ErrorActionPreference = 'Stop'

$controllerLocation = '{{.ControllerLocation}}' | ConvertFrom-Json
$controllerNumber = '{{.ControllerNumber}}' | ConvertFrom-Json

Get-VMHardDiskDrive -VMName '{{.VMName}}' -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation | Remove-VMHardDiskDrive