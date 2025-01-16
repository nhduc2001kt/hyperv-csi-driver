/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

// Supported access modes.
const (
	SingleNodeWriter     = csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
	MultiNodeMultiWriter = csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER
)

var (
	// controllerCaps represents the capability of controller service.
	controllerCaps = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_MODIFY_VOLUME,
	}
)

const trueStr = "true"
const isManagedByDriver = trueStr

// ControllerService represents the controller service of CSI driver.
type ControllerService struct {
	// inFlight              *internal.InFlight
	options               *Options
	// modifyVolumeCoalescer coalescer.Coalescer[modifyVolumeRequest, int32]
	// rpc.UnimplementedModifyServer
	csi.UnimplementedControllerServer
}

// NewControllerService creates a new controller service.
func NewControllerService(o *Options) *ControllerService {
	return &ControllerService{
		options:               o,
		// inFlight:              internal.NewInFlight(),
		// modifyVolumeCoalescer: newModifyVolumeCoalescer(c, o),
	}
}
