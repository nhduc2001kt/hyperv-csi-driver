package mounter

import (
	mountutils "k8s.io/mount-utils"
)

// NodeMounter implements Mounter.
// A superstruct of SafeFormatAndMount.
type Mounter interface {
	mountutils.Interface
}

// NodeMounter implements Mounter.
// A superstruct of SafeFormatAndMount.
type NodeMounter struct {
	*mountutils.SafeFormatAndMount
}

// NewNodeMounter returns a new intsance of NodeMounter.
func NewNodeMounter(hostprocess bool) (Mounter, error) {
	var safeMounter *mountutils.SafeFormatAndMount
	var err error

	if hostprocess {
		safeMounter, err = NewSafeMounterV2()
	} else {
		safeMounter, err = NewSafeMounter()
	}

	if err != nil {
		return nil, err
	}
	return &NodeMounter{safeMounter}, nil
}

