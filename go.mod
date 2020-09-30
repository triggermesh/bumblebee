module github.com/triggermesh/bumblebee

go 1.14

require (
	cloud.google.com/go v0.61.0 // indirect
	github.com/aws/aws-sdk-go v1.31.12 // indirect
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/go-containerregistry v0.1.3 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sys v0.0.0-20200720211630-cb9d2d5c5666 // indirect
	golang.org/x/tools v0.0.0-20200924205911-8a9a89368bd3 // indirect
	google.golang.org/genproto v0.0.0-20200722002428-88e341933a54 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.18.0
	knative.dev/networking v0.0.0-20200922180040-a71b40c69b15
	knative.dev/pkg v0.0.0-20200929211029-1e373a9e5dea
	knative.dev/serving v0.18.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)
