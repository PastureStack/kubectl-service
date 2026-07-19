package helm

import (
	"fmt"

	"github.com/PastureStack/kubectl-service/kubectl"
	log "github.com/sirupsen/logrus"
)

func deleteHelmStackLegacyHelm2(stack *Stack) error {
	releases, err := listReleasesLegacyHelm2()
	if err != nil {
		log.Errorf("Error obtaining helm releases %v", err)
		return err
	}
	releaseFound := false
	brokenRelease := false
	for _, r := range releases {
		if r.Name == stack.Name {
			if r.Status != "DEPLOYED" && r.Status != "SUPERSEDED" {
				brokenRelease = true
			}
			releaseFound = true
			break
		}
	}
	if !releaseFound {
		return nil
	}
	args := []string{"delete", stack.Name}
	output := runCommand(legacyHelm2Command, args...)
	if brokenRelease {
		log.Infof("Tried to delete %s, err: %v", stack.Name, output.Err)
		return nil
	}
	if output.ExitCode > 0 {
		return fmt.Errorf("%s", output.StdErr)
	}
	if output.Err != nil {
		return output.Err
	}
	if stack.Namespace != "" {
		err = kubectl.DeleteNamespace(stack.Namespace)
	}
	return err
}

func rollbackHelmStackLegacyHelm2(stack *Stack) error {
	args := []string{"rollback", stack.Name, "0"}
	output := runCommand(legacyHelm2Command, args...)
	if output.ExitCode > 0 {
		return fmt.Errorf("%s", output.StdErr)
	}
	return output.Err
}
