# ods-pipeline-helm

[![Tests](https://github.com/opendevstack/ods-pipeline-helm/actions/workflows/main.yaml/badge.svg)](https://github.com/opendevstack/ods-pipeline-helm/actions/workflows/main.yaml)

Tekton task for use with [ODS Pipeline](https://github.com/opendevstack/ods-pipeline) to deploy Helm charts.

## Usage

```yaml
tasks:
- name: build
  taskRef:
    resolver: git
    params:
    - { name: url, value: https://github.com/opendevstack/ods-pipeline-helm.git }
    - { name: revision, value: v0.3.0 }
    - { name: pathInRepo, value: tasks/deploy.yaml }
    workspaces:
    - { name: source, workspace: shared-workspace }
```

See the [documentation](/docs/deploy.adoc) for details and available parameters.

The task supports encrypted secrets via the [`helm-secrets`](https://github.com/jkroepke/helm-secrets) plugin. If you want to use this feature, please see the [Working with Helm secrets](/docs/helm-secrets.adoc) documentation.

## About this repository

`docs` and `tasks` are generated directories from recipes located in `build`. See the `Makefile` target for how everything fits together.
