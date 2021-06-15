module github.com/triggermesh/bumblebee

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/code-generator v0.19.7 // indirect
	knative.dev/networking v0.0.0-20210603073844-5521a8b92648
	knative.dev/pkg v0.0.0-20210602095030-0e61d6763dd6
	knative.dev/serving v0.22.2
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)
