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
    version: 1.0.0
  - catalog: my-playground-catalog
    name: deploy-me
    reference: 1.0.1-a7663534964e4051d3ed957981c4f7885d60d15f
    releaseOperatorDeploy: true
    version: 1.0.1
  date: "2020-04-27T12:00:00Z"
  state: active
```

The only part of the release that this operator cares for is the list of components in the spec. It will make sure to deploy all of the components that have `releaseOperatorDeploy` set to true. So in the example above, it will ignore `ignore-me` and deploy `deploy-me` on the CP.

#### What does deploying mean?

So as you can see from the example above, each component can have a subset of the following fields:
* catalog: which catalog to take the component from? (e.g. control-plane-catalog, control-plane-test-catalog)
* name: name of the component.
* reference: reference of the component. Enables testing of untagged commits/branches.
* releaseOperatorDeploy: controls if this operator will deploy the component.
* version: version of the component.

On each reconcile loop release-operator does the following:
1. Iterate over all of the releases on the CP.
1. Undeploy all components currently not referred by any release.
1. For each release, deploy all the components that have `releaseOperatorDeploy` set to `true`.

Let's go into a little more details of what deploying a component actually means. For each component, release-operator will create an App CR on the CP. For example, the App CR for `deploy-me` component of the release above will look like this:

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

#### Release status

In the releases's status, you can find a `Ready` field that will tell you the current state of the release. Once a release has all of its components deployed (even the ones that release-operator is not accountable for), the value of `Ready` will change to `true`.

The status of a release is also being exported as a Prometheus metric.

