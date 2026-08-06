package events

import (
	"github.com/PastureStack/kubectl-service/helm"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/client"
	log "github.com/sirupsen/logrus"
)

var (
	installHelmStack  = helm.InstallHelmStack
	upgradeHelmStack  = helm.UpgradeHelmStack
	deleteHelmStack   = helm.DeleteHelmStack
	rollbackHelmStack = helm.RollbackHelmStack
)

func installCatalog(event *events.Event, cli *client.RancherClient) (map[string]interface{}, error) {
	stack := decodeHelmStack(event, cli, false)
	notes, err := installHelmStack(stack)
	if err != nil {
		log.Errorf("Error installing helm stack: %s error: %v", stack.Name, err)
	}
	return map[string]interface{}{
		"outputs": map[string]string{
			"notes": notes,
		},
	}, err
}

func upgradeCatalog(event *events.Event, cli *client.RancherClient) (map[string]interface{}, error) {
	stack := decodeHelmStack(event, cli, true)
	notes, err := upgradeHelmStack(stack)
	if err != nil {
		log.Errorf("Error upgrading helm stack: %s error: %v", stack.Name, err)
	}
	return map[string]interface{}{
		"outputs": map[string]string{
			"notes": notes,
		},
	}, err
}

func removeCatalog(event *events.Event, cli *client.RancherClient) (map[string]interface{}, error) {
	stack := decodeHelmStack(event, cli, false)
	err := deleteHelmStack(stack)
	if err != nil {
		log.Errorf("Error removing helm stack: %s error:  %v", stack.Name, err)
	}
	return nil, err
}

func rollbackCatalog(event *events.Event, cli *client.RancherClient) (map[string]interface{}, error) {
	stack := decodeHelmStack(event, cli, false)
	err := rollbackHelmStack(stack)
	if err != nil {
		log.Errorf("Error rolling helm stack: %s error: %v", stack.Name, err)
	}
	return nil, err
}
