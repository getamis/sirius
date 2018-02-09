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
	"github.com/spf13/cobra"
)

var (
	host string
	port int
)

var HealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health is a grpc client to health check remote grpc server",
	Long:  `Health is a grpc client to invoke livess and readiness functions of remote grpc server`,
}

func init() {
	HealthCmd.AddCommand(LivenessCmd)
	HealthCmd.AddCommand(ReadinessCmd)

	// add persistent flags for all subcommands
	HealthCmd.PersistentFlags().StringVar(&host, "host", "localhost", "Host of remote gRPC server")
	HealthCmd.PersistentFlags().IntVar(&port, "port", 8080, "Port of remote gRPC server")
}
