package metadata

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"
)

type KubernetesAPIClient func() (kubernetes.Interface, error)

func DefaultKubernetesAPIClient(kubeconfig string) KubernetesAPIClient {
	return func() (clientset kubernetes.Interface, err error) {
		var config *rest.Config
		if kubeconfig != "" {
			config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
				&clientcmd.ConfigOverrides{},
			).ClientConfig()
			if err != nil {
				return nil, err
			}
		} else {
			config, err = rest.InClusterConfig()
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}

			if errors.Is(err, os.ErrNotExist) {
				klog.InfoS("InClusterConfig failed to read token file, retrieving file from sandbox mount point")
				// CONTAINER_SANDBOX_MOUNT_POINT env is set upon container creation in containerd v1.6+
				// it provides the absolute host path to the container volume.
				sandboxMountPoint := os.Getenv("CONTAINER_SANDBOX_MOUNT_POINT")
				if sandboxMountPoint == "" {
					return nil, errors.New("CONTAINER_SANDBOX_MOUNT_POINT environment variable is not set")
				}

				tokenFile := filepath.Join(sandboxMountPoint, "var", "run", "secrets", "kubernetes.io", "serviceaccount", "token")
				rootCAFile := filepath.Join(sandboxMountPoint, "var", "run", "secrets", "kubernetes.io", "serviceaccount", "ca.crt")

				token, tokenErr := os.ReadFile(tokenFile)
				if err != nil {
					return nil, tokenErr
				}

				tlsClientConfig := rest.TLSClientConfig{}
				if _, certErr := cert.NewPool(rootCAFile); err != nil {
					return nil, fmt.Errorf("expected to load root CA config from %s, but got err: %w", rootCAFile, certErr)
				} else {
					tlsClientConfig.CAFile = rootCAFile
				}

				config = &rest.Config{
					Host:            "https://" + net.JoinHostPort(os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")),
					TLSClientConfig: tlsClientConfig,
					BearerToken:     string(token),
					BearerTokenFile: tokenFile,
				} 
			}
		}

		config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
		config.ContentType = "application/vnd.kubernetes.protobuf"
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		return clientset, nil
	}
}
