package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type CouchbaseImageJobArgs struct {
	LinuxImageResource *project.Resource
	ImageRegistry      *image.Registry
	ResourceRegistry   *project.ResourceRegistry
}

func CouchbaseImageJob(args *CouchbaseImageJobArgs) *project.Resource {
	resourceName := project.ResourceName("couchbase-image")
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
			Repository: "def_at/couchbase-image",
		},
	}

	steps := `RUN set -exm \
# install curl \
&& apt-get update -y \
&& apt-get install -y curl \
\
# cleanup \
&& apt-get clean \
&& COUCHBASE_HOST='127.0.0.1' \
&& COUCHBASE_PORT=8091 \
&& COUCHBASE_URL="http://$COUCHBASE_HOST:$COUCHBASE_PORT" \
&& COUCHBASE_MEMORY_QUOTA=2000 \
&& COUCHBASE_USER=Uconcourse \
&& COUCHBASE_PASSWORD=Pconcourse0 \
\
&& /entrypoint.sh couchbase-server & \
\
&& sleep 16 \
\
&& curl -v -X POST $COUCHBASE_URL/pools/default -d memoryQuota=$COUCHBASE_MEMORY_QUOTA -d indexMemoryQuota=$COUCHBASE_MEMORY_QUOTA \
&& curl -v $COUCHBASE_URL/node/controller/setupServices -d services=kv%2cn1ql%2Cindex \
&& curl -v $COUCHBASE_URL/settings/web -d port=$COUCHBASE_PORT_LOCAL -d username=$COUCHBASE_USER -d password=$COUCHBASE_PASSWORD \
\
&& fg 1`

	job := BuildImage(
		&BuildImageArgs{
			ResourceRegistry: args.ResourceRegistry,
			PrepareImage:     image.Couchbase,
			From:             args.LinuxImageResource,
			Name:             "couchbase",
			DockerFileSteps:  steps,
			Image:            imageResource,
		})
	job.AddToGroup(project.SystemGroup)

	imageResource.NeedJobs(job)

	return imageResource
}
