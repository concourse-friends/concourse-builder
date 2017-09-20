package sdpExample

import (
	"bytes"
	"testing"

	"github.com/concourse-friends/concourse-builder/template/sdp"
	"github.com/concourse-friends/concourse-builder/test"
	"github.com/stretchr/testify/assert"
)

var buildImageScript = `
          H4sIAAAAAAAC/4STb2vbMBDGX1ef4qkXKAw8Mxh7l0JI3C60JcXp9mYbqSKda1HPEpKS/iEfftgm8Z94\
          2Uudnvvd3SPdh/No42y0VkVExRZr7jKWLBYP40fzIh8Zc+QR0itj02/x9GY1Sa6XY283xJhK8RPhO4LR\
          bDG9iZOr+W28ms2TACE/ii8f4vtlgN/MZ1SwMxKZRnCfE3cEZ0io9A1dDLRFH4CyxS23iq9zCvaYXl6N\
          U+TgM4JUloTX9g0vGVmqY1o8k01VTnCejAO3g7S6pnLgpS/09Que3pWB81YVT9DpICxgZy2rUp47Yqlq\
          23WVLO5W87vJdfwfQxrh8OBz3xvWktFOVdN6jTVh40hCFdVlSYPI+cbRyR6DUXMXYDxGUAlavb4qj89V\
          wp9nqSxCA2PJcEuyRpwP/4s9QBj07qKPDaHkptpiufieTOsnVQVGzXHJpGYAcFQ9aqkqhTAIk3Zu9GlY\
          LHVBjAmJpg9WmVy51noyXGJ2ePFactkJdRyIf0xuD3OXDdGW54d4N7Gce4B3ytBy0Qijcl+jvqP1z26V\
          FtyflB5384/S3V2upqq+44BmB28RSlz8Ki6w229RGEoSWhJ29UKFIpXH5f8GAAD//2JqijKaBAAA |\`

var expected = `groups:
- name: all
  jobs:
  - branches
  - curl-image
  - fly-image
  - git-image
  - self-update
- name: images
  jobs:
  - curl-image
  - fly-image
  - git-image
- name: sys
  jobs:
  - curl-image
  - fly-image
resource_types:
- name: git-multibranch
  type: docker-image
  source:
    repository: cfcommunity/git-multibranch-resource
resources:
- name: concourse-builder-git
  type: git
  source:
    uri: git@github.com:concourse-friends/concourse-builder.git
    branch: master
    private_key: private-key
- name: curl-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/curl-image
    tag: master
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: fly-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/fly-image
    tag: master
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: git-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/git-image
    tag: master
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: go-image
  type: docker-image
  source:
    repository: golang
  check_every: 24h
- name: target-git
  type: git-multibranch
  source:
    uri: git@github.com:target.git
    private_key: private-key
    branches: master|release[/].*|feature[/].*|task[/].*
- name: ubuntu-image
  type: docker-image
  source:
    repository: ubuntu
  check_every: 24h
jobs:
- name: curl-image
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
    - get: ubuntu-image
      trigger: true
  - task: prepare
    image: ubuntu-image
    config:
      platform: linux
      inputs:
      - name: concourse-builder-git
      params:
        DOCKERFILE_DIR: concourse-builder-git/docker/curl
        FROM_IMAGE: ubuntu
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > script.sh \
          && chmod 755 script.sh \
          && ./script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: curl-image
    params:
      build: prepared
    get_params:
      skip_download: true
- name: fly-image
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
      passed:
      - curl-image
    - get: curl-image
      trigger: true
      passed:
      - curl-image
      params:
        save: true
  - task: prepare
    image: curl-image
    config:
      platform: linux
      inputs:
      - name: concourse-builder-git
      params:
        DOCKERFILE_DIR: concourse-builder-git/docker/fly
        EVAL: echo ENV FLY_VERSION=` + "`" + `curl http://concourse.com/api/v1/info | awk -F
          ',' ' { print $1 } ' | awk -F ':' ' { print $2 } '` + "`" + `
        FROM_IMAGE: registry.com/concourse-builder/curl-image:master
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > script.sh \
          && chmod 755 script.sh \
          && ./script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: fly-image
    params:
      build: prepared
      load_base: curl-image
    get_params:
      skip_download: true
- name: git-image
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
      passed:
      - curl-image
    - get: curl-image
      trigger: true
      passed:
      - curl-image
      params:
        save: true
    - get: ubuntu-image
      trigger: true
      passed:
      - curl-image
  - task: prepare
    image: ubuntu-image
    config:
      platform: linux
      inputs:
      - name: concourse-builder-git
      params:
        DOCKERFILE_DIR: concourse-builder-git/docker/git
        FROM_IMAGE: registry.com/concourse-builder/curl-image:master
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > script.sh \
          && chmod 755 script.sh \
          && ./script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: git-image
    params:
      build: prepared
      load_base: curl-image
    get_params:
      skip_download: true
- name: branches
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
      passed:
      - fly-image
      - git-image
    - get: fly-image
      trigger: true
      passed:
      - fly-image
    - get: git-image
      trigger: true
      passed:
      - git-image
    - get: go-image
      trigger: true
    - get: target-git
      trigger: true
  - task: obtain branches
    image: git-image
    config:
      platform: linux
      inputs:
      - name: target-git
      params:
        GIT_PRIVATE_KEY: private-key
        GIT_REPO_DIR: target-git
        OUTPUT_DIR: branches
      run:
        path: /bin/git/obtain_branches.sh
      outputs:
      - name: branches
        path: branches
  - task: prepare pipelines
    image: go-image
    config:
      platform: linux
      inputs:
      - name: branches
      - name: concourse-builder-git
      params:
        BRANCH: branch
        BRANCHES_FILE: branches/branches
        PIPELINES: pipelines
      run:
        path: concourse-builder-git/foo
      outputs:
      - name: pipelines
        path: pipelines
  - task: create missing pipelines
    image: fly-image
    config:
      platform: linux
      inputs:
      - name: pipelines
      params:
        CONCOURSE_PASSWORD: password
        CONCOURSE_URL: http://concourse.com
        CONCOURSE_USER: user
        PIPELINES: pipelines
      run:
        path: /bin/fly/create_missing_pipelines.sh
  - task: remove not needed pipelines
    image: fly-image
    config:
      platform: linux
      inputs:
      - name: pipelines
      params:
        BRANCHES_DIR: branches
        CONCOURSE_PASSWORD: password
        CONCOURSE_URL: http://concourse.com
        CONCOURSE_USER: user
        PIPELINE_REGEX: .*-sdpb$
        PIPELINES: pipelines
      run:
        path: /bin/fly/remove_not_needed_pipelines.sh
- name: self-update
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
      passed:
      - fly-image
      - git-image
    - get: fly-image
      trigger: true
      passed:
      - fly-image
    - get: go-image
      trigger: true
  - task: check
    image: fly-image
    config:
      platform: linux
      params:
        CONCOURSE_URL: http://concourse.com
      run:
        path: /bin/fly/check_version.sh
  - task: prepare pipelines
    image: go-image
    config:
      platform: linux
      inputs:
      - name: concourse-builder-git
      params:
        BRANCH: branch
        PIPELINES: pipelines
      run:
        path: concourse-builder-git/foo
      outputs:
      - name: pipelines
        path: pipelines
  - task: update pipelines
    image: fly-image
    config:
      platform: linux
      inputs:
      - name: pipelines
      params:
        CONCOURSE_PASSWORD: password
        CONCOURSE_URL: http://concourse.com
        CONCOURSE_USER: user
        PIPELINES: pipelines
      run:
        path: /bin/fly/set_pipelines.sh
`

func TestSdp(t *testing.T) {
	prj, err := sdp.GenerateProject(&testSpecification{})
	assert.NoError(t, err)

	yml := &bytes.Buffer{}

	err = prj.Pipelines[0].Save(yml)
	assert.NoError(t, err)

	test.AssertEqual(t, expected, yml.String())
}
