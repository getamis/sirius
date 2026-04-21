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
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQContainer struct {
	dockerContainer *Container
	URL             string
}

func (container *RabbitMQContainer) Start() error {
	return container.dockerContainer.Start()
}

func (container *RabbitMQContainer) Suspend() error {
	return container.dockerContainer.Suspend()
}

func (container *RabbitMQContainer) Stop() error {
	return container.dockerContainer.Stop()
}

func NewRabbitMQContainer() (*RabbitMQContainer, error) {
	containerPort := 5672
	hostPort := 5673
	guiPort := 15672
	endpoint := fmt.Sprintf("amqp://guest:guest@127.0.0.1:%d", hostPort)
	checker := func(c *Container) error {
		return retry(10, 5*time.Second, func() error {
			conn, err := amqp.Dial(endpoint)
			if err != nil {
				return err
			}
			defer conn.Close()
			return nil
		})
	}
	container := &RabbitMQContainer{
		dockerContainer: NewDockerContainer(
			ImageRepository("rabbitmq"),
			ImageTag("3.13-management"),
			HostPortBindings(
				PortBinding{ContainerPort: fmt.Sprintf("%d", containerPort), HostPort: fmt.Sprintf("%d", hostPort)},
				PortBinding{ContainerPort: fmt.Sprintf("%d", guiPort), HostPort: fmt.Sprintf("%d", guiPort)},
			),
			ExposePorts(fmt.Sprintf("%d", containerPort), fmt.Sprintf("%d", guiPort)),
			HealthChecker(checker),
		),
	}

	container.URL = endpoint

	return container, nil
}
