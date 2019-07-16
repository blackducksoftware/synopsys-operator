package v1

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
)

type size struct{}

func (s size) GetSize(name string) map[string]*types.Size {
	switch name {
	case "small":
		return map[string]*types.Size{
			"authentication": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.AuthenticationContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 1024,
						MaxMem: 1024,
					},
				},
			},
			"binaryscanner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.BinaryScannerContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 2048,
						MaxMem: 2048,
					},
				},
			},
			"cfssl": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.CfsslContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 640,
						MaxMem: 640,
					},
				},
			},
			"documentation": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.DocumentationContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 512,
						MaxMem: 512,
					},
				},
			},
			"jobrunner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.JobrunnerContainerName: {
						MinCPU: 1,
						MaxCPU: 1,
						MinMem: 4608,
						MaxMem: 4608,
					},
				},
			},
			"rabbitmq": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.RabbitMQContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 1024,
						MaxMem: 1024,
					},
				},
			},
			"registration": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.RegistrationContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 640,
						MaxMem: 640,
					},
				},
			},
			"scan": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.ScanContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 2560,
						MaxMem: 2560,
					},
				},
			},
			"solr": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.SolrContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 640,
						MaxMem: 640,
					},
				},
			},
			"uploadcache": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.UploadCacheContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 512,
						MaxMem: 512,
					},
				},
			},
			"webapp-logstash": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.WebappContainerName: {
						MinCPU: 1,
						MaxCPU: 1,
						MinMem: 2560,
						MaxMem: 2560,
					},
					types.LogstashContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 1024,
						MaxMem: 1024,
					},
				},
			},
			"webserver": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.WebserverContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 512,
						MaxMem: 512,
					},
				},
			},
			"zookeeper": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.ZookeeperContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 640,
						MaxMem: 640,
					},
				},
			},
			"postgres": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					types.PostgresContainerName: {
						MinCPU: 0,
						MaxCPU: 0,
						MinMem: 2048,
						MaxMem: 2048,
					},
				},
			},
		}
	}
	return nil
}

func init() {
	store.Register(types.SizeV1, &size{})
}
