package library

import (
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type GitRepo struct {
	// URI to the git repo
	URI string `yaml:",omitempty"`

	// Private key the has access to the repo
	PrivateKey string `yaml:"private_key,omitempty"`
}

type GitSource struct {
	// Git repo and credentials
	Repo *GitRepo

	// branch name
	Branch string `yaml:",omitempty"`
}

func (gs *GitSource) ModelSource() interface{} {
	return &resource.GitSource{
		URI:        gs.Repo.URI,
		PrivateKey: gs.Repo.PrivateKey,
		Branch:     gs.Branch,
	}
}

var ConcourseBuilderGitName project.ResourceName = "concourse-builder-git"

func RegisterConcourseBuilderGit(resourceRegistry *project.ResourceRegistry, source *GitSource) {
	if resourceRegistry.GetResource(ConcourseBuilderGitName) != nil {
		return
	}

	concourseBuilderGit := &project.Resource{
		Name:   ConcourseBuilderGitName,
		Type:   resource.GitResourceType.Name,
		Source: source,
	}

	resourceRegistry.MustRegister(concourseBuilderGit)
}
