package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TODO: default namespace should be same scope for the request
const defaultNamespace = "default"

// AddNamespaceFlags adds the namespace-related flags:
// * --namespace
// * --all-namespaces
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"Namespace to use. (default \"default\")",
	)

	if allowAll {
		flags.Bool(
			"all-namespaces",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace",
		)
	}
}

// GetNamespace returns namespace from command specified by flag
func GetNamespace(cmd *cobra.Command) (string, error) {
	namespace := cmd.Flag("namespace").Value.String()
	if cmd.Flags().Lookup("all-namespaces") == nil {
		if namespace == "" {
			return defaultNamespace, nil
		}
		return namespace, nil
	}

	all, err := cmd.Flags().GetBool("all-namespaces")
	if err != nil {
		return "", err
	}
	if all {
		namespace = ""
	}
	return namespace, nil
}
