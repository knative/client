module knative.dev/client

require (
	github.com/google/go-containerregistry v0.0.0-20200413145205-82d30a103c0a // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.14.1-0.20200523184044-78d7fbb41f8a
	knative.dev/pkg v0.0.0-20200522212244-870993f63e81
	knative.dev/serving v0.14.1-0.20200524222346-2b805814b468
	knative.dev/test-infra v0.0.0-20200522180958-6a0a9b9d893a
	sigs.k8s.io/yaml v1.2.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------

// Fix for `[` in help messages and shell completion code
// See https://github.com/spf13/cobra/pull/899
replace (
	github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20191021105835-a78788917390

	k8s.io/api => k8s.io/api v0.17.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.4
	k8s.io/client-go => k8s.io/client-go v0.17.4
)

go 1.13
