package hypervwinrmimpl

import (
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/winrm"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/winrm/winrmimpl"
	"k8s.io/klog/v2"
)

type hypervClientImpl struct {
	winrmClient winrm.WinRMClient
}

func NewClient(opts *options.Options) (hyperv.HyperVClient, error) {
	c := newHyperVConfig(opts)

	klog.V(4).InfoS("HyperV HypervWinRmClient configured for HyperV API operations", "args", util.SanitizeRequest(c))

	winrmClient, err := winrmimpl.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return &hypervClientImpl{
		winrmClient: winrmClient,
	}, nil
}
