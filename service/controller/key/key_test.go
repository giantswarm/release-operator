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

var testOperators = []releasev1alpha1.ReleaseSpecComponent{
	{
		Name:    "test-operator",
		Version: "1.0.0",
	},
	{
		Name:    "abc-operator",
		Version: "123.0.0",
	},
	{
		Name:    "other-operator",
		Version: "2.0.0",
	},
	{
		Name:    "not-exist-operator",
		Version: "7.0.0",
	},
}

func Test_AppReferenced(t *testing.T) {
	testCases := []struct {
		name           string
		app            applicationv1alpha1.App
		operators      map[string]releasev1alpha1.ReleaseSpecComponent
		expectedResult bool
	}{
		{
			name: "case 0: app is referenced",
			app:  ConstructApp(testOperators[1]),
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
				BuildAppName(testOperators[1]): testOperators[1],
				BuildAppName(testOperators[2]): testOperators[2],
			},
			expectedResult: true,
		},
		{
			name: "case 1: app is not referenced",
			app:  ConstructApp(testOperators[1]),
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
				BuildAppName(testOperators[2]): testOperators[2],
			},
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := AppReferenced(tc.app, tc.operators)

			if !cmp.Equal(result, tc.expectedResult) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedResult, result))
			}
		})
	}
}

func Test_ConstructApp(t *testing.T) {
	testCases := []struct {
		name        string
		operator    releasev1alpha1.ReleaseSpecComponent
		expectedApp applicationv1alpha1.App
	}{
		{
			name: "case 0: operator has only a version (no reference)",
			operator: releasev1alpha1.ReleaseSpecComponent{
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
			name: "case 1: the operator's ref is being used as an app version",
			operator: releasev1alpha1.ReleaseSpecComponent{
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
			operator: releasev1alpha1.ReleaseSpecComponent{
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

			resultApp := ConstructApp(tc.operator)

			if !cmp.Equal(resultApp, tc.expectedApp) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApp, resultApp))
			}
		})
	}
}

func Test_ExtractAllOperators(t *testing.T) {
	testCases := []struct {
		name              string
		releases          releasev1alpha1.ReleaseList
		expectedOperators map[string]releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: extracts all operators",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[0],
								testOperators[1],
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[2],
							},
						},
					},
				},
			},
			expectedOperators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
				BuildAppName(testOperators[1]): testOperators[1],
				BuildAppName(testOperators[2]): testOperators[2],
			},
		},
		{
			name: "case 1: does not extract non-operators",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[0],
								{
									Name:    "something-else",
									Version: "9.0.0",
								},
							},
						},
					},
				},
			},
			expectedOperators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultOperators := ExtractAllOperators(tc.releases)

			if !cmp.Equal(resultOperators, tc.expectedOperators) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOperators, resultOperators))
			}
		})
	}
}

func Test_ExtractOperators(t *testing.T) {
	testCases := []struct {
		name              string
		components        []releasev1alpha1.ReleaseSpecComponent
		expectedOperators []releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: extracts the operators",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testOperators[0],
				testOperators[1],
				{Name: "something-totally-else"},
			},
			expectedOperators: []releasev1alpha1.ReleaseSpecComponent{
				testOperators[0],
				testOperators[1],
			},
		},
		{
			name: "case 0: does not extract other things",
			components: []releasev1alpha1.ReleaseSpecComponent{
				testOperators[0],
				{Name: "something-totally-else"},
			},
			expectedOperators: []releasev1alpha1.ReleaseSpecComponent{
				testOperators[0],
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultOperators := ExtractOperators(tc.components)

			if !cmp.Equal(resultOperators, tc.expectedOperators) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOperators, resultOperators))
			}
		})
	}
}

func Test_IsOperator(t *testing.T) {
	testCases := []struct {
		name           string
		operator       releasev1alpha1.ReleaseSpecComponent
		expectedOutput bool
	}{
		{
			name:           "case 0: is an operator",
			operator:       releasev1alpha1.ReleaseSpecComponent{Name: "i-am-operator"},
			expectedOutput: true,
		},
		{
			name:           "case 1: only contains the word operator",
			operator:       releasev1alpha1.ReleaseSpecComponent{Name: "icontainoperator"},
			expectedOutput: false,
		},
		{
			name:           "case 2: is not an operator",
			operator:       releasev1alpha1.ReleaseSpecComponent{Name: "ignoreme"},
			expectedOutput: false,
		},
		{
			name:           "case 3: ignores chart-operator",
			operator:       releasev1alpha1.ReleaseSpecComponent{Name: "chart-operator"},
			expectedOutput: false,
		},
		{
			name:           "case 4: ignores app-operator",
			operator:       releasev1alpha1.ReleaseSpecComponent{Name: "app-operator"},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := IsOperator(tc.operator)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}

func Test_OperatorCreated(t *testing.T) {
	testCases := []struct {
		name           string
		operator       releasev1alpha1.ReleaseSpecComponent
		apps           []applicationv1alpha1.App
		expectedOutput bool
	}{
		{
			name:           "case 0: operator is created",
			operator:       testOperators[0],
			apps:           []applicationv1alpha1.App{ConstructApp(testOperators[0])},
			expectedOutput: true,
		},
		{
			name:           "case 1: operator is not created",
			operator:       testOperators[0],
			apps:           []applicationv1alpha1.App{},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := OperatorCreated(tc.operator, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}

func Test_OperatorDeployed(t *testing.T) {
	deployedApp := ConstructApp(testOperators[0])
	deployedApp.Status.Release.Status = AppStatusDeployed

	testCases := []struct {
		name           string
		operator       releasev1alpha1.ReleaseSpecComponent
		apps           []applicationv1alpha1.App
		expectedOutput bool
	}{
		{
			name:           "case 0: operator is created and deployed",
			operator:       testOperators[0],
			apps:           []applicationv1alpha1.App{deployedApp},
			expectedOutput: true,
		},
		{
			name:           "case 1: operator is created and not deployed",
			operator:       testOperators[0],
			apps:           []applicationv1alpha1.App{ConstructApp(testOperators[0])},
			expectedOutput: false,
		},
		{
			name:           "case 1: operator is not created",
			operator:       testOperators[0],
			apps:           []applicationv1alpha1.App{},
			expectedOutput: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := OperatorDeployed(tc.operator, tc.apps)

			if !cmp.Equal(result, tc.expectedOutput) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOutput, result))
			}
		})
	}
}
