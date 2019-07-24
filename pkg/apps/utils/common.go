package utils

import "fmt"

// GetVersionLabel will return the label including the version
func GetVersionLabel(componentMame string, name string, version string) map[string]string {
	m := GetLabel(componentMame, name)
	m["version"] = version
	return m
}

// GetLabel will return the label
func GetLabel(componentMame string, name string) map[string]string {
	return map[string]string{
		"app":       "blackduck",
		"name":      name,
		"component": componentMame,
	}
}

// GetResourceName returns the name of the resource
func GetResourceName(name string, appName string, defaultName string) string {
	if len(appName) == 0 {
		return fmt.Sprintf("%s-%s", name, defaultName)
	}

	if len(defaultName) == 0 {
		return fmt.Sprintf("%s-%s", name, appName)
	}

	return fmt.Sprintf("%s-%s-%s", name, appName, defaultName)
}
