package hypervwinrmimpl

import (
	"context"
	"encoding/json"
	"text/template"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
)

type existsVHDArgs struct {
	Path string
}

var existsVHDTemplate = template.Must(template.New("ExistsVHD").Parse(`
$ErrorActionPreference = 'Stop'
$path='{{.Path}}'

if (Test-Path $path) {
	$exists = ConvertTo-Json -InputObject @{Exists=$true}
	$exists
} else {
	$exists = ConvertTo-Json -InputObject @{Exists=$false}
	$exists
}
`))

func (c *hypervClientImpl) VHDExists(ctx context.Context, path string) (result hyperv.VHDExists, err error) {
	err = c.winrmClient.RunScriptWithResult(ctx, existsVHDTemplate, existsVHDArgs{
		Path: path,
	}, &result)

	return result, err
}

type createOrUpdateVHDArgs struct {
	Source     string
	SourceVm   string
	SourceDisk int
	VHDJson    string
}

var createOrUpdateVHDTemplate = template.Must(template.New("CreateOrUpdateVHD").Parse(`
$ErrorActionPreference = 'Stop'

Import-Module Hyper-V
$source='{{.Source}}'
$sourceVm='{{.SourceVm}}'
$sourceDisk={{.SourceDisk}}
$vhd = '{{.VHDJson}}' | ConvertFrom-Json
$vhdType = [Microsoft.VHD.PowerShell.VHDType]$vhd.VHDType

function Get-TarPath {
	if (Get-Command "tar" -ErrorAction SilentlyContinue) {
		return "tar"
	} elseif (Test-Path "$env:SystemRoot\system32\tar.exe") {
		return "$env:SystemRoot\system32\tar.exe"
	} else {
		return ""
	}
}

function Get-7ZipPath {
	if (Get-Command "7z" -ErrorAction SilentlyContinue) {
		return "7z"
	} elseif (Test-Path "$env:ProgramFiles\7-Zip\7z.exe") {
		return "$env:ProgramFiles\7-Zip\7z.exe"
	} elseif (Test-Path "${env:ProgramFiles(x86)}\7-Zip\7z.exe") {
		return "${env:ProgramFiles(x86)}\7-Zip\7z.exe"
	} else {
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

        get-item *.zip | % {
			$tempPath = join-path $FolderPath "temp"

			$7zPath = Get-7ZipPath
			if ($7zPath) {
				$command = """$7zPath"" x ""$($_.FullName)"" -o""$tempPath""" 
				& cmd.exe /C $command
			} else {
				Add-Type -AssemblyName System.IO.Compression.FileSystem
    			if (!(Test-Path $tempPath)) {
        			New-Item -ItemType Directory -Force -Path $tempPath
    			}
            	[System.IO.Compression.ZipFile]::ExtractToDirectory($_.FullName, $tempPath)
			}

			$vhdPath = Get-ChildItem $tempPath *"Virtual Hard Disks"* -Recurse -Directory

            if ($vhdPath -and (Test-Path $vhdPath.FullName)) {
        		Move-Item "$($vhdPath.FullName)\*.*" $FolderPath
			} else {
				Move-Item "$tempPath\*.*" $FolderPath
			}

			Remove-Item $tempPath -Force -Recurse
			Remove-Item $_.FullName -Force
        }

        get-item *.7z | % {
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
			} else {
				Move-Item "$tempPath\*.*" $FolderPath
			}

			Remove-Item $tempPath -Force -Recurse
			Remove-Item $_.FullName -Force
        }

        get-item *.box | % {
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
			} else {
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
        $targetName = $targetName.Substring(0,$targetName.LastIndexOf('.')).split('\')[-1]
        Get-ChildItem -Path "$pathDirectory\$sourceVm\Virtual Hard Disks" |?{$_.BaseName.StartsWith($sourceVm)} | %{
            $targetNamePath = "$($pathDirectory)\$($_.Name.Replace($sourceVm, $targetName))"
            Move-Item $_.FullName $targetNamePath
        }

        Remove-Item "$pathDirectory\$sourceVm" -Force -Recurse
        Get-VHD -path $vhd.Path
    } elseif ($source) {
        Push-Location $pathDirectory
        
        if (Test-Uri -Url $source) {
            Get-FileFromUri -Url $source -FolderPath $pathDirectory
        }
        else {
            Copy-Item $source "$pathDirectory\$pathFilename" -Force
        }

        Expand-Downloads -FolderPath $pathDirectory

        Pop-Location
    } else {
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
            } else {
                $NewVHDArgs.LogicalSectorSizeBytes = 512 #this is the default size
            }

			if ($vhd.Size -gt 0) {
					$NewVHDArgs.SizeBytes = [math]::ceiling($vhd.Size/$NewVHDArgs.LogicalSectorSizeBytes)*$NewVHDArgs.LogicalSectorSizeBytes
			} else {
				throw "VHD Size must be specified for - $($vhd.Path)"
			}
		}

        New-VHD @NewVHDArgs
    }
}
`))

func (c *hypervClientImpl) CreateOrUpdateVHD(ctx context.Context, path string, source string, sourceVm string, sourceDisk int, vhdType hyperv.VHDType, parentPath string, size uint64, blockSize uint32, logicalSectorSize uint32, physicalSectorSize uint32) (err error) {
	vhdJson, err := json.Marshal(hyperv.VHD{
		Path:               path,
		VHDType:            vhdType,
		ParentPath:         parentPath,
		Size:               size,
		BlockSize:          blockSize,
		LogicalSectorSize:  logicalSectorSize,
		PhysicalSectorSize: physicalSectorSize,
	})

	if err != nil {
		return err
	}

	err = c.winrmClient.RunFireAndForgetScript(ctx, createOrUpdateVHDTemplate, createOrUpdateVHDArgs{
		Source:     source,
		SourceVm:   sourceVm,
		SourceDisk: sourceDisk,
		VHDJson:    string(vhdJson),
	})

	return err
}

type resizeVHDArgs struct {
	Path string
	Size uint64
}

var resizeVHDTemplate = template.Must(template.New("ResizeVHD").Parse(`
$ErrorActionPreference = 'Stop'
$vhd = Get-VHD -Path '{{.Path}}'
if ($vhd.Size -ne {{.Size}}){
	Resize-VHD -Path '{{.Path}}' -SizeBytes {{.Size}}
}
`))

func (c *hypervClientImpl) ResizeVHD(ctx context.Context, path string, size uint64) (err error) {
	err = c.winrmClient.RunFireAndForgetScript(ctx, resizeVHDTemplate, resizeVHDArgs{
		Path: path,
		Size: size,
	})

	return err
}

type getVHDArgs struct {
	Path string
}

var getVHDTemplate = template.Must(template.New("GetVHD").Parse(`
$ErrorActionPreference = 'Stop'
$path='{{.Path}}'

$vhdObject = $null
if (Test-Path $path) {
	$vhdObject = Get-VHD -path $path | %{ @{
		Path=$_.Path;
		BlockSize=$_.BlockSize;
		LogicalSectorSize=$_.LogicalSectorSize;
		PhysicalSectorSize=$_.PhysicalSectorSize;
		ParentPath=$_.ParentPath;
		FileSize=$_.FileSize;
		Size=$_.Size;
		MinimumSize=$_.MinimumSize;
		Attached=$_.Attached;
		DiskNumber=$_.DiskNumber;
		Number=$_.Number;
		FragmentationPercentage=$_.FragmentationPercentage;
		Alignment=$_.Alignment;
		DiskIdentifier=$_.DiskIdentifier;
		VHDType=$_.VHDType;
		VHDFormat=$_.VHDFormat;
	}}
}

if ($vhdObject){
	$vhd = ConvertTo-Json -InputObject $vhdObject
	$vhd
} else {
	"{}"
}
`))

func (c *hypervClientImpl) GetVHD(ctx context.Context, path string) (result hyperv.VHD, err error) {
	err = c.winrmClient.RunScriptWithResult(ctx, getVHDTemplate, getVHDArgs{
		Path: path,
	}, &result)

	return result, err
}

type deleteVHDArgs struct {
	Path string
}

var deleteVHDTemplate = template.Must(template.New("DeleteVHD").Parse(`
$ErrorActionPreference = 'Stop'

$targetDirectory = (split-path '{{.Path}}' -Parent)
$targetName = (split-path '{{.Path}}' -Leaf)
$targetName = $targetName.Substring(0,$targetName.LastIndexOf('.')).split('\')[-1]

Get-ChildItem -Path $targetDirectory |?{$_.BaseName.StartsWith($targetName)} | %{
	Remove-Item $_.FullName -Force
}
`))

func (c *hypervClientImpl) DeleteVHD(ctx context.Context, path string) (err error) {
	err = c.winrmClient.RunFireAndForgetScript(ctx, deleteVHDTemplate, deleteVHDArgs{
		Path: path,
	})

	return err
}
