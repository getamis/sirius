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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	client "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

func TestEtcdContainer(t *testing.T) {
	container, _ := NewEtcdContainer()
	assert.NotNil(t, container)
	assert.NoError(t, container.Start())
	defer container.Stop()

	cfg := client.Config{
		Endpoints: []string{container.URL},
		// set timeout per request to fail fast when the target endpoint is unavailable
		DialTimeout: time.Second,
	}

	c, err := client.New(cfg)
	assert.NoError(t, err, "should be no error")

	kapi := client.NewKV(c)
	_, err = kapi.Put(context.Background(), "/foo", "bar")
	assert.NoError(t, err, "should be no error")
}
