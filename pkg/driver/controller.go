package driver

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/cloud"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/driver/internal"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util/template"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// Supported access modes.
const (
	SingleNodeWriter     = csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
	MultiNodeMultiWriter = csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER
)

var (
	// controllerCaps represents the capability of controller service.
	controllerCaps = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		// csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		// csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		// csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		// csi.ControllerServiceCapability_RPC_MODIFY_VOLUME,
	}
)

// const trueStr = "true"
// const isManagedByDriver = trueStr

// ControllerService represents the controller service of CSI driver.
type ControllerService struct {
	inFlight *internal.InFlight
	options  *options.Options
	cloud    cloud.Cloud
	// modifyVolumeCoalescer coalescer.Coalescer[modifyVolumeRequest, int32]
	// rpc.UnimplementedModifyServer
	csi.UnimplementedControllerServer
}

// NewControllerService creates a new controller service.
func NewControllerService(c cloud.Cloud, o *options.Options) *ControllerService {
	return &ControllerService{
		cloud:    c,
		options:  o,
		inFlight: internal.NewInFlight(),
		// modifyVolumeCoalescer: newModifyVolumeCoalescer(c, o),
	}
}

func (d *ControllerService) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.V(4).InfoS("ControllerGetCapabilities: called", "args", req)

	caps := make([]*csi.ControllerServiceCapability, 0, len(controllerCaps))
	for _, capability := range controllerCaps {
		c := &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: capability,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (d *ControllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(4).InfoS("CreateVolume: called", "args", util.SanitizeRequest(req))
	if err := validateCreateVolumeRequest(req); err != nil {
		return nil, err
	}
	volSizeBytes, err := getVolSizeBytes(req)
	if err != nil {
		return nil, err
	}
	volName := req.GetName()
	volCap := req.GetVolumeCapabilities()

	multiAttach := false
	for _, c := range volCap {
		if c.GetAccessMode().GetMode() == MultiNodeMultiWriter && isBlock(c) {
			klog.V(4).InfoS("CreateVolume: multi-attach is enabled", "volumeID", volName)
			multiAttach = true
		}
	}
	_ = multiAttach

	// check if a request is already in-flight
	if ok := d.inFlight.Insert(volName); !ok {
		msg := fmt.Sprintf("Create volume request for %s is already in progress", volName)
		return nil, status.Error(codes.Aborted, msg)
	}
	defer d.inFlight.Delete(volName)

	var (
		vhdType         = hyperv.VHDTypeFixed
		vhdFormat       = hyperv.VHDFormatVHDX
		vhdBlockSize    uint32
		tags            = map[string]string{}
		inodeSize       string
		bytesPerInode   string
		numberOfInodes  string
		ext4BigAlloc    bool
		ext4ClusterSize string
	)

	tProps := new(template.PVProps)
	_ = tProps

	for key, value := range req.GetParameters() {
		switch strings.ToLower(key) {
		case VHDTypeKey:
			vhdType, err = hyperv.StringToVHDType(value)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse invalid VHD type: %v", err)
			}
		case VHDFormatKey:
			vhdFormat, err = hyperv.StringToVHDFormat(value)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse invalid VHD format: %v", err)
			}
		case VHDBlockSizeKey:
			parseBlockSizeKey, parseBlockSizeKeyErr := strconv.ParseInt(value, 10, 32)
			if parseBlockSizeKeyErr != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse invalid block size: %v", err)
			}
			vhdBlockSize = uint32(parseBlockSizeKey)
		case KubernetesPVCNameKey:
			tags[PVCNameTag] = value
			tProps.PVCName = value
		case KubernetesPVCNamespaceKey:
			tags[PVCNamespaceTag] = value
			tProps.PVCNamespace = value
		case KubernetesPVNameKey:
			tags[PVNameTag] = value
			tProps.PVName = value
		case InodeSizeKey:
			if isAlphanumeric := util.StringIsAlphanumeric(value); !isAlphanumeric {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse inodeSize (%s): %v", value, err)
			}
			inodeSize = value
		case BytesPerInodeKey:
			if isAlphanumeric := util.StringIsAlphanumeric(value); !isAlphanumeric {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse bytesPerInode (%s): %v", value, err)
			}
			bytesPerInode = value
		case NumberOfInodesKey:
			if isAlphanumeric := util.StringIsAlphanumeric(value); !isAlphanumeric {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse numberOfInodes (%s): %v", value, err)
			}
			numberOfInodes = value
		case Ext4BigAllocKey:
			ext4BigAlloc = util.IsTrue(value)
		case Ext4ClusterSizeKey:
			if isAlphanumeric := util.StringIsAlphanumeric(value); !isAlphanumeric {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse ext4ClusterSize (%s): %v", value, err)
			}
			ext4ClusterSize = value
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Invalid parameter key %s for CreateVolume", key)
		}
	}

	// TODO: treat Mutable Parameters
	_, err = parseModifyVolumeParameters(req.GetMutableParameters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid mutable parameter: %v", err)
	}

	snapshotID := ""
	volumeSource := req.GetVolumeContentSource()
	if volumeSource != nil {
		if _, ok := volumeSource.GetType().(*csi.VolumeContentSource_Snapshot); !ok {
			return nil, status.Error(codes.InvalidArgument, "Unsupported volumeContentSource type")
		}
		sourceSnapshot := volumeSource.GetSnapshot()
		if sourceSnapshot == nil {
			return nil, status.Error(codes.InvalidArgument, "Error retrieving snapshot from the volumeContentSource")
		}
		snapshotID = sourceSnapshot.GetSnapshotId()
	}
	_ = snapshotID

	// TODO: handle Accessibility Requirements

	responseCtx := map[string]string{}

	if vhdBlockSize > 0 {
		responseCtx[VHDBlockSizeKey] = strconv.Itoa(int(vhdBlockSize))
		if err = validateFormattingOption(volCap, VHDBlockSizeKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}
	if len(inodeSize) > 0 {
		responseCtx[InodeSizeKey] = inodeSize
		if err = validateFormattingOption(volCap, InodeSizeKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}
	if len(bytesPerInode) > 0 {
		responseCtx[BytesPerInodeKey] = bytesPerInode
		if err = validateFormattingOption(volCap, BytesPerInodeKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}
	if len(numberOfInodes) > 0 {
		responseCtx[NumberOfInodesKey] = numberOfInodes
		if err = validateFormattingOption(volCap, NumberOfInodesKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}
	if ext4BigAlloc {
		responseCtx[Ext4BigAllocKey] = "true"
		if err = validateFormattingOption(volCap, Ext4BigAllocKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}
	if len(ext4ClusterSize) > 0 {
		responseCtx[Ext4ClusterSizeKey] = ext4ClusterSize
		if err = validateFormattingOption(volCap, Ext4ClusterSizeKey, FileSystemConfigs); err != nil {
			return nil, err
		}
	}

	// Fill volume tags
	if d.options.KubernetesClusterID != "" {
		resourceLifecycleTag := ResourceLifecycleTagPrefix + d.options.KubernetesClusterID
		tags[resourceLifecycleTag] = ResourceLifecycleOwned
		tags[NameTag] = d.options.KubernetesClusterID + "-dynamic-" + volName
		tags[KubernetesClusterTag] = d.options.KubernetesClusterID
	}

	// TODO: validate tags

	input := &cloud.CreateHyperVVHDInput{
		Name: volName,
		// Source:             snapshotID,
		// SourceVm:           sourceVm,
		// SourceDisk:         sourceDisk,
		Type: vhdType,
		// ParentPath:         parentPath,
		Size:      uint64(volSizeBytes),
		BlockSize: vhdBlockSize,
		Format:    vhdFormat,
		// LogicalSectorSize:  logicalSectorSize,
		// PhysicalSectorSize: physicalSectorSize,
	}
	output, err := d.cloud.CreateHyperVVHD(ctx, input)
	if err != nil {
		var errCode codes.Code
		switch {
		// TODO: handle error
		// case errors.Is(err, cloud.ErrNotFound):
		// 	errCode = codes.NotFound
		// case errors.Is(err, cloud.ErrIdempotentParameterMismatch), errors.Is(err, cloud.ErrAlreadyExists):
		// 	errCode = codes.AlreadyExists
		default:
			errCode = codes.Internal
		}
		return nil, status.Errorf(errCode, "Could not create volume %q: %v", volName, err)
	}
	return newCreateVolumeResponse(output), nil
}

func (d *ControllerService) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(4).InfoS("DeleteVolume: called", "args", util.SanitizeRequest(req))
	if err := validateDeleteVolumeRequest(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()

	// check if a request is already in-flight
	if ok := d.inFlight.Insert(volumeID); !ok {
		msg := fmt.Sprintf(internal.VolumeOperationAlreadyExistsErrorMsg, volumeID)
		return nil, status.Error(codes.Aborted, msg)
	}
	defer d.inFlight.Delete(volumeID)

	input := &cloud.DeleteHyperVVHDInput{
		Path: volumeID,
	}
	if _, err := d.cloud.DeleteHyperVVHD(ctx, input); err != nil {
		// if errors.Is(err, cloud.ErrNotFound) {
		// 	klog.V(4).InfoS("DeleteVolume: volume not found, returning with success")
		// 	return &csi.DeleteVolumeResponse{}, nil
		// }
		// TODO: handle error not found
		return nil, status.Errorf(codes.Internal, "Could not delete volume ID %q: %v", volumeID, err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (d *ControllerService) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.V(4).InfoS("ControllerPublishVolume: called", "args", util.SanitizeRequest(req))
	if err := validateControllerPublishVolumeRequest(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	nodeID := req.GetNodeId()

	if !d.inFlight.Insert(volumeID + nodeID) {
		return nil, status.Error(codes.Aborted, fmt.Sprintf(internal.VolumeOperationAlreadyExistsErrorMsg, volumeID))
	}
	defer d.inFlight.Delete(volumeID + nodeID)

	input := cloud.AttachHyperVVHDInput{
		VmID:    nodeID,
		VHDPath: volumeID,
	}
	output, err := d.cloud.AttachHyperVVHD(ctx, &input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not attach volume %q to node %q: %v", volumeID, nodeID, err)
	}

	pvInfo := map[string]string{
		ControllerNumberKey:   strconv.Itoa(int(output.ControllerNumber)),
		ControllerLocationKey: strconv.Itoa(int(output.ControllerLocation)),
	}
	return &csi.ControllerPublishVolumeResponse{PublishContext: pvInfo}, nil
}

func (d *ControllerService) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.V(4).InfoS("ControllerUnpublishVolume: called", "args", util.SanitizeRequest(req))

	if err := validateControllerUnpublishVolumeRequest(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	nodeID := req.GetNodeId()

	if !d.inFlight.Insert(volumeID + nodeID) {
		return nil, status.Error(codes.Aborted, fmt.Sprintf(internal.VolumeOperationAlreadyExistsErrorMsg, volumeID))
	}
	defer d.inFlight.Delete(volumeID + nodeID)

	klog.V(2).InfoS("ControllerUnpublishVolume: detaching", "volumeID", volumeID, "nodeID", nodeID)
	input := cloud.DetachHyperVVHDInput{
		VmID:    nodeID,
		VHDPath: volumeID,
	}
	output, err := d.cloud.DetachHyperVVHD(ctx, &input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not attach volume %q to node %q: %v", volumeID, nodeID, err)
	}
	if output == nil {
		klog.InfoS("ControllerUnpublishVolume: attachment not found", "volumeID", volumeID, "nodeID", nodeID)
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func newCreateVolumeResponse(output *cloud.CreateHyperVVHDOutput) *csi.CreateVolumeResponse {
	var src *csi.VolumeContentSource
	// if output.SnapshotID != "" {
	// 	src = &csi.VolumeContentSource{
	// 		Type: &csi.VolumeContentSource_Snapshot{
	// 			Snapshot: &csi.VolumeContentSource_SnapshotSource{
	// 				SnapshotId: disk.SnapshotID,
	// 			},
	// 		},
	// 	}
	// }
	// TODO: handle snapshot

	segments := map[string]string{
		// WellKnownZoneTopologyKey: disk.AvailabilityZone
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      output.Path,
			CapacityBytes: util.GiBToBytes(8), // TODO handle disk size
			// VolumeContext: ctx, // TODO handle volume ctx
			AccessibleTopology: []*csi.Topology{
				{
					Segments: segments,
				},
			},
			ContentSource: src,
		},
	}
}

func validateFormattingOption(volumeCapabilities []*csi.VolumeCapability, paramName string, fsConfigs map[string]fileSystemConfig) error {
	for _, volCap := range volumeCapabilities {
		if isBlock(volCap) {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("Cannot use %s with block volume", paramName))
		}

		mountVolume := volCap.GetMount()
		if mountVolume == nil {
			return status.Error(codes.InvalidArgument, "CreateVolume: mount is nil within volume capability")
		}

		fsType := mountVolume.GetFsType()
		if supported := fsConfigs[fsType].isParameterSupported(paramName); !supported {
			return status.Errorf(codes.InvalidArgument, "Cannot use %s with fstype %s", paramName, fsType)
		}
	}

	return nil
}

func validateCreateVolumeRequest(req *csi.CreateVolumeRequest) error {
	volName := req.GetName()
	if len(volName) == 0 {
		return status.Error(codes.InvalidArgument, "Volume name not provided")
	}

	volCaps := req.GetVolumeCapabilities()
	if len(volCaps) == 0 {
		return status.Error(codes.InvalidArgument, "Volume capabilities not provided")
	}

	if !isValidVolumeCapabilities(volCaps) {
		return status.Error(codes.InvalidArgument, "Volume capabilities not supported")
	}
	return nil
}

func validateDeleteVolumeRequest(req *csi.DeleteVolumeRequest) error {
	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	return nil
}

func validateControllerPublishVolumeRequest(req *csi.ControllerPublishVolumeRequest) error {
	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	if len(req.GetNodeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Node ID not provided")
	}

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	if !isValidCapability(volCap) {
		return status.Error(codes.InvalidArgument, "Volume capability not supported")
	}
	return nil
}

func validateControllerUnpublishVolumeRequest(req *csi.ControllerUnpublishVolumeRequest) error {
	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	if len(req.GetNodeId()) == 0 {
		return status.Error(codes.InvalidArgument, "Node ID not provided")
	}

	return nil
}

func isValidVolumeCapabilities(v []*csi.VolumeCapability) bool {
	for _, c := range v {
		if !isValidCapability(c) {
			return false
		}
	}
	return true
}

func isValidCapability(c *csi.VolumeCapability) bool {
	accessMode := c.GetAccessMode().GetMode()

	//nolint:exhaustive
	switch accessMode {
	case SingleNodeWriter:
		return true

	case MultiNodeMultiWriter:
		if isBlock(c) {
			return true
		} else {
			klog.InfoS("isValidCapability: access mode is only supported for block devices", "accessMode", accessMode)
			return false
		}

	default:
		klog.InfoS("isValidCapability: access mode is not supported", "accessMode", accessMode)
		return false
	}
}

func isBlock(capability *csi.VolumeCapability) bool {
	_, isBlk := capability.GetAccessType().(*csi.VolumeCapability_Block)
	return isBlk
}

func getVolSizeBytes(req *csi.CreateVolumeRequest) (int64, error) {
	var volSizeBytes int64
	capRange := req.GetCapacityRange()
	if capRange == nil {
		volSizeBytes = cloud.DefaultVolumeSize
	} else {
		volSizeBytes = util.RoundUpBytes(capRange.GetRequiredBytes())
		maxVolSize := capRange.GetLimitBytes()
		if maxVolSize > 0 && maxVolSize < volSizeBytes {
			return 0, status.Error(codes.InvalidArgument, "After round-up, volume size exceeds the limit specified")
		}
	}
	return volSizeBytes, nil
}
