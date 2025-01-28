package hypervwinrmimpl

import (
	"context"
	"encoding/json"
	"text/template"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
)

type createVMHardDiskDriveArgs struct {
	VMHardDiskDriveJson string
}

var createVMHardDiskDriveTemplate = template.Must(template.New("CreateVMHardDiskDrive").Parse(`
$ErrorActionPreference = 'Stop'
Import-Module Hyper-V
$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json

$NewVMHardDiskDriveArgs = @{
	VMName=$vmHardDiskDrive.VMName
	ControllerType=$vmHardDiskDrive.ControllerType
	ControllerNumber=$vmHardDiskDrive.ControllerNumber
	ControllerLocation=$vmHardDiskDrive.ControllerLocation
	Path=$vmHardDiskDrive.Path
	ResourcePoolName=$vmHardDiskDrive.ResourcePoolName
	SupportPersistentReservations=$vmHardDiskDrive.SupportPersistentReservations
	MaximumIops=$_.MaximumIops;
	MinimumIops=$_.MinimumIops;
	QosPolicyId=$_.QosPolicyId;
	OverrideCacheAttributes=$vmHardDiskDrive.OverrideCacheAttributes
	AllowUnverifiedPaths=$true
}

if ($vmHardDiskDrive.DiskNumber -lt 4294967295){
	$NewVMHardDiskDriveArgs.DiskNumber=$vmHardDiskDrive.DiskNumber
}

Add-VMHardDiskDrive @NewVMHardDiskDriveArgs
`))

func (c *hypervClientImpl) CreateVMHardDiskDrive(
	ctx context.Context,
	vmName string,
	controllerType hyperv.ControllerType,
	controllerNumber int32,
	controllerLocation int32,
	path string,
	diskNumber uint32,
	resourcePoolName string,
	supportPersistentReservations bool,
	maximumIops uint64,
	minimumIops uint64,
	qosPolicyId string,
	overrideCacheAttributes hyperv.CacheAttributes,

) (err error) {
	vmHardDiskDriveJson, err := json.Marshal(hyperv.VMHardDiskDrive{
		VMName:                        vmName,
		ControllerType:                controllerType,
		ControllerNumber:              controllerNumber,
		ControllerLocation:            controllerLocation,
		Path:                          path,
		DiskNumber:                    diskNumber,
		ResourcePoolName:              resourcePoolName,
		SupportPersistentReservations: supportPersistentReservations,
		MaximumIops:                   maximumIops,
		MinimumIops:                   minimumIops,
		QosPolicyId:                   qosPolicyId,
		OverrideCacheAttributes:       overrideCacheAttributes,
	})

	if err != nil {
		return err
	}

	err = c.winrmClient.RunFireAndForgetScript(ctx, createVMHardDiskDriveTemplate, createVMHardDiskDriveArgs{
		VMHardDiskDriveJson: string(vmHardDiskDriveJson),
	})

	return err
}

type getVMHardDiskDrivesArgs struct {
	VMName string
}

var getVMHardDiskDrivesTemplate = template.Must(template.New("GetVMHardDiskDrives").Parse(`
$ErrorActionPreference = 'Stop'
$vmHardDiskDrivesObject = @(Get-VM -Name '{{.VMName}}*' | ?{$_.Name -eq '{{.VMName}}' } | Get-VMHardDiskDrive | %{ @{
	ControllerType=$_.ControllerType;
	ControllerNumber=$_.ControllerNumber;
	ControllerLocation=$_.ControllerLocation;
	Path=$_.Path;
	DiskNumber=if ($_.DiskNumber -eq $null) { 4294967295 } else { $_.DiskNumber };
	ResourcePoolName=$_.PoolName;
	SupportPersistentReservations=$_.SupportPersistentReservations;
	MaximumIops=$_.MaximumIops;
	MinimumIops=$_.MinimumIops;
	QosPolicyId=$_.QosPolicyId;	
	OverrideCacheAttributes=$_.WriteHardeningMethod;
}})

if ($vmHardDiskDrivesObject) {
	$vmHardDiskDrives = ConvertTo-Json -InputObject $vmHardDiskDrivesObject
	$vmHardDiskDrives
} else {
	"[]"
}
`))

func (c *hypervClientImpl) GetVMHardDiskDrives(ctx context.Context, vmName string) (result []hyperv.VMHardDiskDrive, err error) {
	result = make([]hyperv.VMHardDiskDrive, 0)

	err = c.winrmClient.RunScriptWithResult(ctx, getVMHardDiskDrivesTemplate, getVMHardDiskDrivesArgs{
		VMName: vmName,
	}, &result)

	return result, err
}

type updateVMHardDiskDriveArgs struct {
	VMName              string
	ControllerNumber    int32
	ControllerLocation  int32
	VMHardDiskDriveJson string
}

var updateVMHardDiskDriveTemplate = template.Must(template.New("UpdateVMHardDiskDrive").Parse(`
$ErrorActionPreference = 'Stop'
Import-Module Hyper-V
$vmHardDiskDrive = '{{.VMHardDiskDriveJson}}' | ConvertFrom-Json

$vmHardDiskDrivesObject = @(Get-VM -Name '{{.VMName}}*' | ?{$_.Name -eq '{{.VMName}}' } | Get-VMHardDiskDrive -ControllerLocation {{.ControllerLocation}} -ControllerNumber {{.ControllerNumber}} )

if (!$vmHardDiskDrivesObject){
	throw "VM hard disk drive does not exist - {{.ControllerLocation}} {{.ControllerNumber}}"
}

$SetVMHardDiskDriveArgs = @{}
$SetVMHardDiskDriveArgs.VMName=$vmHardDiskDrivesObject.VMName
$SetVMHardDiskDriveArgs.ControllerType=$vmHardDiskDrivesObject.ControllerType
$SetVMHardDiskDriveArgs.ControllerLocation=$vmHardDiskDrivesObject.ControllerLocation
$SetVMHardDiskDriveArgs.ControllerNumber=$vmHardDiskDrivesObject.ControllerNumber
$SetVMHardDiskDriveArgs.ToControllerLocation=$vmHardDiskDrive.ControllerLocation
$SetVMHardDiskDriveArgs.ToControllerNumber=$vmHardDiskDrive.ControllerNumber
$SetVMHardDiskDriveArgs.Path=$vmHardDiskDrive.Path
if ($vmHardDiskDrive.DiskNumber -lt 4294967295){
	$SetVMHardDiskDriveArgs.DiskNumber=$vmHardDiskDrive.DiskNumber
}
if ($vmHardDiskDrivesObject.ResourcePoolName -ne $vmHardDiskDrive.ResourcePoolName) {
	if ($vmHardDiskDrive.ResourcePoolName) {
		$SetVMHardDiskDriveArgs.ResourcePoolName=$vmHardDiskDrive.ResourcePoolName
	} else {
		throw "Unable to remove resource pool $($vmHardDiskDrive.ResourcePoolName) from hard disk drive $(ConvertTo-Json -InputObject $vmHardDiskDrivesObject)"
	}
}
$SetVMHardDiskDriveArgs.SupportPersistentReservations=$vmHardDiskDrive.SupportPersistentReservations
$SetVMHardDiskDriveArgs.MaximumIops=$vmHardDiskDrive.MaximumIops
$SetVMHardDiskDriveArgs.MinimumIops=$vmHardDiskDrive.MinimumIops
$SetVMHardDiskDriveArgs.QosPolicyId=$vmHardDiskDrive.QosPolicyId
$SetVMHardDiskDriveArgs.OverrideCacheAttributes=$vmHardDiskDrive.OverrideCacheAttributes	
$SetVMHardDiskDriveArgs.AllowUnverifiedPaths=$true

Set-VMHardDiskDrive @SetVMHardDiskDriveArgs

`))

func (c *hypervClientImpl) UpdateVMHardDiskDrive(
	ctx context.Context,
	vmName string,
	controllerNumber int32,
	controllerLocation int32,
	controllerType hyperv.ControllerType,
	toControllerNumber int32,
	toControllerLocation int32,
	path string,
	diskNumber uint32,
	resourcePoolName string,
	supportPersistentReservations bool,
	maximumIops uint64,
	minimumIops uint64,
	qosPolicyId string,
	overrideCacheAttributes hyperv.CacheAttributes,
) (err error) {
	vmHardDiskDriveJson, err := json.Marshal(hyperv.VMHardDiskDrive{
		VMName:                        vmName,
		ControllerType:                controllerType,
		ControllerNumber:              toControllerNumber,
		ControllerLocation:            toControllerLocation,
		Path:                          path,
		DiskNumber:                    diskNumber,
		ResourcePoolName:              resourcePoolName,
		SupportPersistentReservations: supportPersistentReservations,
		MaximumIops:                   maximumIops,
		MinimumIops:                   minimumIops,
		QosPolicyId:                   qosPolicyId,
		OverrideCacheAttributes:       overrideCacheAttributes,
	})

	if err != nil {
		return err
	}

	err = c.winrmClient.RunFireAndForgetScript(ctx, updateVMHardDiskDriveTemplate, updateVMHardDiskDriveArgs{
		VMName:              vmName,
		ControllerNumber:    controllerNumber,
		ControllerLocation:  controllerLocation,
		VMHardDiskDriveJson: string(vmHardDiskDriveJson),
	})

	return err
}

type deleteVMHardDiskDriveArgs struct {
	VMName             string
	ControllerNumber   int32
	ControllerLocation int32
}

var deleteVMHardDiskDriveTemplate = template.Must(template.New("DeleteVMHardDiskDrive").Parse(`
$ErrorActionPreference = 'Stop'

@(Get-VMHardDiskDrive -VMName '{{.VMName}}' -ControllerNumber {{.ControllerNumber}} -ControllerLocation {{.ControllerLocation}}) | Remove-VMHardDiskDrive
`))

func (c *hypervClientImpl) DeleteVMHardDiskDrive(ctx context.Context, vmname string, controllerNumber int32, controllerLocation int32) (err error) {
	err = c.winrmClient.RunFireAndForgetScript(ctx, deleteVMHardDiskDriveTemplate, deleteVMHardDiskDriveArgs{
		VMName:             vmname,
		ControllerNumber:   controllerNumber,
		ControllerLocation: controllerLocation,
	})

	return err
}

func (c *hypervClientImpl) CreateOrUpdateVMHardDiskDrives(ctx context.Context, vmName string, hardDiskDrives []hyperv.VMHardDiskDrive) (err error) {
	currentHardDiskDrives, err := c.GetVMHardDiskDrives(ctx, vmName)
	if err != nil {
		return err
	}

	currentHardDiskDrivesLength := len(currentHardDiskDrives)
	desiredHardDiskDrivesLength := len(hardDiskDrives)

	for i := currentHardDiskDrivesLength - 1; i > desiredHardDiskDrivesLength-1; i-- {
		currentHardDiskDrive := currentHardDiskDrives[i]
		err = c.DeleteVMHardDiskDrive(ctx, vmName, currentHardDiskDrive.ControllerNumber, currentHardDiskDrive.ControllerLocation)
		if err != nil {
			return err
		}
	}

	if currentHardDiskDrivesLength > desiredHardDiskDrivesLength {
		currentHardDiskDrivesLength = desiredHardDiskDrivesLength
	}

	for i := 0; i <= currentHardDiskDrivesLength-1; i++ {
		currentHardDiskDrive := currentHardDiskDrives[i]
		hardDiskDrive := hardDiskDrives[i]

		err = c.UpdateVMHardDiskDrive(
			ctx,
			vmName,
			currentHardDiskDrive.ControllerNumber,
			currentHardDiskDrive.ControllerLocation,
			hardDiskDrive.ControllerType,
			hardDiskDrive.ControllerNumber,
			hardDiskDrive.ControllerLocation,
			hardDiskDrive.Path,
			hardDiskDrive.DiskNumber,
			hardDiskDrive.ResourcePoolName,
			hardDiskDrive.SupportPersistentReservations,
			hardDiskDrive.MaximumIops,
			hardDiskDrive.MinimumIops,
			hardDiskDrive.QosPolicyId,
			hardDiskDrive.OverrideCacheAttributes,
		)
		if err != nil {
			return err
		}
	}

	for i := currentHardDiskDrivesLength - 1 + 1; i <= desiredHardDiskDrivesLength-1; i++ {
		hardDiskDrive := hardDiskDrives[i]
		err = c.CreateVMHardDiskDrive(
			ctx,
			vmName,
			hardDiskDrive.ControllerType,
			hardDiskDrive.ControllerNumber,
			hardDiskDrive.ControllerLocation,
			hardDiskDrive.Path,
			hardDiskDrive.DiskNumber,
			hardDiskDrive.ResourcePoolName,
			hardDiskDrive.SupportPersistentReservations,
			hardDiskDrive.MaximumIops,
			hardDiskDrive.MinimumIops,
			hardDiskDrive.QosPolicyId,
			hardDiskDrive.OverrideCacheAttributes,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
