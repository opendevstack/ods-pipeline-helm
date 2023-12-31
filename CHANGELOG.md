# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Note that changes which ONLY affect documentation or the testsuite will not be
listed in the changelog.

## [Unreleased]

## [0.4.1] - 2023-11-13

### Fixed

- Fix invocation of Helm status for remote projects (see https://github.com/opendevstack/ods-pipeline-helm/commit/4699c455f990b32e4420dfe45436761d75a5f710)

## [0.4.0] - 2023-11-06

### Added

- Publish target namespace as Tekton result ([#10](https://github.com/opendevstack/ods-pipeline-helm/issues/10))

## [0.3.0] - 2023-10-30

### Changed

- Retain Helm release record ([#9](https://github.com/opendevstack/ods-pipeline-helm/pull/9)). Note that this removes the previously retained "release" txt file. The data in the new YAML file contains all previous information and more.

## [0.2.0] - 2023-10-09

### Changed

- Migrate from Tekton v1beta1 resources to v1 ([#7](https://github.com/opendevstack/ods-pipeline-helm/pull/7))

## [0.1.0] - 2023-09-29

Initial version.

NOTE: This version is based on v0.13.2 of the task `ods-deploy-helm` in the [ods-pipeline](https://github.com/opendevstack/ods-pipeline) repository.
