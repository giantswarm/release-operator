package app

import (
	"context"
	"io/ioutil"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	versionedfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/release-operator/service/controller/key"
)

func Test_newUpdateChange(t *testing.T) {
	fooApp := &applicationv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: applicationv1alpha1.AppSpec{
			Name:    "bar",
			Version: "v1.0.0",
		},
	}

	testCases := []struct {
		name          string
		obj           interface{}
		currentState  interface{}
		desiredState  interface{}
		expectedAppCR interface{}
		errorMatcher  func(err error) bool
	}{
		{
			name:          "case 0: interface nil arguments",
			currentState:  nil,
			desiredState:  nil,
			expectedAppCR: nil,
			errorMatcher:  key.IsWrongTypeError,
		},
		{
			name:          "case 1: invalid currentState",
			currentState:  "invalid value",
			desiredState:  &applicationv1alpha1.App{},
			expectedAppCR: nil,
			errorMatcher:  key.IsWrongTypeError,
		},
		{
			name:          "case 2: invalid desiredState",
			currentState:  &applicationv1alpha1.App{},
			desiredState:  "invalid value",
			expectedAppCR: nil,
			errorMatcher:  key.IsWrongTypeError,
		},
		{
			name:          "case 3: nil desiredState returns nil",
			currentState:  (*applicationv1alpha1.App)(nil),
			desiredState:  (*applicationv1alpha1.App)(nil),
			expectedAppCR: (*applicationv1alpha1.App)(nil),
			errorMatcher:  nil,
		},
		{
			name:          "case 4: nil currentState returns nil",
			currentState:  (*applicationv1alpha1.App)(nil),
			desiredState:  fooApp,
			expectedAppCR: (*applicationv1alpha1.App)(nil),
			errorMatcher:  nil,
		},
		{
			name:          "case 5: empty currentState returns nil",
			currentState:  &applicationv1alpha1.App{},
			desiredState:  fooApp,
			expectedAppCR: (*applicationv1alpha1.App)(nil),
			errorMatcher:  nil,
		},
		{
			name:          "case 6: same currentState returns nil",
			currentState:  fooApp,
			desiredState:  fooApp,
			expectedAppCR: (*applicationv1alpha1.App)(nil),
			errorMatcher:  nil,
		},
		{
			name: "case 7: different currentState returns desiredState",
			currentState: &applicationv1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
				Spec: applicationv1alpha1.AppSpec{
					Name:    "bar",
					Version: "v1.0.1",
				},
			},
			desiredState:  fooApp,
			expectedAppCR: fooApp,
			errorMatcher:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			var logger micrologger.Logger
			{
				c := micrologger.Config{
					IOWriter: ioutil.Discard,
				}
				logger, err = micrologger.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}
			c := Config{
				G8sClient: versionedfake.NewSimpleClientset(),
				K8sClient: kubernetesfake.NewSimpleClientset(),
				Logger:    logger,
			}
			r, err := New(c)
			if err != nil {
				t.Fatal(err)
			}

			change, err := r.newUpdateChange(context.Background(), tc.obj, tc.currentState, tc.desiredState)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if change != tc.expectedAppCR {
				t.Fatalf("wrong expectedAppCR: got %v, want %v", change, tc.expectedAppCR)
			}
		})
	}
}
