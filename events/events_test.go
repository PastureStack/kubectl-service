package events

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/PastureStack/kubectl-service/helm"
	revents "github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/client"
	log "github.com/sirupsen/logrus"
)

type fakePublishOperations struct {
	created *client.Publish
	err     error
}

func (f *fakePublishOperations) List(*client.ListOpts) (*client.PublishCollection, error) {
	return nil, errors.New("not implemented")
}

func (f *fakePublishOperations) Create(value *client.Publish) (*client.Publish, error) {
	copyValue := *value
	f.created = &copyValue
	return &copyValue, f.err
}

func (f *fakePublishOperations) Update(*client.Publish, interface{}) (*client.Publish, error) {
	return nil, errors.New("not implemented")
}

func (f *fakePublishOperations) ById(string) (*client.Publish, error) {
	return nil, errors.New("not implemented")
}

func (f *fakePublishOperations) Delete(*client.Publish) error {
	return errors.New("not implemented")
}

func testClient(publish *fakePublishOperations) *client.RancherClient {
	return &client.RancherClient{Publish: publish}
}

func catalogEvent() *revents.Event {
	return &revents.Event{
		Name:         "kubernetesStack.create",
		ID:           "event-1",
		ReplyTo:      "reply.event-1",
		ResourceID:   "stack-1",
		ResourceType: "stack",
		Data: map[string]interface{}{
			"environment": map[string]interface{}{
				"name": "sample-stack",
				"data": map[string]interface{}{
					"fields": map[string]interface{}{
						"namespace": "sample-namespace",
						"templates": map[string]interface{}{
							"deployment.yaml": "kind: Deployment",
							"ignored":         42,
						},
					},
				},
			},
			"processData": map[string]interface{}{
				"templates": map[string]interface{}{
					"upgrade.yaml": "kind: Service",
				},
			},
		},
	}
}

func assertStack(t *testing.T, stack *helm.Stack, expectedFiles map[string]string) {
	t.Helper()
	if stack.Name != "sample-stack" || stack.Namespace != "sample-namespace" {
		t.Fatalf("unexpected stack identity: %#v", stack)
	}
	if !reflect.DeepEqual(stack.Files, expectedFiles) {
		t.Fatalf("unexpected stack files: %#v", stack.Files)
	}
}

func TestDecodeHelmStackInstallAndUpgrade(t *testing.T) {
	event := catalogEvent()
	assertStack(t, decodeHelmStack(event, nil, false), map[string]string{"deployment.yaml": "kind: Deployment"})
	assertStack(t, decodeHelmStack(event, nil, true), map[string]string{"upgrade.yaml": "kind: Service"})
}

func TestCatalogHandlersPublishSuccessfulReplies(t *testing.T) {
	originalInstall := installHelmStack
	originalUpgrade := upgradeHelmStack
	originalDelete := deleteHelmStack
	originalRollback := rollbackHelmStack
	t.Cleanup(func() {
		installHelmStack = originalInstall
		upgradeHelmStack = originalUpgrade
		deleteHelmStack = originalDelete
		rollbackHelmStack = originalRollback
	})

	installHelmStack = func(stack *helm.Stack) (string, error) {
		assertStack(t, stack, map[string]string{"deployment.yaml": "kind: Deployment"})
		return "install notes", nil
	}
	upgradeHelmStack = func(stack *helm.Stack) (string, error) {
		assertStack(t, stack, map[string]string{"upgrade.yaml": "kind: Service"})
		return "upgrade notes", nil
	}
	deleteHelmStack = func(stack *helm.Stack) error {
		assertStack(t, stack, map[string]string{"deployment.yaml": "kind: Deployment"})
		return nil
	}
	rollbackHelmStack = func(stack *helm.Stack) error {
		assertStack(t, stack, map[string]string{"deployment.yaml": "kind: Deployment"})
		return nil
	}

	tests := []struct {
		name          string
		handler       func(*revents.Event, *client.RancherClient) error
		expectedNotes string
	}{
		{name: "create", handler: create, expectedNotes: "install notes"},
		{name: "upgrade", handler: upgrade, expectedNotes: "upgrade notes"},
		{name: "remove", handler: remove},
		{name: "rollback", handler: rollback},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publish := &fakePublishOperations{}
			if err := test.handler(catalogEvent(), testClient(publish)); err != nil {
				t.Fatalf("handler failed: %v", err)
			}
			if publish.created == nil || publish.created.Name != "reply.event-1" {
				t.Fatalf("reply was not published: %#v", publish.created)
			}
			if test.expectedNotes != "" {
				outputs, ok := publish.created.Data["outputs"].(map[string]string)
				if !ok || outputs["notes"] != test.expectedNotes {
					t.Fatalf("unexpected notes payload: %#v", publish.created.Data)
				}
			}
		})
	}
}

func TestCatalogFailureLogDoesNotIncludeUntrustedStackName(t *testing.T) {
	originalDelete := deleteHelmStack
	logger := log.StandardLogger()
	originalOutput := logger.Out
	originalFormatter := logger.Formatter
	var output bytes.Buffer
	deleteHelmStack = func(*helm.Stack) error {
		return errors.New("delete failed")
	}
	logger.SetOutput(&output)
	logger.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	t.Cleanup(func() {
		deleteHelmStack = originalDelete
		logger.SetOutput(originalOutput)
		logger.SetFormatter(originalFormatter)
	})

	event := catalogEvent()
	environment := event.Data["environment"].(map[string]interface{})
	environment["name"] = "sample-stack\r\nforged-log-entry"
	_, err := removeCatalog(event, nil)
	if err == nil {
		t.Fatal("expected delete failure")
	}
	if strings.Contains(output.String(), "sample-stack") || strings.Contains(output.String(), "forged-log-entry") {
		t.Fatalf("failure log contains untrusted stack name: %q", output.String())
	}
	if !strings.Contains(output.String(), "Helm stack removal failed") {
		t.Fatalf("failure log is missing the operation context: %q", output.String())
	}
}

func TestWrapPublishesSanitizedFailureAndPropagatesPublishError(t *testing.T) {
	event := catalogEvent()
	publishFailure := errors.New("publish failed")
	publish := &fakePublishOperations{err: publishFailure}
	handlerFailure := errors.New("catalog operation failed")
	err := wrap(event, testClient(publish), func(*revents.Event, *client.RancherClient) (map[string]interface{}, error) {
		return nil, handlerFailure
	})
	if !errors.Is(err, publishFailure) {
		t.Fatalf("publish failure was not propagated: %v", err)
	}
	if publish.created == nil || publish.created.Transitioning != "error" || publish.created.TransitioningMessage != handlerFailure.Error() {
		t.Fatalf("unexpected failure reply: %#v", publish.created)
	}
}

func TestPingAndFinishUpgradeReplyRules(t *testing.T) {
	emptyReply := catalogEvent()
	emptyReply.ReplyTo = ""
	if err := ping(emptyReply, nil); err != nil {
		t.Fatalf("ping without reply target failed: %v", err)
	}

	for _, handler := range []func(*revents.Event, *client.RancherClient) error{ping, finishUpgrade} {
		publish := &fakePublishOperations{}
		if err := handler(catalogEvent(), testClient(publish)); err != nil {
			t.Fatalf("reply handler failed: %v", err)
		}
		if publish.created == nil || !reflect.DeepEqual(publish.created.PreviousIds, []string{"event-1"}) {
			t.Fatalf("unexpected reply: %#v", publish.created)
		}
	}
}

func TestMapAndStringHelpersRejectMalformedPaths(t *testing.T) {
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"text":  "value",
			"count": 2,
		},
		"wrong": "not-a-map",
	}

	gotMap := getMap(data, "nested")
	if !reflect.DeepEqual(gotMap, map[string]interface{}{"text": "value", "count": 2}) {
		t.Fatalf("unexpected map: %#v", gotMap)
	}
	gotMap["text"] = "changed"
	if data["nested"].(map[string]interface{})["text"] != "value" {
		t.Fatal("getMap returned an alias instead of a top-level copy")
	}
	if getMap(data, "missing") != nil || getMap(data, "wrong") != nil {
		t.Fatal("malformed map path was accepted")
	}

	if got := getStringMap(data, "nested"); !reflect.DeepEqual(got, map[string]string{"text": "value"}) {
		t.Fatalf("unexpected string map: %#v", got)
	}
	if getStringMap(data, "missing") != nil || getStringMap(data, "wrong") != nil {
		t.Fatal("malformed string-map path was accepted")
	}

	if got := getString(data, "nested", "text"); got != "value" {
		t.Fatalf("unexpected string: %q", got)
	}
	for _, keys := range [][]string{{}, {"missing"}, {"wrong", "text"}, {"nested", "count"}} {
		if got := getString(data, keys...); got != "" {
			t.Fatalf("malformed string path %v returned %q", keys, got)
		}
	}
}

func TestNewReplyCopiesEventIdentity(t *testing.T) {
	reply := newReply(catalogEvent())
	if reply.Name != "reply.event-1" || reply.ResourceId != "stack-1" || reply.ResourceType != "stack" || !reflect.DeepEqual(reply.PreviousIds, []string{"event-1"}) {
		t.Fatalf("unexpected reply identity: %#v", reply)
	}
}
