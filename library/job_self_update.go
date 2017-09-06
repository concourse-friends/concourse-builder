package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/jinzhu/copier"
)

type SelfUpdateJobArgs struct {
	LinuxImageResource      *project.Resource
	ConcourseBuilderGit     *project.Resource
	ImageRegistry           *image.Registry
	GoImage                 *project.Resource
	ResourceRegistry        *project.ResourceRegistry
	Concourse               *primitive.Concourse
	Environment             map[string]interface{}
	GenerateProjectLocation project.IRun
}

func SelfUpdateJob(args *SelfUpdateJobArgs) *project.Job {
	flyImageJobArgs := &FlyImageJobArgs{}
	copier.Copy(flyImageJobArgs, args)

	flyImage := FlyImageJob(flyImageJobArgs)

	flyImageResource := args.ResourceRegistry.JobResource(flyImage, true, nil)

	taskCheck := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "check",
		Image:    flyImageResource,
		Run: &primitive.Location{
			Volume: &primitive.Directory{
				Root: "/bin/fly",
			},
			RelativePath: "check_version.sh",
		},
		Environment: map[string]interface{}{},
	}

	args.Concourse.PublicAccessEnvironment(taskCheck.Environment)

	goImageResource := args.ResourceRegistry.JobResource(args.GoImage, true, nil)

	pipelinesDir := &project.TaskOutput{
		Directory: "pipelines",
	}

	taskPrepare := &project.TaskStep{
		Platform:    model.LinuxPlatform,
		Name:        "prepare pipelines",
		Image:       goImageResource,
		Run:         args.GenerateProjectLocation,
		Environment: args.Environment,
		Outputs: []project.IOutput{
			pipelinesDir,
		},
	}

	taskPrepare.Environment["PIPELINES"] = "pipelines"

	taskUpdate := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "update pipelines",
		Image:    flyImageResource,
		Run: &primitive.Location{
			Volume: &primitive.Directory{
				Root: "/bin/fly",
			},
			RelativePath: "set_pipelines.sh",
		},
		Environment: map[string]interface{}{
			"PIPELINES": &primitive.Location{
				Volume: pipelinesDir,
			},
		},
	}
	args.Concourse.Environment(taskUpdate.Environment)

	updateJob := &project.Job{
		Name:   project.JobName("self-update"),
		Groups: project.JobGroups{},
		Steps: project.ISteps{
			taskCheck,
			taskPrepare,
			taskUpdate,
		},
	}
	return updateJob
}
