// Copyright © 2018 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/getamis/sirius/health"
)

const (
	defaultDialTimeout = 5 * time.Second
)

// ReadinessCmd represents the readiness command
var ReadinessCmd = &cobra.Command{
	Use:          "readiness",
	Short:        "readiness checks the readiness of remote grpc server",
	Long:         `readiness checks the readiness of remote grpc server`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		grpcAddr := fmt.Sprintf("%s:%d", host, port)

		// dial remote server
		ctx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
		defer cancel()
		conn, err := grpc.DialContext(ctx, grpcAddr,
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		if err != nil {
			return err
		}
		defer conn.Close()
		c := health.NewHealthCheckServiceClient(conn)
		_, err = c.Readiness(ctx, nil)
		return err
	},
}
