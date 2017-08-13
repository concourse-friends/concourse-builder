package library

import (
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
)

type SelfUpdateJobArgs struct {
	Concourse                 *Concourse
	ConcourseBuilderGitSource *GitSource
	Environment               map[string]interface{}
	ImageRegistry             *ImageRegistry
	PipelineLocation          project.IRun
	ResourceRegistry          *project.ResourceRegistry
	Tag                       ImageTag
}

func SelfUpdateJob(args *SelfUpdateJobArgs) *project.Job {
	flyImage, _ := FlyImageJob(&FlyImageJobArgs{
		Concourse:                 args.Concourse,
		ConcourseBuilderGitSource: args.ConcourseBuilderGitSource,
		ImageRegistry:             args.ImageRegistry,
		ResourceRegistry:          args.ResourceRegistry,
		Tag:                       args.Tag,
	})

	flyImageResource := &project.JobResource{
		Name:    flyImage.Name,
		Trigger: true,
	}

	taskCheck := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "check",
		Image:    flyImageResource,
		Run: &Location{
			Volume: &Directory{
				Root: "/bin",
			},
			RelativePath: "check_version.sh",
		},
		Params: map[string]interface{}{
			"CONCOURSE_URL": args.Concourse.URL,
		},
	}

	args.ResourceRegistry.MustRegister(GoImage)

	goImageResource := &project.JobResource{
		Name:    GoImage.Name,
		Trigger: true,
	}

	pipelinesDir := &TaskOutput{
		Directory: "pipelines",
	}

	taskPrepare := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "prepare pipelines",
		Image:    goImageResource,
		Run:      args.PipelineLocation,
		Params:   args.Environment,
		Outputs: []project.IOutput{
			pipelinesDir,
		},
	}

	taskPrepare.Params["PIPELINES"] = "pipelines"

	taskUpdate := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "update pipelines",
		Image:    flyImageResource,
		Run: &Location{
			Volume: &Directory{
				Root: "/bin",
			},
			RelativePath: "set_pipelines.sh",
		},
		Params: map[string]interface{}{
			"PIPELINES": &Location{
				Volume: pipelinesDir,
			},
			"CONCOURSE_URL":      args.Concourse.URL,
			"CONCOURSE_USER":     args.Concourse.User,
			"CONCOURSE_PASSWORD": args.Concourse.Password,
		},
	}

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