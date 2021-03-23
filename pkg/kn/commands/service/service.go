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
	"context"
	"fmt"
	"io"
	"time"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/wait"

	"github.com/spf13/cobra"
)

const (
	// How often to retry in case of an optimistic lock error when replacing a service (--force)
	MaxUpdateRetries = 3
)

func NewServiceCommand(p *commands.KnParams) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:     "service",
		Short:   "Manage Knative services",
		Aliases: []string{"ksvc", "services"},
	}
	serviceCmd.AddCommand(NewServiceListCommand(p))
	serviceCmd.AddCommand(NewServiceDescribeCommand(p))
	serviceCmd.AddCommand(NewServiceCreateCommand(p))
	serviceCmd.AddCommand(NewServiceDeleteCommand(p))
	serviceCmd.AddCommand(NewServiceUpdateCommand(p))
	serviceCmd.AddCommand(NewServiceApplyCommand(p))
	serviceCmd.AddCommand(NewServiceExportCommand(p))
	serviceCmd.AddCommand(NewServiceImportCommand(p))
	return serviceCmd
}

func waitForService(ctx context.Context, client clientservingv1.KnServingClient, serviceName string, out io.Writer, timeout int) error {
	err, duration := client.WaitForService(ctx, serviceName, time.Duration(timeout)*time.Second, wait.SimpleMessageCallback(out))
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%7.3fs Ready to serve.\n", float64(duration.Round(time.Millisecond))/float64(time.Second))
	return nil
}

func showUrl(ctx context.Context, client clientservingv1.KnServingClient, serviceName string, originalRevision string, what string, out io.Writer) error {
	service, err := client.GetService(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("cannot fetch service '%s' in namespace '%s' for extracting the URL: %w", serviceName, client.Namespace(ctx), err)
	}

	url := service.Status.URL.String()

	newRevision := service.Status.LatestReadyRevisionName
	if (originalRevision != "" && originalRevision == newRevision) || originalRevision == "unchanged" {
		fmt.Fprintf(out, "Service '%s' with latest revision '%s' (unchanged) is available at URL:\n%s\n", serviceName, newRevision, url)
	} else {
		fmt.Fprintf(out, "Service '%s' %s to latest revision '%s' is available at URL:\n%s\n", serviceName, what, newRevision, url)
	}

	return nil
}

func newServingClient(p *commands.KnParams, namespace, dir string) (clientservingv1.KnServingClient, error) {
	if dir != "" {
		return p.NewGitopsServingClient(namespace, dir)
	}
	return p.NewServingClient(namespace)
}
