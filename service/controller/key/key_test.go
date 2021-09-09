package key

import (
	"strconv"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/v2/pkg/project"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Catalog:               "first",
		Name:                  "test",
		ReleaseOperatorDeploy: true,
		Version:               "1.0.0",
	},
	{
		Catalog:               "second",
		Name:                  "abc",
		ReleaseOperatorDeploy: true,
		Version:               "123.0.0",
	},
	{
		Catalog:               "third",
		Name:                  "other",
		ReleaseOperatorDeploy: true,
		Version:               "2.0.0",
	},
}

func Test_AppReferenced(t *testing.T) {
	testCases := []struct {
		name           string
		app            applicationv1alpha1.App
		components     map[string]releasev1alpha1.ReleaseSpecComponent
		expectedResult bool
	}{
		{
			name: "case 0: app is referenced",
			app:  ConstructApp(testComponents[1]),
			components: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testComponents[0]): testComponents[0],
				BuildAppName(testComponents[1]): testComponents[1],
				BuildAppName(testComponents[2]): testComponents[2],
			},
			expectedResult: true,
		},
		{
			name: "case 1: app is not referenced",
			app:  ConstructApp(testComponents[1]),
			components: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testComponents[0]): testComponents[0],
				BuildAppName(testComponents[2]): testComponents[2],
			},
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := AppReferenced(tc.app, tc.components)

			if !cmp.Equal(result, tc.expectedResult) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedResult, result))
			}
		})
	}
}

func Test_ConstructApp(t *testing.T) {
	testCases := []struct {
		name        string
		component   releasev1alpha1.ReleaseSpecComponent
		expectedApp applicationv1alpha1.App
	}{
		{
			name: "case 0: component has only a version (no reference)",
			component: releasev1alpha1.ReleaseSpecComponent{
				Name:    "test-operator",
				Version: "1.0.0",
			},
			expectedApp: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-operator-1.0.0",
					Namespace: Namespace,
					Labels: map[string]string{
						// TALK to team batman to find correct version!
						LabelAppOperatorVersion: "0.0.0",
						LabelManagedBy:          project.Name(),
					},
				},
				Spec: applicationv1alpha1.AppSpec{
					KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "test-operator",
					Namespace: Namespace,
					Version:   "1.0.0",
				},
			},
		},
		{
			name: "case 1: the component's ref is being used as an app version",
			component: releasev1alpha1.ReleaseSpecComponent{
				Name:      "test-operator",
				Version:   "1.0.0",
				Reference: "hello",
			},
			expectedApp: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-operator-1.0.0",
					Namespace: Namespace,
					Labels: map[string]string{
						// TALK to team batman to find correct version!
						LabelAppOperatorVersion: "0.0.0",
						LabelManagedBy:          project.Name(),
					},
				},
				Spec: applicationv1alpha1.AppSpec{
					KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "test-operator",
					Namespace: Namespace,
					Version:   "hello",
				},
			},
		},
		{
			name: "case 2: passes the component's catalog to the app",
			component: releasev1alpha1.ReleaseSpecComponent{
				Name:    "test-operator",
				Version: "1.0.0",
				Catalog: "the-catalog",
			},
			expectedApp: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-operator-1.0.0",
					Namespace: Namespace,
					Labels: map[string]string{
						// TALK to team batman to find correct version!
						LabelAppOperatorVersion: "0.0.0",
						LabelManagedBy:          project.Name(),
					},
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: "the-catalog",
					KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "test-operator",
					Namespace: Namespace,
					Version:   "1.0.0",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApp := ConstructApp(tc.component)

			if !cmp.Equal(resultApp, tc.expectedApp) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApp, resultApp))
			}
		})
	}
}

func Test_ExcludeDeletedRelease(t *testing.T) {
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

			resultReleases := ExcludeDeletedRelease(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
		})
	}
}

func Test_ExcludeDeprecatedUnusedRelease(t *testing.T) {
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

			resultReleases := ExcludeUnusedDeprecatedReleases(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
		})
	}
}

func Test_ExtractComponents(t *testing.T) {
	testCases := []struct {
		name               string
		releases           releasev1alpha1.ReleaseList
		expectedcomponents map[string]releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: extracts all components with ReleaseOperatorDeploy set",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testComponents[0],
								testComponents[1],
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testComponents[2],
							},
						},
					},
				},
			},
			expectedcomponents: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testComponents[0]): testComponents[0],
				BuildAppName(testComponents[1]): testComponents[1],
				BuildAppName(testComponents[2]): testComponents[2],
			},
		},
		{
			name: "case 1: ignores components with ReleaseOperatorDeploy set to false",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testComponents[0],
								{
									Catalog:               "hello-catalog",
									Name:                  "hello",
									ReleaseOperatorDeploy: false,
									Version:               "7.0.0",
								},
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testComponents[1],
							},
						},
					},
				},
			},
			expectedcomponents: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testComponents[0]): testComponents[0],
				BuildAppName(testComponents[1]): testComponents[1],
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultcomponents := ExtractComponents(tc.releases)

			if !cmp.Equal(resultcomponents, tc.expectedcomponents) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedcomponents, resultcomponents))
			}
		})
	}
}

func Test_FilterComponents(t *testing.T) {
	testCases := []struct {
		name               string
		components         []releasev1alpha1.ReleaseSpecComponent
		expectedcomponents []releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: filters all components with ReleaseOperatorDeploy set",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				testComponents[1],
			},
			expectedcomponents: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				testComponents[1],
			},
		},
		{
			name: "case 1: ignores components with ReleaseOperatorDeploy set to false",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				{
					Catalog:               "goodbye-catalog",
					Name:                  "goodbye",
					ReleaseOperatorDeploy: false,
					Version:               "7.0.0",
				},
			},
			expectedcomponents: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultcomponents := FilterComponents(tc.components)

			if !cmp.Equal(resultcomponents, tc.expectedcomponents) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedcomponents, resultcomponents))
			}
		})
	}
}

func Test_IsSameApp(t *testing.T) {
	testCases := []struct {
		name           string
		component      releasev1alpha1.ReleaseSpecComponent
		app            applicationv1alpha1.App
		expectedOutput bool
	}{
		{
			name:      "case 0: component is app",
			component: testComponents[0],
			app: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: BuildAppName(testComponents[0]),
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: testComponents[0].Catalog,
					Version: GetComponentRef(testComponents[0]),
				},
			},
			expectedOutput: true,
		},
		{
			name:      "case 1: component has different name",
			component: testComponents[0],
			app: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "something-else",
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: testComponents[0].Catalog,
					Version: GetComponentRef(testComponents[0]),
				},
			},
			expectedOutput: false,
		},
		{
			name:      "case 2: component has different version",
			component: testComponents[0],
			app: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: BuildAppName(testComponents[0]),
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: testComponents[0].Catalog,
					Version: "something-else",
				},
			},
			expectedOutput: false,
		},
		{
			name:      "case 3: component has different reference",
			component: testComponents[0],
			app: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: BuildAppName(testComponents[0]),
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: testComponents[0].Catalog,
					Version: "not-hello",
				},
			},
			expectedOutput: false,
		},
		{
			name:      "case 4: component has different catalog",
			component: testComponents[0],
			app: applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: BuildAppName(testComponents[0]),
				},
				Spec: applicationv1alpha1.AppSpec{
					Catalog: "something-else",
					Version: GetComponentRef(testComponents[0]),
				},
			},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := IsSameApp(tc.component, tc.app)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}

func Test_componentAppCreated(t *testing.T) {
	testCases := []struct {
		name           string
		component      releasev1alpha1.ReleaseSpecComponent
		apps           []applicationv1alpha1.App
		expectedOutput bool
	}{
		{
			name:           "case 0: component is created",
			component:      testComponents[0],
			apps:           []applicationv1alpha1.App{ConstructApp(testComponents[0])},
			expectedOutput: true,
		},
		{
			name:           "case 1: component is not created",
			component:      testComponents[0],
			apps:           []applicationv1alpha1.App{},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := ComponentAppCreated(tc.component, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}

func Test_componentAppDeployed(t *testing.T) {
	deployedApp := ConstructApp(testComponents[0])
	deployedApp.Status.Release.Status = AppStatusDeployed

	testCases := []struct {
		name           string
		component      releasev1alpha1.ReleaseSpecComponent
		apps           []applicationv1alpha1.App
		expectedOutput bool
	}{
		{
			name:           "case 0: component is created and deployed",
			component:      testComponents[0],
			apps:           []applicationv1alpha1.App{deployedApp},
			expectedOutput: true,
		},
		{
			name:           "case 1: component is created and not deployed",
			component:      testComponents[0],
			apps:           []applicationv1alpha1.App{ConstructApp(testComponents[0])},
			expectedOutput: false,
		},
		{
			name:           "case 1: component is not created",
			component:      testComponents[0],
			apps:           []applicationv1alpha1.App{},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := ComponentAppDeployed(tc.component, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}
