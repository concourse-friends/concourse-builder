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
		project.ResourceTypeGroup,
	},
}

type BuildImageArgs struct {
	ResourceRegistry   *project.ResourceRegistry
	PrepareImage       *project.Resource
	PrepareEnvironment map[string]interface{}
	From               *project.Resource
	Name               string
	TargetName         string
	DockerFileResource project.IValue
	DockerFileSteps    string
	Image              *project.Resource
	BuildArgs          map[string]interface{}
	PreprepareSteps    project.ISteps
	SourceDirs         []interface{}
	Eval               string
}

func taskPrepare(args *BuildImageArgs) *project.TaskStep {
	preparedDir := &project.TaskOutput{
		Directory: "prepared",
	}

	prepareImageResource := args.ResourceRegistry.JobResource(args.PrepareImage, true, nil)

	fromImageResource := args.ResourceRegistry.JobResource(args.From, true, nil)

	taskPrepare := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "prepare",
		Image:    prepareImageResource,
		Environment: map[string]interface{}{
			"FROM_IMAGE": &primitive.Location{
				Volume: fromImageResource,
			},
			"TARGET_NAME": &primitive.Location{
				Volume: args.TargetName,
			},
		},
		Outputs: []project.IOutput{
			preparedDir,
		},
	}

	script := `#!/usr/bin/env bash
ROOT=` + "`pwd`" + `

set -ex

CHECK_ARGS=true
TARGET_NAME_STR=""

if [ -z "$DOCKERFILE_DIR" -a -z "$DOCKERFILE_STEPS" ]
then
	echo "Please specify DOCKERFILE_DIR or DOCKERFILE_STEPS env variable"
	echo "DOCKERFILE_DIR specifies the directory where the dockerfile steps are"
	echo "DOCKERFILE_STEPS is a base64 gzip string of the dockerfile steps"
	CHECK_ARGS=false
fi

if [ -z "$FROM_IMAGE" ]
then
	echo "Please specify FROM_IMAGE env variable"
	echo "It specifies the repository to be used in the FROM clause"
	CHECK_ARGS=false
fi

if [ "$CHECK_ARGS" == "false" ]
then
	exit 1
fi

mkdir -p prepared
if [ ! -z "$DOCKERFILE_DIR" ]
then
	cp $DOCKERFILE_DIR/* prepared
fi

for SOURCE_DIR in $SOURCE_DIRS
do
	mkdir -p prepared/$SOURCE_DIR
	cp -R $SOURCE_DIR/. prepared/$SOURCE_DIR
done

if [ ! -z "$TARGET_NAME" ]
then
	TARGET_NAME_STR=" as $TARGET_NAME"
fi

cd  prepared

REPOSITORY=$(cat $ROOT/$FROM_IMAGE/repository)
TAG=$(cat $ROOT/$FROM_IMAGE/tag)
echo FROM $REPOSITORY:$TAG$TARGET_NAME_STR > Dockerfile
echo >> Dockerfile

if [ ! -z "$EVAL" ]
then
    eval "$EVAL" >> Dockerfile
fi

echo >> Dockerfile
if [ ! -z "$DOCKERFILE_DIR" -a -e $ROOT/$DOCKERFILE_DIR/steps ]
then
    cat $ROOT/$DOCKERFILE_DIR/steps >> Dockerfile
fi

if [ ! -z "$DOCKERFILE_STEPS" ]
then
    echo "$DOCKERFILE_STEPS" | tr -d '\n' | base64 --decode | gzip -cfd >> Dockerfile
fi
`

	taskPrepare.Run, taskPrepare.Arguments = EncodeScript(script)

	for k, v := range args.PrepareEnvironment {
		taskPrepare.Environment[k] = v
	}

	if args.DockerFileResource != nil {
		taskPrepare.Environment["DOCKERFILE_DIR"] = args.DockerFileResource
	}

	if args.DockerFileSteps != "" {
		taskPrepare.Environment["DOCKERFILE_STEPS"] = GZipBase64Lines(args.DockerFileSteps, "\n")
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
