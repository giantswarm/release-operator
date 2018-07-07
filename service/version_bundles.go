package service

import (
	"github.com/giantswarm/release-operator/service/controller/v1"
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v1.VersionBundle())

	return versionBundles
}
