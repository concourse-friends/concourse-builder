package image

import (
	"time"

	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

var Gcc = &project.Resource{
	Name: "gcc-image",
	Type: resource.ImageResourceType.Name,
	Source: &Source{
		Registry:   DockerHub,
		Repository: "gcc",
	},
	CheckInterval: model.Duration(24 * time.Hour),
}
