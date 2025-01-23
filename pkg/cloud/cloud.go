package cloud

import (
	"context"
	"fmt"
	"path/filepath"

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

type Cloud interface {
	CreateHyperVVHD(context.Context, *CreateHyperVVHDInput) (*CreateHyperVVHDOutput, error)
	DeleteHyperVVHD(context.Context, *DeleteHyperVVHDInput) (*DeleteHyperVVHDOutput, error)
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
	vhdPath := util.WinPath(filepath.Join(c.vhdBasePath, vhdFile))
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
