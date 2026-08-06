package helm

import "fmt"

type stackBackend interface {
	Name() string
	InstallStack(stack *Stack) (string, error)
	UpgradeStack(stack *Stack) (string, error)
	DeleteStack(stack *Stack) error
	RollbackStack(stack *Stack) error
	ListReleases() ([]Release, error)
}

type legacyHelm2Backend struct{}

var activeBackend stackBackend = legacyHelm2Backend{}

func ConfigureBackend(name string) error {
	var selected stackBackend
	switch name {
	case "", LegacyHelm2BackendName:
		selected = legacyHelm2Backend{}
	case Helm4BackendName:
		selected = helm4Backend{}
	default:
		return fmt.Errorf("unsupported Helm backend %q; use %s or %s", name, LegacyHelm2BackendName, Helm4BackendName)
	}

	activeBackend = selected
	return nil
}

func ActiveBackendName() string {
	return activeBackend.Name()
}

func InstallHelmStack(stack *Stack) (string, error) {
	if err := validateStackIdentity(stack); err != nil {
		return "", err
	}
	return activeBackend.InstallStack(stack)
}

func UpgradeHelmStack(stack *Stack) (string, error) {
	if err := validateStackIdentity(stack); err != nil {
		return "", err
	}
	return activeBackend.UpgradeStack(stack)
}

func DeleteHelmStack(stack *Stack) error {
	if err := validateStackIdentity(stack); err != nil {
		return err
	}
	return activeBackend.DeleteStack(stack)
}

func RollbackHelmStack(stack *Stack) error {
	if err := validateStackIdentity(stack); err != nil {
		return err
	}
	return activeBackend.RollbackStack(stack)
}

func ListReleases() ([]Release, error) {
	return activeBackend.ListReleases()
}

func (legacyHelm2Backend) Name() string {
	return LegacyHelm2BackendName
}

func (legacyHelm2Backend) InstallStack(stack *Stack) (string, error) {
	args := []string{"install"}
	return executeHelmCreateUpgradeTask(stack, args, false)
}

func (legacyHelm2Backend) UpgradeStack(stack *Stack) (string, error) {
	args := []string{"upgrade"}
	return executeHelmCreateUpgradeTask(stack, args, true)
}

func (legacyHelm2Backend) DeleteStack(stack *Stack) error {
	return deleteHelmStackLegacyHelm2(stack)
}

func (legacyHelm2Backend) RollbackStack(stack *Stack) error {
	return rollbackHelmStackLegacyHelm2(stack)
}

func (legacyHelm2Backend) ListReleases() ([]Release, error) {
	return listReleasesLegacyHelm2()
}

type helm4Backend struct{}

func (helm4Backend) Name() string {
	return Helm4BackendName
}

func (helm4Backend) InstallStack(stack *Stack) (string, error) {
	return executeHelm4CreateUpgradeTask(stack, false)
}

func (helm4Backend) UpgradeStack(stack *Stack) (string, error) {
	return executeHelm4CreateUpgradeTask(stack, true)
}

func (helm4Backend) DeleteStack(stack *Stack) error {
	return deleteHelmStackHelm4(stack)
}

func (helm4Backend) RollbackStack(stack *Stack) error {
	return rollbackHelmStackHelm4(stack)
}

func (helm4Backend) ListReleases() ([]Release, error) {
	return listReleasesHelm4()
}
