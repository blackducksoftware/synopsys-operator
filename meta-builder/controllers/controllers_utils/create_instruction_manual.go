package controllers_utils

import (
	"github.com/yashbhutwala/kb-synopsys-operator/flying-dutchman"
	"gopkg.in/yaml.v2"
)

func CreateInstructionManual(instructionManualLocation string) (*flying_dutchman.RuntimeObjectDepencyYaml, error) {
	// Read Dependency YAML File into Struct
	dependencyYamlBytes, err := HttpGet(instructionManualLocation)
	if err != nil {
		return nil, err
	}

	dependencyYamlStruct := &flying_dutchman.RuntimeObjectDepencyYaml{}
	err = yaml.Unmarshal(dependencyYamlBytes, dependencyYamlStruct)
	if err != nil {
		return nil, err
	}
	return dependencyYamlStruct, nil
}
