package flying_dutchman

// Dependency Resources from YAML file
type RuntimeObjectDependency struct {
	Obj           string `yaml:"obj"`
	IsDependentOn string `yaml:"isdependenton"`
}
type RuntimeObjectDepencyYaml struct {
	Groups       map[string][]string       `yaml:"runtimeobjectsgroupings"`
	Dependencies []RuntimeObjectDependency `yaml:"runtimeobjectdependencies"`
}
