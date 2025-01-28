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
	// VHDPath represents key for path to the existing virtual hard disk file(s) that is
	// being created or being copied to. If a filename or relative path is specified, the
	// virtual hard disk path is calculated relative to the current working directory.
	// Depending on the source selected, the path will be used to determine where to copy
	// source vhd/vhdx/vhds file to.
	VHDPath = "path"

	// VHDPath represents key for VHD type to use. Valid values to use are Unknown, Fixed,
	// Dynamic, Differencing.
	VHDTypeKey = "type"

	// VHDFormatKey represents key for the format of the virtual hard disk to be created.
	VHDFormatKey = "format"

	// VHDBlockSize represents key for the block size, in bytes, of the virtual hard disk to
	// be created.
	VHDBlockSize = "blockSize"

	// VHDSize for represents key the maximum size, in bytes, of the virtual hard disk to be
	// created.
	VHDSize = "size"

	// KubernetesPVCNameKey contains name of the PVC for which is a volume provisioned.
	KubernetesPVCNameKey = "csi.storage.k8s.io/pvc/name"

	// KubernetesPVCNamespaceKey contains namespace of the PVC for which is a volume provisioned.
	KubernetesPVCNamespaceKey = "csi.storage.k8s.io/pvc/namespace"

	// KubernetesPVNameKey contains name of the final PV that will be used for the dynamically
	// provisioned volume.
	KubernetesPVNameKey = "csi.storage.k8s.io/pv/name"
)


