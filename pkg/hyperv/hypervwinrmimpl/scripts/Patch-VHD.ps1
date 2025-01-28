$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$source = '{{.Source}}'
$sourceVm = '{{.SourceVm}}'
$sourceDisk = '{{.SourceDisk}}' | ConvertFrom-Json
$vhd = '{{.VHDJson}}' | ConvertFrom-Json
$vhdType = [Microsoft.VHD.PowerShell.VHDType]$vhd.VHDType

function Get-TarPath {
  if (Get-Command "tar" -ErrorAction SilentlyContinue) {
    return "tar"
  }
  elseif (Test-Path "$env:SystemRoot\system32\tar.exe") {
    return "$env:SystemRoot\system32\tar.exe"
  }
  else {
    return ""
  }
}

function Get-7ZipPath {
  if (Get-Command "7z" -ErrorAction SilentlyContinue) {
    return "7z"
  }
  elseif (Test-Path "$env:ProgramFiles\7-Zip\7z.exe") {
    return "$env:ProgramFiles\7-Zip\7z.exe"
  }
  elseif (Test-Path "${env:ProgramFiles(x86)}\7-Zip\7z.exe") {
    return "${env:ProgramFiles(x86)}\7-Zip\7z.exe"
  }
  else {
    return ""
  }
}

function Expand-Downloads {
  param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]
    [Alias('Folder')]
    $FolderPath
  )
  process {
    Push-Location $FolderPath

    Get-Item *.zip | ForEach-Object {
      $tempPath = join-path $FolderPath "temp"

      $7zPath = Get-7ZipPath
      if ($7zPath) {
        $command = """$7zPath"" x ""$($_.FullName)"" -o""$tempPath""" 
        & cmd.exe /C $command
      }
      else {
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        if (!(Test-Path $tempPath)) {
          New-Item -ItemType Directory -Force -Path $tempPath
        }
        [System.IO.Compression.ZipFile]::ExtractToDirectory($_.FullName, $tempPath)
      }

      $vhdPath = Get-ChildItem $tempPath *"Virtual Hard Disks"* -Recurse -Directory

      if ($vhdPath -and (Test-Path $vhdPath.FullName)) {
        Move-Item "$($vhdPath.FullName)\*.*" $FolderPath
      }
      else {
        Move-Item "$tempPath\*.*" $FolderPath
      }

      Remove-Item $tempPath -Force -Recurse
      Remove-Item $_.FullName -Force
    }

    Get-Item *.7z | ForEach-Object {
      $7zPath = Get-7ZipPath
      if (-not $7zPath) {
        throw "7z.exe needed"
      }
      $tempPath = join-path $FolderPath "temp"
      $command = """$7zPath"" x ""$($_.FullName)"" -o""$tempPath""" 
      & cmd.exe /C $command

      $vhdPath = Get-ChildItem $tempPath *"Virtual Hard Disks"* -Recurse -Directory

      if ($vhdPath -and (Test-Path $vhdPath.FullName)) {
        Move-Item "$($vhdPath.FullName)\*.*" $FolderPath
      }
      else {
        Move-Item "$tempPath\*.*" $FolderPath
      }

      Remove-Item $tempPath -Force -Recurse
      Remove-Item $_.FullName -Force
    }

    Get-Item *.box | ForEach-Object {
      $tarPath = Get-TarPath
      if (-not $tarPath) {
        throw "tar.exe needed"
      }
      $tempPath = join-path $FolderPath "temp"

      if (!(Test-Path $tempPath)) {
        New-Item -ItemType Directory -Force -Path $tempPath
      }
      $command = """$tarPath"" -C ""$tempPath"" -x -f ""$($_.FullName)"""
      & cmd.exe /C $command

      $vhdPath = Get-ChildItem $tempPath *"Virtual Hard Disks"* -Recurse -Directory

      if ($vhdPath -and (Test-Path $vhdPath.FullName)) {
        Move-Item "$($vhdPath.FullName)\*.*" $FolderPath
      }
      else {
        Move-Item "$tempPath\*.*" $FolderPath
      }

      Remove-Item $tempPath -Force -Recurse
      Remove-Item $_.FullName -Force
    }

    Pop-Location
  }
}

function Get-FileFromUri {
  param(
    [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline = $true, ValueFromPipelineByPropertyName = $true)]
    [string]
    [Alias('Uri')]
    $Url,
    [Parameter(Mandatory = $false, Position = 1)]
    [string]
    [Alias('Folder')]
    $FolderPath
  )
  process {
    $req = [System.Net.HttpWebRequest]::Create($Url)
    $req.Method = "HEAD"
    $response = $req.GetResponse()
    $fUri = $response.ResponseUri
    $filename = [System.IO.Path]::GetFileName($fUri.LocalPath)
    $response.Close()

    $origExt = [System.IO.Path]::GetExtension($Url)
    $newExt = [System.IO.Path]::GetExtension($filename)
    if ($newExt -ne $origExt) {
      $filename += $origExt
    }

    $destination = (Get-Item -Path ".\" -Verbose).FullName
    if ($FolderPath) { $destination = $FolderPath }
    if ($destination.EndsWith('\')) {
      $destination += $filename
    }
    else {
      $destination += '\' + $filename
    }
    $webclient = New-Object System.Net.WebClient
    $webclient.DownloadFile($fUri.AbsoluteUri, $destination)
  }
}

function Test-Uri {
  param(
    [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline = $true, ValueFromPipelineByPropertyName = $true)]
    [string]
    [Alias('Uri')]
    $Url
  )
  process {
    $testUri = $Url -as [System.URI]
    $null -ne $testUri.AbsoluteURI -and $testUri.Scheme -match '[http|https]' -and ($testUri.ToString().ToLower().StartsWith("http://") -or $testUri.ToString().ToLower().StartsWith("https://"))
  }
}

if ($vhd -and !(Test-Path $vhd.Path)) {
  $pathDirectory = [System.IO.Path]::GetDirectoryName($vhd.Path)
  $pathFilename = [System.IO.Path]::GetFileName($vhd.Path)

  if (!(Test-Path $pathDirectory)) {
    New-Item -ItemType Directory -Force -Path $pathDirectory
  }

  if ($sourceVm) {
    Export-VM -Name $sourceVm -Path $pathDirectory
    $targetName = (split-path $vhd.Path -Leaf)
    $targetName = $targetName.Substring(0, $targetName.LastIndexOf('.')).split('\')[-1]
    Get-ChildItem -Path "$pathDirectory\$sourceVm\Virtual Hard Disks" | Where-Object { $_.BaseName.StartsWith($sourceVm) } | ForEach-Object {
      $targetNamePath = "$($pathDirectory)\$($_.Name.Replace($sourceVm, $targetName))"
      Move-Item $_.FullName $targetNamePath
    }

    Remove-Item "$pathDirectory\$sourceVm" -Force -Recurse
    Get-VHD -path $vhd.Path
  }
  elseif ($source) {
    Push-Location $pathDirectory
        
    if (Test-Uri -Url $source) {
      Get-FileFromUri -Url $source -FolderPath $pathDirectory
    }
    else {
      Copy-Item $source "$pathDirectory\$pathFilename" -Force
    }

    Expand-Downloads -FolderPath $pathDirectory

    Pop-Location
  }
  else {
    $NewVHDArgs = @{}
    $NewVHDArgs.Path = $vhd.Path

    if ($sourceDisk) {
      $NewVHDArgs.SourceDisk = $sourceDisk
    }
    elseif ($vhdType -eq [Microsoft.VHD.PowerShell.VHDType]::Differencing) {
      $NewVHDArgs.Differencing = $true
      $NewVHDArgs.ParentPath = $vhd.ParentPath
    }
    else {
      if ($vhdType -eq [Microsoft.VHD.PowerShell.VHDType]::Dynamic) {
        $NewVHDArgs.Dynamic = $true
      }
      elseif ($vhdType -eq [Microsoft.VHD.PowerShell.VHDType]::Fixed) {
        $NewVHDArgs.Fixed = $true
      }

      if ($vhd.BlockSize -gt 0) {
        $NewVHDArgs.BlockSizeBytes = $vhd.BlockSize
      }

      if ($vhd.PhysicalSectorSize -gt 0) {
        $NewVHDArgs.PhysicalSectorSizeBytes = $vhd.PhysicalSectorSize
      }

      if ($vhd.LogicalSectorSize -gt 0) {
        $NewVHDArgs.LogicalSectorSizeBytes = $vhd.LogicalSectorSize
      }
      else {
        $NewVHDArgs.LogicalSectorSizeBytes = 512 #this is the default size
      }

      if ($vhd.Size -gt 0) {
        $NewVHDArgs.SizeBytes = [math]::ceiling($vhd.Size / $NewVHDArgs.LogicalSectorSizeBytes) * $NewVHDArgs.LogicalSectorSizeBytes
      }
      else {
        throw "VHD Size must be specified for - $($vhd.Path)"
      }
    }

    New-VHD @NewVHDArgs
  }
}