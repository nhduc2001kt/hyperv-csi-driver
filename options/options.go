package options

import (
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util/types/mode"
	flag "github.com/spf13/pflag"
)

// constants for default command line flag values.
const (
	DefaultCSIEndpoint        = "unix://tmp/csi.sock"
	DefaultWinRMUser          = "Administrator"
	DefaultWinRMHost          = "127.0.0.1"
	DefaultWinRMPort          = 5986
	DefaultWinRMTimeout       = "30s"
	DefaultWinRMAllowInsecure = false
)

type Options struct {
	Mode mode.Mode

	// Kubeconfig is an absolute path to a kubeconfig file.
	// If empty, the in-cluster config will be loaded.
	Kubeconfig string

	// Endpoint is the endpoint for the CSI driver server
	Endpoint string

	// KubernetesClusterID is the ID of the kubernetes cluster.
	KubernetesClusterID string

	// WinRMUser is the username for WinRM connection
	WinRMUser string

	// WinRMPassword is the password for WinRM connection
	WinRMPassword string

	// WinRMHost is the host for WinRM connection
	WinRMHost string

	// WinRMUseHTTPS indicates whether to use HTTPS for WinRM connection
	WinRMUseHTTPS bool

	// WinRMPort is the port for WinRM connection
	WinRMPort int

	// WinRMTimeout is the timeout for WinRM connection
	WinRMTimeout string

	// WinRMAllowInsecure indicates whether to allow insecure WinRM connections
	WinRMAllowInsecure bool

	// WindowsHostProcess indicates whether the driver is running in a Windows privileged container
	WindowsHostProcess bool
}

func (o *Options) AddFlags(f *flag.FlagSet) {
	f.StringVar(&o.Kubeconfig, "kubeconfig", "", "Absolute path to a kubeconfig file. The default is the empty string, which causes the in-cluster config to be used")
	f.StringVar(&o.Endpoint, "endpoint", DefaultCSIEndpoint, "Endpoint for the CSI driver server")
	f.StringVar(&o.KubernetesClusterID, "kubernetes-cluster-id", "", "ID of the kubernetes cluster")
	f.StringVar(&o.WinRMUser, "winrm-user", DefaultWinRMUser, "Username for WinRM connection")
	f.StringVar(&o.WinRMPassword, "winrm-password", "", "Password for WinRM connection")
	f.StringVar(&o.WinRMHost, "winrm-host", DefaultWinRMHost, "Host for WinRM connection")
	f.BoolVar(&o.WinRMUseHTTPS, "winrm-use-https", true, "Indicates whether to use HTTPS for WinRM connection")
	f.IntVar(&o.WinRMPort, "winrm-port", DefaultWinRMPort, "Port for WinRM connection")
	f.StringVar(&o.WinRMTimeout, "winrm-timeout", DefaultWinRMTimeout, "Timeout for WinRM connection")
	f.BoolVar(&o.WinRMAllowInsecure, "winrm-allow-insecure", DefaultWinRMAllowInsecure, "Indicates whether to allow insecure WinRM connections")

	if o.Mode == mode.AllMode || o.Mode == mode.NodeMode {
		f.BoolVar(&o.WindowsHostProcess, "windows-host-process", false, "ALPHA: Indicates whether the driver is running in a Windows privileged container")
	}
}

func (o *Options) Validate() error {
	return nil
}
