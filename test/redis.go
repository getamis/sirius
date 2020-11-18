// Copyright 2017 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/getamis/sirius/log"
	redis "gopkg.in/redis.v5"
)

const (
	DefaultRedisPort = "6379"
)

type RedisOptions struct {
	Host string
	Port string
}

func (o RedisOptions) Endpoint() string {
	return net.JoinHostPort(o.Host, o.Port)
}

// UpdateHostFromContainer updates the redis host field according to the current environment
//
// If we're inside the container, we need to override the hostname
// defined in the option.
// If not, we should use the default value 127.0.0.1 because we will need to connect to the host port.
// please note that the TEST_REDIS_HOST can be overridden.
func (o *RedisOptions) UpdateHostFromContainer(c *Container) error {
	if IsInsideContainer() {
		inspectedContainer, err := c.dockerClient.InspectContainer(c.container.ID)
		if err != nil {
			return err
		}
		o.Host = inspectedContainer.NetworkSettings.IPAddress
	}
	return nil
}

type RedisContainer struct {
	*Container
	Options  RedisOptions
	Endpoint string
}

func (c *RedisContainer) Start() error {
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

func (c *RedisContainer) Teardown() error {
	if c.Container != nil && c.Container.Started {
		return c.Container.Stop()
	}

	return nil
}

// RedisOptions returns the redis options that will be used for the test
// cases to connect to.
func LoadRedisOptions() RedisOptions {
	options := RedisOptions{
		Host: "localhost",
		Port: DefaultRedisPort,
	}
	if host, ok := os.LookupEnv("TEST_REDIS_HOST"); ok {
		options.Host = host
	}
	if val, ok := os.LookupEnv("TEST_REDIS_PORT"); ok {
		options.Port = val
	}
	return options
}

func SetupRedis() (*RedisContainer, error) {
	options := LoadRedisOptions()

	// Explicit redis host is specified
	if _, ok := os.LookupEnv("TEST_REDIS_HOST"); ok {
		return &RedisContainer{
			Options:  options,
			Endpoint: options.Endpoint(),
		}, nil
	}

	c := NewRedisContainer(options)

	if err := c.Start(); err != nil {
		return nil, err
	}

	return c, nil
}

func NewRedisHealthChecker(options RedisOptions) ContainerCallback {
	return func(c *Container) error {
		if IsInsideContainer() {
			if err := options.UpdateHostFromContainer(c); err != nil {
				return err
			}
		}

		return retry(10, 1*time.Second, func() error {
			log.Debug("Checking redis status", "endpoint", options.Endpoint())
			c := redis.NewClient(&redis.Options{
				Addr: options.Endpoint(),
			})
			if c == nil {
				return fmt.Errorf("failed to connect to %s", options.Endpoint())
			}
			return c.Ping().Err()
		})
	}
}

func NewRedisContainer(options RedisOptions, containerOptions ...Option) *RedisContainer {
	checker := NewRedisHealthChecker(options)

	if IsInsideContainer() {
		containerOptions = append(containerOptions, ExposePorts(DefaultRedisPort))
	} else {
		// redis container port always expose the server port on 8000
		// bind the redis default port to the custom port on the host.
		containerOptions = append(containerOptions, ExposePorts(DefaultRedisPort))
		containerOptions = append(containerOptions, HostPortBindings(PortBinding{DefaultRedisPort + "/tcp", options.Port}))
	}

	// Create the container, please note that the container is not started yet.
	return &RedisContainer{
		Options: options,
		Container: NewDockerContainer(
			// this is to keep some flexibility for passing extra container options..
			// however if we literally use "..." in the method call, an error
			// "too many arguments" will raise.
			append([]Option{
				ImageRepository("redis"),
				ImageTag("6-alpine"),
				DockerEnv([]string{}),
				HealthChecker(checker),
			}, containerOptions...)...,
		),
		Endpoint: options.Endpoint(),
	}
}
