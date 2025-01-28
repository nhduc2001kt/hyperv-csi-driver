package hypervwinrmimpl

import (
	"context"
	_ "embed"
	"encoding/json"
	"text/template"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
)

var (
	//go:embed scripts/Attach-VMHardDiskDrive.ps1
	attachVMHardDiskDriveFile string
	//go:embed scripts/Create-VMHardDiskDrive.ps1
	createVMHardDiskDriveFile string
	//go:embed scripts/Get-VMHardDiskDrives.ps1
	getVMHardDiskDrivesFile string
	//go:embed scripts/Get-VMHardDiskDrivesByID.ps1
	getVMHardDiskDrivesByIDFile string
	//go:embed scripts/Update-VMHardDiskDrive.ps1
	updateVMHardDiskDriveFile string
	//go:embed scripts/Delete-VMHardDiskDrive.ps1
	deleteVMHardDiskDriveFile string
)

var (
	attachVMHardDiskDriveTemplate   = template.Must(template.New("AttachVMHardDiskDrive").Parse(attachVMHardDiskDriveFile))
	createVMHardDiskDriveTemplate   = template.Must(template.New("CreateVMHardDiskDrive").Parse(createVMHardDiskDriveFile))
	getVMHardDiskDrivesTemplate     = template.Must(template.New("GetVMHardDiskDrives").Parse(getVMHardDiskDrivesFile))
	getVMHardDiskDrivesByIDTemplate = template.Must(template.New("GetVMHardDiskDrivesByID").Parse(getVMHardDiskDrivesByIDFile))
	updateVMHardDiskDriveTemplate   = template.Must(template.New("UpdateVMHardDiskDrive").Parse(updateVMHardDiskDriveFile))
	deleteVMHardDiskDriveTemplate   = template.Must(template.New("DeleteVMHardDiskDrive").Parse(deleteVMHardDiskDriveFile))
)

type attachVMHardDiskDriveArgs struct {
	VMHardDiskDriveJson string
}

type createVMHardDiskDriveArgs struct {
	VMHardDiskDriveJson string
}

type getVMHardDiskDrivesArgs struct {
	VMName string
}

type getVMHardDiskDrivesByIDArgs struct {
	ID string
}

type updateVMHardDiskDriveArgs struct {
	VMName              string
	ControllerNumber    int32
	ControllerLocation  int32
	VMHardDiskDriveJson string
}

type deleteVMHardDiskDriveArgs struct {
	VMName             string
	ControllerNumber   int32
	ControllerLocation int32
}

func (c *hypervClientImpl) AttachVMHardDiskDrive(
	ctx context.Context,
	vmName string,
	controllerType hyperv.ControllerType,
	path string,
) (err error) {
	vmHardDiskDriveJson, err := json.Marshal(hyperv.VMHardDiskDrive{
		VMName:         vmName,
		ControllerType: controllerType,
		Path:           path,
	})

	if err != nil {
		return err
	}

	err = c.winrmClient.RunFireAndForgetScript(ctx, attachVMHardDiskDriveTemplate, attachVMHardDiskDriveArgs{
		VMHardDiskDriveJson: string(vmHardDiskDriveJson),
	})

	return err
}

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

func (c *hypervClientImpl) GetVMHardDiskDrives(ctx context.Context, vmName string) (result []hyperv.VMHardDiskDrive, err error) {
	result = make([]hyperv.VMHardDiskDrive, 0)

	err = c.winrmClient.RunScriptWithResult(ctx, getVMHardDiskDrivesTemplate, getVMHardDiskDrivesArgs{
		VMName: vmName,
	}, &result)

	return result, err
}

func (c *hypervClientImpl) GetVMHardDiskDrivesByID(ctx context.Context, vmID string) (result []hyperv.VMHardDiskDrive, err error) {
	result = make([]hyperv.VMHardDiskDrive, 0)

	err = c.winrmClient.RunScriptWithResult(ctx, getVMHardDiskDrivesByIDTemplate, getVMHardDiskDrivesByIDArgs{
		ID: vmID,
	}, &result)

	return result, err
}

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
