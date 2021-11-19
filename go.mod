module github.com/giantswarm/release-operator/v2

go 1.14

require (
	github.com/giantswarm/apiextensions-application v0.0.0-20211118184941-0e4a8fce3437
	github.com/giantswarm/config-controller v0.4.1-0.20211119173856-8e0598526059
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v6 v6.0.0
	github.com/giantswarm/k8smetadata v0.6.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v6 v6.0.0
	github.com/google/go-cmp v0.5.6
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.9.0
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1 // indirect
	k8s.io/api v0.20.12
	k8s.io/apimachinery v0.20.12
	k8s.io/client-go v0.20.12
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
)
