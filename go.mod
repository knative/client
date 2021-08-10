module knative.dev/client

go 1.16

require (
	github.com/google/go-cmp v0.5.6
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/cli-runtime v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/code-generator v0.20.7
	knative.dev/eventing v0.25.0
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/networking v0.0.0-20210803181815-acdfd41c575c
	knative.dev/pkg v0.0.0-20210803160015-21eb4c167cc5
	knative.dev/serving v0.25.0
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3
