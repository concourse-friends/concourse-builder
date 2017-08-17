package library

import (
	"github.com/concourse-friends/concourse-builder/resource"
)

type GitMultiSource struct {
	// Git repo and credentials
	Repo *GitRepo
}

func (gms *GitMultiSource) ModelSource() interface{} {
	return &resource.GitSource{
		URI:        gms.Repo.URI,
		PrivateKey: gms.Repo.PrivateKey,
	}
}
