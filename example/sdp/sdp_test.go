package sdpExample

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/concourse-friends/concourse-builder/template/sdp"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
)

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
    tag: "1.8"
  check_every: 24h
- name: ubuntu-image
  type: docker-image
  source:
    repository: ubuntu
    tag: "16.04"
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
        FROM_IMAGE: ubuntu:16.04
      run:
        path: concourse-builder-git/scripts/docker_image_prepare.sh
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
        EVAL: echo ENV FLY_VERSION=`+"`"+`curl http://concourse.com/api/v1/info | awk -F
          ',' ' { print $1 } ' | awk -F ':' ' { print $2 } '`+"`"+`
        FROM_IMAGE: registry.com/concourse-builder/curl-image:master
      run:
        path: concourse-builder-git/scripts/docker_image_prepare.sh
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
        path: concourse-builder-git/scripts/docker_image_prepare.sh
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
    - get: fly-image
      trigger: true
      passed:
      - fly-image
    - get: git-image
      trigger: true
      passed:
      - git-image
  - task: obtain branches
    image: git-image
    config:
      platform: linux
      params:
        BRANCH: branch
        PIPELINES: pipelines
      run:
        path: /bin/obtain_branches.sh
      outputs:
      - name: branches
        path: branches
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
        path: /bin/create_missing_pipelines.sh
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
        path: /bin/check_version.sh
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
        path: /bin/set_pipelines.sh
`

func ContextDiff(a, b string) string {
	diff := difflib.ContextDiff{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: "actual",
		ToFile:   "expected",
		Context:  20,
		Eol:      "\n",
	}
	result, _ := difflib.GetContextDiffString(diff)
	return result
}

func TestSdp(t *testing.T) {
	prj, err := sdp.GenerateProject(&testSpecification{})
	assert.NoError(t, err)

	yml := &bytes.Buffer{}

	err = prj.Pipelines[0].Save(yml)
	assert.NoError(t, err)

	if expected != yml.String() {
		log.Printf("\n" + strings.Replace(yml.String(), "`", "`+\"`\"+`", -1))
	}

	assert.Equal(t, expected, yml.String(), ContextDiff(expected, yml.String()))
}
