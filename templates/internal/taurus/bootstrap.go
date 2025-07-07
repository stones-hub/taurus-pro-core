package taurus

var (
	Container *Components
)

// buildComponents builds all components
// configPath is the path to the configuration file or directory
// env is the environment file
func BuildComponents(configPath, env string) (func(), error) {
	var (
		cleanup func()
		err     error
	)

	// build Components
	Container, cleanup, err = buildComponents(&ConfigOptions{
		ConfigPath:  configPath,
		Env:         env,
		PrintEnable: true,
	})
	if err != nil {
		return nil, err
	}

	return cleanup, nil
}
