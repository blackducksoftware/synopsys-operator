package model

type PerceptorConfig struct {
	HubHost             string
	HubUser             string
	HubUserPassword     string
	ConcurrentScanLimit int
	UseMockMode         bool
}

type PerceptorScannerConfig struct {
	HubHost         string
	HubPort         int
	HubUser         string
	HubUserPassword string
}

type PodPerceiverConfig struct {
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
}

type ImagePerceiverConfig struct {
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
}

type PerceptorImagefacadeConfig struct {
	Dockerusername string
	Dockerpassword string
}

type PifTesterConfig struct {
}
