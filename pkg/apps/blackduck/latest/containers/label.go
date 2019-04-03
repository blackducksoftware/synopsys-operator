package containers

// GetVersionLabel will return the label including the version
func (c *Creater) GetVersionLabel(name string) map[string]string {
	m := c.GetLabel(name)
	m["version"] = c.hubSpec.Version
	return m
}

// GetLabel will return the label
func (c *Creater) GetLabel(name string) map[string]string {
	// var tier string
	// switch strings.ToLower(name) {
	// case "postgres":
	// 	tier = "backend"
	// case "uploadcache":
	// 	tier = "uploadcache"
	// default:
	// 	tier = "frontend"
	// }
	return map[string]string{
		"app":       "blackduck",
		"component": name,
		//"tier":      tier,
	}
}
