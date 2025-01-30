package driver

const (
	DriverName = "hyperv.csi.k8s.io"
)

// constants for node k8s API use.
const (
	// AgentNotReadyNodeTaintKey contains the key of taints to be removed on driver startup.
	AgentNotReadyNodeTaintKey = "hyperv.csi.k8s.io/agent-not-ready"
)

// constants for volume tags and their values.
const (
	// ResourceLifecycleTagPrefix is prefix of tag for provisioned EBS volume that
	// marks them as owned by the cluster. Used only when --cluster-id is set.
	ResourceLifecycleTagPrefix = "kubernetes.io/cluster/"

	// ResourceLifecycleOwned is the value we use when tagging resources to indicate
	// that the resource is considered owned and managed by the cluster,
	// and in particular that the lifecycle is tied to the lifecycle of the cluster.
	// From k8s.io/legacy-cloud-providers/aws/tags.go.
	ResourceLifecycleOwned = "owned"

	// NameTag is tag applied to provisioned EBS volume for backward compatibility with
	// in-tree volume plugin. Used only when --cluster-id is set.
	NameTag = "Name"

	// KubernetesClusterTag is tag applied to provisioned EBS volume for backward compatibility with
	// in-tree volume plugin. Used only when --cluster-id is set.
	// See https://github.com/kubernetes/cloud-provider-aws/blob/release-1.20/pkg/providers/v1/tags.go#L38-L41.
	KubernetesClusterTag = "KubernetesCluster"

	// PVCNameTag is tag applied to provisioned EBS volume for backward compatibility
	// with in-tree volume plugin. Value of the tag is PVC name. It is applied only when
	// the external provisioner sidecar is started with --extra-create-metadata=true and
	// thus provides such metadata to the CSI driver.
	PVCNameTag = "kubernetes.io/created-for/pvc/name"

	// PVCNamespaceTag is tag applied to provisioned EBS volume for backward compatibility
	// with in-tree volume plugin. Value of the tag is PVC namespace. It is applied only when
	// the external provisioner sidecar is started with --extra-create-metadata=true and
	// thus provides such metadata to the CSI driver.
	PVCNamespaceTag = "kubernetes.io/created-for/pvc/namespace"

	// PVNameTag is tag applied to provisioned EBS volume for backward compatibility
	// with in-tree volume plugin. Value of the tag is PV name. It is applied only when
	// the external provisioner sidecar is started with --extra-create-metadata=true and
	// thus provides such metadata to the CSI driver.
	PVNameTag = "kubernetes.io/created-for/pv/name"
)

// constants of keys in volume parameters.
const (
	// VHDTypeKey represents key for VHD type to use. Valid values to use are Unknown, Fixed,
	// Dynamic, Differencing.
	VHDTypeKey = "type"

	// VHDFormatKey represents key for the format of the virtual hard disk to be created.
	VHDFormatKey = "format"

	// VHDBlockSizeKey represents key for the block size, in bytes, of the virtual hard disk to
	// be created.
	VHDBlockSizeKey = "blockSize"

	// InodeSizeKey configures the inode size when formatting a volume.
	InodeSizeKey = "inodesize"

	// BytesPerInodeKey configures the `bytes-per-inode` when formatting a volume.
	BytesPerInodeKey = "bytesperinode"

	// NumberOfInodesKey configures the `number-of-inodes` when formatting a volume.
	NumberOfInodesKey = "numberofinodes"

	// Ext4ClusterSizeKey enables the bigalloc option when formatting an ext4 volume.
	Ext4BigAllocKey = "ext4bigalloc"

	// Ext4ClusterSizeKey configures the cluster size when formatting an ext4 volume with the bigalloc option enabled.
	Ext4ClusterSizeKey = "ext4clustersize"

	// KubernetesPVCNameKey contains name of the PVC for which is a volume provisioned.
	KubernetesPVCNameKey = "csi.storage.k8s.io/pvc/name"

	// KubernetesPVCNamespaceKey contains namespace of the PVC for which is a volume provisioned.
	KubernetesPVCNamespaceKey = "csi.storage.k8s.io/pvc/namespace"

	// KubernetesPVNameKey contains name of the final PV that will be used for the dynamically
	// provisioned volume.
	KubernetesPVNameKey = "csi.storage.k8s.io/pv/name"
)

// constants of keys in PublishContext.
const (
	// ControllerNumberKey represents key for the controller number to use when attaching the
	// virtual hard disk to the virtual machine.
	ControllerNumberKey = "controllerNumber"

	// ControllerLocationKey represents key for the controller location to use when attaching
	// the virtual hard disk to the virtual machine.
	ControllerLocationKey = "controllerLocation"
)

// constants of keys in VolumeContext.
const (
	// VolumeAttributePartition represents key for partition config in VolumeContext
	// this represents the partition number on a device used to mount.
	VolumeAttributePartition = "partition"
)

// constants for fstypes.
const (
	// FSTypeExt3 represents the ext3 filesystem type.
	FSTypeExt3 = "ext3"
	// FSTypeExt4 represents the ext4 filesystem type.
	FSTypeExt4 = "ext4"
	// FSTypeXfs represents the xfs filesystem type.
	FSTypeXfs = "xfs"
	// FSTypeNtfs represents the ntfs filesystem type.
	FSTypeNtfs = "ntfs"
)

var (
	ValidFSTypes = map[string]struct{}{
		FSTypeExt3: {},
		FSTypeExt4: {},
		// FSTypeXfs:  {},
		// FSTypeNtfs: {},
	}
)

type fileSystemConfig struct {
	NotSupportedParams map[string]struct{}
}

func (fsConfig fileSystemConfig) isParameterSupported(paramName string) bool {
	_, notSupported := fsConfig.NotSupportedParams[paramName]
	return !notSupported
}


var (
	FileSystemConfigs = map[string]fileSystemConfig{
		FSTypeExt3: {
			NotSupportedParams: map[string]struct{}{
				// Ext4BigAllocKey:    {},
				// Ext4ClusterSizeKey: {},
			},
		},
		FSTypeExt4: {
			NotSupportedParams: map[string]struct{}{},
		},
		FSTypeXfs: {
			NotSupportedParams: map[string]struct{}{
				// BytesPerInodeKey:   {},
				// NumberOfInodesKey:  {},
				// Ext4BigAllocKey:    {},
				// Ext4ClusterSizeKey: {},
			},
		},
		FSTypeNtfs: {
			NotSupportedParams: map[string]struct{}{
				// BlockSizeKey:       {},
				// InodeSizeKey:       {},
				// BytesPerInodeKey:   {},
				// NumberOfInodesKey:  {},
				// Ext4BigAllocKey:    {},
				// Ext4ClusterSizeKey: {},
			},
		},
	}
)

