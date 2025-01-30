$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$vm = Get-VM -Id '{{.ID}}'
$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json

$vmHardDiskDriveObject = @( $vm | Get-VMHardDiskDrive | Where-Object { 
		$_.Path -eq $vmHardDiskDrive.Path
	} | ForEach-Object { 
		@{
			ControllerType                = $_.ControllerType;
			ControllerNumber              = $_.ControllerNumber;
			ControllerLocation            = $_.ControllerLocation;
			Path                          = $_.Path;
			DiskNumber                    = if ($_.DiskNumber -eq $null) { 4294967295 } else { $_.DiskNumber };
			ResourcePoolName              = $_.PoolName;
			SupportPersistentReservations = $_.SupportPersistentReservations;
			MaximumIops                   = $_.MaximumIops;
			MinimumIops                   = $_.MinimumIops;
			QosPolicyId                   = $_.QosPolicyId;	
			OverrideCacheAttributes       = $_.WriteHardeningMethod;
		} 
	}
)

if ($vmHardDiskDriveObject) {
	$out = ConvertTo-Json -InputObject $vmHardDiskDriveObject[0]
	$out
	return
}

$NewVMHardDiskDriveArgs = @{
	VMName         = $vm.Name
	ControllerType = $vmHardDiskDrive.ControllerType
	Path           = $vmHardDiskDrive.Path
}
Add-VMHardDiskDrive @NewVMHardDiskDriveArgs

$vmHardDiskDriveObject = @( $vm | Get-VMHardDiskDrive | Where-Object { 
		$_.Path -eq $vmHardDiskDrive.Path
	} | ForEach-Object { 
		@{
			ControllerType                = $_.ControllerType;
			ControllerNumber              = $_.ControllerNumber;
			ControllerLocation            = $_.ControllerLocation;
			Path                          = $_.Path;
			DiskNumber                    = if ($_.DiskNumber -eq $null) { 4294967295 } else { $_.DiskNumber };
			ResourcePoolName              = $_.PoolName;
			SupportPersistentReservations = $_.SupportPersistentReservations;
			MaximumIops                   = $_.MaximumIops;
			MinimumIops                   = $_.MinimumIops;
			QosPolicyId                   = $_.QosPolicyId;	
			OverrideCacheAttributes       = $_.WriteHardeningMethod;
		} 
	}
)

if ($vmHardDiskDriveObject) {
	$out = ConvertTo-Json -InputObject $vmHardDiskDriveObject[0]
	$out
}
else {
	"{}"
}
