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

// CreateHyperVVHDInput represents the input for CreateHyperVVHD.
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

// GetHyperVVHDInput represents the input for GetHyperVVHD.
type GetHyperVVHDInput struct {
	Path string
}

// GetHyperVVHDOutput represents the output for GetHyperVVHD.
type GetHyperVVHDOutput struct {
	Name               string
	Type               hyperv.VHDType
	Format             hyperv.VHDFormat
	ParentPath         string
	Size               uint64
	BlockSize          uint32
	LogicalSectorSize  uint32
	PhysicalSectorSize uint32
}

// CreateHyperVVHDOutput represents the output for CreateHyperVVHD.
type CreateHyperVVHDOutput struct {
	Path string
}

// DeleteHyperVVHDInput represents the input for DeleteHyperVVHD.
type DeleteHyperVVHDInput struct {
	Path string
}

// DeleteHyperVVHDOutput represents the output for DeleteHyperVVHD.
type DeleteHyperVVHDOutput struct{}

// AttachHyperVVHDInput represents the input for AttachHyperVVHD.
type AttachHyperVVHDInput struct {
	VmID    string
	VHDPath string
}

// AttachHyperVVHDOutput represents the output for AttachHyperVVHD.
type AttachHyperVVHDOutput struct {
	ControllerNumber   int32
	ControllerLocation int32
}

// DetachHyperVVHDInput represents the input for DetachHyperVVHD.
type DetachHyperVVHDInput struct {
	VmID    string
	VHDPath string
}

// DetachHyperVVHDOutput represents the output for DetachHyperVVHD.
type DetachHyperVVHDOutput struct {
	ControllerNumber   int32
	ControllerLocation int32
}

type Cloud interface {
	GetHyperVVHD(context.Context, *GetHyperVVHDInput) (*GetHyperVVHDOutput, error)
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

func (c *cloud) GetHyperVVHD(ctx context.Context, i *GetHyperVVHDInput) (*GetHyperVVHDOutput, error) {
	klog.V(4).InfoS("GetHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient

	vhd, err := client.GetVHD(ctx, i.Path)
	if err != nil {
		return nil, err
	}
	if vhd.Path != i.Path {
		return nil, fmt.Errorf("VHD not found: %s", i.Path)
	}

	return &GetHyperVVHDOutput{
		Name:               vhd.Path,
		Type:               vhd.VHDType,
		Format:             vhd.VHDFormat,
		ParentPath:         vhd.ParentPath,
		Size:               vhd.Size,
		BlockSize:          vhd.BlockSize,
		LogicalSectorSize:  vhd.LogicalSectorSize,
		PhysicalSectorSize: vhd.PhysicalSectorSize,
	}, nil
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

	res, err := client.AttachVMHardDiskDrive(
		ctx,
		i.VmID,
		hyperv.ControllerTypeSCSI,
		i.VHDPath,
	)
	if err != nil {
		return nil, err
	}

	return &AttachHyperVVHDOutput{
		ControllerNumber:   res.ControllerNumber,
		ControllerLocation: res.ControllerLocation,
	}, nil
}

func (c *cloud) DetachHyperVVHD(ctx context.Context, i *DetachHyperVVHDInput) (*DetachHyperVVHDOutput, error) {
	klog.V(4).InfoS("DetachHyperVVHD: called", "args", util.SanitizeRequest(i))

	client := c.hypervClient

	err := client.DetachVMHardDiskDrive(ctx, i.VmID, i.VHDPath)
	if err != nil {
		return nil, err
	}

	return &DetachHyperVVHDOutput{}, nil
}
