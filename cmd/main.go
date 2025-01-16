package main

import (
	"fmt"
	"os"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/cloud/metadata"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/driver"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/mounter"
	flag "github.com/spf13/pflag"
	"k8s.io/component-base/featuregate"
	logsapi "k8s.io/component-base/logs/api/v1"
	json "k8s.io/component-base/logs/json"
	"k8s.io/klog/v2"
)

var (
	featureGate = featuregate.NewFeatureGate()
)

func main() {
	fs := flag.NewFlagSet("hyperv-csi-driver", flag.ExitOnError)
	if err := logsapi.RegisterLogFormat(logsapi.JSONLogFormat, json.Factory{}, logsapi.LoggingBetaOptions); err != nil {
		klog.ErrorS(err, "failed to register JSON log format")
	}

	var (
		version = fs.Bool("version", false, "Print the version and exit.")
		args    = os.Args[1:]
		cmd     = string(driver.AllMode)
		options = driver.Options{}
	)

	c := logsapi.NewLoggingConfiguration()
	err := logsapi.AddFeatureGates(featureGate)
	if err != nil {
		klog.ErrorS(err, "failed to add feature gates")
	}
	logsapi.AddFlags(c, fs)

	switch cmd {
	case string(driver.ControllerMode), string(driver.NodeMode), string(driver.AllMode):
		options.Mode = driver.Mode(cmd)
	default:
		klog.Errorf("Unknown driver mode %s: Expected %s, %s or %s", cmd, driver.ControllerMode, driver.NodeMode, driver.AllMode)
		klog.FlushAndExit(klog.ExitFlushTimeout, 0)
	}

	options.AddFlags(fs)

	if err := fs.Parse(args); err != nil {
		klog.ErrorS(err, "Failed to parse options")
		klog.FlushAndExit(klog.ExitFlushTimeout, 0)
	}
	if err := options.Validate(); err != nil {
		klog.ErrorS(err, "Invalid options")
		klog.FlushAndExit(klog.ExitFlushTimeout, 0)
	}

	err = logsapi.ValidateAndApply(c, featureGate)
	if err != nil {
		klog.ErrorS(err, "failed to validate and apply logging configuration")
	}

	if *version {
		versionInfo, versionErr := driver.GetVersionJSON()
		if versionErr != nil {
			klog.ErrorS(err, "failed to get version")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		//nolint:forbidigo // Print version info without klog/timestamp
		fmt.Println(versionInfo)
		os.Exit(0)
	}

	cfg := metadata.MetadataServiceConfig{
		K8sAPIClient:      metadata.DefaultKubernetesAPIClient(options.Kubeconfig),
	}

	m, err := mounter.NewNodeMounter(options.WindowsHostProcess)
	if err != nil {
		klog.ErrorS(err, "failed to create node mounter")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	drv, err := driver.NewDriver(&options, m, k8sClient)
	if err != nil {
		klog.ErrorS(err, "failed to create driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	if err := drv.Run(); err != nil {
		klog.ErrorS(err, "failed to run driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
