package resource

import (
	"bytes"
	"fmt"

	"github.com/concourse-friends/concourse-builder/model"
	"gopkg.in/yaml.v2"
)

// An object that tracks collection of resources by name
type TypeRegistry struct {
	types map[model.ResourceTypeName]*model.ResourceType
}

func (r *TypeRegistry) MustRegisterType(resourceType *model.ResourceType) {
	res, ok := r.types[resourceType.Name]
	if ok {
		current, err := yaml.Marshal(res)
		if err != nil {
			panic(err.Error())
		}
		new, err := yaml.Marshal(resourceType)
		if err != nil {
			panic(err.Error())
		}

		if bytes.Compare(current, new) != 0 {
			panic(fmt.Sprintf(
				"Resource type with name %s is already registered with different content", resourceType.Name))
		}
		return
	}

	r.types[resourceType.Name] = resourceType
}

func (r *TypeRegistry) RegisterType(resourceTypeName model.ResourceTypeName) *model.ResourceType {
	return r.types[resourceTypeName]
}

var GlobalTypeRegistry = initTypeRegistry()

func initTypeRegistry() *TypeRegistry {
	registry := &TypeRegistry{
		types: make(map[model.ResourceTypeName]*model.ResourceType),
	}

	return registry
}
