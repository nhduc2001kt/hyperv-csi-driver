package driver

// constants for default command line flag values.
const (
	DefaultCSIEndpoint = "unix://tmp/csi.sock"
)

const (
	DriverName               = "hyperv.csi.aws.com"
)

// constants for node k8s API use.
const (
	// AgentNotReadyNodeTaintKey contains the key of taints to be removed on driver startup.
	AgentNotReadyNodeTaintKey = "hyperv.csi.aws.com/agent-not-ready"
)