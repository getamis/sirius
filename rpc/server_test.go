//    Copyright 2018 AMIS Technologies
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package rpc

import (
	"testing"

	"github.com/getamis/sirius/rpc/mocks"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type handler struct {
	mocks.API
}

func (_m *handler) Shutdown() {
	_m.Called()
}

var _ = Describe("RPC Server", func() {
	Context("API has Shutdown function", func() {
		It("should execute Shutdown if exists", func() {
			mockHandler := new(handler)
			mockHandler.On("Shutdown").Return(nil).Once()
			mockHandler.On("Bind", mock.Anything).Return(nil).Once()
			server := NewServer(APIs(mockHandler))
			server.Shutdown()
			mockHandler.AssertExpectations(GinkgoT())
		})
	})

	Context("API has no Shutdown function", func() {
		It("should be okay even without Shutdown function", func() {
			mockHandler := new(mocks.API)
			mockHandler.On("Bind", mock.Anything).Return(nil).Once()
			server := NewServer(APIs(mockHandler))
			server.Shutdown()
			mockHandler.AssertExpectations(GinkgoT())
		})
	})
})

func TestServerSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RPC Server Suite")
}
