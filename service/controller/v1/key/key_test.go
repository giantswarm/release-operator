package key

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/versionbundle"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func urlMustParse(v string) *versionbundle.URL {
	u, err := url.Parse(v)
	if err != nil {
		panic(err)
	}

	return &versionbundle.URL{
		URL: u,
	}
}

func Test_ToConfigMap(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject apiv1.ConfigMap
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestName",
					Namespace: "release-operator",
				},
			},
			expectedObject: apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestName",
					Namespace: "release-operator",
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &apiv1.Pod{},
			errorMatcher: IsWrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToConfigMap(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedObject) {
				t.Fatalf("Custom Object == %q, want %q", result, tc.expectedObject)
			}
		})
	}
}

func Test_ToIndexReleases(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject []versionbundle.IndexRelease
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestName",
					Namespace: "release-operator",
				},
				Data: map[string]string{
					"indexblob": `- active: true
  authorities:
  - endpoint: http://aws-operator:8000
    name: aws-operator
    version: 0.1.0
  - endpoint: http://cert-operator:8000
    name: cert-operator
    provider: kvm
    version: 0.1.0
  - endpoint: http://cluster-operator:8000
    name: cluster-operator
    provider: aws
    version: 0.5.0
  date: 2018-01-01T12:00:00Z
  version: 1.0.0`,
				},
			},
			expectedObject: []versionbundle.IndexRelease{
				{
					Active: true,
					Authorities: []versionbundle.Authority{
						{
							Endpoint: urlMustParse("http://aws-operator:8000"),
							Name:     "aws-operator",
							Version:  "0.1.0",
						},
						{
							Endpoint: urlMustParse("http://cert-operator:8000"),
							Name:     "cert-operator",
							Provider: "kvm",
							Version:  "0.1.0",
						},
						{
							Endpoint: urlMustParse("http://cluster-operator:8000"),
							Name:     "cluster-operator",
							Provider: "aws",
							Version:  "0.5.0",
						},
					},
					Date:    time.Date(2018, time.January, 1, 12, 00, 0, 0, time.UTC),
					Version: "1.0.0",
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &apiv1.Pod{},
			errorMatcher: IsWrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToIndexReleases(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedObject) {
				t.Fatalf("Custom Object == %v, want %v", result, tc.expectedObject)
			}
		})
	}
}

func Test_toIndexReleaseFromString(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedObject []versionbundle.IndexRelease
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: `- active: true
  authorities:
  - endpoint: http://aws-operator:8000
    name: aws-operator
    version: 0.1.0
  - endpoint: http://cert-operator:8000
    name: cert-operator
    provider: kvm
    version: 0.1.0
  - endpoint: http://cluster-operator:8000
    name: cluster-operator
    provider: aws
    version: 0.5.0
  date: 2018-01-01T12:00:00Z
  version: 1.0.0`,
			expectedObject: []versionbundle.IndexRelease{
				{
					Active: true,
					Authorities: []versionbundle.Authority{
						{
							Endpoint: urlMustParse("http://aws-operator:8000"),
							Name:     "aws-operator",
							Version:  "0.1.0",
						},
						{
							Endpoint: urlMustParse("http://cert-operator:8000"),
							Name:     "cert-operator",
							Provider: "kvm",
							Version:  "0.1.0",
						},
						{
							Endpoint: urlMustParse("http://cluster-operator:8000"),
							Name:     "cluster-operator",
							Provider: "aws",
							Version:  "0.5.0",
						},
					},
					Date:    time.Date(2018, time.January, 1, 12, 00, 0, 0, time.UTC),
					Version: "1.0.0",
				},
			},
		},
		{
			name:         "case 1: empty string",
			input:        "",
			errorMatcher: IsEmptyValueError,
		},
		{
			name:         "case 2: invalid indexblob",
			input:        "something: invalid",
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toIndexReleasesFromString(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedObject) {
				t.Fatalf("Custom Object == %v, want %v", result, tc.expectedObject)
			}
		})
	}
}
