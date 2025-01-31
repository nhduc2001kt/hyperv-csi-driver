//go:build linux
// +build linux

package mounter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
	mountutils "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	nvmeDiskPartitionSuffix = "p"
	diskPartitionSuffix     = ""
)

// constants of paths
const (
	// classBlockPath represents the path to block devices.
	classSCSIDevicePath = "/sys/class/scsi_device"

	// classSCSIHostPath represents the path to SCSI host.
	classSCSIHostPath = "/sys/class/scsi_host"

	// classSCSIBlockDevicePath represents the path to SCSI block devices.
	classSCSIBlockDevicePath = "device/block"

	// devicePath represents the path to block devices.
	devicePath = "/dev"
)

func NewSafeMounter() (*mountutils.SafeFormatAndMount, error) {
	return &mountutils.SafeFormatAndMount{
		Interface: mountutils.New(""),
		Exec:      utilexec.New(),
	}, nil
}

func NewSafeMounterV2() (*mountutils.SafeFormatAndMount, error) {
	return nil, errors.New("NewSafeMounterV2 is not supported on this platform")
}

// FindDevicePath finds path of device and verifies its existence
func (m *NodeMounter) FindDevicePath(devicePath, partition string) (string, error) {
	canonicalDevicePath := ""

	// If the given path exists, the device MAY be nvme. Further, it MAY be a
	// symlink to the nvme device path like:
	// | $ stat /dev/xvdba
	// | File: ‘/dev/xvdba’ -> ‘nvme1n1’
	// Since these are maybes, not guarantees, the search for the nvme device
	// path below must happen and must rely on volume ID
	exists, err := m.PathExists(devicePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if path %q exists: %w", devicePath, err)
	}

	if !exists {
		return "", fmt.Errorf("device path %q does not exist", devicePath)
	}

	stat, lstatErr := os.Lstat(devicePath)
	if lstatErr != nil {
		return "", fmt.Errorf("failed to lstat %q: %w", devicePath, err)
	}

	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		canonicalDevicePath, err = filepath.EvalSymlinks(devicePath)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate symlink %q: %w", devicePath, err)
		}
	} else {
		canonicalDevicePath = devicePath
	}

	klog.V(5).InfoS("The canonical device path was resolved", "devicePath", devicePath, "cacanonicalDevicePath", canonicalDevicePath)
	// strippedVolumeName := strings.ReplaceAll(volumeID, "-", "")
	// if err = verifyVolumeSerialMatch(canonicalDevicePath, strippedVolumeName, execRunner); err != nil {
	// 	return "", err
	// }
	return m.appendPartition(canonicalDevicePath, partition), nil
}

// This function is mirrored in ./sanity_test.go to make sure sanity test covered this block of code
// Please mirror the change to func MakeFile in ./sanity_test.go.
func (m *NodeMounter) PathExists(path string) (bool, error) {
	return mountutils.PathExists(path)
}

// IsBlockDevice checks if the given path is a block device.
func (m *NodeMounter) IsBlockDevice(fullPath string) (bool, error) {
	var st unix.Stat_t
	err := unix.Stat(fullPath, &st)
	if err != nil {
		return false, err
	}

	return (st.Mode & unix.S_IFMT) == unix.S_IFBLK, nil
}

// IsCorruptedMnt return true if err is about corrupted mount point.
func (m *NodeMounter) IsCorruptedMnt(err error) bool {
	return mountutils.IsCorruptedMnt(err)
}

// GetSCSIBlockDevicePath returns the block device path for the given SCSI device.
func (m *NodeMounter) GetSCSIBlockDevicePath(host *int, bus *int, target *int, lun *int) (string, error) {
	hostStr := util.ItoaOrDefault(host, `\d+`)
	busStr := util.ItoaOrDefault(bus, `\d+`)
	targetStr := util.ItoaOrDefault(target, `\d+`)
	lunStr := util.ItoaOrDefault(lun, `\d+`)
	pattern := fmt.Sprintf("^%s:%s:%s:%s$", hostStr, busStr, targetStr, lunStr)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	// Read the SCSI device directory
	scsiDevices, err := os.ReadDir(classSCSIDevicePath)
	if err != nil {
		return "", err
	}
	klog.V(4).Infof("found %d SCSI devices on host", len(scsiDevices))

	// Iterate through each SCSI device
	for _, device := range scsiDevices {
		klog.V(4).Infof("checking SCSI device %s", device.Name())

		dName := device.Name()
		if !regex.MatchString(dName) {
			klog.V(4).Infof("skipping non-matching SCSI device %s", dName)
			continue
		}

		// Construct path to block devices
		blockPath := filepath.Join(classSCSIDevicePath, dName, classSCSIBlockDevicePath)

		// Read the block device directory
		blockDevices, err := os.ReadDir(blockPath)
		if err != nil {
			klog.V(4).Infof("failed to read block device directory %s: %v", blockPath, err)
			continue
		}
		klog.V(4).Infof("found %d device entries on SCSI device %s", len(blockDevices), dName)

		// Print found block devices
		for _, blockDevice := range blockDevices {
			klog.V(4).Infof("checking block device %s", blockDevice.Name())

			devName := filepath.Join(devicePath, blockDevice.Name())
			ok, err := m.IsBlockDevice(devName)
			if ok && err == nil {
				klog.V(4).Infof("found block device %s", devName)
				return devName, nil
			}

			klog.V(4).Infof("skipping non-block device %s", devName)
		}
	}

	// Define the pattern to match
	return "", errors.New("no block device found for SCSI device")
}

// This function is mirrored in ./sanity_test.go to make sure sanity test covered this block of code
// Please mirror the change to func MakeFile in ./sanity_test.go.
func (m *NodeMounter) MakeFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	if err = f.Close(); err != nil {
		return err
	}
	return nil
}

// This function is mirrored in ./sanity_test.go to make sure sanity test covered this block of code
// Please mirror the change to func MakeFile in ./sanity_test.go.
func (m *NodeMounter) MakeDir(path string) error {
	err := os.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}

// GetDeviceNameFromMount returns the volume ID for a mount path.
func (m *NodeMounter) GetDeviceNameFromMount(mountPath string) (string, int, error) {
	return mountutils.GetDeviceNameFromMount(m, mountPath)
}

// Resize resizes the filesystem of the given devicePath.
func (m *NodeMounter) Resize(devicePath, deviceMountPath string) (bool, error) {
	return mountutils.NewResizeFs(m.Exec).Resize(devicePath, deviceMountPath)
}

// NeedResize checks if the filesystem of the given devicePath needs to be resized.
func (m *NodeMounter) NeedResize(devicePath string, deviceMountPath string) (bool, error) {
	return mountutils.NewResizeFs(m.Exec).NeedResize(devicePath, deviceMountPath)
}

// Unstage unmounts the given path.
func (m *NodeMounter) Unstage(path string) error {
	err := mountutils.CleanupMountPoint(path, m, false)
	// Ignore the error when it contains "not mounted", because that indicates the
	// world is already in the desired state
	//
	// mount-utils attempts to detect this on its own but fails when running on
	// a read-only root filesystem, which our manifests use by default
	if err == nil || strings.Contains(fmt.Sprint(err), "not mounted") {
		return nil
	} else {
		return err
	}
}

// Unpublish unmounts the given path.
func (m *NodeMounter) Unpublish(path string) error {
	// On linux, unpublish and unstage both perform an unmount
	return m.Unstage(path)
}

// PreparePublishTarget creates the target directory for the volume to be mounted.
func (m *NodeMounter) PreparePublishTarget(target string) error {
	klog.V(4).InfoS("NodePublishVolume: creating dir", "target", target)
	if err := m.MakeDir(target); err != nil {
		return fmt.Errorf("could not create dir %q: %w", target, err)
	}
	return nil
}

// appendPartition appends the partition to the device path.
func (m *NodeMounter) appendPartition(devicePath, partition string) string {
	if partition == "" {
		return devicePath
	}

	if strings.HasPrefix(devicePath, "/dev/nvme") {
		return devicePath + nvmeDiskPartitionSuffix + partition
	}

	return devicePath + diskPartitionSuffix + partition
}

// // execRunner is a helper to inject exec.Comamnd().CombinedOutput() for verifyVolumeSerialMatch
// // Tests use a mocked version that does not actually execute any binaries.
// func execRunner(name string, arg ...string) ([]byte, error) {
// 	return exec.Command(name, arg...).CombinedOutput()
// }

// // verifyVolumeSerialMatch checks the volume serial of the device against the expected volume.
// func verifyVolumeSerialMatch(canonicalDevicePath string, strippedVolumeName string, execRunner func(string, ...string) ([]byte, error)) error {
// 	// In some rare cases, a race condition can lead to the /dev/disk/by-id/ symlink becoming out of date
// 	// See https://github.com/kubernetes-sigs/aws-ebs-csi-driver/issues/1224 for more info
// 	// Attempt to use lsblk to double check that the nvme device selected was the correct volume
// 	output, err := execRunner("lsblk", "--noheadings", "--ascii", "--nodeps", "--output", "SERIAL", canonicalDevicePath)

// 	if err == nil {
// 		// Look for an EBS volume ID in the output, compare all matches against what we expect
// 		// (in some rare cases there may be multiple matches due to lsblk printing partitions)
// 		// If no volume ID is in the output (non-Nitro instances, SBE devices, etc) silently proceed
// 		volumeRegex := regexp.MustCompile(`vol[a-z0-9]+`)
// 		for _, volume := range volumeRegex.FindAllString(string(output), -1) {
// 			klog.V(6).InfoS("Comparing volume serial", "canonicalDevicePath", canonicalDevicePath, "expected", strippedVolumeName, "actual", volume)
// 			if volume != strippedVolumeName {
// 				return fmt.Errorf("refusing to mount %s because it claims to be %s but should be %s", canonicalDevicePath, volume, strippedVolumeName)
// 			}
// 		}
// 	} else {
// 		// If the command fails (for example, because lsblk is not available), silently ignore the error and proceed
// 		klog.V(5).ErrorS(err, "Ignoring lsblk failure", "canonicalDevicePath", canonicalDevicePath, "strippedVolumeName", strippedVolumeName)
// 	}

// 	return nil
// }
