package library

import (
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type GradleImageJobArgs struct {
	*CurlImageJobArgs
}

func GradleImageJob(args *GradleImageJobArgs) (*project.Resource, *project.Job) {
	resourceName := project.ResourceName("gradle-image")
	image := args.ResourceRegistry.GetResource(resourceName)
	if image != nil {
		return image, image.NeededJobs[0]
	}

	image = &project.Resource{
		Name: resourceName,
		Type: resource.ImageResourceType.Name,
		Source: &ImageSource{
			Tag:        args.Tag,
			Registry:   args.ImageRegistry,
			Repository: "concourse-builder/curl-image",
		},
	}

	dockerSteps := &Location{
		Volume: &project.JobResource{
			Name:    ConcourseBuilderGitName,
			Trigger: true,
		},
		RelativePath: "docker/gradle",
	}

	args.ResourceRegistry.MustRegister(UbuntuImage)

	job := BuildImage(
		UbuntuImage,
		UbuntuImage,
		&BuildImageArgs{
			Name:               "gradle",
			DockerFileResource: dockerSteps,
			Image:              image.Name,
		})
	job.AddToGroup(project.SystemGroup)

	image.NeededJobs = project.Jobs{job}
	args.ResourceRegistry.MustRegister(image)

	return image, job
}
