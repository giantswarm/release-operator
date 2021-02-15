package configs

import (
	"sort"
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	apiexlabels "github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/v2/pkg/project"
	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Catalog: "test-catalog",
		Name:    "test",
		Version: "1.0.0",
	},
	{
		Catalog: "test-catalog",
		Name:    "abc",
		Version: "123.0.0",
	},
	{
		Catalog: "test-catalog",
		Name:    "other",
		Version: "2.0.0",
	},
}

func Test_calculateMissingConfigs(t *testing.T) {
	testCases := []struct {
		name            string
		operators       map[string]releasev1alpha1.ReleaseSpecComponent
		configs         corev1alpha1.ConfigList
		expectedConfigs corev1alpha1.ConfigList
	}{
		{
			name: "case 0: an config is missing",
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				key.BuildConfigName(testComponents[0]): testComponents[0],
				key.BuildConfigName(testComponents[1]): testComponents[1],
				key.BuildConfigName(testComponents[2]): testComponents[2],
			},
			configs: corev1alpha1.ConfigList{
				Items: []corev1alpha1.Config{
					configForComponent(testComponents[0]),
				},
			},

			expectedConfigs: corev1alpha1.ConfigList{
				Items: []corev1alpha1.Config{
					key.ConstructConfig(testComponents[1]),
					key.ConstructConfig(testComponents[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultConfigs := calculateMissingConfigs(tc.operators, tc.configs)

			sortSlice := func(slice []corev1alpha1.Config) {
				sort.SliceStable(slice, func(i, j int) bool { return slice[i].Name < slice[j].Name })
			}

			// calculateMissingConfigs iterates over a map -> random order
			sortSlice(tc.expectedConfigs.Items)
			sortSlice(resultConfigs.Items)

			if !cmp.Equal(resultConfigs, tc.expectedConfigs) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedConfigs, resultConfigs))
			}
		})
	}
}

func Test_calculateObsoleteConfigs(t *testing.T) {
	testCases := []struct {
		name            string
		operators       map[string]releasev1alpha1.ReleaseSpecComponent
		configs         corev1alpha1.ConfigList
		expectedConfigs corev1alpha1.ConfigList
	}{
		{
			name: "case 0: there is an obsolete config",
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				key.BuildConfigName(testComponents[0]): testComponents[0],
				key.BuildConfigName(testComponents[1]): testComponents[1],
			},
			configs: corev1alpha1.ConfigList{
				Items: []corev1alpha1.Config{
					key.ConstructConfig(testComponents[0]),
					key.ConstructConfig(testComponents[1]),
					key.ConstructConfig(testComponents[2]),
				},
			},
			expectedConfigs: corev1alpha1.ConfigList{
				Items: []corev1alpha1.Config{
					key.ConstructConfig(testComponents[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultConfigs := calculateObsoleteConfigs(tc.operators, tc.configs)

			if !cmp.Equal(resultConfigs, tc.expectedConfigs) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedConfigs, resultConfigs))
			}
		})
	}
}

func configForComponent(operator releasev1alpha1.ReleaseSpecComponent) corev1alpha1.Config {
	return corev1alpha1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: key.BuildConfigName(operator),
			Labels: map[string]string{
				apiexlabels.ConfigControllerVersion: "0.0.0",
				key.LabelManagedBy:                  project.Name(),
			},
		},
		Spec: corev1alpha1.ConfigSpec{
			App: corev1alpha1.ConfigSpecApp{
				Catalog: operator.Catalog,
				Name:    operator.Name,
				Version: operator.Version,
			},
		},
	}
}
