package driver

import (
	flag "github.com/spf13/pflag"
)

type Options struct {
	Mode Mode

	// Kubeconfig is an absolute path to a kubeconfig file.
	// If empty, the in-cluster config will be loaded.
	Kubeconfig string

	// Endpoint is the endpoint for the CSI driver server
	Endpoint string

	// WindowsHostProcess indicates whether the driver is running in a Windows privileged container
	WindowsHostProcess bool
}

func (o *Options) AddFlags(f *flag.FlagSet) {
	f.StringVar(&o.Kubeconfig, "kubeconfig", "", "Absolute path to a kubeconfig file. The default is the empty string, which causes the in-cluster config to be used")
	f.StringVar(&o.Endpoint, "endpoint", DefaultCSIEndpoint, "Endpoint for the CSI driver server")

	if o.Mode == AllMode || o.Mode == NodeMode {
		f.BoolVar(&o.WindowsHostProcess, "windows-host-process", false, "ALPHA: Indicates whether the driver is running in a Windows privileged container")
	}
}

func (o *Options) Validate() error {
	return nil
}
