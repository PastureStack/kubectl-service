package helm

type stackBackend interface {
	Name() string
	InstallStack(stack *Stack) (string, error)
	UpgradeStack(stack *Stack) (string, error)
	DeleteStack(stack *Stack) error
	RollbackStack(stack *Stack) error
	ListReleases() ([]Release, error)
}

var activeBackend stackBackend = helm4Backend{}

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
