package utils

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
