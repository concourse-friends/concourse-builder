package library

import (
	"fmt"

	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
	"github.com/jinzhu/copier"
)

type ServiceImageJobArgs struct {
	LinuxImageResource     *project.Resource
	CuneiformImageResource *project.Resource
	ConcourseBuilderGit    *project.Resource
	ImageRegistry          *image.Registry
	ResourceRegistry       *project.ResourceRegistry
}

func ServiceImageJob(args *ServiceImageJobArgs) *project.Resource {
	resourceName := project.ResourceName("service-image")
	imageResource := args.ResourceRegistry.GetResource(resourceName)
	if imageResource != nil {
		return imageResource
	}

	imageResource = &project.Resource{
		Name:  resourceName,
		Type:  resource.ImageResourceType.Name,
		Scope: project.PipelineScope,
		Source: &image.Source{
			Registry:   args.ImageRegistry,
			Repository: "dev_at",
		},
	}

	steps := `
########    COUCHBASE    ########

RUN apt-get update && \
    apt-get install -yq runit wget python-httplib2 chrpath \
    lsof lshw sysstat net-tools numactl && \
    apt-get autoremove && apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ARG CB_VERSION=5.0.0
ARG CB_RELEASE_URL=http://packages.couchbase.com/releases
ARG CB_PACKAGE=couchbase-server-enterprise_5.0.0-ubuntu16.04_amd64.deb
ARG CB_SHA256=bc3b65c78793b819ecba87c330bd1bcc0a2edec214c597069c8eb7e34505eb69

ENV PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Create Couchbase user with UID 1000(1001) (necessary to match default
# boot2docker UID)
RUN groupadd -g 1001 couchbase && useradd couchbase -u 1001 -g couchbase -M

# Install couchbase
RUN wget -N $CB_RELEASE_URL/$CB_VERSION/$CB_PACKAGE && \
    echo "$CB_SHA256  $CB_PACKAGE" | sha256sum -c - && \
    dpkg -i ./$CB_PACKAGE && rm -f ./$CB_PACKAGE

# Add runit script for couchbase-server
#COPY scripts/run /etc/service/couchbase-server/run
    echo "#!/bin/sh\n\nexec 2>&1\n\n# Create directories where couchbase stores its data" > /etc/service/couchbase-server/run && \
    echo "cd /opt/couchbase" >> /etc/service/couchbase-server/run && \
    echo "mkdir -p var/lib/couchbase var/lib/couchbase/config var/lib/couchbase/data var/lib/couchbase/stats var/lib/couchbase/logs var/lib/moxi" >> /etc/service/couchbase-server/run && \
    echo "\nchown -R couchbase:couchbase var" >> /etc/service/couchbase-server/run && \
    echo "exec chpst -ucouchbase /opt/couchbase/bin/couchbase-server -- -kernel global_enable_tracing false -noinput" >> /etc/service/couchbase-server/run && \
    chmod 775 /etc/service/couchbase-server/run

# Add dummy script for commands invoked by cbcollect_info that
# make no sense in a Docker container
#COPY scripts/dummy.sh /usr/local/bin/
RUN echo "#!/bin/sh\n\necho \"Running in Docker container - \$0 not available\"" > /usr/local/bin/dummy.sh && \
    chmod 775 /usr/local/bin/dummy.sh
RUN ln -s dummy.sh /usr/local/bin/iptables-save && \
    ln -s dummy.sh /usr/local/bin/lvdisplay && \
    ln -s dummy.sh /usr/local/bin/vgdisplay && \
    ln -s dummy.sh /usr/local/bin/pvdisplay

# Fix curl RPATH
RUN chrpath -r '$ORIGIN/../lib' /opt/couchbase/bin/curl

# Add bootstrap script and Peernova hooks
#COPY scripts/entrypoint.sh /
RUN echo "#!/bin/bash\n\nset -e\n\nnohup /peernova/couchbase/run.sh &\n\n[[ \"\$1\" == \"couchbase-server\" ]] && {" > /entrypoint.sh && \
    echo "    echo \"Starting Couchbase Server -- Web UI available at http://<ip>:8091 and logs available in /opt/couchbase/var/lib/couchbase/logs\"" >> /entrypoint.sh && \
    echo "    exec /usr/sbin/runsvdir-start\n}\n\nexec \"\$@\"" >> /entrypoint.sh && \
    chmod 775 /entrypoint.sh

# 8091: Couchbase Web console, REST/HTTP interface
# 8092: Views, queries, XDCR
# 8093: Query services (4.0+)
# 8094: Full-text Search (4.5+)
# 11207: Smart client library data node access (SSL)
# 11210: Smart client library/moxi data node access
# 11211: Legacy non-smart client library data node access
# 18091: Couchbase Web console, REST/HTTP interface (SSL)
# 18092: Views, query, XDCR (SSL)
# 18093: Query services (SSL) (4.0+)
EXPOSE 8091 8092 8093 8094 11207 11210 11211 18091 18092 18093
VOLUME /opt/couchbase/var

# Peernova hooks
RUN mkdir -p /peernova/couchbase
RUN echo "COUCHBASE_USER='peernova'" > /peernova/couchbase/env && \
    echo "COUCHBASE_PASSWORD='peernova'" >> /peernova/couchbase/env && \
    echo "COUCHBASE_HOST='127.0.0.1'" >> /peernova/couchbase/env && \
    echo "COUCHBASE_PORT=8091" >> /peernova/couchbase/env && \
    echo "COUCHBASE_MEMORY_QUOTA=2048" >> /peernova/couchbase/env
RUN echo "#!/bin/bash\n\nset -em\n\nsleep 10\n" > /peernova/couchbase/run.sh && \
    echo ". /peernova/couchbase/env\n" >> /peernova/couchbase/run.sh && \
    echo "/opt/couchbase/bin/curl -v -X POST http://\$COUCHBASE_HOST:\$COUCHBASE_PORT/pools/default -d memoryQuota=\$COUCHBASE_MEMORY_QUOTA -d indexMemoryQuota=\$COUCHBASE_MEMORY_QUOTA" >> /peernova/couchbase/run.sh && \
    echo "/opt/couchbase/bin/curl -v http://\$COUCHBASE_HOST:\$COUCHBASE_PORT/node/controller/setupServices -d services=kv%2cn1ql%2Cindex" >> /peernova/couchbase/run.sh && \
    echo "/opt/couchbase/bin/curl -v http://\$COUCHBASE_HOST:\$COUCHBASE_PORT/settings/web -d port=\$COUCHBASE_PORT -d username=\$COUCHBASE_USER -d password=\$COUCHBASE_PASSWORD" >> /peernova/couchbase/run.sh && \
    echo "/opt/couchbase/bin/curl -v http://\$COUCHBASE_HOST:\$COUCHBASE_PORT/settings/indexes -d 'storageMode=memory_optimized' -d username=\$COUCHBASE_USER -d password=\$COUCHBASE_PASSWORD" >> /peernova/couchbase/run.sh && \
    echo "\nrm -f /peernova/couchbase/env\n" >> /peernova/couchbase/run.sh && \
    chmod 775 /peernova/couchbase/run.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["couchbase-server"]


########    RIAK KV    ########`

	job := BuildImage(
		&BuildImageArgs{
			ResourceRegistry:   args.ResourceRegistry,
			PrepareImage:       args.LinuxImageResource,
			From:               args.CuneiformImageResource,
			Name:               "service",
			DockerFileResource: steps,
			Image:              imageResource,
		})
	job.AddToGroup(project.SystemGroup)

	imageResource.NeedJobs(job)

	return imageResource
}
