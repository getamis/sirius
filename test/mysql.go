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
	"fmt"
	"time"

	"github.com/getamis/sirius/database/mysql"
	"github.com/getamis/sirius/log"
)

type MySQLContainer struct {
	dockerContainer *Container
	URL             string
}

func (container *MySQLContainer) Start() error {
	return container.dockerContainer.Start()
}

func (container *MySQLContainer) Suspend() error {
	return container.dockerContainer.Suspend()
}

func (container *MySQLContainer) Stop() error {
	return container.dockerContainer.Stop()
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
	Port:     3306,
	Database: "db0",

	Host: "127.0.0.1",
}

func NewMySQLContainer(options MySQLOptions, containerOptions ...Option) (*MySQLContainer, error) {

	// We use this connection string to verify the mysql container is ready.
	connectionString, _ := mysql.ToConnectionString(
		mysql.Connector(mysql.DefaultProtocol, options.Host, fmt.Sprintf("%d", options.Port)),
		mysql.Database(options.Database),
		mysql.UserInfo(options.Username, options.Password),
	)

	// Once the mysql container is ready, we will create the database if it does not exist.
	checker := func(c *Container) error {
		return retry(10, 5*time.Second, func() error {
			db, err := sql.Open("mysql", connectionString)
			if err != nil {
				return err
			}
			defer db.Close()
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", options.Database))
			return err
		})
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
				Ports(options.Port),
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

	container.URL = connectionString

	return container, nil
}
