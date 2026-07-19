package helm

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

func ActiveBackendName() string {
	return activeBackend.Name()
}

func InstallHelmStack(stack *Stack) (string, error) {
	return activeBackend.InstallStack(stack)
}

func UpgradeHelmStack(stack *Stack) (string, error) {
	return activeBackend.UpgradeStack(stack)
}

func DeleteHelmStack(stack *Stack) error {
	return activeBackend.DeleteStack(stack)
}

func RollbackHelmStack(stack *Stack) error {
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
