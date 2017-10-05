package project

import (
	"io"
	"log"
	"sort"

	"github.com/concourse-friends/concourse-builder/model"
	yaml "gopkg.in/yaml.v2"
)

type AllJobsGroupOption int

const (
	AllJobsGroupNone = iota
	AllJobsGroupFirst
	AllJobsGroupLast
)

type Pipeline struct {
	Name PipelineName

	AllJobsGroup AllJobsGroupOption
	Jobs         Jobs

	ResourceRegistry *ResourceRegistry
}

type Pipelines []*Pipeline

func NewPipeline() *Pipeline {
	return &Pipeline{
		ResourceRegistry: NewResourceRegistry(),
	}
}

func (p *Pipeline) JobsFor(checkJobs JobsSet) (Jobs, error) {
	jobs := make(map[JobName]*Job)
	for job := checkJobs.Pop(); job != nil; job = checkJobs.Pop() {
		log.Printf("Check job %s resources", job.Name)
		jobs[job.Name] = job

		resources, err := job.InputResources()
		if err != nil {
			return nil, err
		}

		for _, resource := range resources {
			projectResource := p.ResourceRegistry.MustGetResource(resource.Name)
			for _, resJob := range projectResource.NeededJobs() {
				job.AddJobToRunAfter(resJob)
				if _, exists := jobs[resJob.Name]; exists {
					continue
				}
				checkJobs[resJob] = struct{}{}
				jobs[resJob.Name] = resJob
			}
		}
	}

	sliceJobs := make(Jobs, 0, len(jobs))
	for _, job := range jobs {
		sliceJobs = append(sliceJobs, job)
	}

	return sliceJobs, nil
}

func (p *Pipeline) AllJobs() (Jobs, error) {
	checkJobs := make(JobsSet)

	for _, job := range p.Jobs {
		checkJobs[job] = struct{}{}
	}

	return p.JobsFor(checkJobs)
}

func (p *Pipeline) ModelGroups(allJobs Jobs) (model.Groups, error) {
	groups := make(map[*JobGroup][]string)

	jobsThatHaveNoGroup := false

	allJobNames := make([]string, 0, len(allJobs))
	for _, job := range allJobs {
		allJobNames = append(allJobNames, string(job.Name))

		if len(job.Groups) == 0 {
			jobsThatHaveNoGroup = true
		}

		for _, group := range job.Groups {
			groups[group] = append(groups[group], string(job.Name))
		}
	}
	sort.Strings(allJobNames)

	groupsByOrder := JobGroups{}
	for group := range groups {
		groupsByOrder = append(groupsByOrder, group)
	}

	groupsByOrder, err := SortJobGroups(groupsByOrder)
	if err != nil {
		return nil, err
	}

	if p.AllJobsGroup != AllJobsGroupNone && (jobsThatHaveNoGroup || len(groups) > 1) {
		all := &JobGroup{
			Name: "all",
		}

		if p.AllJobsGroup == AllJobsGroupFirst {
			groupsByOrder = append(JobGroups{all}, groupsByOrder...)
		} else /* if p.AllJobsGroup == AllJobsGroupLast */ {
			groupsByOrder = append(groupsByOrder, all)
		}
		groups[all] = allJobNames
	}

	modelGroups := model.Groups{}
	for _, group := range groupsByOrder {
		modelJobs := model.JobNames{}

		jobNames := groups[group]
		sort.Strings(jobNames)
		for _, jobName := range jobNames {
			modelJobs = append(modelJobs, model.JobName(jobName))
		}

		modelGroup := model.Group{
			Name: model.GroupName(group.Name),
			Jobs: modelJobs,
		}

		modelGroups = append(modelGroups, &modelGroup)
	}

	return modelGroups, nil
}

func resources(allJobs Jobs) (JobResources, error) {
	var jobResources JobResources
	for _, job := range allJobs {
		resources, err := job.Resources()
		if err != nil {
			return nil, err
		}
		jobResources = append(jobResources, resources...)
	}

	return jobResources.Deduplicate(), nil
}

func (p *Pipeline) ModelResourceTypes(info *ScopeInfo, allJobs Jobs) (model.ResourceTypes, error) {
	jobResources, err := resources(allJobs)
	if err != nil {
		return nil, err
	}

	typesSet := make(map[ResourceTypeName]struct{})

	for _, jobResource := range jobResources {
		res := p.ResourceRegistry.MustGetResource(jobResource.Name)
		resourceType := GlobalTypeRegistry.RegisterType(res.Type)
		if resourceType == nil {
			continue
		}

		typesSet[resourceType.Name] = struct{}{}
	}

	types := make([]string, 0, len(typesSet))
	for k := range typesSet {
		types = append(types, string(k))
	}

	sort.Strings(types)

	var resourceTypes model.ResourceTypes
	for _, tp := range types {
		resourceType := GlobalTypeRegistry.RegisterType(ResourceTypeName(tp))
		if !resourceType.IsSystem() {
			modelResourceType := resourceType.Model(info)
			resourceTypes = append(resourceTypes, modelResourceType)
		}
	}

	return resourceTypes, nil
}

func (p *Pipeline) ModelResources(info *ScopeInfo, allJobs Jobs) (model.Resources, error) {
	jobResources, err := resources(allJobs)
	if err != nil {
		return nil, err
	}

	var resources model.Resources
	for _, res := range jobResources {
		modelResource := res.Model(info, p.ResourceRegistry)
		resources = append(resources, modelResource)
	}

	return resources, nil
}

func (p *Pipeline) ModelJobs(allJobs Jobs) (model.Jobs, error) {
	columns, err := allJobs.SortByColumns()
	if err != nil {
		return nil, err
	}

	modelJobs := make(model.Jobs, 0, len(allJobs))

	for i, column := range columns {
		for _, job := range column {
			modelJob, err := job.Model(columns[:i])
			if err != nil {
				return nil, err
			}
			modelJobs = append(modelJobs, modelJob)
		}
	}

	return modelJobs, nil
}

func (p *Pipeline) Save(team TeamName, installation InstallationName, writer io.Writer) error {
	info := &ScopeInfo{
		Pipeline:     p.Name,
		Team:         team,
		Installation: installation,
	}

	allJobs, err := p.AllJobs()
	if err != nil {
		return err
	}

	groups, err := p.ModelGroups(allJobs)
	if err != nil {
		return err
	}

	resourceTypes, err := p.ModelResourceTypes(info, allJobs)
	if err != nil {
		return err
	}

	resources, err := p.ModelResources(info, allJobs)
	if err != nil {
		return err
	}

	jobs, err := p.ModelJobs(allJobs)
	if err != nil {
		return err
	}

	pipeline := model.Pipeline{
		Groups:        groups,
		ResourceTypes: resourceTypes,
		Resources:     resources,
		Jobs:          jobs,
	}

	result, err := yaml.Marshal(pipeline)
	if err != nil {
		return err
	}

	_, err = writer.Write(result)
	if err != nil {
		return err
	}

	return nil
}
