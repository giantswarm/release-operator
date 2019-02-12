package app

import (
	"reflect"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"
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
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-1"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app"
					appCR.Namespace = "non-default"
				}),
			},
			inputName:      "app",
			inputNamespace: "non-default",
			expectedAppCR: newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
				appCR.Name = "app"
				appCR.Namespace = "non-default"
			}),
			expectedOK: true,
		},
		{
			name: "case 1: nonexistent namespace",
			inputAppCRs: []*applicationv1alpha1.App{
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-0"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-1"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-2"
					appCR.Namespace = "default"
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
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-0"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-1"
					appCR.Namespace = "default"
				}),
				newTestAppCRFromFilled(func(appCR *applicationv1alpha1.App) {
					appCR.Name = "app-2"
					appCR.Namespace = "default"
				}),
			},
			inputName:      "app",
			inputNamespace: "default",
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
