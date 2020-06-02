module knative.dev/client

require (
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.15.1-0.20200528220601-a61a6784a053
	knative.dev/pkg v0.0.0-20200528142800-1c6815d7e4c9
	knative.dev/serving v0.15.1-0.20200601175503-4eab87b2ad07
	sigs.k8s.io/yaml v1.2.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------

// Fix for `[` in help messages and shell completion code
// See https://github.com/spf13/cobra/pull/899
replace (
	github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20191021105835-a78788917390
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
)

go 1.13
