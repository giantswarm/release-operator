package status

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/v3/service/controller/key"
)

const (
	// from https://github.com/giantswarm/kvm-operator/blob/9dc5f0d8075731c600e2852b27e71dbc2e91015d/service/controller/key/key.go#L123
	PodWatcherLabel = "kvm-operator.giantswarm.io/pod-watcher"
	// from https://github.com/giantswarm/kvm-operator/blob/9dc5f0d8075731c600e2852b27e71dbc2e91015d/service/controller/key/key.go#L101
	KVMVersionBundleVersionAnnotation = "kvm-operator.giantswarm.io/version-bundle"
	// from https://github.com/giantswarm/kvm-operator/blob/eee64f540cae53d530628d50e54883b636d0693f/pkg/label/label.go#L17
	KVMOperatorVersionLabel = "kvm-operator.giantswarm.io/version"
)

func (r *Resource) getKVMOperatorVersionPodsExist(ctx context.Context, operatorVersion string) (bool, error) {
	pods, err := r.k8sClient.K8sClient().CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", PodWatcherLabel, key.ProviderOperatorKVM),
	})
	if err != nil {
		return false, microerror.Mask(err)
	}

	for _, pod := range pods.Items {
		if pod.Annotations[KVMVersionBundleVersionAnnotation] == operatorVersion ||
			pod.Labels[KVMOperatorVersionLabel] == operatorVersion {
			return true, nil
		}
	}

	return false, nil
}
