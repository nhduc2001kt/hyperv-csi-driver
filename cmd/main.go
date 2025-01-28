package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/cloud"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/cloud/metadata"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/driver"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hvkvp/hvkvpimpl"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/mounter"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util/types/mode"
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
		cmd     = string(mode.AllMode)
		options = options.Options{}
	)

	c := logsapi.NewLoggingConfiguration()
	err := logsapi.AddFeatureGates(featureGate)
	if err != nil {
		klog.ErrorS(err, "failed to add feature gates")
	}
	logsapi.AddFlags(c, fs)

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		cmd = os.Args[1]
		args = os.Args[2:]
	}

	switch cmd {
	case "pre-stop-hook":
	case "hv-kvp-daemon":
		hvKVP := hvkvpimpl.NewHyperVKVP()
		err := hvKVP.RunDaemon(context.Background())
		if err != nil {
			klog.ErrorS(err, "failed to get domain name")
		}

		klog.FlushAndExit(klog.ExitFlushTimeout, 0)
	default:
		options.Mode, err = mode.StringToMode(cmd)
		if err != nil {
			klog.ErrorS(err, "Failed to parse mode")
			klog.FlushAndExit(klog.ExitFlushTimeout, 0)
		}
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
		K8sAPIClient: metadata.DefaultKubernetesAPIClient(options.Kubeconfig),
	}

	cloud, err := cloud.NewCloud(&options)
	if err != nil {
		klog.ErrorS(err, "failed to create cloud service")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	m, err := mounter.NewNodeMounter(options.WindowsHostProcess)
	if err != nil {
		klog.ErrorS(err, "failed to create node mounter")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	k8sClient, err := cfg.K8sAPIClient()
	if err != nil {
		klog.V(2).InfoS("Failed to setup k8s client", "err", err)
	}

	drv, err := driver.NewDriver(cloud, &options, m, k8sClient)
	if err != nil {
		klog.ErrorS(err, "failed to create driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	if err := drv.Run(); err != nil {
		klog.ErrorS(err, "failed to run driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
