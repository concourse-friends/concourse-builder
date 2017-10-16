package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type RiakKVImageJobArgs struct {
	LinuxImageResource *project.Resource
	ImageRegistry      *image.Registry
	ResourceRegistry   *project.ResourceRegistry
}

func RiakKVImageJob(args *RiakKVImageJobArgs) *project.Resource {
	resourceName := project.ResourceName("riak-kv-image")
	imageResource := args.ResourceRegistry.GetResource(resourceName)
	if imageResource != nil {
		return imageResource
	}

	imageResource = &project.Resource{
		Name:  resourceName,
		Type:  resource.ImageResourceType.Name,
		Scope: project.TeamScope,
		Source: &image.Source{
			Registry:   args.ImageRegistry,
			Repository: "dev_at/riak-kv-image",
		},
	}

	steps := `RUN set -exm \
# install curl \
&& apt-get update -y \
&& apt-get install -y curl \
\
# cleanup \
&& apt-get clean \
\
&& CLUSTER_NAME=riak-kv \
\
&& echo "storage_backend = leveldb" > /etc/riak/user.conf \
&& sed -i 's/^search = off/search = on/g' /etc/riak/riak.conf \
&& service riak restart \
`

	job := BuildImage(
		&BuildImageArgs{
			ResourceRegistry: args.ResourceRegistry,
			PrepareImage:     image.RiakKV,
			From:             args.LinuxImageResource,
			Name:             "riak-kv",
			DockerFileSteps:  steps,
			Image:            imageResource,
		})
	job.AddToGroup(project.SystemGroup)

	imageResource.NeedJobs(job)

	return imageResource
}
