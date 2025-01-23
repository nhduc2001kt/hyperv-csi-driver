package winrmimpl

import (
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
)

type winrmConfig struct {
	user     string
	password string
	host     string
	port     int
	https    bool
	insecure bool

	krbRealm  string
	krbSpn    string
	krbConfig string
	krbCCache string

	ntlm bool

	tlsServerName string
	caCert        []byte
	cert          []byte
	key           []byte

	timeout string
}

func newWinRMConfig(opts *options.Options) winrmConfig {
	cfg := winrmConfig{
		user:     opts.WinRMUser,
		password: opts.WinRMPassword,
		host:     opts.WinRMHost,
		port:     opts.WinRMPort,
		https:    opts.WinRMUseHTTPS,
		insecure: opts.WinRMAllowInsecure,
		timeout:  opts.WinRMTimeout,
	}

	return cfg
}
