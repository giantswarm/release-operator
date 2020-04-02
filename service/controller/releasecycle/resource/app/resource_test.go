package app

import (
	"reflect"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/release-operator/pkg/testfixture"
)

const (
	defaultApp       = "app"
	defaultNamespace = "default"
)

func Test_getAppCR(t *testing.T) {
	testCases := []struct {
		name           string
		inputAppCRs    []*applicationv1alpha1.App
		inputName      string
		inputNamespace string
		expectedAppCR  *applicationv1alpha1.App
		expectedOK     bool
	}{
		{
			name: "case 0: same name different namespace",
			inputAppCRs: []*applicationv1alpha1.App{
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = defaultApp
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-8"
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = defaultApp
					appCR.Namespace = "non-default"
				}),
			},
			inputName:      defaultApp,
			inputNamespace: "non-default",
			expectedAppCR: testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
				appCR.Name = defaultApp
				appCR.Namespace = "non-default"
			}),
			expectedOK: true,
		},
		{
			name: "case 1: nonexistent namespace",
			inputAppCRs: []*applicationv1alpha1.App{
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-0"
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-1"
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-2"
					appCR.Namespace = defaultNamespace
				}),
			},
			inputName:      "app-2",
			inputNamespace: "no-such-namespace",
			expectedAppCR:  nil,
			expectedOK:     false,
		},
		{
			name: "case 2: nonexistent name",
			inputAppCRs: []*applicationv1alpha1.App{
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-0"
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-1"
					appCR.Namespace = defaultNamespace
				}),
				testfixture.NewAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-2"
					appCR.Namespace = defaultNamespace
				}),
			},
			inputName:      defaultApp,
			inputNamespace: defaultNamespace,
			expectedAppCR:  nil,
			expectedOK:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appCR, ok := getAppCR(tc.inputAppCRs, tc.inputNamespace, tc.inputName)

			if !reflect.DeepEqual(appCR, tc.expectedAppCR) {
				t.Fatalf("\n\n%s\n", cmp.Diff(appCR, tc.expectedAppCR))
			}
			if !reflect.DeepEqual(ok, tc.expectedOK) {
				t.Fatalf("\n\n%s\n", cmp.Diff(ok, tc.expectedOK))
			}
		})
	}
}
