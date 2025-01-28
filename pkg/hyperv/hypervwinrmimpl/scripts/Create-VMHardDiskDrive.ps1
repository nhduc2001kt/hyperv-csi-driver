$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json
$NewVMHardDiskDriveArgs = @{
	VMName                        = $vmHardDiskDrive.VMName
	ControllerType                = $vmHardDiskDrive.ControllerType
	ControllerNumber              = $vmHardDiskDrive.ControllerNumber
	ControllerLocation            = $vmHardDiskDrive.ControllerLocation
	Path                          = $vmHardDiskDrive.Path
	ResourcePoolName              = $vmHardDiskDrive.ResourcePoolName
	SupportPersistentReservations = $vmHardDiskDrive.SupportPersistentReservations
	MaximumIops                   = $_.MaximumIops;
	MinimumIops                   = $_.MinimumIops;
	QosPolicyId                   = $_.QosPolicyId;
	OverrideCacheAttributes       = $vmHardDiskDrive.OverrideCacheAttributes
	AllowUnverifiedPaths          = $true
}
if ($vmHardDiskDrive.DiskNumber -lt 4294967295) {
	$NewVMHardDiskDriveArgs.DiskNumber = $vmHardDiskDrive.DiskNumber
}

Add-VMHardDiskDrive @NewVMHardDiskDriveArgs