package key

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	yaml "gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
)

const (
	ChartConfigAPIVersion           = "core.giantswarm.io"
	ChartConfigKind                 = "ChartConfig"
	ChartConfigVersionbundleVersion = "0.2.0"
)

func ChartChannel(version, operatorname string) string {
	channelversion := strings.Replace(version, ".", "-", -1)
	return fmt.Sprintf("%s-%s", operatorname, channelversion)
}

func ChartConfigName(operatorname string) string {
	return fmt.Sprintf("%s-chartconfig", operatorname)
}

func ChartName(operatorname string) string {
	return fmt.Sprintf("%s-chart", operatorname)
}

func ToConfigMap(v interface{}) (apiv1.ConfigMap, error) {
	customObjectPointer, ok := v.(*apiv1.ConfigMap)
	if !ok {
		return apiv1.ConfigMap{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.ConfigMap{}, v)
	}

	return *customObjectPointer, nil
}

func ToIndexReleases(v interface{}) ([]versionbundle.IndexRelease, error) {
	cm, err := ToConfigMap(v)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	indexBlob := cm.Data["indexblob"]

	indexReleases, err := toIndexReleasesFromString(indexBlob)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return indexReleases, nil
}

func toIndexReleasesFromString(indexBlob string) ([]versionbundle.IndexRelease, error) {
	if indexBlob == "" {
		return nil, microerror.Maskf(emptyValueError, "empty value cannot be converted to IndexReleases")
	}

	var indexReleases []versionbundle.IndexRelease
	err := yaml.Unmarshal([]byte(indexBlob), &indexReleases)
	if err != nil {
		return nil, microerror.Maskf(invalidConfigError, "unable to parse release index blob: %#v", err)
	}
	return indexReleases, nil
}
