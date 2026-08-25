package main

import (
	"context"
	"fmt"
	"os"

	"github.com/PastureStack/kubectl-service/events"
	"github.com/PastureStack/kubectl-service/healthcheck"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var VERSION = "dev"

var (
	startHealthCheck  = healthcheck.Start
	startEventHandler = events.StartEventHandler
	logFatalf         = logrus.Fatalf
)

func newApp() *cli.Command {
	return &cli.Command{
		Name:    "kubectl-service",
		Version: VERSION,
		Action:  launch,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "platform-url",
				Usage:   "URL for the control-platform API",
				Sources: cli.EnvVars("PLATFORM_URL", "CATTLE_URL"),
			},
			&cli.StringFlag{
				Name:    "platform-access-key",
				Usage:   "Control-platform API access key",
				Sources: cli.EnvVars("PLATFORM_ACCESS_KEY", "CATTLE_ACCESS_KEY"),
			},
			&cli.StringFlag{
				Name:    "platform-secret-key",
				Usage:   "Control-platform API secret key",
				Sources: cli.EnvVars("PLATFORM_SECRET_KEY", "CATTLE_SECRET_KEY"),
			},
			&cli.IntFlag{
				Name:    "worker-count",
				Value:   50,
				Usage:   "Number of workers for handling events",
				Sources: cli.EnvVars("WORKER_COUNT"),
			},
			&cli.IntFlag{
				Name:    "health-check-port",
				Value:   10240,
				Usage:   "Port to configure an HTTP health check listener on",
				Sources: cli.EnvVars("HEALTH_CHECK_PORT"),
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug logs",
				Sources: cli.EnvVars("DEBUG"),
			},
			&cli.StringFlag{
				Name:    "locale",
				Value:   "en-US",
				Usage:   "Operator message locale: en-US or zh-TW",
				Sources: cli.EnvVars("PASTURESTACK_LOCALE"),
			},
		},
	}
}

func main() {
	if err := newApp().Run(context.Background(), os.Args); err != nil {
		logFatalf("Fatal exit: %v", err)
	}
}

func launch(_ context.Context, ctx *cli.Command) error {
	hcPort := ctx.Int("health-check-port")

	locale := ctx.String("locale")
	if locale != "en-US" && locale != "zh-TW" {
		return fmt.Errorf("unsupported locale %q; use en-US or zh-TW", locale)
	}
	url := ctx.String("platform-url")
	accessKey := ctx.String("platform-access-key")
	secretKey := ctx.String("platform-secret-key")
	workers := ctx.Int("worker-count")

	if ctx.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	go func() {
		logFatalf("%s: %v", operatorMessage(locale, "healthcheck-exit"), startHealthCheck(hcPort))
	}()

	return startEventHandler(url, accessKey, secretKey, workers)
}

func operatorMessage(locale, key string) string {
	messages := map[string]map[string]string{
		"en-US": {"healthcheck-exit": "Health check exited with an error"},
		"zh-TW": {"healthcheck-exit": "健康檢查因錯誤而停止"},
	}
	return messages[locale][key]
}
