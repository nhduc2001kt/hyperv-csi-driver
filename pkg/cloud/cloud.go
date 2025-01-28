package cloud

import (
	"context"
	"fmt"

	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv/hypervwinrmimpl"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"k8s.io/klog/v2"
)

// Defaults.
const (
	// DefaultVolumeSize represents the default volume size.
	DefaultVolumeSize int64 = 100 * util.GiB

	// DefaultVHDBasePath represents the default VHD base path
	DefaultVHDBasePath = "C:\\ProgramData\\Microsoft\\Windows\\Virtual Hard Disks"
)

type CreateHyperVVHDInput struct {
	Name               string
	Source             string
	SourceVm           string
	SourceDisk         int
	Type               hyperv.VHDType
	Format             hyperv.VHDFormat
	ParentPath         string
	Size               uint64
	BlockSize          uint32
	LogicalSectorSize  uint32
	PhysicalSectorSize uint32
}

type CreateHyperVVHDOutput struct {
	Path string
}

type DeleteHyperVVHDInput struct {
	Path string
}

type DeleteHyperVVHDOutput struct{}

type AttachHyperVVHDInput struct {
	VmID    string
	VHDPath string
}

type AttachHyperVVHDOutput struct {
	ControllerNumber   int32
	ControllerLocation int32
}

type DetachHyperVVHDInput struct {
	VmID    string
	VHDPath string
}

type DetachHyperVVHDOutput struct {
	ControllerNumber   int32
	ControllerLocation int32
}

type Cloud interface {
	CreateHyperVVHD(context.Context, *CreateHyperVVHDInput) (*CreateHyperVVHDOutput, error)
	DeleteHyperVVHD(context.Context, *DeleteHyperVVHDInput) (*DeleteHyperVVHDOutput, error)
	AttachHyperVVHD(context.Context, *AttachHyperVVHDInput) (*AttachHyperVVHDOutput, error)
	DetachHyperVVHD(context.Context, *DetachHyperVVHDInput) (*DetachHyperVVHDOutput, error)
}

type CloudConfig interface {
}

// NewCloud returns a new instance of Docker client
// It panics if session is invalid.
func NewCloud(opts *options.Options) (Cloud, error) {
	hypervClient, err := hypervwinrmimpl.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return &cloud{
		hypervClient: hypervClient,
		vhdBasePath:  DefaultVHDBasePath,
	}, nil
}

type cloud struct {
	hypervClient hyperv.HyperVClient
	vhdBasePath  string
}

func (c *cloud) CreateHyperVVHD(ctx context.Context, i *CreateHyperVVHDInput) (*CreateHyperVVHDOutput, error) {
	klog.V(4).InfoS("CreateHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient
	vhdFile := fmt.Sprintf("%s%s", i.Name, hyperv.VHDFormatExtension[i.Format])
	vhdPath := util.JoinWinPath(c.vhdBasePath, vhdFile)
	err := client.CreateOrUpdateVHD(
		ctx,
		vhdPath,
		i.Source,
		i.SourceVm,
		i.SourceDisk,
		i.Type,
		i.ParentPath,
		i.Size,
		i.BlockSize,
		i.LogicalSectorSize,
		i.PhysicalSectorSize,
	)
	if err != nil {
		return nil, err
	}

	return &CreateHyperVVHDOutput{
		Path: vhdPath,
	}, nil
}

func (c *cloud) DeleteHyperVVHD(ctx context.Context, i *DeleteHyperVVHDInput) (*DeleteHyperVVHDOutput, error) {
	klog.V(4).InfoS("DeleteHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient
	err := client.DeleteVHD(ctx, i.Path)
	if err != nil {
		return nil, err
	}

	return &DeleteHyperVVHDOutput{}, nil
}

func (c *cloud) AttachHyperVVHD(ctx context.Context, i *AttachHyperVVHDInput) (*AttachHyperVVHDOutput, error) {
	klog.V(4).InfoS("AttachHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient

	vm, err := client.GetVMByID(ctx, i.VmID)
	if err != nil {
		return nil, err
	}

	err = client.AttachVMHardDiskDrive(
		ctx,
		vm.Name,
		hyperv.ControllerTypeSCSI,
		i.VHDPath,
	)
	if err != nil {
		return nil, err
	}

	vmDisks, err := client.GetVMHardDiskDrives(ctx, vm.Name)
	if err != nil {
		return nil, err
	}

	for _, disk := range vmDisks {
		if disk.Path == i.VHDPath {
			return &AttachHyperVVHDOutput{
				ControllerNumber:   disk.ControllerNumber,
				ControllerLocation: disk.ControllerLocation,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to attach VHD %s to VM %s", i.VHDPath, vm.Name)
}

func (c *cloud) DetachHyperVVHD(ctx context.Context, i *DetachHyperVVHDInput) (*DetachHyperVVHDOutput, error) {
	klog.V(4).InfoS("DetachHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient

	vm, err := client.GetVMByID(ctx, i.VmID)
	if err != nil {
		return nil, err
	}

	vmDisks, err := client.GetVMHardDiskDrives(ctx, vm.Name)
	if err != nil {
		return nil, err
	}

	for _, disk := range vmDisks {
		if disk.Path == i.VHDPath {
			err = client.DeleteVMHardDiskDrive(
				ctx,
				vm.Name,
				disk.ControllerNumber,
				disk.ControllerLocation,
			)
			if err != nil {
				return nil, err
			}

			return &DetachHyperVVHDOutput{
				ControllerLocation: disk.ControllerLocation,
				ControllerNumber:   disk.ControllerNumber,
			}, nil
		}
	}

	return nil, nil
}
