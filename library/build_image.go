package library

import (
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
	Name               string
	DockerFileResource project.IParamValue
	Image              project.ResourceName
	BuildArgs          map[string]interface{}
}

func BuildImage(args *BuildImageArgs) *project.Job {
	imageResource := &project.JobResource{
		Name: args.Image,
	}

	ubuntuImageResource := &project.JobResource{
		Name:    UbuntuImage.Name,
		Trigger: true,
	}

	preparedDir := &TaskOutput{
		Directory: "prepared",
	}

	taskPrepare := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "prepare",
		Image:    ubuntuImageResource,
		Run: &Location{
			Volume: &project.JobResource{
				Name:    ConcourseBuilderGitName,
				Trigger: true,
			},
			RelativePath: "scripts/docker_image_prepare.sh",
		},
		Params: map[string]interface{}{
			"DOCKER_STEPS": args.DockerFileResource,
			"FROM_IMAGE":   (*FromParam)(UbuntuImage),
		},
		Outputs: []project.IOutput{
			preparedDir,
		},
	}

	putImage := &project.PutStep{
		JobResource: imageResource,
		Params: &ImagePutParams{
			FromImage: ubuntuImageResource,
			Build: &Location{
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
		Steps: project.ISteps{
			taskPrepare,
			putImage,
		},
	}
	return imageJob
}