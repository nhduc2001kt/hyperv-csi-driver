$ErrorActionPreference = 'Stop'

$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json
$vm = Get-VM -Id '{{.ID}}'

Get-VMHardDiskDrive -VMName $vm.Name | Where-Object { 
  $_.Path -eq $vmHardDiskDrive.Path
} | Remove-VMHardDiskDrive