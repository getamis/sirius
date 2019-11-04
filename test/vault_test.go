package test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestNewVaultContainer(t *testing.T) {
	putPath := "secret/data/mysecret/test"
	options := LoadVaultOptions()
	container, err := NewVaultContainer(options)
	assert.NoError(t, err, "vault handle should be created.")

	err = container.Start()
	assert.NoError(t, err)

	config := vaultApi.DefaultConfig()
	config.Address = options.Endpoint()
	client, err := vaultApi.NewClient(config)
	assert.NoError(t, err)

	client.SetToken(options.Token)
	secret := make(map[string]interface{})
	secret["data"] = map[string]interface{}{
		"value": "pass",
	}
	_, err = client.Logical().Write(putPath, secret)
	assert.NoError(t, err)

	res, err := client.Logical().List("secret/metadata/mysecret")
	assert.NoError(t, err)
	assert.NotNil(t, res)

	_, ok := res.Data["keys"]
	assert.True(t, ok)

	err = container.Teardown()
	assert.NoError(t, err)
}
