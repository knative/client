// Copyright © 2019 The Knative Authors
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

package revision

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
	v1 "knative.dev/client/pkg/serving/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// NewRevisionDeleteCommand represent 'revision delete' command
func NewRevisionDeleteCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags
	// prune filter, used with "-p"
	var pruneFilter string
	var pruneAll bool
	RevisionDeleteCommand := &cobra.Command{
		Use:   "delete NAME [NAME ...]",
		Short: "Delete revisions",
		Example: `
  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde

  # Delete all unreferenced revisions
  kn revision delete --prune-all

  # Delete all unreferenced revisions for a given service 'mysvc'
  kn revision delete --prune mysvc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prune := cmd.Flags().Changed("prune")
			argsLen := len(args)
			if argsLen < 1 && !pruneAll && !prune {
				return errors.New("'kn revision delete' requires one or more revision name")
			}
			if argsLen > 0 && pruneAll {
				return errors.New("'kn revision delete' with --prune-all flag requires no arguments")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}
			// Create list filters
			var params []v1.ListConfig
			if prune {
				params = append(params, v1.WithService(pruneFilter))
			}
			if prune || pruneAll {
				args, err = getUnreferencedRevisionNames(params, client)
				if err != nil {
					return err
				}
				if len(args) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No unreferenced revisions found.\n")
					return nil
				}
			}

			errs := []string{}
			for _, name := range args {
				timeout := time.Duration(0)
				if waitFlags.Wait {
					timeout = time.Duration(waitFlags.TimeoutInSeconds) * time.Second
				}
				err = client.DeleteRevision(context.TODO(), name, timeout)
				if err != nil {
					errs = append(errs, err.Error())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Revision '%s' deleted in namespace '%s'.\n", name, namespace)
				}
			}
			if len(errs) > 0 {
				return errors.New("Error: " + strings.Join(errs, "\nError: "))
			}
			return nil
		},
	}
	flags := RevisionDeleteCommand.Flags()
	flags.StringVar(&pruneFilter, "prune", "", "Remove unreferenced revisions for a given service in a namespace.")
	flags.BoolVar(&pruneAll, "prune-all", false, "Remove all unreferenced revisions in a namespace.")
	commands.AddNamespaceFlags(RevisionDeleteCommand.Flags(), false)
	waitFlags.AddConditionWaitFlags(RevisionDeleteCommand, commands.WaitDefaultTimeout, "delete", "revision", "deleted")
	return RevisionDeleteCommand
}

// Return unreferenced revision names
func getUnreferencedRevisionNames(lConfig []v1.ListConfig, client v1.KnServingClient) ([]string, error) {
	revisionList, err := client.ListRevisions(context.TODO(), lConfig...)
	if err != nil {
		return []string{}, err
	}
	// Sort revisions by namespace, service, generation (in this order)
	sortRevisions(revisionList)
	revisionNames := []string{}
	for _, revision := range revisionList.Items {
		if revision.GetRoutingState() != servingv1.RoutingStateActive {
			revisionNames = append(revisionNames, revision.Name)
		}
	}
	return revisionNames, nil
}
