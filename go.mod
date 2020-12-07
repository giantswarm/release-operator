module github.com/giantswarm/release-operator/v2

go 1.14

require (
	github.com/giantswarm/apiextensions/v2 v2.6.2
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.3.4
	github.com/giantswarm/operatorkit/v2 v2.0.0
	github.com/google/go-cmp v0.5.4
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	sigs.k8s.io/cluster-api v0.3.7
	sigs.k8s.io/controller-runtime v0.6.4
)

// v3.3.13 is required by github.com/spf13/viper v1.3.2, 1.6.2, 1.4.0 and github.com/bketelsen/crypt@v0.0.3-0.20200106085610-5cbc8cc4026c. Can remove this replace when updated.
replace github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible

replace (
	sigs.k8s.io/cluster-api v0.3.7 => github.com/giantswarm/cluster-api v0.3.7
	sigs.k8s.io/cluster-api-provider-azure v0.4.6 => github.com/giantswarm/cluster-api-provider-azure v0.4.6
)
