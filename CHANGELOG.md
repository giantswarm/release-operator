# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

### Changed

- Detect running AWS and Azure clusters using `Cluster` CRs instead of provider-specific CRs

## [2.1.0] - 2020-09-16

### Added

- Set `InUse` field of `release` CRs.

### Changed

- Remove components for deprecated releases if no cluster is using the release.

### Added

- Add monitoring labels and add basic labels

## Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.
- Don't error when app was not found while deleting app.
- Deleted Release namespace from logging since releases are not namespaced.

## [2.0.0] - 2020-07-23

### Added

- Added functionality for watching Release CRs and creating App CRs to ensure all required components for the release are running.

### Changed

- No longer ensure Release CRD.

## [1.0.3] 2020-04-22

### Changed

- Use release.Revision in Helm chart for Helm 3 support.

## [1.0.2] 2020-04-21

### Fixed

- Push to china registry on tag.

## [1.0.1] 2020-04-15

### Fixed

- Fix version in project.go file.

## [1.0.0] 2020-04-15

### Changed

- Deploy as a unique app in app collection.

## [0.2.2] 2020-04-02

### Fixed

- Fix version in Helm templates.

## [0.2.1] 2020-04-02

### Fixed

- Set proper project version according to released tag.

## [0.2.0] 2020-04-02

### Changed

- Switch from dep to Go modules.
- Use latest architect orb.


[Unreleased]: https://github.com/giantswarm/release-operator/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/giantswarm/release-operator/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/giantswarm/release-operator/compare/v1.0.3...v2.0.0
[1.0.3]: https://github.com/giantswarm/release-operator/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/giantswarm/release-operator/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/release-operator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/release-operator/compare/v1.0.0...v1.0.0
[0.2.2]: https://github.com/giantswarm/release-operator/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/release-operator/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/release-operator/releases/tag/v0.2.0
