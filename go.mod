module knative.dev/client

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.5 // indirect; indirect needed by knative serving
	github.com/google/go-containerregistry v0.0.0-20191029173801-50b26ee28691 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20191016110246-af539daaa43a
	k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762
	k8s.io/cli-runtime v0.0.0-20191016113937-7693ce2cae74
	k8s.io/client-go v0.0.0-20191016110837-54936ba21026
	knative.dev/pkg v0.0.0-20191022181926-0b19b4ad9139
	knative.dev/serving v0.10.0
	knative.dev/test-infra v0.0.0-20191106182702-a6a34418ff33
	sigs.k8s.io/yaml v1.1.0
)

// Temporary pinning certain libraries. Please check periodically, whether these are still needed
// ----------------------------------------------------------------------------------------------

// Fix for `[` in help messages and shell completion code
// See https://github.com/spf13/cobra/pull/899
replace github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20191021105835-a78788917390
