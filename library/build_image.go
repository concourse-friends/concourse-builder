package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

var ImagesGroup = &project.JobGroup{
	Name: "images",
	Before: project.JobGroups{
		project.SystemGroup,
	},
}

type BuildImageArgs struct {
	ConcourseBuilderGit *project.Resource
	ResourceRegistry    *project.ResourceRegistry
	Prepare             *project.Resource
	From                *project.Resource
	Name                string
	DockerFileResource  project.IEnvironmentValue
	Image               *project.Resource
	BuildArgs           map[string]interface{}
	PreprepareSteps     project.ISteps
	SourceDirs          []interface{}
	Eval                string
}

func taskPrepare(args *BuildImageArgs) *project.TaskStep {
	preparedDir := &project.TaskOutput{
		Directory: "prepared",
	}

	prepareImageResource := args.ResourceRegistry.JobResource(args.Prepare, true, nil)

	taskPrepare := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "prepare",
		Image:    prepareImageResource,
		Run: &primitive.Location{
			Volume:       args.ResourceRegistry.JobResource(args.ConcourseBuilderGit, true, nil),
			RelativePath: "scripts/docker_image_prepare.sh",
		},
		Environment: map[string]interface{}{
			"DOCKERFILE_DIR": args.DockerFileResource,
			"FROM_IMAGE":     (*image.FromParam)(args.From),
		},
		Outputs: []project.IOutput{
			preparedDir,
		},
	}

	if len(args.SourceDirs) > 0 {
		taskPrepare.Environment["SOURCE_DIRS"] = primitive.Array(args.SourceDirs)
	}

	if args.Eval != "" {
		taskPrepare.Environment["EVAL"] = args.Eval
	}

	return taskPrepare
}

func BuildImage(args *BuildImageArgs) *project.Job {
	taskPrepare := taskPrepare(args)

	imageSource := args.From.Source.(*image.Source)
	public := imageSource.Registry.Public()

	fromImageResource := args.ResourceRegistry.JobResource(args.From, true, nil)

	if !public {
		fromImageResource.GetParams = &resource.ImageGetParams{
			Save: true,
		}
	}

	preparedDir := taskPrepare.Outputs[0]

	putImage := &project.PutStep{
		Resource: args.Image,
		Params: &image.PutParams{
			FromImage: fromImageResource,
			Load:      !public,
			Build: &primitive.Location{
				RelativePath: preparedDir.Path(),
			},
			BuildArgs: args.BuildArgs,
		},
		GetParams: &resource.ImageGetParams{
			SkipDownload: true,
		},
	}

	imageJob := &project.Job{
		Name: project.JobName(args.Name + "-image"),
		Groups: project.JobGroups{
			ImagesGroup,
		},
		Steps: args.PreprepareSteps,
	}

	imageJob.Steps = append(imageJob.Steps, taskPrepare, putImage)

	return imageJob
}
