package kubectl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultKubernetesServer        = "http://localhost:8080"
	legacyInternalKubernetesServer = "http://kubernetes.kubernetes.rancher.internal"
	serviceKubernetesServer        = "https://kubernetes.default.svc"
	clusterLocalKubernetesServer   = "https://kubernetes.default.svc.cluster.local"
	defaultDeleteTimeout           = 30 * time.Second
)

var kubernetesNamespacePattern = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$`)

func DeleteNamespace(namespace string) error {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil
	}
	if len(namespace) > 63 || !kubernetesNamespacePattern.MatchString(namespace) {
		return fmt.Errorf("invalid Kubernetes namespace")
	}

	endpoint, err := approvedKubernetesEndpoint()
	if err != nil {
		return err
	}
	endpoint.Path = "/api/v1/namespaces/" + namespace
	endpoint.RawPath = "/api/v1/namespaces/" + url.PathEscape(namespace)

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}
	resp, err := namespaceHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted ||
		resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		if resp.StatusCode == http.StatusNotFound {
			log.Info("Kubernetes namespace was already absent")
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return nil
	}

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("delete Kubernetes namespace failed: HTTP %d", resp.StatusCode)
}

func approvedKubernetesEndpoint() (*url.URL, error) {
	server := strings.TrimRight(strings.TrimSpace(os.Getenv("PLATFORM_KUBERNETES_SERVER")), "/")
	if server == "" {
		server = strings.TrimRight(strings.TrimSpace(os.Getenv("SERVER")), "/")
	}
	if server == "" {
		server = strings.TrimRight(strings.TrimSpace(os.Getenv("KUBE_SERVER")), "/")
	}
	if server == "" {
		server = defaultKubernetesServer
	}

	switch server {
	case defaultKubernetesServer:
		return &url.URL{Scheme: "http", Host: "localhost:8080"}, nil
	case legacyInternalKubernetesServer:
		return &url.URL{Scheme: "http", Host: "kubernetes.kubernetes.rancher.internal"}, nil
	case serviceKubernetesServer:
		return &url.URL{Scheme: "https", Host: "kubernetes.default.svc"}, nil
	case clusterLocalKubernetesServer:
		return &url.URL{Scheme: "https", Host: "kubernetes.default.svc.cluster.local"}, nil
	default:
		return nil, fmt.Errorf("Kubernetes server URL is not an approved internal endpoint")
	}
}

func newNamespaceHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultDeleteTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

var namespaceHTTPClient = newNamespaceHTTPClient()
