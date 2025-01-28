$ErrorActionPreference = 'Stop'

$path = '{{.Path}}'
$vhdObject = $null

if (Test-Path $path) {
  $vhdObject = Get-VHD -path $path | ForEach-Object { @{
      Path                    = $_.Path;
      BlockSize               = $_.BlockSize;
      LogicalSectorSize       = $_.LogicalSectorSize;
      PhysicalSectorSize      = $_.PhysicalSectorSize;
      ParentPath              = $_.ParentPath;
      FileSize                = $_.FileSize;
      Size                    = $_.Size;
      MinimumSize             = $_.MinimumSize;
      Attached                = $_.Attached;
      DiskNumber              = $_.DiskNumber;
      Number                  = $_.Number;
      FragmentationPercentage = $_.FragmentationPercentage;
      Alignment               = $_.Alignment;
      DiskIdentifier          = $_.DiskIdentifier;
      VHDType                 = $_.VHDType;
      VHDFormat               = $_.VHDFormat;
    } 
  }
}

if ($vhdObject) {
  $vhd = ConvertTo-Json -InputObject $vhdObject
  $vhd
}
else {
  "{}"
}