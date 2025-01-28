package hyperv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type VHDType int

const (
	VHDTypeUnknown      VHDType = 0
	VHDTypeFixed        VHDType = 2
	VHDTypeDynamic      VHDType = 3
	VHDTypeDifferencing VHDType = 4
)

const (
	VHDTypeUnknownText      = "Unknown"
	VHDTypeFixedText        = "Fixed"
	VHDTypeDynamicText      = "Dynamic"
	VHDTypeDifferencingText = "Differencing"
)

var VHDTypename = map[VHDType]string{
	VHDTypeUnknown:      VHDTypeUnknownText,
	VHDTypeFixed:        VHDTypeFixedText,
	VHDTypeDynamic:      VHDTypeDynamicText,
	VHDTypeDifferencing: VHDTypeDifferencingText,
}

var VHDTypevalue = map[string]VHDType{
	strings.ToLower(VHDTypeUnknownText):      VHDTypeUnknown,
	strings.ToLower(VHDTypeFixedText):        VHDTypeFixed,
	strings.ToLower(VHDTypeDynamicText):      VHDTypeDynamic,
	strings.ToLower(VHDTypeDifferencingText): VHDTypeDifferencing,
}

var (
	// ErrInvalidVHDType
	ErrInvalidVHDType = errors.New("the provided VHD type string is not a valid type")
	// ErrInvalidVHDFormat
	ErrInvalidVHDFormat = errors.New("the provided VHD format string is not a valid format")
)

func (x VHDType) String() string {
	return VHDTypename[x]
}

func StringToVHDType(x string) (VHDType, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return VHDType(integerValue), nil
	}

	if value, exist := VHDTypevalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return VHDTypeUnknown, ErrInvalidVHDType
}

func (d *VHDType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(d.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *VHDType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = VHDType(i)
			return nil
		}

		return err
	}
	*d, err = StringToVHDType(s)
	return err
}

type VHDFormat int

const (
	VHDFormatUnknown VHDFormat = 0
	VHDFormatVHD     VHDFormat = 2 // extension ".vhd"
	VHDFormatVHDX    VHDFormat = 3 // extension ".vhdx"
	VHDFormatVHDSet  VHDFormat = 4 // extension ".vhds"
)

const (
	VHDFormatUnknownText = "Unknown"
	VHDFormatVHDText     = "VHD"
	VHDFormatVHDXText    = "VHDX"
	VHDFormatVHDSetText  = "VHDSet"
)

var VHDFormatName = map[VHDFormat]string{
	VHDFormatUnknown: VHDFormatUnknownText,
	VHDFormatVHD:     VHDFormatVHDText,
	VHDFormatVHDX:    VHDFormatVHDXText,
	VHDFormatVHDSet:  VHDFormatVHDSetText,
}

var VHDFormatExtension = map[VHDFormat]string{
	VHDFormatUnknown: "",
	VHDFormatVHD:     ".vhd",
	VHDFormatVHDX:    ".vhdx",
	VHDFormatVHDSet:  ".vhds",
}

var VHDFormatValue = map[string]VHDFormat{
	strings.ToLower(VHDFormatUnknownText): VHDFormatUnknown,
	strings.ToLower(VHDFormatVHDText):     VHDFormatVHD,
	strings.ToLower(VHDFormatVHDXText):    VHDFormatVHDX,
	strings.ToLower(VHDFormatVHDSetText):  VHDFormatVHDSet,
}

func (x VHDFormat) String() string {
	return VHDFormatName[x]
}

func StringToVHDFormat(x string) (VHDFormat, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return VHDFormat(integerValue), nil
	}

	if value, exist := VHDFormatValue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return VHDFormatUnknown, ErrInvalidVHDFormat
}

func (d *VHDFormat) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(d.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *VHDFormat) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = VHDFormat(i)
			return nil
		}

		return err
	}
	*d, err = StringToVHDFormat(s)
	return err
}

type VHDExists struct {
	Exists bool
}

type VHD struct {
	Path                    string
	BlockSize               uint32
	LogicalSectorSize       uint32
	PhysicalSectorSize      uint32
	ParentPath              string
	FileSize                uint64
	Size                    uint64
	MinimumSize             uint64
	Attached                bool
	DiskNumber              int
	Number                  int
	FragmentationPercentage int
	Alignment               int
	DiskIdentifier          string
	VHDType                 VHDType
	VHDFormat               VHDFormat
}

type HyperVVHDClient interface {
	VHDExists(ctx context.Context, path string) (result VHDExists, err error)
	CreateOrUpdateVHD(ctx context.Context, path string, source string, sourceVm string, sourceDisk int, vhdType VHDType, parentPath string, size uint64, blockSize uint32, logicalSectorSize uint32, physicalSectorSize uint32) (err error)
	ResizeVHD(ctx context.Context, path string, size uint64) (err error)
	GetVHD(ctx context.Context, path string) (result VHD, err error)
	DeleteVHD(ctx context.Context, path string) (err error)
}
