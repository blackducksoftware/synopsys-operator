package containers

// GetVersionLabel will return the label including the version
func (c *Creater) GetVersionLabel(name string) map[string]string {
	m := c.GetLabel(name)
	m["version"] = c.hubSpec.Version
	return m
}

// GetLabel will return the label
func (c *Creater) GetLabel(name string) map[string]string {
	return map[string]string{
		"app":       "blackduck",
		"component": name,
	}
}
