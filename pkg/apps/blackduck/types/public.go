package types

type PublicVersion struct {
	Size       ComponentName
	RCs        map[string]PublicRC
	Services   []ComponentName
	ConfigMaps []ComponentName
	Secrets    []ComponentName
	PVC        []ComponentName
}

type PublicRC struct {
	Identifier ComponentName
	Container  map[ContainerName]string
}
