package key

import (
	"strconv"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/pkg/project"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Name:    "test-operator",
		Version: "1.0.0",
		Catalog: "first",
	},
	{
		Name:    "abc-operator",
		Version: "123.0.0",
		Catalog: "second",
	},
	{
		Name:    "other-operator",
		Version: "2.0.0",
		Catalog: "third",
	},
	{
		Name:    "not-exist-operator",
		Version: "7.0.0",
		Catalog: "fourth",
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
					Name:      "test-operator-1.0.0-hackathon",
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
					Name:      "test-operator-1.0.0-hackathon",
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
					Name:      "test-operator-1.0.0-hackathon",
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

func Test_ExtractComponents(t *testing.T) {
	testCases := []struct {
		name               string
		releases           releasev1alpha1.ReleaseList
		expectedcomponents map[string]releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: extracts all operators",
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
			name: "case 1: does not extract non-operators",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testComponents[0],
								{
									Name:    "something-else",
									Version: "9.0.0",
								},
							},
						},
					},
				},
			},
			expectedcomponents: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testComponents[0]): testComponents[0],
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
			name: "case 0: extracts the operators",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				testComponents[1],
				{Name: "something-totally-else"},
			},
			expectedcomponents: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				testComponents[1],
			},
		},
		{
			name: "case 0: does not extract other things",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testComponents[0],
				{Name: "something-totally-else"},
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

func Test_IsOperator(t *testing.T) {
	testCases := []struct {
		name           string
		component      releasev1alpha1.ReleaseSpecComponent
		expectedOutput bool
	}{
		{
			name:           "case 0: is an operator",
			component:      releasev1alpha1.ReleaseSpecComponent{Name: "i-am-operator"},
			expectedOutput: true,
		},
		{
			name:           "case 1: only contains the word operator",
			component:      releasev1alpha1.ReleaseSpecComponent{Name: "icontainoperator"},
			expectedOutput: false,
		},
		{
			name:           "case 2: is not an operator",
			component:      releasev1alpha1.ReleaseSpecComponent{Name: "ignoreme"},
			expectedOutput: false,
		},
		{
			name:           "case 3: ignores chart-operator",
			component:      releasev1alpha1.ReleaseSpecComponent{Name: "chart-operator"},
			expectedOutput: false,
		},
		{
			name:           "case 4: ignores app-component",
			component:      releasev1alpha1.ReleaseSpecComponent{Name: "app-operator"},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := IsOperator(tc.component)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
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

func Test_componentCreated(t *testing.T) {
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

			result := ComponentCreated(tc.component, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}

func Test_componentDeployed(t *testing.T) {
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

			result := ComponentDeployed(tc.component, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}
