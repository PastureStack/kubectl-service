package kubectl

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func useNamespaceTransport(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	original := namespaceHTTPClient
	client := newNamespaceHTTPClient()
	client.Transport = transport
	namespaceHTTPClient = client
	t.Cleanup(func() {
		namespaceHTTPClient = original
	})
}

func response(request *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

func TestDeleteNamespaceUsesApprovedKubernetesHTTPAPI(t *testing.T) {
	requests := []string{}
	useNamespaceTransport(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.Method+" "+request.URL.String())
		if request.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", request.Method)
		}
		if request.URL.Host != "kubernetes.kubernetes.rancher.internal" {
			t.Fatalf("unexpected host %s", request.URL.Host)
		}
		if request.URL.EscapedPath() != "/api/v1/namespaces/demo-ns" {
			t.Fatalf("unexpected path %s", request.URL.EscapedPath())
		}
		return response(request, http.StatusAccepted, ""), nil
	}))

	t.Setenv("SERVER", legacyInternalKubernetesServer)
	t.Setenv("KUBE_SERVER", "")
	t.Setenv("PLATFORM_KUBERNETES_SERVER", "")

	if err := DeleteNamespace("demo-ns"); err != nil {
		t.Fatalf("delete namespace failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one request, got %#v", requests)
	}
}

func TestDeleteNamespaceTreatsNotFoundAsSuccess(t *testing.T) {
	useNamespaceTransport(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return response(request, http.StatusNotFound, "not found"), nil
	}))

	t.Setenv("SERVER", defaultKubernetesServer)
	t.Setenv("KUBE_SERVER", "")
	t.Setenv("PLATFORM_KUBERNETES_SERVER", "")

	if err := DeleteNamespace("already-gone"); err != nil {
		t.Fatalf("expected 404 to be non-fatal, got %v", err)
	}
}

func TestApprovedKubernetesEndpointUsesCompatibilityFallback(t *testing.T) {
	t.Setenv("SERVER", "")
	t.Setenv("KUBE_SERVER", serviceKubernetesServer)
	t.Setenv("PLATFORM_KUBERNETES_SERVER", "")

	endpoint, err := approvedKubernetesEndpoint()
	if err != nil {
		t.Fatalf("approved service endpoint failed: %v", err)
	}
	if endpoint.String() != serviceKubernetesServer {
		t.Fatalf("unexpected endpoint %q", endpoint.String())
	}
}

func TestApprovedKubernetesEndpointPrefersModernSetting(t *testing.T) {
	t.Setenv("PLATFORM_KUBERNETES_SERVER", clusterLocalKubernetesServer)
	t.Setenv("SERVER", "http://169.254.169.254")
	t.Setenv("KUBE_SERVER", "http://example.invalid")

	endpoint, err := approvedKubernetesEndpoint()
	if err != nil {
		t.Fatalf("approved modern endpoint failed: %v", err)
	}
	if endpoint.String() != clusterLocalKubernetesServer {
		t.Fatalf("unexpected endpoint %q", endpoint.String())
	}
}

func TestDeleteNamespaceRejectsUnapprovedServers(t *testing.T) {
	servers := []string{
		"file:///tmp/kube",
		"http://169.254.169.254",
		"http://kubernetes.kubernetes.rancher.internal.evil.invalid",
		legacyInternalKubernetesServer + "/proxy",
		"http://user:secret@kubernetes.kubernetes.rancher.internal",
	}
	for _, server := range servers {
		t.Run(server, func(t *testing.T) {
			t.Setenv("SERVER", server)
			t.Setenv("KUBE_SERVER", "")
			t.Setenv("PLATFORM_KUBERNETES_SERVER", "")
			if err := DeleteNamespace("demo"); err == nil {
				t.Fatal("expected unapproved server to fail")
			}
		})
	}
}

func TestDeleteNamespaceRejectsInvalidNamespace(t *testing.T) {
	t.Setenv("SERVER", defaultKubernetesServer)
	t.Setenv("PLATFORM_KUBERNETES_SERVER", "")
	for _, namespace := range []string{"demo/ns", "../metadata", "UPPERCASE", strings.Repeat("a", 64)} {
		t.Run(namespace, func(t *testing.T) {
			if err := DeleteNamespace(namespace); err == nil {
				t.Fatal("expected invalid namespace to fail")
			}
		})
	}
}

func TestDeleteNamespaceDoesNotFollowRedirects(t *testing.T) {
	requests := []string{}
	useNamespaceTransport(t, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.URL.String())
		redirect := response(request, http.StatusFound, "redirect")
		redirect.Header.Set("Location", "http://169.254.169.254/latest/meta-data")
		return redirect, nil
	}))

	t.Setenv("SERVER", legacyInternalKubernetesServer)
	t.Setenv("KUBE_SERVER", "")
	t.Setenv("PLATFORM_KUBERNETES_SERVER", "")

	err := DeleteNamespace("demo")
	if err == nil || !strings.Contains(err.Error(), "HTTP 302") {
		t.Fatalf("expected redirect to be rejected, got %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("redirect was followed: %#v", requests)
	}
}
