package sdp

import (
	"github.com/concourse-friends/concourse-builder/library"
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type BranchesJobArgs struct {
	*library.GitImageJobArgs
	*library.FlyImageJobArgs
	TargetGitRepo           *library.GitRepo
	Environment             map[string]interface{}
	GenerateProjectLocation project.IRun
}

func taskObtainBranches(args *BranchesJobArgs, branchesDir *library.TaskOutput) *project.TaskStep {
	gitImage, _ := library.GitImageJob(args.GitImageJobArgs)

	gitImageResource := &project.JobResource{
		Name:    gitImage.Name,
		Trigger: true,
	}

	targetGitResource := &project.Resource{
		Name: "target-git",
		Type: resource.GitMultibranchResourceType.Name,
		Source: &library.GitMultiSource{
			Repo: args.TargetGitRepo,
		},
	}

	args.GitImageJobArgs.ResourceRegistry.MustRegister(targetGitResource)

	targetGitJobResource := &project.JobResource{
		Name:    targetGitResource.Name,
		Trigger: true,
	}

	params := make(map[string]interface{})
	params["GIT_REPO_DIR"] = &library.Location{
		Volume: targetGitJobResource,
	}
	params["OUTPUT_DIR"] = branchesDir.Path()

	task := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "obtain branches",
		Image:    gitImageResource,
		Run: &library.Location{
			Volume: &library.Directory{
				Root: "/bin",
			},
			RelativePath: "obtain_branches.sh",
		},

		Params: params,
		Outputs: []project.IOutput{
			branchesDir,
		},
	}

	return task
}

func taskPreparePipelines(args *BranchesJobArgs, branchesDir *library.TaskOutput, pipelinesDir *library.TaskOutput) *project.TaskStep {
	args.GitImageJobArgs.ResourceRegistry.MustRegister(library.GoImage)

	goImageResource := &project.JobResource{
		Name:    library.GoImage.Name,
		Trigger: true,
	}

	params := make(map[string]interface{})
	for k, v := range args.Environment {
		params[k] = v
	}

	params[BranchesFileEnvVar] = &library.Location{
		Volume:       branchesDir,
		RelativePath: "branches",
	}

	task := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "prepare pipelines",
		Image:    goImageResource,
		Run:      args.GenerateProjectLocation,
		Params:   params,
		Outputs: []project.IOutput{
			pipelinesDir,
		},
	}

	return task
}

func taskCreateMissingPipelines(args *BranchesJobArgs, pipelinesDir *library.TaskOutput) *project.TaskStep {
	flyImage, _ := library.FlyImageJob(args.FlyImageJobArgs)

	flyImageResource := &project.JobResource{
		Name:    flyImage.Name,
		Trigger: true,
	}

	task := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "create missing pipelines",
		Image:    flyImageResource,
		Run: &library.Location{
			Volume: &library.Directory{
				Root: "/bin",
			},
			RelativePath: "create_missing_pipelines.sh",
		},
		Params: map[string]interface{}{
			"PIPELINES": &library.Location{
				Volume: pipelinesDir,
			},
			"CONCOURSE_URL":      args.Concourse.URL,
			"CONCOURSE_USER":     args.Concourse.User,
			"CONCOURSE_PASSWORD": args.Concourse.Password,
		},
	}

	return task
}

func taskRemoveNotNeededPipelines(args *BranchesJobArgs, pipelinesDir *library.TaskOutput, branchesDir *library.TaskOutput) *project.TaskStep {
	flyImage, _ := library.FlyImageJob(args.FlyImageJobArgs)

	flyImageResource := &project.JobResource{
		Name:    flyImage.Name,
		Trigger: true,
	}

	task := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "remove not needed pipelines",
		Image:    flyImageResource,
		Run: &library.Location{
			Volume: &library.Directory{
				Root: "/bin",
			},
			RelativePath: "remove_not_needed_pipelines.sh",
		},
		Params: map[string]interface{}{
			"PIPELINES": &library.Location{
				Volume: pipelinesDir,
			},
			"CONCOURSE_URL":      args.Concourse.URL,
			"CONCOURSE_USER":     args.Concourse.User,
			"CONCOURSE_PASSWORD": args.Concourse.Password,
			"BRANCHES_DIR":       branchesDir.Path(),
			"PIPELINE_REGEX":	  "*sdpd*",
		},
	}

	return task
}

func BranchesJob(args *BranchesJobArgs) *project.Job {
	branchesDir := &library.TaskOutput{
		Directory: "branches",
	}

	taskObtainBranches := taskObtainBranches(args, branchesDir)

	pipelinesDir := &library.TaskOutput{
		Directory: "pipelines",
	}

	taskPreparePipelines := taskPreparePipelines(args, branchesDir, pipelinesDir)

	taskCreateMissingPipelines := taskCreateMissingPipelines(args, pipelinesDir)

	taskRemoveNotNeededPipelines := taskRemoveNotNeededPipelines(args, pipelinesDir, branchesDir)

	branchesJob := &project.Job{
		Name:   project.JobName("branches"),
		Groups: project.JobGroups{},
		Steps: project.ISteps{
			taskObtainBranches,
			taskPreparePipelines,
			taskCreateMissingPipelines,
			taskRemoveNotNeededPipelines,
		},
	}

	return branchesJob
}
