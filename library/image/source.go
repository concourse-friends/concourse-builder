package image

import (
	"fmt"
	"path"

	"github.com/concourse-friends/concourse-builder/project"
	"github.com/concourse-friends/concourse-builder/resource"
)

type Source struct {
	Registry   *Registry
	Repository string
	Tag        Tag
	Name       string
}

func (im *Source) ModelSource(scope project.Scope, info *project.ScopeInfo) interface{} {
	repository := im.Repository
	if im.Registry.Domain != "" {
		repository = path.Join(im.Registry.Domain, repository)
	}

	tag := info.Scope(scope, "-")
	if tag != "" && im.Tag == "" {
		tag = tag[:len(tag)-1]
	} else {
		tag = tag + string(im.Tag)
	}

	source := &resource.ImageSource{
		Repository: repository,
		Tag:        string(ConvertToImageTag(tag)),
		TargetName: im.Name
	}

	if im.Registry.AwsAccessKeyId != "" || im.Registry.AwsSecretAccessKey != "" {
		if im.Registry.AwsAccessKeyId == "" || im.Registry.AwsSecretAccessKey == "" {
			return fmt.Errorf(
				"For ImageRegistry AwsAccessKeyId and AwsSecretAccessKey make sense only as pair",
				im.Registry.Domain)
		}

		source.AwsAccessKeyID = im.Registry.AwsAccessKeyId
		source.AwsSecretAccessKey = im.Registry.AwsSecretAccessKey
	}

	return source
}
