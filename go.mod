module knative.dev/client

go 1.15

require (
	github.com/google/go-cmp v0.5.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/code-generator v0.21.2
	knative.dev/eventing v0.23.1-0.20210623160544-0cb787308255
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/networking v0.0.0-20210626000544-78822ee81f36
	knative.dev/pkg v0.0.0-20210625194144-4cdacd04734a
	knative.dev/serving v0.23.1
	sigs.k8s.io/yaml v1.2.0
)
