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
	"time"

	"google.golang.org/grpc"

	"github.com/getamis/sirius/log"
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
