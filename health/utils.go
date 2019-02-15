// Copyright 2018 AMIS Technologies
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

package health

import (
	"context"
	"net/http"
	"time"

	"github.com/getamis/sirius/log"
	"google.golang.org/grpc"
)

const (
	defaultDialTimeout = 5 * time.Second
)

func GRPCServerHealthChecker(addr, serviceName string) CheckFn {
	return func(ctx context.Context) error {
		dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
		defer cancel()
		conn, err := grpc.DialContext(dialCtx, addr,
			grpc.WithInsecure(),
			grpc.WithBlock())
		if err != nil {
			log.Error("Failed to dial grpc server", "addr", addr, "serviceName", serviceName, "err", err)
			return err
		}
		conn.Close()
		return nil
	}
}

func CheckHealth(ctx context.Context, checkFns []CheckFn) error {
	if len(checkFns) == 0 {
		return nil
	}
	errCh := make(chan error, len(checkFns))
	for _, checker := range checkFns {
		go func(checker CheckFn) {
			errCh <- checker(ctx)
		}(checker)
	}
	for range checkFns {
		if err := <-errCh; err != nil {
			log.Error("Failed to check readiness", "err", err)
			return err
		}
	}
	return nil
}

// SetLivenessAndReadiness sets the liveness and readiness route in http server mux
func SetLivenessAndReadiness(mux *http.ServeMux, checkFns ...CheckFn) {
	mux.HandleFunc("/readiness", func(rw http.ResponseWriter, r *http.Request) {
		if err := CheckHealth(r.Context(), checkFns); err != nil {
			log.Error("Failed to check readiness", "err", err)
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		rw.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/liveness", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
}
