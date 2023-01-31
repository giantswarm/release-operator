[![CircleCI](https://circleci.com/gh/giantswarm/release-operator.svg?style=shield&circle-token=7b8b0735a20cc338a802eda120ae33db180bf263)](https://circleci.com/gh/giantswarm/release-operator)

# release-operator

As an example of a Release CR:

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

More information on how this operator works can be found [here](docs/workflow.md).

# How to generate the CRD

When you change anything in the `api` directory you might need to regenerate the Release CRD and the generated boilerplate code.
You can do that by running `make generate`.
