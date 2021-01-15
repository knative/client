module knative.dev/client

go 1.14

require (
	github.com/google/go-cmp v0.5.4
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.1-0.20200715031239-b95db644ed1c
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	gopkg.in/ini.v1 v1.56.0 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/cli-runtime v0.18.12
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.12
	knative.dev/eventing v0.20.1-0.20210115015220-1c94ef83b6b5
	knative.dev/hack v0.0.0-20210114150620-4422dcadb3c8
	knative.dev/networking v0.0.0-20210113172032-07a8160d1971
	knative.dev/pkg v0.0.0-20210114223020-f0ea5e6b9c4e
	knative.dev/serving v0.20.1-0.20210115004319-84421cde1553
	sigs.k8s.io/yaml v1.2.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------
replace (
	k8s.io/api => k8s.io/api v0.18.12
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.12
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.12
	k8s.io/client-go => k8s.io/client-go v0.18.12
	k8s.io/code-generator => k8s.io/code-generator v0.18.12
)

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.12
