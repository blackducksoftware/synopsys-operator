package flyingdutchman

// RuntimeObjectDependency configure the run object dependency
type RuntimeObjectDependency struct {
	Obj           string   `yaml:"obj"`
	IsDependentOn []string `yaml:"isdependenton"`
}

// RuntimeObjectDependencyYaml configure the run object dependency from yaml
type RuntimeObjectDependencyYaml struct {
	Groups       map[string][]string       `yaml:"runtimeobjectsgroupings"`
	Dependencies []RuntimeObjectDependency `yaml:"runtimeobjectdependencies"`
}
