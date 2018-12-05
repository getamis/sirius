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

package postgresql

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgresql Options", func() {
	Describe("Test options", func() {
		Context("Connector", func() {
			It("should match", func() {
				address := "127.0.0.1"
				port := "3306"
				expectedOptions := &options{
					Address: address,
					Port:    port,
				}

				options := defaultOptions()
				fn := Connector(address, port)
				fn(options)

				Expect(options.Address).Should(Equal(expectedOptions.Address))
				Expect(options.Port).Should(Equal(expectedOptions.Port))
				gotConntectionString, _ := ToConnectionString(fn)
				Expect(gotConntectionString).Should(Equal(options.String()))
			})
		})

		Context("UserInfo", func() {
			It("should match", func() {
				username := "foo"
				password := "foo's password"
				expectedOptions := &options{
					UserName: username,
					Password: password,
				}

				options := defaultOptions()
				fn := UserInfo(username, password)
				fn(options)

				Expect(options.UserName).Should(Equal(expectedOptions.UserName))
				Expect(options.Password).Should(Equal(expectedOptions.Password))
				gotConntectionString, _ := ToConnectionString(fn)
				Expect(gotConntectionString).Should(Equal(options.String()))
			})
		})

		Context("SSLMode", func() {
			It("should enable", func() {
				options := defaultOptions()
				fn := SSLMode("enable")
				fn(options)

				Expect(options.SSLMode).Should(Equal("enable"))
				gotConntectionString, _ := ToConnectionString(fn)
				Expect(gotConntectionString).Should(Equal(options.String()))
			})
		})

		Context("Database", func() {
			It("should match", func() {
				dbName := "This is a database name"
				expectedOptions := &options{
					DatabaseName: dbName,
				}

				options := defaultOptions()
				fn := Database(dbName)
				fn(options)

				Expect(options.DatabaseName).Should(Equal(expectedOptions.DatabaseName))
				gotConntectionString, _ := ToConnectionString(fn)
				Expect(gotConntectionString).Should(Equal(options.String()))
			})
		})

		Context("ToConnectionString", func() {
			It("should fail, because wrong option type", func() {
				wrongOption := "wrong option"
				gotConntectionString, err := ToConnectionString(wrongOption)
				Expect(gotConntectionString).Should(BeEmpty())
				Expect(err.Error()).Should(Equal(fmt.Sprintf("Invalid option: %v", wrongOption)))
			})
		})
	})
})

func TestPostgreSQLOptionsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgreSQL Options Suite")
}
