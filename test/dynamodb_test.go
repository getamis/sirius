package test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestNewDynamodbContainer(t *testing.T) {
	options := LoadDynamodbOptions()
	container, err := NewDynamodbContainer(options)
	assert.NoError(t, err, "dynamodb handle should be created.")

	err = container.Start()
	assert.NoError(t, err)

	if IsInsideContainer() {
		err = options.UpdateHostFromContainer(container)
		assert.NoError(t, err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(options.Region),
		Endpoint:    aws.String(options.Endpoint()),
		Credentials: credentials.NewStaticCredentials("FAKE", "FAKE", "FAKE"),
	}))

	ddb := dynamodb.New(sess)
	assert.NotNil(t, ddb)

	assert.NoError(t, container.Stop())

}
