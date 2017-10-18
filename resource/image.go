package resource

import (
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
)

// Image resource type
var ImageResourceType = &project.ResourceType{
	// The name
	Name: "docker-image",

	// The type
	Type: model.SystemResourceTypeName,
}

// Image resource source
type ImageSource struct {
	// Image repository
	Repository string

	// Docker target name
	TargetName string

	// Image tag
	Tag string `yaml:",omitempty"`

	// Is the repo insecure
	Insecure bool `yaml:",omitempty"`

	// Optional. AWS access key to use for acquiring ECR credentials.
	AwsAccessKeyID string `yaml:"aws_access_key_id,omitempty"`

	// Optional. AWS secret key to use for acquiring ECR credentials.
	AwsSecretAccessKey string `yaml:"aws_secret_access_key,omitempty"`
}

func (im ImageSource) ResourceName() project.ResourceName {
	return ""
}

func (im ImageSource) ResourceScope() project.Scope {
	return project.PipelineScope
}

func (im ImageSource) ModelSource(scope project.Scope, info *project.ScopeInfo) interface{} {
	return im
}

func (im ImageSource) NeededJobs() project.Jobs {
	return nil
}

type ImageGetParams struct {
	Save         bool `yaml:",omitempty"`
	SkipDownload bool `yaml:"skip_download,omitempty"`
}

type ImagePutParams struct {
	Build     string                 `yaml:",omitempty"`
	BuildArgs map[string]interface{} `yaml:"build_args,omitempty"`
	LoadBase  string                 `yaml:"load_base,omitempty"`
}

func init() {
	project.GlobalTypeRegistry.MustRegisterType(ImageResourceType)
}
