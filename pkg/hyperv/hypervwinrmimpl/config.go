package hypervwinrmimpl

import (
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
)

type hyperVConfig struct {
	user     string
	password string
	host     string
	port     int
	https    bool
	insecure bool

	// krbRealm  string
	// krbSpn    string
	// krbConfig string
	// krbCCache string

	// ntlm bool

	// tlsServerName string
	// caCert        []byte
	// cert          []byte
	// key           []byte

	// scriptPath string
	timeout    string
}

func newHyperVConfig(opts *options.Options) hyperVConfig {
	cfg := hyperVConfig{
		user:     opts.WinRMUser,
		password: opts.WinRMPassword,
		host:     opts.WinRMHost,
		port:     opts.WinRMPort,
		insecure: opts.WinRMAllowInsecure,
		https:    opts.WinRMUseHTTPS,
		timeout:  opts.WinRMTimeout,
	}

	return cfg
}
