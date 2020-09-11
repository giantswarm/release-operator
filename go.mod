module github.com/giantswarm/release-operator

go 1.13

require (
	github.com/giantswarm/apiextensions v0.4.15
	// github.com/giantswarm/apiextensions v0.4.20
	github.com/giantswarm/apiextensions/v2 v2.1.1-0.20200911094856-08012f5d7754
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit/v2 v2.0.0
	github.com/google/go-cmp v0.5.2
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/viper v1.7.1
	// giantswarm/cluster-api-provider-azure v0.4.6
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	sigs.k8s.io/cluster-api-provider-azure v0.4.6
	sigs.k8s.io/controller-runtime v0.6.1
)

// v3.3.17 is required by sigs.k8s.io/controller-runtime v0.5.2. Can remove this replace when updated.
replace github.com/coreos/etcd v3.3.17+incompatible => github.com/coreos/etcd v3.3.24+incompatible

replace (
	sigs.k8s.io/cluster-api v0.3.7 => github.com/giantswarm/cluster-api v0.3.7
	sigs.k8s.io/cluster-api-provider-azure v0.4.6 => github.com/giantswarm/cluster-api-provider-azure v0.4.6
)
