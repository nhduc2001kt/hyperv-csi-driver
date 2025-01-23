package driver

import (
	"context"
	"fmt"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/cloud"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/mounter"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/util/types/mode"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)


type Driver struct {
	controller *ControllerService
	node       *NodeService
	srv        *grpc.Server
	options    *options.Options
	csi.UnimplementedIdentityServer
}

func NewDriver(c cloud.Cloud, o *options.Options, m mounter.Mounter, k kubernetes.Interface) (*Driver, error) {
	klog.InfoS("Driver Information", "Driver", DriverName, "Version", driverVersion)

	// if err := ValidateDriverOptions(o); err != nil {
	// 	return nil, fmt.Errorf("invalid driver options: %w", err)
	// }
	// TODO: validate options

	driver := &Driver{
		options: o,
	}

	switch o.Mode {
	case mode.ControllerMode:
		driver.controller = NewControllerService(c, o)
	case mode.NodeMode:
		driver.node = NewNodeService(o, m, k)
	case mode.AllMode:
		driver.controller = NewControllerService(c, o)
		driver.node = NewNodeService(o, m, k)
	default:
		return nil, fmt.Errorf("unknown mode: %s", o.Mode)
	}

	return driver, nil
}

func (d *Driver) Run() error {
	scheme, addr, err := util.ParseEndpoint(d.options.Endpoint, d.options.WindowsHostProcess)
	if err != nil {
		return err
	}

	listener, err := net.Listen(scheme, addr)
	if err != nil {
		return err
	}

	logErr := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			klog.ErrorS(err, "GRPC error")
		}
		return resp, err
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logErr),
	}

	// if d.options.EnableOtelTracing {
	// 	opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	// }

	d.srv = grpc.NewServer(opts...)
	csi.RegisterIdentityServer(d.srv, d)

	switch d.options.Mode {
	case mode.ControllerMode:
		csi.RegisterControllerServer(d.srv, d.controller)
		// rpc.RegisterModifyServer(d.srv, d.controller)
	case mode.NodeMode:
		csi.RegisterNodeServer(d.srv, d.node)
	case mode.AllMode:
		csi.RegisterControllerServer(d.srv, d.controller)
		csi.RegisterNodeServer(d.srv, d.node)
		// rpc.RegisterModifyServer(d.srv, d.controller)
	default:
		return fmt.Errorf("unknown mode: %s", d.options.Mode)
	}

	klog.V(4).InfoS("Listening for connections", "address", listener.Addr())
	return d.srv.Serve(listener)
}

func (d *Driver) Stop() {
	d.srv.Stop()
}
