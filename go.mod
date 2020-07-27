module github.com/triggermesh/transformation-prototype

go 1.14

require (
	github.com/cloudevents/sdk-go v1.0.0
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/google/licenseclassifier v0.0.0-20200708223521-3d09a0ea2f39
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/qntfy/jsonparser v1.0.2 // indirect
	github.com/qntfy/kazaam v3.4.8+incompatible
	go.uber.org/zap v1.14.1
	gopkg.in/qntfy/kazaam.v3 v3.4.8 // indirect
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.0
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	knative.dev/pkg v0.0.0-20200723060257-ae9c3f7fa8d3
	knative.dev/serving v0.16.0
	knative.dev/test-infra v0.0.0-20200722142057-3ca910b5a25e
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)
