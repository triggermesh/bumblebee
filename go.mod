module github.com/triggermesh/transformation

go 1.14

require (
	cloud.google.com/go v0.61.0 // indirect
	github.com/aws/aws-sdk-go v1.31.12 // indirect
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/go-containerregistry v0.1.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sys v0.0.0-20200720211630-cb9d2d5c5666 // indirect
	golang.org/x/tools v0.0.0-20200721223218-6123e77877b2 // indirect
	google.golang.org/genproto v0.0.0-20200722002428-88e341933a54 // indirect
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200702222342-ea4d6e985ba0
	knative.dev/serving v0.16.0
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)
