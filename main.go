package main

import (
	"fmt"
	"os"

	"github.com/PastureStack/kubectl-service/events"
	"github.com/codegangsta/cli"
	"github.com/rancher/swarm-agent/healthcheck"
	"github.com/sirupsen/logrus"
)

var VERSION = "dev"

func main() {
	app := cli.NewApp()
	app.Name = "kubectl-service"
	app.Version = VERSION
	app.Action = launch

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "platform-url",
			Usage:  "URL for the control-platform API",
			EnvVar: "PLATFORM_URL,CATTLE_URL",
		},
		cli.StringFlag{
			Name:   "platform-access-key",
			Usage:  "Control-platform API access key",
			EnvVar: "PLATFORM_ACCESS_KEY,CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "platform-secret-key",
			Usage:  "Control-platform API secret key",
			EnvVar: "PLATFORM_SECRET_KEY,CATTLE_SECRET_KEY",
		},
		cli.IntFlag{
			Name:   "worker-count",
			Value:  50,
			Usage:  "Number of workers for handling events",
			EnvVar: "WORKER_COUNT",
		},
		cli.IntFlag{
			Name:   "health-check-port",
			Value:  10240,
			Usage:  "Port to configure an HTTP health check listener on",
			EnvVar: "HEALTH_CHECK_PORT",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Enable debug logs",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "locale",
			Value:  "en-US",
			Usage:  "Operator message locale: en-US or zh-TW",
			EnvVar: "PASTURESTACK_LOCALE",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf("Fatal exit: %v", err)
	}
}

func launch(ctx *cli.Context) error {
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
		logrus.Fatalf("%s: %v", operatorMessage(locale, "healthcheck-exit"), healthcheck.StartHealthCheck(hcPort))
	}()

	return events.StartEventHandler(url, accessKey, secretKey, workers)
}

func operatorMessage(locale, key string) string {
	messages := map[string]map[string]string{
		"en-US": {"healthcheck-exit": "Health check exited with an error"},
		"zh-TW": {"healthcheck-exit": "健康檢查因錯誤而停止"},
	}
	return messages[locale][key]
}
