module knative.dev/client

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	github.com/google/go-containerregistry v0.0.0-20190910142231-b02d448a3705 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/cli-runtime v0.16.4
	k8s.io/client-go v0.16.4
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200122022923-4e81bc3c320f
	knative.dev/serving v0.12.0
	knative.dev/test-infra v0.0.0-20200129221128-ee08b33b3cf0
	sigs.k8s.io/yaml v1.1.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------

// Fix for `[` in help messages and shell completion code
// See https://github.com/spf13/cobra/pull/899
replace github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20191021105835-a78788917390

go 1.13
