package hvkvp

import "context"

const (
	// HyerVKVPInfoKeyTag is the tag for the Hyper-V key value pairs info key.
	HyerVKVPInfoKeyTag = "kvpkey"
)

type HyperVKVPInfo struct {
	VirtualMachineName string `kvpkey:"VirtualMachineName"`
	VirtualMachineID   string `kvpkey:"VirtualMachineId"`
	HostName           string `kvpkey:"HostName"`
}

type HyperVKVP interface {
	InitFile() error
	WaitDaemonPool(context.Context, int) error
	ReadPool(context.Context, int) (*HyperVKVPInfo, error)
	RunDaemon(context.Context) error
}
