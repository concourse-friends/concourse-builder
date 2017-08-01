package library

import (
	"path"

	"github.com/concourse-friends/concourse-builder/project"
)

// Location is a struct that directs to file or directory inside a volume
type Location struct {
	Volume       IVolume
	RelativePath string
}

func (l *Location) Path() string {
	if l.Volume != nil {
		return path.Join(l.Volume.Path(), l.RelativePath)
	}
	return l.RelativePath
}

func (l *Location) InputResource() *project.JobResource {
	if res, ok := l.Volume.(*project.JobResource); ok {
		return res
	}
	return nil
}

func (l *Location) OutputName() string {
	if res, ok := l.Volume.(*TaskOutput); ok {
		return res.Name()
	}
	return ""
}

func (l *Location) Value() interface{} {
	return l.Path()
}
