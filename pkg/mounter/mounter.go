package mounter

import (
	mountutils "k8s.io/mount-utils"
)

// NodeMounter implements Mounter.
// A superstruct of SafeFormatAndMount.
type Mounter interface {
	mountutils.Interface

	FormatAndMountSensitiveWithFormatOptions(source string, target string, fstype string, options []string, sensitiveOptions []string, formatOptions []string) error
	IsBlockDevice(fullPath string) (bool, error)
	IsCorruptedMnt(err error) bool
	CountSCSIHosts() (int, error)
	CountSCSIDevices() (int, error)
	GetSCSIBlockDevicePath(host *int, bus *int, target *int, lun *int) (string, error)
	GetDeviceNameFromMount(mountPath string) (string, int, error)
	FindDevicePath(devicePath, partition string) (string, error)
	PathExists(path string) (bool, error)
	MakeFile(path string) error
	MakeDir(path string) error
	NeedResize(devicePath string, deviceMountPath string) (bool, error)
	Resize(devicePath, deviceMountPath string) (bool, error)
	Unstage(path string) error
	Unpublish(path string) error
	PreparePublishTarget(target string) error
}

// NodeMounter implements Mounter.
// A superstruct of SafeFormatAndMount.
type NodeMounter struct {
	*mountutils.SafeFormatAndMount
}

// NewNodeMounter returns a new intsance of NodeMounter.
func NewNodeMounter(hostprocess bool) (Mounter, error) {
	var safeMounter *mountutils.SafeFormatAndMount
	var err error

	if hostprocess {
		safeMounter, err = NewSafeMounterV2()
	} else {
		safeMounter, err = NewSafeMounter()
	}

	if err != nil {
		return nil, err
	}
	return &NodeMounter{safeMounter}, nil
}
