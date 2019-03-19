package commands

import "github.com/spf13/pflag"

// AddNamespaceFlags adds the namespace-related flags:
// * --namespace
// * --all-namespaces
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this request",
	)

	if allowAll {
		flags.Bool(
			"all-namespaces",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace",
		)
	}
}
