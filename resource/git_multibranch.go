package resource

import "github.com/concourse-friends/concourse-builder/model"

type GitMultibranchSource struct {
	// URI to the git repo
	URI string `yaml:",omitempty"`

	// Private key the has access to the repo
	PrivateKey string `yaml:"private_key,omitempty"`
}

// The git multibranch resource type
var GitMultibranchResourceType = &model.ResourceType{
	// The name
	Name: "git-multibranch",
	// The type
	Type: model.DockerImageType,

	// The official repo
	Source: &ImageSource{
		// Image repository
		Repository: "cfcommunity/git-multibranch-resource",
	},
}

func init() {
	GlobalTypeRegistry.MustRegisterType(GitMultibranchResourceType)
}
