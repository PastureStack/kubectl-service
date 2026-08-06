package kubectl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultKubernetesServer = "http://localhost:8080"
	defaultDeleteTimeout    = 30 * time.Second
)

func DeleteNamespace(namespace string) error {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil
	}

	server := strings.TrimRight(strings.TrimSpace(os.Getenv("SERVER")), "/")
	if server == "" {
		server = strings.TrimRight(strings.TrimSpace(os.Getenv("KUBE_SERVER")), "/")
	}
	if server == "" {
		server = defaultKubernetesServer
	}

	endpoint, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("invalid Kubernetes server URL %q: %v", server, err)
	}
	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return fmt.Errorf("unsupported Kubernetes server URL scheme %q", endpoint.Scheme)
	}
	if endpoint.Host == "" {
		return fmt.Errorf("Kubernetes server URL %q is missing host", server)
	}
	basePath := strings.TrimRight(endpoint.Path, "/")
	baseRawPath := strings.TrimRight(endpoint.EscapedPath(), "/")
	endpoint.Path = basePath + "/api/v1/namespaces/" + namespace
	endpoint.RawPath = baseRawPath + "/api/v1/namespaces/" + url.PathEscape(namespace)
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

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
			log.Infof("Namespace %s was already absent", namespace)
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("delete namespace %s failed: HTTP %d: %s", namespace, resp.StatusCode, strings.TrimSpace(string(body)))
}

var namespaceHTTPClient = &http.Client{Timeout: defaultDeleteTimeout}
