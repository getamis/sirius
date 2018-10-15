package test

import (
	"net"
	"os"
)

const DefaultDynamodbPort = "8000"

type DynamodbOptions struct {
	Host string
	Port string
}

type DynamodbContainer struct {
	*Container
	Options  DynamodbOptions
	Endpoint string
}

// LoadDynamodbOptions returns the dynamodb options that will be used for the test
// cases to connect to.
func LoadDynamodbOptions() DynamodbOptions {
	options := DynamodbOptions{
		Host: "localhost",
		Port: DefaultDynamodbPort,
	}
	if host, ok := os.LookupEnv("TEST_DYNAMODB_HOST"); ok {
		options.Host = host
	}
	if val, ok := os.LookupEnv("TEST_DYNAMODB_PORT"); ok {
		options.Port = val
	}
	return options
}

func SetupDynamodb() (*DynamodbContainer, error) {
	options := LoadDynamodbOptions()

	// Explicit dynamodb host is specified
	if _, ok := os.LookupEnv("TEST_DYNAMODB_HOST"); ok {
		return &DynamodbContainer{
			Options:  options,
			Endpoint: "http://" + net.JoinHostPort(options.Host, options.Port),
		}
	}

	c, err := NewDynamodbContainer(options)
	return c, err
}

func NewDynamodbContainer(options DynamodbOptions, containerOptions ...Option) (*DynamodbContainer, error) {
	if IsInsideContainer() {
		containerOptions = append(containerOptions, ExposePorts(DefaultDynamodbPort))
	} else {
		// dynamodb container port always expose the server port on 8000
		// bind the dynamodb default port to the custom port on the host.
		containerOptions = append(containerOptions, ExposePorts(DefaultDynamodbPort))
		containerOptions = append(containerOptions, HostPortBindings(PortBinding{DefaultDynamodbPort + "/tcp", options.Port}))
	}

	// Create the container, please note that the container is not started yet.
	return &DynamodbContainer{
		Options: options,
		Container: NewDockerContainer(
			// this is to keep some flexibility for passing extra container options..
			// however if we literally use "..." in the method call, an error
			// "too many arguments" will raise.
			append([]Option{
				ImageRepository("amazon/dynamodb-local"),
				ImageTag("latest"),
				DockerEnv([]string{}),
			}, containerOptions...)...,
		),
		Endpoint: "http://" + net.JoinHostPort(options.Host, options.Port),
	}, nil
}
