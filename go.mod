module knative.dev/client

go 1.15

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
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/code-generator v0.19.7
	knative.dev/eventing v0.20.1-0.20210127134430-ea452b5b6655
	knative.dev/hack v0.0.0-20210120165453-8d623a0af457
	knative.dev/networking v0.0.0-20210123150554-8b522a9049a1
	knative.dev/pkg v0.0.0-20210125222030-6040b3af4803
	knative.dev/serving v0.20.1-0.20210123202654-e61294b2ca32
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3
