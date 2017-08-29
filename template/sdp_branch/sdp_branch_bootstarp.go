package sdpBranch

import (
	"github.com/concourse-friends/concourse-builder/library"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/project"
)

type BootstrapSpecification interface {
	Branch() *primitive.GitBranch
	Concourse() (*primitive.Concourse, error)
	DeployImageRegistry() (*library.ImageRegistry, error)
	ConcourseBuilderGitSource() (*library.GitSource, error)
	GenerateProjectLocation(resourceRegistry *project.ResourceRegistry) (project.IRun, error)
	Environment() (map[string]interface{}, error)
}

func GenerateBootstrapProject(specification BootstrapSpecification) (*project.Project, error) {
	mainPipeline := project.NewPipeline()
	mainPipeline.AllJobsGroup = project.AllJobsGroupFirst
	mainPipeline.Name = project.ConvertToPipelineName(specification.Branch().Name() + "-sdpb")

	concourseBuilderGitSource, err := specification.ConcourseBuilderGitSource()
	if err != nil {
		return nil, err
	}

	imageRegistry, err := specification.DeployImageRegistry()
	if err != nil {
		return nil, err
	}

	concourse, err := specification.Concourse()
	if err != nil {
		return nil, err
	}

	generateProjectLocation, err := specification.GenerateProjectLocation(mainPipeline.ResourceRegistry)
	if err != nil {
		return nil, err
	}

	environment, err := specification.Environment()
	if err != nil {
		return nil, err
	}

	selfUpdateJob := library.SelfUpdateJob(&library.SelfUpdateJobArgs{
		ConcourseBuilderGitSource: concourseBuilderGitSource,
		ImageRegistry:             imageRegistry,
		ResourceRegistry:          mainPipeline.ResourceRegistry,
		Tag:                       library.ConvertToImageTag(concourseBuilderGitSource.Branch),
		Concourse:                 concourse,
		Environment:               environment,
		GenerateProjectLocation:   generateProjectLocation,
	})

	mainPipeline.Jobs = project.Jobs{
		selfUpdateJob,
	}

	prj := &project.Project{
		Pipelines: project.Pipelines{
			mainPipeline,
		},
	}

	return prj, nil
}
