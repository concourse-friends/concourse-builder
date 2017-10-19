package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/model"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/jinzhu/copier"
)

type SharedResourcesArgs struct {
	ConcourseBuilderGit *project.Resource
	ImageRegistry       *image.Registry
	ResourceRegistry    *project.ResourceRegistry
	Concourse           *primitive.Concourse

	// optional
	LinuxImageResource  *project.Resource
	GradleImageResource *project.Resource
}

func SharedResources(args *SharedResourcesArgs) *project.Job {
	taskDummy := &project.TaskStep{
		Platform: model.LinuxPlatform,
		Name:     "dummy",
		Image:    args.ResourceRegistry.JobResource(image.Alpine, true, nil),
		Run: &primitive.Location{
			Volume: &primitive.Directory{
				Root: "/bin",
			},
			RelativePath: "echo",
		},
		Environment: make(map[string]interface{}),
	}

	dummyResourceImageJobArgs := &DummyResourceImageJobArgs{}
	copier.Copy(dummyResourceImageJobArgs, args)
	dummyResourceImage := DummyResourceJob(dummyResourceImageJobArgs)
	dummyResourceImageResource := args.ResourceRegistry.JobResource((*project.Resource)(dummyResourceImage), true, nil)
	taskDummy.Environment["DUMMY_RESOURCE_IMAGE"] = &primitive.Location{
		Volume: dummyResourceImageResource,
	}

	if args.LinuxImageResource != nil {
		flyImageJobArgs := &FlyImageJobArgs{}
		copier.Copy(flyImageJobArgs, args)
		flyImage := FlyImageJob(flyImageJobArgs)
		flyImageResource := args.ResourceRegistry.JobResource(flyImage, true, nil)
		taskDummy.Environment["FLY_IMAGE"] = &primitive.Location{
			Volume: flyImageResource,
		}
	}

	if args.LinuxImageResource != nil {
		awsImageJobArgs := &AwsImageJobArgs{}
		copier.Copy(awsImageJobArgs, args)
		awsImage := AwsImageJob(awsImageJobArgs)
		awsImageResource := args.ResourceRegistry.JobResource(awsImage, true, nil)
		taskDummy.Environment["AWS_IMAGE"] = &primitive.Location{
			Volume: awsImageResource,
		}
	}

	if args.LinuxImageResource != nil {
		gitImageJobArgs := &GitImageJobArgs{}
		copier.Copy(gitImageJobArgs, args)
		gitImage := GitImageJob(gitImageJobArgs)
		gitImageResource := args.ResourceRegistry.JobResource(gitImage, true, nil)
		taskDummy.Environment["GIT_IMAGE"] = &primitive.Location{
			Volume: gitImageResource,
		}
	}

	if args.GradleImageResource != nil {
		gradleImageJobArgs := &GradleImageJobArgs{}
		copier.Copy(gradleImageJobArgs, args)
		gradleImage := GradleImageJob(gradleImageJobArgs)
		gradleImageResource := args.ResourceRegistry.JobResource(gradleImage, true, nil)
		taskDummy.Environment["GRADLE_IMAGE"] = &primitive.Location{
			Volume: gradleImageResource,
		}
	}

	dummyJob := &project.Job{
		Name:   project.JobName("dummy"),
		Groups: project.JobGroups{},
		Steps: project.ISteps{
			taskDummy,
		},
	}

	return dummyJob
}
