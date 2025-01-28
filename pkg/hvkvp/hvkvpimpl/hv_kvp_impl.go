package hvkvpimpl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode/utf16"
	"unsafe"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hvkvp"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util/types/addressfamily"
	syscall "golang.org/x/sys/unix"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type hypervKVPImpl struct {
	osName        string
	osMajor       string
	osMinor       string
	osVersion     string
	osBuild       string
	processorArch string
	fileInfos     []kvpFileInfo
	licVersion    string

	fd int
}

type kvpFileInfo struct {
	fname   string
	records map[string]string
}

func NewHyperVKVP() hvkvp.HyperVKVP {
	fileInfos := make([]kvpFileInfo, HyperVKVPPoolCount)
	for i := 0; i < HyperVKVPPoolCount; i++ {
		fileInfos[i] = kvpFileInfo{
			fname:   filepath.Join(HyperVKVPConfigLoc, fmt.Sprintf("%s%d", HyperVKVPPoolFilePrefix, i)),
			records: make(map[string]string),
		}
	}

	return &hypervKVPImpl{
		fileInfos:  fileInfos,
		licVersion: "Unknown",
	}
}

func (h *hypervKVPImpl) WaitDaemonPool(ctx context.Context, pool int) error {
	filePath := filepath.Join(
		HyperVKVPConfigLoc,
		fmt.Sprintf("%s%d", HyperVKVPPoolFilePrefix, pool),
	)

	return wait.PollUntilContextTimeout(
		ctx,
		HyperVKVPPoolFileCheckInterval,
		HyperVKVPPoolFileCheckTimeout,
		false,
		func(ctx context.Context) (bool, error) {
			_, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return false, nil
			}

			return err == nil, err
		},
	)
}

func (h *hypervKVPImpl) ReadPool(ctx context.Context, pool int) (*hvkvp.HyperVKVPInfo, error) {
	err := h.updateMemState(pool, false)
	if err != nil {
		return nil, fmt.Errorf("failed to update memory state for pool %d: %v", pool, err)
	}

	mapInfo := h.fileInfos[pool].records

	info := hvkvp.HyperVKVPInfo{}
	rVal := reflect.ValueOf(&info).Elem()
	numField := rVal.NumField()

	for i := 0; i < numField; i++ {
		field := rVal.Type().Field(i)
		tagKey := field.Tag.Get(hvkvp.HyerVKVPInfoKeyTag)

		if tagKey == "" {
			continue
		}

		val, ok := mapInfo[tagKey]
		if !ok {
			continue
		}

		fieldVal := rVal.Field(i)
		if !fieldVal.IsValid() {
			continue
		}

		if !fieldVal.CanSet() {
			continue
		}

		klog.Infof("Loading field %s with value %s", tagKey, val)

		fieldVal.SetString(val)
	}

	return &info, nil
}

func (h *hypervKVPImpl) getOSInfo() error {
	var uts syscall.Utsname
	err := syscall.Uname(&uts)
	if err != nil {
		return fmt.Errorf("failed to get system information: %v", err)
	}

	osVersion := string(uts.Release[:])
	h.osBuild = osVersion
	h.osName = string(uts.Sysname[:])
	h.processorArch = string(uts.Machine[:])

	if dashIndex := strings.Index(osVersion, "-"); dashIndex != -1 {
		h.osVersion = osVersion[:dashIndex]
	}

	var releaseFile string
	var file *os.File
	filesToTry := []string{OSReleaseFile, SuSEReleaseFile, RedHatReleaseFile}
	for _, filePath := range filesToTry {
		file, err = os.Open(filePath)
		if err == nil {
			// Successfully opened a file
			releaseFile = filePath
			defer file.Close()
			break
		}
	}

	switch releaseFile {
	case OSReleaseFile:
		h.handleOSReleaseFile(file)
	case SuSEReleaseFile, RedHatReleaseFile:
		h.handleOtherReleaseFile(file)
	default:
		return errors.New("OS information not found")
	}

	return nil
}

func (h *hypervKVPImpl) getDomainName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("no IP addresses found for hostname %s", hostname)
	}

	canonicalName := hostname
	ips, err := net.LookupAddr(addrs[0].String())
	if err == nil && len(ips) > 0 {
		canonicalName = ips[0]
	}

	return canonicalName, nil
}

func (h *hypervKVPImpl) handleOSReleaseFile(file *os.File) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Split into key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], `"'`) // Remove quotes

		// Handle the specific keys
		switch key {
		case "NAME":
			h.osName = value
		case "VERSION_ID":
			h.osMajor = value
		}
	}
}

func (h *hypervKVPImpl) handleOtherReleaseFile(file *os.File) {
	// Read up to three lines from the file
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		h.osName = strings.TrimSpace(scanner.Text())
	}

	if scanner.Scan() {
		h.osMajor = strings.TrimSpace(scanner.Text())
	}

	if scanner.Scan() {
		h.osMinor = strings.TrimSpace(scanner.Text())
	}
}

func (h *hypervKVPImpl) InitFile() error {
	if _, err := os.Stat(HyperVKVPConfigLoc); os.IsNotExist(err) {
		err := os.MkdirAll(HyperVKVPConfigLoc, 0755)
		if err != nil {
			return err
		}
	}

	for i, info := range h.fileInfos {
		fname := info.fname
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			file, err := os.Create(fname)
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %v", fname, err)
			}
			file.Close()
		}

		err := h.updateMemState(i, true)
		if err != nil {
			return fmt.Errorf("failed to update memory state for pool %d: %v", i, err)
		}
	}

	return nil
}

func (h *hypervKVPImpl) updateMemState(pool int, lock bool) error {
	if pool < 0 || pool >= HyperVKVPPoolCount {
		return errors.New("invalid pool index")
	}

	klog.Infof("Reading file: %s", h.fileInfos[pool].fname)

	info := h.fileInfos[pool]
	file, err := os.Open(info.fname)
	if err != nil {
		return fmt.Errorf("failed to open file '%v': %v", info.fname, err)
	}
	defer file.Close()

	if lock {
		err = h.acquireLock(int(file.Fd()))
		if err != nil {
			return fmt.Errorf("failed to acquire lock for pool %d: %v", pool, err)
		}
		defer h.releaseLock(int(file.Fd()))
	}

	for {
		// Read data from the file
		size := HyperVKPVExchangeMaxKeySize + HyperVKPVExchangeMaxValueSize
		buffer := make([]byte, size)

		klog.Infof("Reading size: %d", size)
		readBytes, err := file.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			return fmt.Errorf("failed to read file, pool: %d; error: %v", pool, err)
		}

		klog.Infof("Read size: %d", readBytes)

		if readBytes <= 0 {
			break
		}

		if readBytes == int(size) {
			// Deserialize the record
			key := syscall.ByteSliceToString(buffer[:HyperVKPVExchangeMaxKeySize])
			val := syscall.ByteSliceToString(buffer[HyperVKPVExchangeMaxKeySize:])
			h.fileInfos[pool].records[key] = val
		} else {
			return fmt.Errorf("something went wrong with reading file, pool: %d", pool)
		}
	}

	return nil
}

func (h *hypervKVPImpl) acquireLock(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_EX)
}

func (h *hypervKVPImpl) releaseLock(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_UN)
}

// hvKVPMsg represents the Hyper-V key value pairs message
type hvKVPMsg struct {
	Header hvKVPHdr
	Body   [HyperVKVPMsgBodySize]byte
}

// hvKVPHdr represents the header of the hvKVPMsg
type hvKVPHdr struct {
	Operation uint8
	Pool      uint8
	Pad       uint16
}

// hvKVPBody represents the body of the hvKVPMsg
type hvKVPBody interface {
	hvKVPMsgGetBody | hvKVPMsgSetBody | hvKVPMsgDeleteBody | hvKVPMsgEnumerateBody | hvKVPIPAddrValueBody | hvKVPRegisterBody
}

// hvKVPMsgGet represents the
type hvKVPMsgGetBody struct {
	Data hvKVPExchgMsgValue
}

// hvKVPMsgSet represents the
type hvKVPMsgSetBody struct {
	Data hvKVPExchgMsgValue
}

// hvKVPMsgDelete represents the
type hvKVPMsgDeleteBody struct {
	KeySize uint32
	Key     [HyperVKPVExchangeMaxKeySize]uint8
}

// hvKVPRegister represents the
type hvKVPRegisterBody struct {
	Version [HyperVKPVExchangeMaxKeySize]uint8
}

// hvKVPMsgEnumerate represents the
type hvKVPMsgEnumerateBody struct {
	Index uint32
	Data  hvKVPExchgMsgValue
}

// hvKVPIPAddrValue represents the
type hvKVPIPAddrValueBody struct {
	AdapterID   [NetMaxAdapterIDSize]uint16
	AddrFamily  uint8
	DHCPEnabled uint8
	IPAddr      [NetMaxIPAddrSize]uint16
	Subnet      [NetMaxIPAddrSize]uint16
	Gateway     [NetMaxGatewaySize]uint16
	DNSAddr     [NetMaxIPAddrSize]uint16
}

// hvKVPExchgMsgValue represents the
type hvKVPExchgMsgValue struct {
	ValueType uint32
	KeySize   uint32
	ValueSize uint32
	Key       [HyperVKPVExchangeMaxKeySize]uint8
	Value     [HyperVKPVExchangeMaxValueSize]uint8
}

func (h *hypervKVPImpl) RunDaemon(ctx context.Context) error {
	isHandShaking := false

	err := h.getOSInfo()
	if err != nil {
		return err
	}

	_, err = h.getDomainName()
	if err != nil {
		return err
	}

	// err = h.InitFile()
	// if err != nil {
	// 	return err
	// }

	err = h.RegisterWithKernel()
	if err != nil {
		return err
	}
	defer h.UnregisterWithKernel()

	pfd := syscall.PollFd{}
	pfd.Fd = int32(h.fd)

Loop:
	for {
		klog.Info("Making poll call")
		pfd.Events = syscall.POLLIN
		pfd.Revents = 0

		n, err := syscall.Poll([]syscall.PollFd{pfd}, 10000)
		if err != nil || n < 0 {
			klog.Errorf("poll failed %d: %v", n, err)
			if err == syscall.EINVAL {
				return err
			}

			continue
		}

		klog.Info("Making read call")

		size := int(unsafe.Sizeof(hvKVPMsg{}))
		hvMsgBytes := make([]byte, size)
		len, err := syscall.Read(h.fd, hvMsgBytes)
		if err != nil {
			return fmt.Errorf("read failed with length %d: %v", len, err)
		}

		klog.Infof("hvMsgBytes length: %d %d", len, size)

		hvMsg, err := util.DeserializeData[hvKVPMsg](hvMsgBytes)
		if err != nil {
			return fmt.Errorf("failed to deserialize kvp message: %v", err)
		}

		klog.Infof("hvMsg hdr: %v", hvMsg.Header)

		switch hvMsg.Header.Operation {
		case HyperVKVPOpRgister1:
			if !isHandShaking {
				break
			}

			isHandShaking = false
			body, err := deserializeBody[hvKVPRegisterBody](hvMsg.Body)
			if err != nil {
				return fmt.Errorf("failed to deserialize kvp register body: %v", err)
			}

			licVersion := string(body.Version[:])
			if licVersion != "" {
				klog.Infof("hvMsg licVersion: %s", licVersion)
				h.licVersion = licVersion
			}

			continue
		case HyperVKVPOpGetIPInfo:
			body, err := deserializeBody[hvKVPIPAddrValueBody](hvMsg.Body)
			if err != nil {
				return fmt.Errorf("failed to deserialize kvp IP address value body: %v", err)
			}

			err = h.macToIp(body)
			if err != nil {
				return fmt.Errorf("failed to get IP: %v", err)
			}
		case HyperVKVPOpSet:
			body, err := deserializeBody[hvKVPMsgSetBody](hvMsg.Body)
			if err != nil {
				return fmt.Errorf("failed to deserialize kvp IP address value body: %v", err)
			}

			err = h.addOrUpdateKey(int(hvMsg.Header.Pool), body)
			if err != nil {
				return fmt.Errorf("failed to add or update key: %v", err)
			}
		case HyperVKVPOpSetIPInfo:
		case HyperVKVPOpDelete:
		case HyperVKVPOpEnumerate:
			continue
		default:
			break Loop
		}

		klog.Infof("records: %v", h.fileInfos)
	}

	return nil
}

func (h *hypervKVPImpl) RegisterWithKernel() error {
	// Create a new hvKVPMsg
	hvMsg := new(hvKVPMsg)

	// Close previous file descriptor if open
	if h.fd != 0 {
		syscall.Close(h.fd)
	}

	// Open the device file
	fd, err := syscall.Open(HyperVKVPFileDescriptor, syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", HyperVKVPFileDescriptor, err)
	}
	h.fd = fd

	// Register the application with the kernel
	hvMsg.Header.Operation = HyperVKVPOpRgister1
	msgBytes, err := util.SerializeData(hvMsg)
	if err != nil {
		h.fd = 0
		defer syscall.Close(fd)
		return fmt.Errorf("failed to serialize kvp message: %v", err)
	}

	// Write the message to the file descriptor
	n, err := syscall.Write(h.fd, msgBytes)
	if err != nil || n != len(msgBytes) {
		h.fd = 0
		defer syscall.Close(fd)
		return fmt.Errorf("registration to kernel failed: %v", err)
	}

	return nil
}

func (h *hypervKVPImpl) UnregisterWithKernel() error {
	if h.fd == 0 {
		return nil
	}

	return syscall.Close(h.fd)
}

func (h *hypervKVPImpl) macToIp(ipVal *hvKVPIPAddrValueBody) error {
	mac := string(utf16.Decode(ipVal.AdapterID[:]))
	dirEntries, err := os.ReadDir(HyperVKVPNetDir)
	if err != nil {
		return errors.New("failed to open network directory")
	}

	for _, entry := range dirEntries {
		ifName := entry.Name()
		devID := filepath.Join(HyperVKVPNetDir, ifName, "address")

		m, err := util.GetFileFirstLine(devID)
		if err != nil {
			klog.Errorf("failed to read MAC address from %s: %v", devID, err)
			continue
		}

		m = strings.ToUpper(m)
		if m != mac {
			continue
		}

		err = h.getIPInfo(addressfamily.AddressFamilyNone, ifName, ipVal)
		if err != nil {
			klog.Errorf("failed to read IP info from %s: %v", ifName, err)
			continue
		}
	}

	return nil
}

func (h *hypervKVPImpl) addOrUpdateKey(pool int, body *hvKVPMsgSetBody) error {
	keySize := body.Data.KeySize
	valueSize := body.Data.ValueSize

	if keySize > HyperVKPVExchangeMaxKeySize || valueSize > HyperVKPVExchangeMaxValueSize {
		return errors.New("key or value size exceeds maximum size")
	}

	err := h.updateMemState(pool, true)
	if err != nil {
		return fmt.Errorf("failed to update memory state: %v", err)
	}

	bodyKey := string(body.Data.Key[:])
	bodyVal := string(body.Data.Value[:])
	h.fileInfos[pool].records[bodyKey] = bodyVal

	return nil
}

func (h *hypervKVPImpl) getIPInfo(family addressfamily.AddressFamily, ifName string, ipVal *hvKVPIPAddrValueBody) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return errors.New("failed to get network interfaces")
	}

	var netIf net.Interface
	for _, iface := range interfaces {
		if ifName == iface.Name {
			netIf = iface
		}
	}

	// Skip loopback interfaces
	if netIf.Flags&net.FlagLoopback != 0 {
		return errors.New("loopback interface")
	}

	addrs, err := netIf.Addrs()
	if err != nil {
		return fmt.Errorf("failed to get addresses for interface %s: %v", ifName, err)
	}

	for _, addr := range addrs {
		var ip net.IP
		var subnetMask string

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
			subnetMask = v.Mask.String()
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}

		ipv4 := ip.To4()

		// Filter address family
		if family == addressfamily.AddressFamilyIPv4 && ipv4 == nil {
			continue
		}

		// Filter address family
		if family == addressfamily.AddressFamilyIPv6 && ipv4 != nil {
			continue
		}

		// Determine address family
		if ipv4 != nil {
			ipVal.AddrFamily |= uint8(addressfamily.AddressFamilyIPv4)
		} else {
			ipVal.AddrFamily |= uint8(addressfamily.AddressFamilyIPv6)
		}

		// Append IP address
		var buffer strings.Builder
		if buffer.Len() > 0 {
			buffer.WriteString(";")
		}
		buffer.WriteString(ip.String())

		// Collect subnet mask in CIDR format
		if subnetMask != "" && len(ipVal.Subnet) == 0 {
			copy(ipVal.Subnet[:], utf16.Encode([]rune(subnetMask)))
		}
	}

	return nil
}

// deserializeBody is helper function to deserialize KVP message body
func deserializeBody[T hvKVPBody](body [HyperVKVPMsgBodySize]byte) (*T, error) {
	size := int(unsafe.Sizeof(*new(T)))
	return util.DeserializeData[T](body[:size])
}
