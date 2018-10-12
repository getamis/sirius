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
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/getamis/sirius/database/mysql"
	"github.com/getamis/sirius/log"
)

var ErrMySQLNotStarted = errors.New("MySQL container is not started.")

type MySQLContainer struct {
	dockerContainer *Container
	Started         bool
	URL             string
}

func (container *MySQLContainer) Start() error {
	container.Started = true
	return container.dockerContainer.Start()
}

func (container *MySQLContainer) Suspend() error {
	return container.dockerContainer.Suspend()
}

func (container *MySQLContainer) Stop() error {
	container.Started = false
	return container.dockerContainer.Stop()
}

func (container *MySQLContainer) Teardown() error {
	if container.dockerContainer != nil && container.Started {
		container.Started = false
		return container.dockerContainer.Stop()
	}
	return ErrMySQLNotStarted
}

// MigrationOptions for mysql migration container
type MigrationOptions struct {
	ImageRepository string
	ImageTag        string

	MySQLOptions MySQLOptions

	// this command will override the default command.
	// "bundle" "exec" "rake" "db:migrate"
	Command []string
}

// RunMigrationContainer creates the migration container and connects to the
// mysql database to run the migration scripts.
func RunMigrationContainer(options MigrationOptions) error {
	// the default command
	command := []string{"bundle", "exec", "rake", "db:migrate"}
	if len(options.Command) > 0 {
		command = options.Command
	}

	if len(options.ImageTag) == 0 {
		options.ImageTag = "latest"
	}

	container := NewDockerContainer(
		ImageRepository(options.ImageRepository),
		ImageTag(options.ImageTag),
		DockerEnv(
			[]string{
				"RAILS_ENV=customized",
				fmt.Sprintf("HOST=%s", options.MySQLOptions.Host),
				fmt.Sprintf("PORT=%d", options.MySQLOptions.Port),
				fmt.Sprintf("DATABASE=%s", options.MySQLOptions.Database),
				fmt.Sprintf("USERNAME=%s", options.MySQLOptions.Username),
				fmt.Sprintf("PASSWORD=%s", options.MySQLOptions.Password),
			},
		),
		RunOptions(command),
	)

	if err := container.Start(); err != nil {
		log.Error("Failed to start container", "err", err)
		return err
	}

	if err := container.Wait(); err != nil {
		log.Error("Failed to wait container", "err", err)
		return err
	}

	return container.Stop()
}

type MySQLOptions struct {
	// The following options are used in the connection string and the mysql server container itself.
	Username string
	Password string
	Port     int
	Database string

	// The host address that will be used to build the connection string
	Host string
}

var DefaultMySQLOptions = MySQLOptions{
	Username: "root",
	Password: "my-secret-pw",

	// Currently the port will be published to the host.
	Port: 3306,

	// The db we want to run the test
	Database: "db0",

	// the mysql host to be connected from the client
	// if we're running test on the host, we might need to connect to the mysql
	// server via 127.0.0.1:3306. however if we want to run the test inside the container,
	// we need to inspect the IP of the container
	Host: "127.0.0.1",
}

func IsInsideContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/bin/running-in-container"); err == nil {
		return true
	}
	return false
}

func NewMySQLHealthChecker(options MySQLOptions) ContainerCallback {
	return func(c *Container) error {
		// We use this connection string to verify the mysql container is ready.
		connectionString, err := ToMySQLConnectionString(c, options)
		if err != nil {
			return err
		}

		return retry(10, 10*time.Second, func() error {
			db, err := sql.Open("mysql", connectionString)
			if err != nil {
				return err
			}
			defer db.Close()
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", options.Database))
			return err
		})
	}
}

func ToMySQLConnectionString(c *Container, options MySQLOptions) (string, error) {
	// By default we will use the host that is defined in the mysql options
	var host string = options.Host

	// If we're inside the container, we need to override the hostname
	// defined in the option
	if IsInsideContainer() {
		inspectedContainer, err := c.dockerClient.InspectContainer(c.container.ID)
		if err != nil {
			return "", err
		}
		host = inspectedContainer.NetworkSettings.IPAddress
	}

	// We use this connection string to verify the mysql container is ready.
	return mysql.ToConnectionString(
		mysql.Connector(mysql.DefaultProtocol, host, fmt.Sprintf("%d", options.Port)),
		mysql.Database(options.Database),
		mysql.UserInfo(options.Username, options.Password),
	)
}

// setup the mysql connection
// if TEST_MYSQL_HOST is defined, then we will use the connection directly.
// if not, a mysql container will be started
func SetupMySQL() (*MySQLContainer, error) {
	if host, ok := os.LookupEnv("TEST_MYSQL_HOST"); ok {
		port := DefaultMySQLOptions.Port
		if val, ok := os.LookupEnv("TEST_MYSQL_PORT"); ok {
			if p, err := strconv.Atoi(val); err != nil {
				return nil, err
			} else {
				port = p
			}
		}

		database := DefaultMySQLOptions.Database
		if val, ok := os.LookupEnv("TEST_MYSQL_DATABASE"); ok {
			database = val
		}

		username := DefaultMySQLOptions.Username
		if val, ok := os.LookupEnv("TEST_MYSQL_USERNAME"); ok {
			username = val
		}

		password := DefaultMySQLOptions.Password
		if val, ok := os.LookupEnv("TEST_MYSQL_PASSWORD"); ok {
			password = val
		}

		connectionString, err := mysql.ToConnectionString(
			mysql.Connector(mysql.DefaultProtocol, host, fmt.Sprintf("%d", port)),
			mysql.Database(database),
			mysql.UserInfo(username, password),
		)
		if err != nil {
			return nil, err
		}

		return &MySQLContainer{URL: connectionString}, nil
	}

	container, err := NewMySQLContainer(DefaultMySQLOptions)
	if err != nil {
		return nil, err
	}

	if err := container.Start(); err != nil {
		return container, err
	}

	return container, nil
}

func NewMySQLContainer(options MySQLOptions, containerOptions ...Option) (*MySQLContainer, error) {
	// Once the mysql container is ready, we will create the database if it does not exist.
	checker := NewMySQLHealthChecker(options)

	// we should publish the ports only when we're on the host
	if !IsInsideContainer() {
		containerOptions = append(containerOptions, Ports(options.Port))
	}

	// Create the container, please note that the container is not started yet.
	container := &MySQLContainer{
		dockerContainer: NewDockerContainer(
			// this is to keep some flexibility for passing extra container options..
			// however if we literally use "..." in the method call, an error
			// "too many arguments" will raise.
			append([]Option{
				ImageRepository("mysql"),
				ImageTag("5.7"),
				DockerEnv(
					[]string{
						fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", options.Password),
						fmt.Sprintf("MYSQL_DATABASE=%s", options.Database),
					},
				),
				HealthChecker(checker),
			}, containerOptions...)...,
		),
	}

	// FIXME
	connectionString, _ := mysql.ToConnectionString(
		mysql.Connector(mysql.DefaultProtocol, options.Host, fmt.Sprintf("%d", options.Port)),
		mysql.Database(options.Database),
		mysql.UserInfo(options.Username, options.Password),
	)
	container.URL = connectionString
	return container, nil
}
