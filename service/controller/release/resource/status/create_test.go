package status

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/k8sclient/v4/pkg/k8sclienttest"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var testClusters = []tenantCluster{
	{
		ID:               "abc12",
		OperatorVersion:  "1.2.3",
		ProviderOperator: "test-operator",
		ReleaseVersion:   "9.8.7",
	},
	{
		ID:               "def34",
		OperatorVersion:  "4.5.6",
		ProviderOperator: "test-operator",
		ReleaseVersion:   "8.7.6",
	},
	// These next two clusters differ only in their IDs.
	// The versions should be deduplicated in the result map.
	{
		ID:               "ghi56",
		OperatorVersion:  "1.2.1-4",
		ProviderOperator: "another-operator",
		ReleaseVersion:   "9.8.9-1",
	},
	{
		ID:               "jkl78",
		OperatorVersion:  "1.2.1-4",
		ProviderOperator: "another-operator",
		ReleaseVersion:   "9.8.9-1",
	},
}

func Test_consolidateClusterVersions(t *testing.T) {
	testCases := []struct {
		name              string
		clusters          []tenantCluster
		expectedReleases  map[string]bool
		expectedOperators map[string]map[string]bool
	}{
		{
			name:     "case 0: release and operator versions are all present in map",
			clusters: testClusters,
			expectedReleases: map[string]bool{
				"9.8.7":   true,
				"8.7.6":   true,
				"9.8.9-1": true,
			},
			expectedOperators: map[string]map[string]bool{
				"test-operator": map[string]bool{
					"1.2.3": true,
					"4.5.6": true,
				},
				"another-operator": map[string]bool{
					"1.2.1-4": true,
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultReleases, resultOperators := consolidateClusterVersions(tc.clusters)
			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
			if !cmp.Equal(resultOperators, tc.expectedOperators) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOperators, resultOperators))
			}
		})
	}
}

func Test_getKVMOperatorVersionPodsExist(t *testing.T) {
	testCases := []struct {
		name            string
		pods            []runtime.Object
		operatorVersion string
		expectedValue   bool
	}{
		{
			name: "case 0: pre-k8s 1.18 kvm-operator with matching pod",
			pods: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							PodWatcherLabel: "kvm-operator",
						},
						Annotations: map[string]string{
							KVMVersionBundleVersionAnnotation: "1.0.0",
						},
					},
				},
			},
			operatorVersion: "1.0.0",
			expectedValue:   true,
		},
		{
			name: "case 1: k8s 1.18 kvm-operator with matching pod",
			pods: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							PodWatcherLabel:         "kvm-operator",
							KVMOperatorVersionLabel: "1.0.0",
						},
					},
				},
			},
			operatorVersion: "1.0.0",
			expectedValue:   true,
		},
		{
			name: "case 2: pre-k8s 1.18 kvm-operator without matching pod",
			pods: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							PodWatcherLabel: "kvm-operator",
						},
						Annotations: map[string]string{
							KVMVersionBundleVersionAnnotation: "1.0.1",
						},
					},
				},
			},
			operatorVersion: "1.0.0",
			expectedValue:   false,
		},
		{
			name: "case 3: k8s 1.18 kvm-operator without matching pod",
			pods: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							PodWatcherLabel:         "kvm-operator",
							KVMOperatorVersionLabel: "1.0.1",
						},
					},
				},
			},
			operatorVersion: "1.0.0",
			expectedValue:   false,
		},
		{
			name:            "case 4: no pods",
			pods:            nil,
			operatorVersion: "1.0.0",
			expectedValue:   false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			fakeK8sClient := k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
				K8sClient: fake.NewSimpleClientset(tc.pods...),
			})

			r := Resource{
				k8sClient: fakeK8sClient,
			}
			result, err := r.getKVMOperatorVersionPodsExist(context.Background(), tc.operatorVersion)
			if !cmp.Equal(err, nil) {
				t.Fatalf("\n\n%s\n", cmp.Diff(err, nil))
			}
			if !cmp.Equal(result, tc.expectedValue) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedValue, result))
			}
		})
	}
}
