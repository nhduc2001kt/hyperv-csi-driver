$ErrorActionPreference = 'Stop'

$vm = Get-VM -Id '{{.ID}}'
$vmHardDiskDrivesObject = @( $vm | Get-VMHardDiskDrive | ForEach-Object { 
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

if ($vmHardDiskDrivesObject) {
  $vmHardDiskDrives = ConvertTo-Json -InputObject $vmHardDiskDrivesObject
  $vmHardDiskDrives
}
else {
  "[]"
}