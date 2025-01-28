$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json
$controllerLocation = '{{.ControllerLocation}}' | ConvertFrom-Json
$controllerNumber = '{{.ControllerNumber}}' | ConvertFrom-Json

$vm = Get-VM -Name '{{.VMName}}*' | Where-Object { $_.Name -eq '{{.VMName}}' }
$vmHardDiskDrivesObject = @($vm | Get-VMHardDiskDrive -ControllerLocation $controllerLocation -ControllerNumber $controllerNumber )
if (!$vmHardDiskDrivesObject) {
  throw "VM hard disk drive does not exist - $controllerLocation $controllerNumber"
}
	
$SetVMHardDiskDriveArgs = @{}
$SetVMHardDiskDriveArgs.VMName = $vmHardDiskDrivesObject.VMName
$SetVMHardDiskDriveArgs.ControllerType = $vmHardDiskDrivesObject.ControllerType
$SetVMHardDiskDriveArgs.ControllerLocation = $vmHardDiskDrivesObject.ControllerLocation
$SetVMHardDiskDriveArgs.ControllerNumber = $vmHardDiskDrivesObject.ControllerNumber
$SetVMHardDiskDriveArgs.ToControllerLocation = $vmHardDiskDrive.ControllerLocation
$SetVMHardDiskDriveArgs.ToControllerNumber = $vmHardDiskDrive.ControllerNumber
$SetVMHardDiskDriveArgs.Path = $vmHardDiskDrive.Path
if ($vmHardDiskDrive.DiskNumber -lt 4294967295) {
		$SetVMHardDiskDriveArgs.DiskNumber = $vmHardDiskDrive.DiskNumber
}
if ($vmHardDiskDrivesObject.ResourcePoolName -ne $vmHardDiskDrive.ResourcePoolName) {
  if ($vmHardDiskDrive.ResourcePoolName) {
    $SetVMHardDiskDriveArgs.ResourcePoolName = $vmHardDiskDrive.ResourcePoolName
  }
  else {
    throw "Unable to remove resource pool $($vmHardDiskDrive.ResourcePoolName) from hard disk drive $(ConvertTo-Json -InputObject $vmHardDiskDrivesObject)"
  }
}
$SetVMHardDiskDriveArgs.SupportPersistentReservations = $vmHardDiskDrive.SupportPersistentReservations
$SetVMHardDiskDriveArgs.MaximumIops = $vmHardDiskDrive.MaximumIops
$SetVMHardDiskDriveArgs.MinimumIops = $vmHardDiskDrive.MinimumIops
$SetVMHardDiskDriveArgs.QosPolicyId = $vmHardDiskDrive.QosPolicyId
$SetVMHardDiskDriveArgs.OverrideCacheAttributes = $vmHardDiskDrive.OverrideCacheAttributes	
$SetVMHardDiskDriveArgs.AllowUnverifiedPaths = $true
	
Set-VMHardDiskDrive @SetVMHardDiskDriveArgs
	