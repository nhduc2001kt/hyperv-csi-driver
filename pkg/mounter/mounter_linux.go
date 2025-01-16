//go:build linux
// +build linux

package mounter

import (
	"errors"

	mountutils "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

func NewSafeMounter() (*mountutils.SafeFormatAndMount, error) {
	return &mountutils.SafeFormatAndMount{
		Interface: mountutils.New(""),
		Exec:      utilexec.New(),
	}, nil
}

func NewSafeMounterV2() (*mountutils.SafeFormatAndMount, error) {
	return nil, errors.New("NewSafeMounterV2 is not supported on this platform")
}