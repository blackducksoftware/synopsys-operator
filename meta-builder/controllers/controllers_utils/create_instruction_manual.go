package controllers_utils

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/flying-dutchman"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

func CreateInstructionManual(mapOfUniqueIdToDesiredRuntimeObject map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {

	dependencyYamlStruct := &flying_dutchman.RuntimeObjectDependencyYaml{}

	accessor := meta.NewAccessor()

	for uniqueId, desiredRuntimeObject := range mapOfUniqueIdToDesiredRuntimeObject {
		labels, err := accessor.Labels(desiredRuntimeObject)
		if err != nil {
			return nil, err
		}

		group, ok := labels["operator.synopsys.com/group-id"]
		if !ok {
			return nil, fmt.Errorf("couldn't retrieve group label of %s", uniqueId)
		}

		if dependencyYamlStruct.Groups == nil {
			dependencyYamlStruct.Groups = make(map[string][]string)
		}
		dependencyYamlStruct.Groups[group] = append(dependencyYamlStruct.Groups[group], uniqueId)

		dependencies, ok := labels["operator.synopsys.com/group-dependencies"]
		if !ok {
			return nil, fmt.Errorf("couldn't retrieve group dependencies of %s", uniqueId)
		}

		if len(dependencies) > 0 {
			for _, dependency := range strings.Split(dependencies, "_") {

				isDepAlreadyPresent := false
				for _, value := range dependencyYamlStruct.Dependencies {
					if strings.Compare(value.Obj, group) == 0 {
						value.IsDependentOn = append(value.IsDependentOn, strings.TrimSpace(dependency))
						isDepAlreadyPresent = true
						break
					}
				}

				if !isDepAlreadyPresent {
					dependencyYamlStruct.Dependencies = append(dependencyYamlStruct.Dependencies, flying_dutchman.RuntimeObjectDependency{
						Obj:           group,
						IsDependentOn: []string{strings.TrimSpace(dependency)},
					})
				}
			}
		}

	}

	return dependencyYamlStruct, nil
}
