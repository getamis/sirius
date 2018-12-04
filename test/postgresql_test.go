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
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLSetupAndTeardown(t *testing.T) {
	postgresql, err := SetupPostgreSQL()
	assert.NoError(t, err, "postgresql connection handle should be created.")
	assert.NotNil(t, postgresql, "the postgresql container should be returned.")

	db, err := gorm.Open("postgres", postgresql.URL)
	assert.NoError(t, err, "postgresql connection should work")
	db.Close()

	err = postgresql.Teardown()
	assert.NoError(t, err, "postgresql connection handle should be torn down.")
}

func TestPostgreSQLContainer(t *testing.T) {
	if _, ok := os.LookupEnv("TEST_POSTGRESQL_HOST"); ok {
		t.Skip("postgresql container test is ignored when postgresql host is enabled.")
	}

	options := LoadPostgreSQLOptions()

	container, err := NewPostgreSQLContainer(options)
	assert.NoError(t, err, "postgresql container should be created.")
	assert.NotNil(t, container)
	assert.NoError(t, container.Start(), "postgresql container should be started")

	db, err := gorm.Open("postgres", container.URL)
	assert.NoError(t, err, "postgresql connection should work")
	db.Close()

	// stop postgresql
	assert.NoError(t, container.Suspend())
	time.Sleep(100 * time.Millisecond)
	_, err = gorm.Open("postgres", container.URL)
	assert.Error(t, err, "should got error")

	// restart postgresql
	assert.NoError(t, container.Start())
	db, err = gorm.Open("postgres", container.URL)
	assert.NoError(t, err, "should be no error")
	db.Close()

	// close MySQL
	assert.NoError(t, container.Stop())
	time.Sleep(100 * time.Millisecond)

	_, err = gorm.Open("postgres", container.URL)
	assert.Error(t, err, "should got error")
}
