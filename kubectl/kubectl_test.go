package kubectl

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteNamespaceUsesKubernetesHTTPAPI(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath())
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v1/namespaces/demo-ns" {
			t.Fatalf("unexpected path %s", r.URL.EscapedPath())
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	t.Setenv("SERVER", server.URL)
	t.Setenv("KUBE_SERVER", "")

	if err := DeleteNamespace("demo-ns"); err != nil {
		t.Fatalf("delete namespace failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one request, got %#v", requests)
	}
}

func TestDeleteNamespaceTreatsNotFoundAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	t.Setenv("SERVER", server.URL)

	if err := DeleteNamespace("already-gone"); err != nil {
		t.Fatalf("expected 404 to be non-fatal, got %v", err)
	}
}

func TestDeleteNamespaceRejectsInvalidServer(t *testing.T) {
	t.Setenv("SERVER", "file:///tmp/kube")

	if err := DeleteNamespace("demo"); err == nil {
		t.Fatal("expected invalid scheme to fail")
	}
}

func TestDeleteNamespaceKeepsExistingServerPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/proxy/api/v1/namespaces/demo%2Fns" {
			t.Fatalf("unexpected path %s", r.URL.EscapedPath())
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("SERVER", server.URL+"/proxy")

	if err := DeleteNamespace("demo/ns"); err != nil {
		t.Fatalf("delete namespace failed: %v", err)
	}
}
