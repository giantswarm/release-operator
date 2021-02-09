package config

import (
	"sort"
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Name:    "test",
		Version: "1.0.0",
	},
	{
		Name:    "abc",
		Version: "123.0.0",
	},
	{
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

func Test_excludeDeletedRelease(t *testing.T) {
	testCases := []struct {
		name             string
		releases         releasev1alpha1.ReleaseList
		expectedReleases releasev1alpha1.ReleaseList
	}{
		{
			name: "case 0: some releases are being deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "being-deleted",
							DeletionTimestamp: &metav1.Time{},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "not-being-deleted",
							DeletionTimestamp: nil,
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultReleases := excludeDeletedRelease(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
		})
	}
}

func Test_excludeDeprecatedUnusedRelease(t *testing.T) {
	testCases := []struct {
		name             string
		releases         releasev1alpha1.ReleaseList
		expectedReleases releasev1alpha1.ReleaseList
	}{
		{
			name: "case 0: an unused, deprecated release is deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ancient-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateDeprecated,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
		{
			name: "case 1: an unused active release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "active-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateActive,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "active-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "active",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
		{
			name: "case 2: a used deprecated release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateDeprecated,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "deprecated",
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
		{
			name: "case 3: an unused wip release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateWIP,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "wip",
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			fakeK8sClient := k8sclienttest.NewEmpty()

			var newResource *Resource
			{
				c := Config{
					K8sClient: fakeK8sClient,
					Logger:    microloggertest.New(),
				}
				var err error
				newResource, err = New(c)
				if err != nil {
					t.Fatal("expected", nil, "got", err)
				}
			}

			resultReleases := newResource.excludeUnusedDeprecatedReleases(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
		})
	}
}

func configForComponent(operator releasev1alpha1.ReleaseSpecComponent) corev1alpha1.Config {
	return corev1alpha1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: key.BuildConfigName(operator),
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
