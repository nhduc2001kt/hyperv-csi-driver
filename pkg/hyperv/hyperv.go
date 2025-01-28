package hyperv

type HyperVClient interface {
	HyperVVHDClient
	HyperVVMHardDiskDriveClient
}
