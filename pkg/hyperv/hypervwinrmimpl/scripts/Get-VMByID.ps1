$ErrorActionPreference = 'Stop'

$vmObject = Get-VM -Id '{{.ID}}' -ErrorAction SilentlyContinue | ForEach-Object { 
  @{
    Name                                = $_.Name;
    Path                                = $_.Path;
    Generation                          = $_.Generation;
    AutomaticCriticalErrorAction        = $_.AutomaticCriticalErrorAction;
    AutomaticCriticalErrorActionTimeout = $_.AutomaticCriticalErrorActionTimeout;
    AutomaticStartAction                = $_.AutomaticStartAction;
    AutomaticStartDelay                 = $_.AutomaticStartDelay;
    AutomaticStopAction                 = $_.AutomaticStopAction;
    CheckpointType                      = $_.CheckpointType;
    DynamicMemory                       = $_.DynamicMemoryEnabled;
    GuestControlledCacheTypes           = $_.GuestControlledCacheTypes;
    HighMemoryMappedIoSpace             = $_.HighMemoryMappedIoSpace;
    LockOnDisconnect                    = $_.LockOnDisconnect;
    LowMemoryMappedIoSpace              = $_.LowMemoryMappedIoSpace;
    MemoryMaximumBytes                  = $_.MemoryMaximum;
    MemoryMinimumBytes                  = $_.MemoryMinimum;
    MemoryStartupBytes                  = $_.MemoryStartup;
    Notes                               = $_.Notes;
    ProcessorCount                      = $_.ProcessorCount;
    SmartPagingFilePath                 = $_.SmartPagingFilePath;
    SnapshotFileLocation                = $_.SnapshotFileLocation;
    StaticMemory                        = !$_.DynamicMemoryEnabled;
  } 
}

if ($vmObject) {
  $vm = ConvertTo-Json -InputObject $vmObject
  $vm
}
else {
  "{}"
}