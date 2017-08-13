package project

import (
	"fmt"
)

// An object that tracks collection of resources by name
type ResourceRegistry struct {
	resources map[ResourceName]*Resource
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		resources: make(map[ResourceName]*Resource),
	}
}

func (r *ResourceRegistry) MustRegister(resource *Resource) {
	_, ok := r.resources[resource.Name]
	if ok {
		// TODO: check if register more than once, the registered is the same
		return
	}

	r.resources[resource.Name] = resource
}

func (r *ResourceRegistry) GetResource(name ResourceName) *Resource {
	if res, ok := r.resources[name]; ok {
		return res
	}
	return nil
}

func (r *ResourceRegistry) MustGetResource(name ResourceName) *Resource {
	res, ok := r.resources[name]
	if !ok {
		panic(fmt.Sprintf("Resource %s is not found in the registry", name))
	}
	return res
}
