package project

import (
	"sort"

	"github.com/concourse-friends/concourse-builder/model"
)

type IRun interface {
	InputResources() JobResources
	Path() string
}

type IOutput interface {
	Name() string
	Path() string
}

type IParamValue interface {
	Value() interface{}
}

type IParamInput interface {
	OutputName() string
}

type IParamResource interface {
	InputResources() JobResources
}

type TaskStep struct {
	Platform model.Platform
	Name     model.TaskName
	Image    *JobResource
	Run      IRun
	Outputs  []IOutput
	Params   map[string]interface{}
}

func (ts *TaskStep) Model() (model.IStep, error) {
	task := &model.Task{
		Task: ts.Name,
		Config: &model.TaskConfig{
			Platform: ts.Platform,
			Run: &model.TaskRun{
				Path: ts.Run.Path(),
			},
			Params: make(map[string]interface{}),
		},
	}

	if ts.Image != nil {
		task.Image = model.ResourceName(ts.Image.Name)
	}

	runInputResources := ts.Run.InputResources()

	inputsMap := make(map[string]struct{})
	for _, runInputResource := range runInputResources {
		inputsMap[string(runInputResource.Name)] = struct{}{}
	}

	for _, value := range ts.Params {
		if param, ok := value.(IParamInput); ok {
			name := param.OutputName()
			if name != "" {
				inputsMap[name] = struct{}{}
			}
		}
	}

	inputs := make([]string, 0, len(inputsMap))
	for input := range inputsMap {
		inputs = append(inputs, input)
	}

	sort.Strings(inputs)
	for _, name := range inputs {
		task.Config.Inputs = append(task.Config.Inputs, &model.TaskInput{
			Name: model.ResourceName(name),
		})
	}

	for _, output := range ts.Outputs {
		task.Config.Outputs = append(task.Config.Outputs, &model.TaskOutput{
			Name: output.Name(),
			Path: output.Path(),
		})
	}

	for name, value := range ts.Params {
		if param, ok := value.(IParamValue); ok {
			task.Config.Params[name] = param.Value()
		} else {
			task.Config.Params[name] = value
		}
	}

	return task, nil
}

func (ts *TaskStep) InputResources() (JobResources, error) {
	var resources JobResources

	if ts.Image != nil {
		resources = append(resources, ts.Image)
	}

	locationResources := ts.Run.InputResources()
	resources = append(resources, locationResources...)

	for _, value := range ts.Params {
		if param, ok := value.(IParamResource); ok {
			resources = append(resources, param.InputResources()...)
		}
	}

	return resources.Deduplicate(), nil
}

func (ts *TaskStep) OutputResource() (*JobResource, error) {
	return nil, nil
}
