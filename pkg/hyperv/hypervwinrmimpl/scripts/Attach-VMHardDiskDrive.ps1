$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json
$NewVMHardDiskDriveArgs = @{
	VMName         = $vmHardDiskDrive.VMName
	ControllerType = $vmHardDiskDrive.ControllerType
	Path           = $vmHardDiskDrive.Path
}

Add-VMHardDiskDrive @NewVMHardDiskDriveArgs