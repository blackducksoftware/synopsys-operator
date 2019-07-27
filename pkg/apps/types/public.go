package types

// PublicVersion refers to public version configuration
type PublicVersion struct {
	Size       ComponentName
	RCs        map[string]PublicRC
	Services   []ComponentName
	ConfigMaps []ComponentName
	Secrets    []ComponentName
	PVC        []ComponentName
}

// PublicRC refers to public container configuration
type PublicRC struct {
	Identifier ComponentName
	Container  map[ContainerName]string
}
