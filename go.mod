module github.com/giantswarm/release-operator

go 1.14

require (
	github.com/giantswarm/apiextensions/v2 v2.1.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.2-0.20200901165545-4a16870c8303
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit/v2 v2.0.0
	github.com/google/go-cmp v0.5.1
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/viper v1.7.1
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	sigs.k8s.io/controller-runtime v0.6.1
)

replace github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.3
