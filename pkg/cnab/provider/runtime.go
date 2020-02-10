package cnabprovider

import (
	"os"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/cnabio/cnab-go/driver/lookup"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

type Runtime struct {
	*config.Config
	credentials credentials.CredentialProvider
	claims      claims.ClaimProvider
}

func NewRuntime(c *config.Config, claims claims.ClaimProvider, credentials credentials.CredentialProvider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
	}
}

func (d *Runtime) newDriver(driverName string, claimName string, args ActionArguments) (driver.Driver, error) {
	driverImpl, err := lookup.Lookup(driverName)
	if err != nil {
		return driverImpl, err
	}

	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := make(map[string]string)
		// Load any driver-specific config out of the environment
		for env := range configurable.Config() {
			if val, ok := os.LookupEnv(env); ok {
				driverCfg[env] = val
			}
		}

		configurable.SetConfig(driverCfg)
	}

	// Add docker specific options.
	// There is probably a better way to do this?
	if docker, ok := driverImpl.(*docker.Driver); ok {
		mounts := func(config *container.Config, hostConfig *container.HostConfig) error {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			})
			return nil
		}

		docker.AddConfigurationOptions(mounts)
	}

	return driverImpl, err
}
