## Release operator

### How does release-operator work?
Release operator reconciles on release CRs. An example of a valid release looks like this:

```
apiVersion: release.giantswarm.io/v1alpha1
kind: Release
metadata:
  creationTimestamp: null
  name: v11.3.0
spec:
  apps:
  - name: app-1
    version: 1.2.1
  components:
  - name: ignore-me
    releaseOperatorDeploy: false
    version: 1.0.0
  - catalog: my-playground-catalog
    name: deploy-me
    reference: 1.0.1-a7663534964e4051d3ed957981c4f7885d60d15f
    releaseOperatorDeploy: true
    version: 1.0.1
  date: "2020-04-27T12:00:00Z"
  state: active
```

The only part of the release that this operator cares for is the list of components in the spec. It will make sure to deploy all of the components that have `releaseOperatorDeploy`
set to true. So in the example above, it will ignore `ignore-me` and deploy `deploy-me` on the CP.

#### What does deploying mean?

So as you can see from the example above, each component can have a subset of the following fields:
* `catalog`: which catalog to take the component from? (e.g. control-plane-catalog, control-plane-test-catalog)
* `name`: name of the component.
* `reference`: reference of the component. A reference points to a tagged version of a component (e.g. 0.1.0, 0.1.0-1) with an optional SHA suffix
(e.g. 0.1.0-1078ad9d2c15178d1466f79f1a54ebd9c92d9614) to specify a commit. Used for testing and referring to alternative versions of existing components.
* `releaseOperatorDeploy`: controls if this operator will deploy the component.
* `version`: version of the component.

On each reconciliation loop, release-operator does the following:
1. Iterate over all of the Release CRs on the CP.
1. Undeploy all components currently not referenced by any release.
1. Deploy all components referenced by at least one release that have `releaseOperatorDeploy` set to `true`.

Let's go into a little more details of what deploying a component actually means. For each component, release-operator will create an App CR on the CP. For example, the App CR
for `deploy-me` component of the release above will look like this:

```
apiVersion: application.giantswarm.io/v1alpha1
kind: App
metadata:
  creationTimestamp: null
  labels:
    app-operator.giantswarm.io/version: 1.0.0
    giantswarm.io/managed-by: release-operator
  name: deploy-me-1.0.1
  namespace: giantswarm
spec:
  catalog: my-playground-catalog
  name: deploy-me-1.0.1
  namespace: giantswarm
  kubeConfig:
    inCluster: true
  version: 1.0.1-a7663534964e4051d3ed957981c4f7885d60d15f
```

A few key points here:
* the app name is a concatenation of the component name and version.
* reference is being passed through as version in the App CR. If no reference is being used, then release-operator will default to using the component version.
* for every app, `inCluster` is being set to `true` in the `kubeConfig`.

It's also important to notice that `release-operator` is only responsible for creating the App CRs. `app-operator` and `chart-operator` then take over and deploy the corresponding Helm charts.

#### Release status

In the releases's status, you can find a `Ready` field that will tell you the current state of the release. The value changes to `true` once all the App CRs for components marked with `releaseOperatorDeploy` are present on the CP.

The status of a release is being exported as a Prometheus metric. There is also an
[alert](https://github.com/giantswarm/g8s-prometheus/blob/master/helm/g8s-prometheus/prometheus-rules/release.rules.yml) that will page if a release spends more than 30 minutes in a non-ready state.

