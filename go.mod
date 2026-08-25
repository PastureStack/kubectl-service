module github.com/PastureStack/kubectl-service

go 1.27.0

require (
	github.com/rancher/event-subscriber v0.0.0-20160713200406-2e42f210dae2
	github.com/rancher/go-rancher v0.1.1-0.20160626053643-37c17e456a2d
	github.com/sirupsen/logrus v1.10.1
	github.com/urfave/cli/v3 v3.11.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/rancher/event-subscriber => ./third_party/event-subscriber

replace github.com/rancher/go-rancher => ./third_party/go-rancher
