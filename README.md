[![CircleCI](https://circleci.com/gh/giantswarm/release-operator.svg?style=shield&circle-token=7b8b0735a20cc338a802eda120ae33db180bf263)](https://circleci.com/gh/giantswarm/release-operator)

# release-operator

As an example of a Release CR:

```
apiVersion: "release.giantswarm.io/v1alpha1"
kind: "Release"
metadata:
  name: "azure.v0.0.1"
  labels:
    giantswarm.io/managed-by: "app-operator"
    giantswarm.io/provider: "azure"
spec:
  changelog:
    - component: "azure-operator"
      description: "Updated to foo the bar."
      kind: "changed"
  components:
    - name: "azure-operator"
      version: "0.0.1"
  parentVersion: "0.0.0"
  version: "0.0.1"
```

and an example of a ReleaseCycle CR:

```
apiVersion: "release.giantswarm.io/v1alpha1"
kind: "ReleaseCycle"
metadata:
  name: "azure.v0.0.1"
  labels:
    giantswarm.io/managed-by: "opsctl"
    giantswarm.io/provider: "azure"
spec:
  enabledDate: "2019-01-08"
  phase: "enabled"
```