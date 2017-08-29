package project

import (
	"github.com/concourse-friends/concourse-builder/model"
)

type IPutParams interface {
	ModelParams() interface{}
}

type PutStep struct {
	// The resource that will be put
	JobResource *JobResource

	// Additional resource specific parameters
	Params IPutParams

	// Additional resource specific parameters for the get operation that will follow the put operation
	GetParams interface{}
}

func (ps *PutStep) Model() (model.IStep, error) {
	put := &model.Put{
		Put:       model.ResourceName(ps.JobResource.Name),
		GetParams: ps.GetParams,
	}

	if ps.Params != nil {
		put.Params = ps.Params.ModelParams()
	}

	return put, nil
}

func (ps *PutStep) InputResources() (JobResources, error) {
	var resources JobResources

	if ps.Params != nil {
		if res, ok := ps.Params.(IInputResource); ok {
			resources = append(resources, res.InputResources()...)
		}
	}

	return resources, nil
}

func (ps *PutStep) OutputResource() (*JobResource, error) {
	return ps.JobResource, nil
}
