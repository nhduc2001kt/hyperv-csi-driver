package hyperv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	// "log"
	// "path/filepath"
	"strconv"
	"strings"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ControllerType int

const (
	ControllerTypeIDE  ControllerType = 0
	ControllerTypeSCSI ControllerType = 1
)

const (
	ControllerTypeIDEText  = "Ide"
	ControllerTypeSCSIText = "Scsi"
)

var ControllerTypename = map[ControllerType]string{
	ControllerTypeIDE:  ControllerTypeIDEText,
	ControllerTypeSCSI: ControllerTypeSCSIText,
}

var ControllerTypeValue = map[string]ControllerType{
	strings.ToLower(ControllerTypeIDEText):  ControllerTypeIDE,
	strings.ToLower(ControllerTypeSCSIText): ControllerTypeSCSI,
}

var (
	ErrInvalidControllerType = fmt.Errorf("the provided ControllerType string is not a valid type")
)

func (x ControllerType) String() string {
	return ControllerTypename[x]
}

func StringToControllerType(x string) (ControllerType, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return ControllerType(integerValue), nil
	}

	if value, exist := ControllerTypeValue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return ControllerType(-1), ErrInvalidControllerType
}

func (d *ControllerType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(d.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *ControllerType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = ControllerType(i)
			return nil
		}

		return err
	}
	*d, err = StringToControllerType(s)
	return err
}

type CacheAttributes int

const (
	CacheAttributesDefault                 CacheAttributes = 0
	CacheAttributesWriteCacheEnabled       CacheAttributes = 1
	CacheAttributesWriteCacheAndFUAEnabled CacheAttributes = 2
	CacheAttributesWriteCacheDisabled      CacheAttributes = 3
)

var CacheAttributesName = map[CacheAttributes]string{
	CacheAttributesDefault:                 "Default",
	CacheAttributesWriteCacheEnabled:       "WriteCacheEnabled",
	CacheAttributesWriteCacheAndFUAEnabled: "WriteCacheAndFUAEnabled",
	CacheAttributesWriteCacheDisabled:      "WriteCacheDisabled",
}

var CacheAttributesValue = map[string]CacheAttributes{
	"default":                 CacheAttributesDefault,
	"writecacheenabled":       CacheAttributesWriteCacheEnabled,
	"writecacheandfuaenabled": CacheAttributesWriteCacheAndFUAEnabled,
	"writecachedisabled":      CacheAttributesWriteCacheDisabled,
}

func (x CacheAttributes) String() string {
	return CacheAttributesName[x]
}

func ToCacheAttributes(x string) CacheAttributes {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return CacheAttributes(integerValue)
	}
	return CacheAttributesValue[strings.ToLower(x)]
}

func (d *CacheAttributes) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(d.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *CacheAttributes) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = CacheAttributes(i)
			return nil
		}

		return err
	}
	*d = ToCacheAttributes(s)
	return nil
}

// func DiffSuppressVMHardDiskPath(key, old, new string, d *schema.ResourceData) bool {
// 	if new == "" {
// 		// We have not explicitly set a value, so allow any value as we are not tracking it
// 		return true
// 	}

// 	if new == old {
// 		return true
// 	}

// 	// Ignore snapshots otherwise it will change from "c:\\vhdx\\web_server_g2_B63C9D15-F9A3-4F63-A896-FFD80BC7754C.avhdx" -> "c:\\vhdx\\web_server_g2.vhdx"
// 	oldExtension := strings.ToLower(filepath.Ext(old))
// 	newExtension := strings.ToLower(filepath.Ext(new))
// 	if oldExtension == ".avhdx" && newExtension == ".vhdx" {
// 		newName := new[0 : len(new)-len(newExtension)]
// 		return strings.HasPrefix(old, newName+"_")
// 	}

// 	return false
// }

// func ExpandHardDiskDrives(d *schema.ResourceData) ([]VMHardDiskDrive, error) {
// 	expandedHardDiskDrives := make([]VMHardDiskDrive, 0)

// 	if v, ok := d.GetOk("hard_disk_drives"); ok {
// 		hardDiskDrives := v.([]interface{})

// 		for _, hardDiskDrive := range hardDiskDrives {
// 			hardDiskDrive, ok := hardDiskDrive.(map[string]interface{})
// 			if !ok {
// 				return nil, fmt.Errorf("[ERROR][hyperv] hard_disk_drives should be a Hash - was '%+v'", hardDiskDrive)
// 			}

// 			controllerType, err := StringToControllerType(hardDiskDrive["controller_type"].(string))
// 			if err != nil {
// 				return nil, fmt.Errorf("[ERROR][hyperv] controller_type should be a valid ControllerType - was '%+v'", hardDiskDrive["controller_type"])
// 			}

// 			expandedHardDiskDrive := VMHardDiskDrive{
// 				ControllerType:                controllerType,
// 				ControllerNumber:              int32(hardDiskDrive["controller_number"].(int)),
// 				ControllerLocation:            int32(hardDiskDrive["controller_location"].(int)),
// 				Path:                          hardDiskDrive["path"].(string),
// 				DiskNumber:                    uint32(hardDiskDrive["disk_number"].(int)),
// 				ResourcePoolName:              hardDiskDrive["resource_pool_name"].(string),
// 				SupportPersistentReservations: hardDiskDrive["support_persistent_reservations"].(bool),
// 				MaximumIops:                   uint64(hardDiskDrive["maximum_iops"].(int)),
// 				MinimumIops:                   uint64(hardDiskDrive["minimum_iops"].(int)),
// 				QosPolicyId:                   hardDiskDrive["qos_policy_id"].(string),
// 				OverrideCacheAttributes:       ToCacheAttributes(hardDiskDrive["override_cache_attributes"].(string)),
// 			}

// 			expandedHardDiskDrives = append(expandedHardDiskDrives, expandedHardDiskDrive)
// 		}
// 	}

// 	return expandedHardDiskDrives, nil
// }

func FlattenHardDiskDrives(hardDiskDrives *[]VMHardDiskDrive) []interface{} {
	if hardDiskDrives == nil || len(*hardDiskDrives) < 1 {
		return nil
	}

	flattenedHardDiskDrives := make([]interface{}, 0)

	for _, hardDiskDrive := range *hardDiskDrives {
		flattenedHardDiskDrive := make(map[string]interface{})
		flattenedHardDiskDrive["controller_type"] = hardDiskDrive.ControllerType.String()
		flattenedHardDiskDrive["controller_number"] = hardDiskDrive.ControllerNumber
		flattenedHardDiskDrive["controller_location"] = hardDiskDrive.ControllerLocation
		flattenedHardDiskDrive["path"] = hardDiskDrive.Path
		flattenedHardDiskDrive["disk_number"] = hardDiskDrive.DiskNumber
		flattenedHardDiskDrive["resource_pool_name"] = hardDiskDrive.ResourcePoolName
		flattenedHardDiskDrive["support_persistent_reservations"] = hardDiskDrive.SupportPersistentReservations
		flattenedHardDiskDrive["maximum_iops"] = hardDiskDrive.MaximumIops
		flattenedHardDiskDrive["minimum_iops"] = hardDiskDrive.MinimumIops
		flattenedHardDiskDrive["qos_policy_id"] = hardDiskDrive.QosPolicyId
		flattenedHardDiskDrive["override_cache_attributes"] = hardDiskDrive.OverrideCacheAttributes.String()
		flattenedHardDiskDrives = append(flattenedHardDiskDrives, flattenedHardDiskDrive)
	}

	return flattenedHardDiskDrives
}

type VMHardDiskDrive struct {
	VMName                        string
	ControllerType                ControllerType
	ControllerNumber              int32
	ControllerLocation            int32
	Path                          string
	DiskNumber                    uint32
	ResourcePoolName              string
	SupportPersistentReservations bool
	MaximumIops                   uint64
	MinimumIops                   uint64
	QosPolicyId                   string
	OverrideCacheAttributes       CacheAttributes
	// AllowUnverifiedPaths          bool no way of checking if its turned on so always turn on
}

type HyperVVMHardDiskDriveClient interface {
	AttachVMHardDiskDrive(
		ctx context.Context,
		vmName string,
		controllerType ControllerType,
		path string,
	) (err error)
	CreateVMHardDiskDrive(
		ctx context.Context,
		vmName string,
		controllerType ControllerType,
		controllerNumber int32,
		controllerLocation int32,
		path string,
		diskNumber uint32,
		resourcePoolName string,
		supportPersistentReservations bool,
		maximumIops uint64,
		minimumIops uint64,
		qosPolicyId string,
		overrideCacheAttributes CacheAttributes,
	) (err error)
	GetVMHardDiskDrives(ctx context.Context, vmName string) (result []VMHardDiskDrive, err error)
	GetVMHardDiskDrivesByID(ctx context.Context, vmID string) (result []VMHardDiskDrive, err error)
	UpdateVMHardDiskDrive(
		ctx context.Context,
		vmName string,
		controllerNumber int32,
		controllerLocation int32,
		controllerType ControllerType,
		toControllerNumber int32,
		toControllerLocation int32,
		path string,
		diskNumber uint32,
		resourcePoolName string,
		supportPersistentReservations bool,
		maximumIops uint64,
		minimumIops uint64,
		qosPolicyId string,
		overrideCacheAttributes CacheAttributes,
	) (err error)
	DeleteVMHardDiskDrive(ctx context.Context, vmName string, controllerNumber int32, controllerLocation int32) (err error)
	CreateOrUpdateVMHardDiskDrives(ctx context.Context, vmName string, hardDiskDrives []VMHardDiskDrive) (err error)
}
