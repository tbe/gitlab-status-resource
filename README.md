# Gitlab Status Resource

Updates the status of an gitlab commit

## Source configuration

- `gitlab_url`: *Required.* The URL of the Gitlab Instance
- `api_key`: *Required.* The API Key for Gitlab. The Key must belong to a user with at least *Developer* privileges
- `group`: *Required.* The Gitlab group of the repository
- `project`: *Required.* The Project name of the repository
- `verify_ssl`: *Optional.* Verify the SSL Certificate of the Gitlab API Endpoint. Defaults to `true`

## Parameters

- `build_status`: *Required.* The desired build status. Must be one of `pending`, `running`, `success`, `canceled`, `failed`
- `repo`: *Required.* The asset name/ input path of the checked out git repository
- `status_name`: *Optional.* The name of the status. Defaults to `default`


## Example

```yaml
resource_types:
- name: gitlab-status
  type: docker-image
  source:
    repository: tbede/gitlab-resource
    tag: 0.1.0

resources:
- name: source
  type: git
  source:
    branch: master
    uri: https://gitlab.something.lan/mygroup/myrepo
    username: buildbot
    password: ((gitlabtoken))
- name: gitlab-status
  type: gitlab-status
  source:
    api_key: ((gitlabtoken))
    gitlab_url: https://gitlab.something.lan
    verify_ssl: false
    group: mygroup
    project: myrepo

jobs:
- name: build
  plan:
  - get: source
    trigger: true
  - put: gitlab-status
    params:
      build_status: running
      repo: source
  - put: ....
    # your build task is here
    on_success:
      put: gitlab-status
      params:
        build_status: success
        repo: source
    on_failure:
      put: gitlab-status
      params:
        build_status: failed
        repo: source
    on_abort:
      put: gitlab-status
      params:
        build_status: canceled
        repo: source
```
