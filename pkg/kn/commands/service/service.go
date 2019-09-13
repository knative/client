// Copyright © 2018 The Knative Authors
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

package service

import (
	"io"
	"time"

	"knative.dev/client/pkg/kn/commands"
	serving_kn_v1alpha1 "knative.dev/client/pkg/serving/v1alpha1"

	"fmt"

	"github.com/spf13/cobra"
)

const (
	// How often to retry in case of an optimistic lock error when replacing a service (--force)
	MaxUpdateRetries = 3
)

func NewServiceCommand(p *commands.KnParams) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Service command group",
	}
	serviceCmd.AddCommand(NewServiceListCommand(p))
	serviceCmd.AddCommand(NewServiceDescribeCommand(p))
	serviceCmd.AddCommand(NewServiceCreateCommand(p))
	serviceCmd.AddCommand(NewServiceDeleteCommand(p))
	serviceCmd.AddCommand(NewServiceUpdateCommand(p))
	serviceCmd.AddCommand(NewServiceMigrateCommand(p))
	return serviceCmd
}

func waitForService(client serving_kn_v1alpha1.KnClient, serviceName string, out io.Writer, timeout int) error {
	fmt.Fprintf(out, "Waiting for service '%s' to become ready ... ", serviceName)
	flush(out)

	err := client.WaitForService(serviceName, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Fprintln(out)
		return err
	}
	fmt.Fprintln(out, "OK")
	return nil
}
