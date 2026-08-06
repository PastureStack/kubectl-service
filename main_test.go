package main

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	platformhelm "github.com/PastureStack/kubectl-service/helm"
	"github.com/codegangsta/cli"
	"github.com/sirupsen/logrus"
)

func TestOperatorMessagesIncludeTraditionalChinese(t *testing.T) {
	if got := operatorMessage("en-US", "healthcheck-exit"); got == "" {
		t.Fatal("English operator message is empty")
	}
	if got := operatorMessage("zh-TW", "healthcheck-exit"); got != "健康檢查因錯誤而停止" {
		t.Fatalf("unexpected Traditional Chinese operator message: %q", got)
	}
}

func TestNewAppDefinesCompatibilityConfiguration(t *testing.T) {
	previousVersion := VERSION
	VERSION = "test-version"
	t.Cleanup(func() { VERSION = previousVersion })

	app := newApp()
	if app.Name != "kubectl-service" || app.Version != "test-version" {
		t.Fatalf("unexpected application identity: name=%q version=%q", app.Name, app.Version)
	}
	if app.Action == nil {
		t.Fatal("application action is not configured")
	}

	flagNames := make([]string, 0, len(app.Flags))
	for _, configuredFlag := range app.Flags {
		flagNames = append(flagNames, configuredFlag.GetName())
	}
	expected := []string{
		"platform-url",
		"platform-access-key",
		"platform-secret-key",
		"worker-count",
		"health-check-port",
		"debug",
		"locale",
		"helm-backend",
	}
	if !reflect.DeepEqual(flagNames, expected) {
		t.Fatalf("unexpected CLI flags: %#v", flagNames)
	}
}

func contextForLaunch(t *testing.T, values map[string]string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("launch-test", flag.ContinueOnError)
	set.String("platform-url", "", "")
	set.String("platform-access-key", "", "")
	set.String("platform-secret-key", "", "")
	set.Int("worker-count", 50, "")
	set.Int("health-check-port", 10240, "")
	set.Bool("debug", false, "")
	set.String("locale", "en-US", "")
	set.String("helm-backend", "legacy-helm2", "")
	for name, value := range values {
		if err := set.Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func TestLaunchRejectsUnsupportedLocaleBeforeStartingServices(t *testing.T) {
	eventCalls := 0
	healthCalls := 0
	originalEvents := startEventHandler
	originalHealth := startHealthCheck
	startEventHandler = func(string, string, string, int) error {
		eventCalls++
		return nil
	}
	startHealthCheck = func(int) error {
		healthCalls++
		return nil
	}
	t.Cleanup(func() {
		startEventHandler = originalEvents
		startHealthCheck = originalHealth
	})

	err := launch(contextForLaunch(t, map[string]string{"locale": "zh-CN"}))
	if err == nil || !strings.Contains(err.Error(), "unsupported locale") {
		t.Fatalf("unsupported locale returned %v", err)
	}
	if eventCalls != 0 || healthCalls != 0 {
		t.Fatalf("services started for an invalid locale: events=%d health=%d", eventCalls, healthCalls)
	}
}

func TestLaunchRejectsUnsupportedHelmBackendBeforeStartingServices(t *testing.T) {
	eventCalls := 0
	healthCalls := 0
	originalEvents := startEventHandler
	originalHealth := startHealthCheck
	startEventHandler = func(string, string, string, int) error {
		eventCalls++
		return nil
	}
	startHealthCheck = func(int) error {
		healthCalls++
		return nil
	}
	t.Cleanup(func() {
		startEventHandler = originalEvents
		startHealthCheck = originalHealth
	})

	err := launch(contextForLaunch(t, map[string]string{"helm-backend": "auto"}))
	if err == nil || !strings.Contains(err.Error(), "unsupported Helm backend") {
		t.Fatalf("unsupported backend returned %v", err)
	}
	if eventCalls != 0 || healthCalls != 0 {
		t.Fatalf("services started for an invalid backend: events=%d health=%d", eventCalls, healthCalls)
	}
}

func TestLaunchPassesConfigurationAndReportsHealthExit(t *testing.T) {
	originalEvents := startEventHandler
	originalHealth := startHealthCheck
	originalFatalf := logFatalf
	originalLevel := logrus.GetLevel()
	originalBackend := platformhelm.ActiveBackendName()

	healthPort := make(chan int, 1)
	releaseHealth := make(chan struct{})
	fatalMessage := make(chan string, 1)
	eventFailure := errors.New("event router stopped")

	startHealthCheck = func(port int) error {
		healthPort <- port
		<-releaseHealth
		return errors.New("health listener stopped")
	}
	logFatalf = func(format string, arguments ...interface{}) {
		fatalMessage <- fmt.Sprintf(format, arguments...)
	}
	startEventHandler = func(url, accessKey, secretKey string, workers int) error {
		if url != "https://platform.invalid" || accessKey != "access" || secretKey != "secret" || workers != 7 {
			t.Fatalf("unexpected event configuration: url=%q access=%q secret=%q workers=%d", url, accessKey, secretKey, workers)
		}
		if got := platformhelm.ActiveBackendName(); got != platformhelm.Helm4BackendName {
			t.Fatalf("event handler started with Helm backend %q", got)
		}
		return eventFailure
	}
	t.Cleanup(func() {
		startEventHandler = originalEvents
		startHealthCheck = originalHealth
		logFatalf = originalFatalf
		logrus.SetLevel(originalLevel)
		if err := platformhelm.ConfigureBackend(originalBackend); err != nil {
			t.Fatalf("restore Helm backend: %v", err)
		}
	})

	err := launch(contextForLaunch(t, map[string]string{
		"platform-url":        "https://platform.invalid",
		"platform-access-key": "access",
		"platform-secret-key": "secret",
		"worker-count":        "7",
		"health-check-port":   "12040",
		"debug":               "true",
		"locale":              "zh-TW",
		"helm-backend":        "helm4",
	}))
	if !errors.Is(err, eventFailure) {
		t.Fatalf("event failure was not propagated: %v", err)
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Fatalf("debug logging was not enabled: %s", logrus.GetLevel())
	}

	select {
	case port := <-healthPort:
		if port != 12040 {
			t.Fatalf("unexpected health port: %d", port)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("health listener did not start")
	}
	close(releaseHealth)
	select {
	case message := <-fatalMessage:
		if !strings.Contains(message, "健康檢查因錯誤而停止") || !strings.Contains(message, "health listener stopped") {
			t.Fatalf("unexpected health failure message: %q", message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("health listener exit was not reported")
	}
}
