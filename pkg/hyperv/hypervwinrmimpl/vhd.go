package hypervwinrmimpl

import (
	"context"
	_ "embed"
	"encoding/json"
	"text/template"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
)

var (
	//go:embed scripts/Exist-VHD.ps1
	existVHDFile string

	//go:embed scripts/Patch-VHD.ps1
	patchVHDFile string

	//go:embed scripts/Resize-VHD.ps1
	resizeVHDFile string

	//go:embed scripts/Get-VHD.ps1
	getVHDFile string

	//go:embed scripts/Delete-VHD.ps1
	deleteVHDFile string
)

var (
	existVHDTemplate  = template.Must(template.New("ExistVHD").Parse(existVHDFile))
	patchVHDTemplate  = template.Must(template.New("PatchVHD").Parse(patchVHDFile))
	resizeVHDTemplate = template.Must(template.New("ResizeVHD").Parse(resizeVHDFile))
	getVHDTemplate    = template.Must(template.New("GetVHD").Parse(getVHDFile))
	deleteVHDTemplate = template.Must(template.New("DeleteVHD").Parse(deleteVHDFile))
)

type existsVHDArgs struct {
	Path string
}

type createOrUpdateVHDArgs struct {
	Source     string
	SourceVm   string
	SourceDisk int
	VHDJson    string
}

type resizeVHDArgs struct {
	Path string
	Size uint64
}

type getVHDArgs struct {
	Path string
}

type deleteVHDArgs struct {
	Path string
}

func (c *hypervClientImpl) VHDExists(ctx context.Context, path string) (result hyperv.VHDExists, err error) {
	err = c.winrmClient.RunScriptWithResult(ctx, existVHDTemplate, existsVHDArgs{
		Path: path,
	}, &result)

	return result, err
}

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

	err = c.winrmClient.RunFireAndForgetScript(ctx, patchVHDTemplate, createOrUpdateVHDArgs{
		Source:     source,
		SourceVm:   sourceVm,
		SourceDisk: sourceDisk,
		VHDJson:    string(vhdJson),
	})

	return err
}

func (c *hypervClientImpl) ResizeVHD(ctx context.Context, path string, size uint64) (err error) {
	err = c.winrmClient.RunFireAndForgetScript(ctx, resizeVHDTemplate, resizeVHDArgs{
		Path: path,
		Size: size,
	})

	return err
}

func (c *hypervClientImpl) GetVHD(ctx context.Context, path string) (result hyperv.VHD, err error) {
	err = c.winrmClient.RunScriptWithResult(ctx, getVHDTemplate, getVHDArgs{
		Path: path,
	}, &result)

	return result, err
}

func (c *hypervClientImpl) DeleteVHD(ctx context.Context, path string) (err error) {
	err = c.winrmClient.RunFireAndForgetScript(ctx, deleteVHDTemplate, deleteVHDArgs{
		Path: path,
	})

	return err
}
