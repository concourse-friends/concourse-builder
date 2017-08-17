package sdp

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/concourse-friends/concourse-builder/library"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/template/sdp_branch"
	"regexp"
)

type Specification interface {
	Concourse() (*library.Concourse, error)
	DeployImageRegistry() (*library.ImageRegistry, error)
	ConcourseBuilderGitSource() (*library.GitSource, error)
	GenerateProjectLocation(resourceRegistry *project.ResourceRegistry, overrideBranch string) (project.IRun, error)
	TargetGitRepo() (*library.GitRepo, error)
	Environment() (map[string]interface{}, error)
}

const BranchesFileEnvVar = "BRANCHES_FILE"

func BootstrapBranches() ([]string, error) {
	branchesFile, exist := os.LookupEnv(BranchesFileEnvVar)
	if !exist {
		return nil, nil
	}
	branches, err := ioutil.ReadFile(branchesFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(branches), "\n"), err
}

func GenerateProject(specification Specification) (*project.Project, error) {
	prj := &project.Project{}

	branches, err := BootstrapBranches()
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		re, err := regexp.Compile("master|release/.*|feature/.*|task/.*")
		if err != nil {
			return nil, err
		}

		if re.MatchString(branch) {
			branchSpecification := &BranchBootstrapSpecification{
				Specification: specification,
				Branch:        branch,
			}
			project, err := sdpBranch.GenerateBootstrapProject(branchSpecification)
			if err != nil {
				return nil, err
			}
			prj.Pipelines = append(prj.Pipelines, project.Pipelines...)
		}
	}

	mainPipeline := project.NewPipeline()
	mainPipeline.AllJobsGroup = project.AllJobsGroupFirst

	concourseBuilderGitSource, err := specification.ConcourseBuilderGitSource()
	if err != nil {
		return nil, err
	}

	mainPipeline.Name = project.ConvertToPipelineName(concourseBuilderGitSource.Branch + "-sdp")

	imageRegistry, err := specification.DeployImageRegistry()
	if err != nil {
		return nil, err
	}

	concourse, err := specification.Concourse()
	if err != nil {
		return nil, err
	}

	generateProjectLocation, err := specification.GenerateProjectLocation(mainPipeline.ResourceRegistry, "")
	if err != nil {
		return nil, err
	}

	environment, err := specification.Environment()
	if err != nil {
		return nil, err
	}

	curlImageJobArgs := &library.CurlImageJobArgs{
		ConcourseBuilderGitSource: concourseBuilderGitSource,
		ImageRegistry:             imageRegistry,
		ResourceRegistry:          mainPipeline.ResourceRegistry,
		Tag:                       library.ConvertToImageTag(concourseBuilderGitSource.Branch),
	}

	flyImageJobArgs := &library.FlyImageJobArgs{
		CurlImageJobArgs: curlImageJobArgs,
		Concourse:        concourse,
	}

	selfUpdateJob := library.SelfUpdateJob(&library.SelfUpdateJobArgs{
		FlyImageJobArgs:         flyImageJobArgs,
		Environment:             environment,
		GenerateProjectLocation: generateProjectLocation,
	})

	targetGit, err := specification.TargetGitRepo()
	if err != nil {
		return nil, err
	}

	branchesJob := BranchesJob(&BranchesJobArgs{
		GitImageJobArgs: &library.GitImageJobArgs{
			CurlImageJobArgs: curlImageJobArgs,
		},
		FlyImageJobArgs:         flyImageJobArgs,
		TargetGitRepo:           targetGit,
		Environment:             environment,
		GenerateProjectLocation: generateProjectLocation,
	})

	mainPipeline.Jobs = project.Jobs{
		selfUpdateJob,
		branchesJob,
	}

	prj.Pipelines = append(prj.Pipelines, mainPipeline)
	return prj, nil
}
