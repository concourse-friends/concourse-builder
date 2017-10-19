// Copyright (C) 2017 PeerNova, Inc.
//
// All rights reserved.
//
// PeerNova and Cuneiform are trademarks of PeerNova, Inc. References to
// third-party marks or brands are the property of their respective owners.
// No rights or licenses are granted, express or implied, unless set forth in
// a written agreement signed by PeerNova, Inc. You may not distribute,
// disseminate, copy, record, modify, enhance, supplement, create derivative
// works from, adapt, or translate any content contained herein except as
// otherwise expressly permitted pursuant to a written agreement signed by
// PeerNova, Inc.

package library

import (
	"github.com/concourse-friends/concourse-builder/library/image"
	"github.com/concourse-friends/concourse-builder/library/primitive"
	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type AwsImageJobArgs struct {
	LinuxImageResource  *project.Resource
	ConcourseBuilderGit *project.Resource
	ImageRegistry       *image.Registry
	ResourceRegistry    *project.ResourceRegistry
	Concourse           *primitive.Concourse
}

func AwsImageJob(args *AwsImageJobArgs) *project.Resource {
	resourceName := project.ResourceName("aws-image")
	imageResource := args.ResourceRegistry.GetResource(resourceName)

	if imageResource != nil {
		return imageResource
	}

	imageResource = &project.Resource{
		Name:  resourceName,
		Type:  resource.ImageResourceType.Name,
		Scope: project.AllTeamsScope,
		Source: &image.Source{
			Registry:   args.ImageRegistry,
			Repository: "concourse-builder/aws-image",
		},
	}

	args.ResourceRegistry.MustRegister(imageResource)
	dockerSteps := &primitive.Location{
		Volume:       args.ResourceRegistry.JobResource(args.ConcourseBuilderGit, true, nil),
		RelativePath: "docker/aws",
	}

	job := BuildImage(
		&BuildImageArgs{
			ResourceRegistry:   args.ResourceRegistry,
			PrepareImage:       image.Ubuntu,
			From:               args.LinuxImageResource,
			Name:               "aws",
			DockerFileResource: dockerSteps,
			Image:              imageResource,
		})
	job.AddToGroup(project.SystemGroup)

	imageResource.NeedJobs(job)

	return imageResource
}

