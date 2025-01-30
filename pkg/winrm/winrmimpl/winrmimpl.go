package winrmimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dylanmei/iso8601"
	pool "github.com/jolestar/go-commons-pool/v2"
	"k8s.io/klog/v2"

	"github.com/masterzen/winrm"
	"github.com/nhduc2001kt/hyperv-csi-driver/options"
	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/powershell"
	iwinrm "github.com/nhduc2001kt/hyperv-csi-driver/pkg/winrm"
)

func NewClient(opts *options.Options) (iwinrm.WinRMClient, error) {
	ctx := context.Background()
	config := newWinRMConfig(opts)
	factory := pool.NewPooledObjectFactorySimple(
		func(context.Context) (interface{}, error) {
			winrmClient, err := newWinRMClient(&config)

			if err != nil {
				return nil, err
			}

			return winrmClient, nil
		},
	)

	winRmClientPool := pool.NewObjectPoolWithDefaultConfig(ctx, factory)
	winRmClientPool.Config.BlockWhenExhausted = true
	winRmClientPool.Config.MinIdle = 0
	winRmClientPool.Config.MaxIdle = 2
	winRmClientPool.Config.MaxTotal = 5
	winRmClientPool.Config.TimeBetweenEvictionRuns = 10 * time.Second

	return &winrmClient{
		winRmClientPool:  winRmClientPool,
		vars:             "",
		elevatedUser:     config.user,
		elevatedPassword: config.password,
	}, nil
}

// newWinrmClient creates a new communicator implementation over WinRM.
func newWinRMClient(config *winrmConfig) (winrmClient *winrm.Client, err error) {
	addr := fmt.Sprintf("%s:%d", config.host, config.port)
	endpoint, err := parseEndpoint(addr, config.https, config.insecure, config.tlsServerName, config.caCert, config.cert, config.key, config.timeout)
	if err != nil {
		return nil, err
	}

	params := winrm.DefaultParameters

	if config.krbRealm != "" {
		proto := "http"
		if config.https {
			proto = "https"
		}

		params.TransportDecorator = func() winrm.Transporter {
			return &winrm.ClientKerberos{
				Username:  config.user,
				Password:  config.password,
				Hostname:  config.host,
				Port:      config.port,
				Proto:     proto,
				Realm:     config.krbRealm,
				SPN:       config.krbSpn,
				KrbConf:   config.krbConfig,
				KrbCCache: config.krbCCache,
			}
		}
	} else if config.ntlm {
		params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }
	}

	if endpoint.Timeout.Seconds() > 0 {
		params.Timeout = iso8601.FormatDuration(endpoint.Timeout)
	}

	winrmClient, err = winrm.NewClientWithParameters(
		endpoint, config.user, config.password, params)

	if err != nil {
		return nil, err
	}

	return winrmClient, nil
}

func parseEndpoint(addr string, https bool, insecure bool, tlsServerName string, caCert []byte, cert []byte, key []byte, timeout string) (*winrm.Endpoint, error) {
	var host string
	var port int

	if addr == "" {
		return nil, fmt.Errorf("couldn't convert \"\" to an address")
	}
	if !strings.Contains(addr, ":") || (strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]")) {
		host = addr
		port = 5985
	} else {
		shost, sport, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("couldn't convert \"%s\" to an address", addr)
		}
		// Check for IPv6 addresses and reformat appropriately
		host = ipFormat(shost)
		port, err = strconv.Atoi(sport)
		if err != nil {
			return nil, fmt.Errorf("couldn't convert \"%s\" to a port number", sport)
		}
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert \"%s\" to a duration", timeout)
	}

	return &winrm.Endpoint{
		Host:          host,
		Port:          port,
		HTTPS:         https,
		Insecure:      insecure,
		TLSServerName: tlsServerName,
		Cert:          cert,
		Key:           key,
		CACert:        caCert,
		Timeout:       timeoutDuration,
	}, nil
}

// ipFormat formats the IP correctly, so we don't provide IPv6 address in an IPv4 format during node communication.
// We return the ip parameter as is if it's an IPv4 address or a hostname.
func ipFormat(ip string) string {
	ipObj := net.ParseIP(ip)
	// Return the ip/host as is if it's either a hostname or an IPv4 address.
	if ipObj == nil || ipObj.To4() != nil {
		return ip
	}

	return fmt.Sprintf("[%s]", ip)
}

type winrmClient struct {
	winRmClientPool  *pool.ObjectPool
	elevatedUser     string
	elevatedPassword string
	vars             string
}

func (c *winrmClient) RunFireAndForgetScript(ctx context.Context, script *template.Template, args interface{}) error {
	var scriptRendered bytes.Buffer

	klog.V(4).InfoS("Execute: called")
	err := script.Execute(&scriptRendered, args)
	if err != nil {
		return err
	}
	command := scriptRendered.String()

	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)
	if err != nil {
		return err
	}

	klog.V(4).InfoS("RunPowershell: called")
	_, _, _, err = powershell.RunPowershell(winrmClient.(*winrm.Client), c.elevatedUser, c.elevatedPassword, c.vars, command)
	klog.V(4).InfoS("ReturnObject: called")
	errRet := c.winRmClientPool.ReturnObject(ctx, winrmClient)
	if err != nil {
		return err
	}
	if errRet != nil {
		return errRet
	}

	return nil
}

func (c *winrmClient) RunScriptWithResult(ctx context.Context, script *template.Template, args interface{}, result interface{}) (err error) {
	var scriptRendered bytes.Buffer
	err = script.Execute(&scriptRendered, args)

	if err != nil {
		return err
	}

	command := scriptRendered.String()

	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return err
	}

	exitStatus, stdout, stderr, err := powershell.RunPowershell(winrmClient.(*winrm.Client), c.elevatedUser, c.elevatedPassword, c.vars, command)

	err2 := c.winRmClientPool.ReturnObject(ctx, winrmClient)

	if err != nil {
		return err
	}

	if err2 != nil {
		return err2
	}

	stdout = strings.TrimSpace(stdout)

	err = json.Unmarshal([]byte(stdout), &result)
	if err != nil {
		return fmt.Errorf("exitStatus:%d\nstdOut:%s\nstdErr:%s\nerr:%s\ncommand:%s", exitStatus, stdout, stderr, err, command)
	}

	return nil
}

func (c *winrmClient) UploadFile(ctx context.Context, filePath string, remoteFilePath string) (string, error) {
	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return "", err
	}

	remoteFilePath, err = powershell.UploadFile(winrmClient.(*winrm.Client), filePath, remoteFilePath)
	errRet := c.winRmClientPool.ReturnObject(ctx, winrmClient)
	if err != nil {
		return "", err
	}
	if errRet != nil {
		return "", errRet
	}

	return remoteFilePath, nil
}

func (c *winrmClient) UploadDirectory(ctx context.Context, rootPath string, excludeList []string) (remoteRootPath string, remoteAbsoluteFilePaths []string, err error) {
	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return "", []string{}, err
	}

	remoteRootPath, remoteAbsoluteFilePaths, err = powershell.UploadDirectory(winrmClient.(*winrm.Client), rootPath, excludeList)

	err2 := c.winRmClientPool.ReturnObject(ctx, winrmClient)

	if err != nil {
		return "", []string{}, err
	}

	if err2 != nil {
		return "", []string{}, err2
	}

	return remoteRootPath, remoteAbsoluteFilePaths, nil
}

func (c *winrmClient) FileExists(ctx context.Context, remoteFilePath string) (exists bool, err error) {
	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return false, err
	}

	result, err := powershell.FileExists(winrmClient.(*winrm.Client), remoteFilePath)
	errRet := c.winRmClientPool.ReturnObject(ctx, winrmClient)
	if err != nil {
		return false, err
	}
	if errRet != nil {
		return false, errRet
	}

	return result, nil
}

func (c *winrmClient) DirectoryExists(ctx context.Context, remoteDirectoryPath string) (exists bool, err error) {
	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return false, err
	}

	result, err := powershell.DirectoryExists(winrmClient.(*winrm.Client), remoteDirectoryPath)
	errRet := c.winRmClientPool.ReturnObject(ctx, winrmClient)
	if err != nil {
		return false, err
	}
	if errRet != nil {
		return false, errRet
	}

	return result, nil
}

func (c *winrmClient) DeleteFileOrDirectory(ctx context.Context, remotePath string) (err error) {
	winrmClient, err := c.winRmClientPool.BorrowObject(ctx)

	if err != nil {
		return err
	}

	err = powershell.DeleteFileOrDirectory(winrmClient.(*winrm.Client), remotePath)
	errRet := c.winRmClientPool.ReturnObject(ctx, winrmClient)
	if err != nil {
		return err
	}
	if errRet != nil {
		return errRet
	}

	return nil
}
