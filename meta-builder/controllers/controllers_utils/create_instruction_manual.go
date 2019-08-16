package controllers_utils

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/flying-dutchman"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

func CreateInstructionManual(objects map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {

	dependencyYamlStruct := &flying_dutchman.RuntimeObjectDependencyYaml{}

	accessor := meta.NewAccessor()

	for k, v := range objects {
		labels, err := accessor.Labels(v)
		if err != nil {
			return nil, err
		}

		group, ok := labels["operator.synopsys.com/group-id"]
		if !ok {
			return nil, fmt.Errorf("couldn't retrieve group label of %s", k)
		}
		if dependencyYamlStruct.Groups == nil {
			dependencyYamlStruct.Groups = make(map[string][]string)
		}
		dependencyYamlStruct.Groups[group] = append(dependencyYamlStruct.Groups[group], k)

		dependencies, ok := labels["operator.synopsys.com/group-dependencies"]
		if !ok {
			return nil, fmt.Errorf("couldn't retrieve group dependencies of %s", k)
		}
		// TODO have RuntimeObjectDependency take an array of dependencies
		if len(dependencies) > 0 {
			for _, dep := range strings.Split(dependencies, "_") {
				isDepAlreadyPresent := false
				for _, value := range dependencyYamlStruct.Dependencies {
					if strings.Compare(value.Obj, group) == 0 && strings.Compare(value.IsDependentOn, strings.TrimSpace(dep)) == 0 {
						isDepAlreadyPresent = true
						break
					}
				}
				if !isDepAlreadyPresent {
					dependencyYamlStruct.Dependencies = append(dependencyYamlStruct.Dependencies, flying_dutchman.RuntimeObjectDependency{
						Obj:           group,
						IsDependentOn: strings.TrimSpace(dep),
					})
				}
			}
		}
	}

	return dependencyYamlStruct, nil
}
