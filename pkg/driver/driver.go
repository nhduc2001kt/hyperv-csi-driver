package driver

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/mounter"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Mode is the operating mode of the CSI driver.
type Mode string

const (
	// ControllerMode is the mode that only starts the controller service.
	ControllerMode Mode = "controller"
	// NodeMode is the mode that only starts the node service.
	NodeMode Mode = "node"
	// AllMode is the mode that only starts both the controller and the node service.
	AllMode Mode = "all"
)

type Driver struct {
	controller *ControllerService
	node       *NodeService
	// srv        *grpc.Server
	options    *Options
	csi.UnimplementedIdentityServer
}


func NewDriver(o *Options, m mounter.Mounter, k kubernetes.Interface) (*Driver, error) {
	klog.InfoS("Driver Information", "Driver", DriverName, "Version", driverVersion)

	// if err := ValidateDriverOptions(o); err != nil {
	// 	return nil, fmt.Errorf("invalid driver options: %w", err)
	// }

	driver := &Driver{
		options: o,
	}

	switch o.Mode {
	case ControllerMode:
		driver.controller = NewControllerService(o)
	case NodeMode:
		driver.node = NewNodeService(o, m, k)
	case AllMode:
		driver.controller = NewControllerService(o)
		driver.node = NewNodeService(o, m, k)
	default:
		return nil, fmt.Errorf("unknown mode: %s", o.Mode)
	}

	return driver, nil
}