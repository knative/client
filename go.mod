module knative.dev/client

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20200304201134-fcc8ea80e26f // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v0.17.0
	knative.dev/eventing v0.13.4
	knative.dev/pkg v0.0.0-20200323231609-0840da9555a3
	knative.dev/serving v0.13.1-0.20200324164309-ce10070bb7ef
	knative.dev/test-infra v0.0.0-20200324183109-d81c66eea4e6
	sigs.k8s.io/yaml v1.1.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------

// Fix for `[` in help messages and shell completion code
// See https://github.com/spf13/cobra/pull/899
replace github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20191021105835-a78788917390

go 1.13
