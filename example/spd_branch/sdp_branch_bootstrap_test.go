package sdpBranchExample

import (
	"bytes"
	"testing"

	"github.com/concourse-friends/concourse-builder/template/sdp_branch"
	"github.com/concourse-friends/concourse-builder/test"
	"github.com/stretchr/testify/assert"
)

var buildImageScript = `
          H4sIAAAAAAAC/4RUXWvbQBB8zv2KiSpIU1BFofSh4IBxFNckQUFyC6UtzvluZR9RdeLu7HzgH18kYUu2\
          lfRRe7OzO7O7encarqwJ56oIqVhjzu2SJXE8HdyXj/KeMUsOAT0xNvoWja5nw2ScDpxZEWMqwy8EL/D8\
          y3h0HSVXk5todjlJPAT8KJ5Oo7vUwx/mllSwExJLDe8uJ24JtiShsmfs00AbHBKganHNjeLznLwtzUFe\
          Q6fIwi0JUhkSTptnPC7JUBPT4oFMpnKCdVRacNPL1tRUFrzyhb58xuJFlbDOqGIBnfWSeeykY1XGc0ss\
          U127rpL4dja5HY6j/xjSAvuFT9yBWEOltqpW6zTmhJUlCVXUjxUbRM5Xlt7s0fPbNw+DAbwa0On1STl8\
          qhP+PkhlEJQoDZXckGwoTvv3YksgShy8hR9ahoo30wZp/D0ZNSNVBfz2M2VSMwA4qh52UDVClAiSbm74\
          sR8sdUGMCYm2D5ZEd3E6mcbJz4H/XnAHvzqMsDO/sDX8nE2H41dxji/OWT20egp+S/3Vnw7HuMDlbo0a\
          3MVeaM/W6MfwZmdmpZLWPN/F9xMrM3v43ppSdb201XAwpuZcOqU7cnuhx928Unr/B1Grqne8B7OBMwgk\
          zn4XZ9hsTzMIJAktCZvmSgORyePy/wIAAP///BQZEO8EAAA= |\`

var expectedBootstrap = `groups:
- name: all
  jobs:
  - curl-image
  - dummy_resource-image
  - fly-image
  - aws-image
  - self-update
- name: images
  jobs:
  - curl-image
  - dummy_resource-image
  - fly-image
  - aws-image
- name: res_types
  jobs:
  - dummy_resource-image
- name: sys
  jobs:
  - curl-image
  - dummy_resource-image
  - fly-image
  - aws-image
resource_types:
- name: dummy
  type: docker-image
  source:
    repository: registry.com/concourse-builder/dummy_resource-image
    tag: target_branch_image-sdpb
    aws_access_key_id: key
    aws_secret_access_key: secret
resources:
- name: alpine-image
  type: docker-image
  source:
    repository: alpine
  check_every: 24h
- name: concourse-builder-git
  type: git
  source:
    uri: git@github.com:concourse-friends/concourse-builder.git
    branch: master_image
    private_key: private-key
- name: curl-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/curl-image
    tag: target_branch_image-sdpb
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: dummy_resource-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/dummy_resource-image
    tag: target_branch_image-sdpb
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: fly-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/fly-image
    tag: installation
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: aws-image
  type: docker-image
  source:
    repository: registry.com/concourse-builder/aws-image
    tag: installation
    aws_access_key_id: key
    aws_secret_access_key: secret
- name: go-image
  type: docker-image
  source:
    repository: golang
  check_every: 24h
- name: pipeline
  type: dummy
- name: ubuntu-image
  type: docker-image
  source:
    repository: ubuntu
  check_every: 24h
jobs:
- name: curl-image
  plan:
  - get: ubuntu-image
    trigger: true
  - task: prepare
    image: ubuntu-image
    config:
      platform: linux
      inputs:
      - name: ubuntu-image
      params:
        DOCKERFILE_STEPS: |-
          H4sIAAAAAAAC/1TMsQrCUAyF4b1PcUDoIIS8iYPg1iXWKBdiCUmu6NuLlg5dP/5zzpcTUgukb0zDAW3J
          EjPMPQzTMI4QL3pooftNSve21fTZBr+P2VSW7vv0jyvFExR38EuCrV1ZvNhaVvLxGwAA//9QRiyQjwAA
          AA==
        FROM_IMAGE: ubuntu-image
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          mkdir -p /tmp \
          && echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > /tmp/script.sh \
          && cat /tmp/script.sh \
          && echo \
          && chmod 755 /tmp/script.sh \
          && /tmp/script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: curl-image
    params:
      build: prepared
    get_params:
      skip_download: true
- name: dummy_resource-image
  plan:
  - aggregate:
    - get: alpine-image
      trigger: true
    - get: concourse-builder-git
      trigger: true
    - get: ubuntu-image
      trigger: true
  - task: prepare
    image: ubuntu-image
    config:
      platform: linux
      inputs:
      - name: alpine-image
      - name: concourse-builder-git
      params:
        DOCKERFILE_DIR: concourse-builder-git/docker/dummy_resource
        FROM_IMAGE: alpine-image
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          mkdir -p /tmp \
          && echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > /tmp/script.sh \
          && cat /tmp/script.sh \
          && echo \
          && chmod 755 /tmp/script.sh \
          && /tmp/script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: dummy_resource-image
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
      - dummy_resource-image
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
      - name: curl-image
      params:
        DOCKERFILE_DIR: concourse-builder-git/docker/fly
        EVAL: echo ENV FLY_VERSION=` + "`" + `curl http://concourse.com/api/v1/info | awk -F
          ',' ' { print $1 } ' | awk -F ':' ' { print $2 } '` + "`" + `
        FROM_IMAGE: curl-image
      run:
        path: /bin/bash
        args:
        - -c
        - |-
          mkdir -p /tmp \
          && echo \` + buildImageScript + `
              base64 --decode |\
              gzip -cfd > /tmp/script.sh \
          && cat /tmp/script.sh \
          && echo \
          && chmod 755 /tmp/script.sh \
          && /tmp/script.sh
      outputs:
      - name: prepared
        path: prepared
  - put: fly-image
    params:
      build: prepared
      load_base: curl-image
    get_params:
      skip_download: true
- name: self-update
  plan:
  - aggregate:
    - get: concourse-builder-git
      trigger: true
      passed:
      - fly-image
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
        BRANCH: branch_image
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
  - put: pipeline
`

func TestSdpBranchBootstrap(t *testing.T) {
	prj, err := sdpBranch.GenerateBootstrapProject(&testSpecification{})
	assert.NoError(t, err)

	yml := &bytes.Buffer{}

	err = prj.Pipelines[0].Save("team", "installation", yml)
	assert.NoError(t, err)
	test.AssertEqual(t, expectedBootstrap, yml.String())
}
