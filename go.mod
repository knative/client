module knative.dev/client

go 1.15

require (
	github.com/google/go-cmp v0.5.6
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	gopkg.in/ini.v1 v1.56.0 // indirect
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/cli-runtime v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/code-generator v0.20.7
	knative.dev/eventing v0.23.1-0.20210623160544-0cb787308255
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/networking v0.0.0-20210622182128-53f45d6d2cfa
	knative.dev/pkg v0.0.0-20210622173328-dd0db4b05c80
	knative.dev/serving v0.23.1-0.20210623204144-c54eeabbb984
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3
