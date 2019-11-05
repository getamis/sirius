package test

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/getamis/sirius/log"
	vaultApi "github.com/hashicorp/vault/api"
)

const (
	DefaultVaultPort  = "8200"
	DefaultToken      = "TEST-DEFAULT-TOKEN"
	DefaultListenAddr = "0.0.0.0"
)

type VaultOptions struct {
	Host       string
	Port       string
	Token      string
	ListenAddr string
}

func (o VaultOptions) Endpoint() string {
	return "http://" + net.JoinHostPort(o.Host, o.Port)
}

// UpdateHostFromContainer updates the vault host field according to the current environment
//
// If we're inside the container, we need to override the hostname
// defined in the option.
// If not, we should use the default value 127.0.0.1 because we will need to connect to the host port.
// please note that the TEST_VAULT_HOST can be overridden.
func (o *VaultOptions) UpdateHostFromContainer(c *Container) error {
	if IsInsideContainer() {
		inspectedContainer, err := c.dockerClient.InspectContainer(c.container.ID)
		if err != nil {
			return err
		}
		o.Host = inspectedContainer.NetworkSettings.IPAddress
	}
	return nil
}

type VaultContainer struct {
	*Container
	Options  VaultOptions
	Endpoint string
}

func (c *VaultContainer) Start() error {
	err := c.Container.Start()
	if err != nil {
		return err
	}

	if err := c.Options.UpdateHostFromContainer(c.Container); err != nil {
		return err
	}

	c.Endpoint = c.Options.Endpoint()
	return nil
}

func (c *VaultContainer) Teardown() error {
	if c.Container != nil && c.Container.Started {
		return c.Container.Stop()
	}

	return nil
}

// LoadVaultOptions returns the vault options that will be used for the test
// cases to connect to.
func LoadVaultOptions() VaultOptions {
	options := VaultOptions{
		Host:       "localhost",
		Port:       DefaultVaultPort,
		Token:      DefaultToken,
		ListenAddr: DefaultListenAddr,
	}
	if host, ok := os.LookupEnv("TEST_VAULT_HOST"); ok {
		options.Host = host
	}
	if val, ok := os.LookupEnv("TEST_VAULT_PORT"); ok {
		options.Port = val
	}
	if val, ok := os.LookupEnv("TEST_VAULT_TOKEN"); ok {
		options.Token = val
	}
	if val, ok := os.LookupEnv("TEST_LISTEN_ADDR"); ok {
		options.ListenAddr = val
	}
	return options
}

func SetupVault() (*VaultContainer, error) {
	options := LoadVaultOptions()

	// Explicit vault host is specified
	if _, ok := os.LookupEnv("TEST_VAULT_HOST"); ok {
		return &VaultContainer{
			Options:  options,
			Endpoint: options.Endpoint(),
		}, nil
	}

	c, err := NewVaultContainer(options)

	if err := c.Start(); err != nil {
		return c, err
	}

	return c, err
}

func NewVaultHealthChecker(options VaultOptions) ContainerCallback {
	return func(c *Container) error {
		if IsInsideContainer() {
			if err := options.UpdateHostFromContainer(c); err != nil {
				return err
			}
		}

		return retry(10, 1*time.Second, func() error {
			log.Debug("Checking vault status", "endpoint", options.Endpoint())
			config := vaultApi.DefaultConfig()
			config.Address = options.Endpoint()

			client, err := vaultApi.NewClient(config)
			if err != nil {
				log.Error("Failed to create vault client", "err", err)
				return err
			}
			_, err = client.Sys().Health()
			return err
		})
	}
}

func NewVaultContainer(options VaultOptions, containerOptions ...Option) (*VaultContainer, error) {
	// Once the vault container is ready, we will create the database if it does not exist.
	checker := NewVaultHealthChecker(options)

	if IsInsideContainer() {
		containerOptions = append(containerOptions, ExposePorts(DefaultVaultPort))
	} else {
		// vault container port always expose the server port on 8000
		// bind the vault default port to the custom port on the host.
		containerOptions = append(containerOptions, ExposePorts(DefaultVaultPort))
		containerOptions = append(containerOptions, HostPortBindings(PortBinding{DefaultVaultPort + "/tcp", options.Port}))
	}

	// Create the container, please note that the container is not started yet.
	return &VaultContainer{
		Options: options,
		Container: NewDockerContainer(
			// this is to keep some flexibility for passing extra container options..
			// however if we literally use "..." in the method call, an error
			// "too many arguments" will raise.
			append([]Option{
				ImageRepository("vault"),
				ImageTag("1.0.3"),
				DockerEnv([]string{
					fmt.Sprintf("VAULT_DEV_ROOT_TOKEN_ID=%s", options.Token),
					fmt.Sprintf("VAULT_DEV_LISTEN_ADDRESS=%s", net.JoinHostPort(options.ListenAddr, options.Port)),
				}),
				HealthChecker(checker),
			}, containerOptions...)...,
		),
		Endpoint: "http://" + net.JoinHostPort(options.Host, options.Port),
	}, nil
}
