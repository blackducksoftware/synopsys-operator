package v1

import (
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

type size struct{}

// GetSize returns the size for each container
func (s size) GetSize(name string) map[string]*types.Size {
	switch strings.ToUpper(name) {
	case "SMALL":
		return map[string]*types.Size{
			"authentication": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.AuthenticationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"binaryscanner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.BinaryScannerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"cfssl": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.CfsslContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"documentation": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.DocumentationContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"jobrunner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.JobrunnerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(4608),
						MaxMem: util.IntToInt32(4608),
					},
				},
			},
			"rabbitmq": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RabbitMQContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"registration": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RegistrationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"scan": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ScanContainerName: {
						MinMem: util.IntToInt32(2560),
						MaxMem: util.IntToInt32(2560),
					},
				},
			},
			"solr": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.SolrContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"uploadcache": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.UploadCacheContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"webapp-logstash": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebappContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(2560),
						MaxMem: util.IntToInt32(2560),
					},
					blackduck.LogstashContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"webserver": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebserverContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"zookeeper": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ZookeeperContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"postgres": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.PostgresContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(3072),
						MaxMem: util.IntToInt32(3072),
					},
				},
			},
		}
	case "MEDIUM":
		return map[string]*types.Size{
			"authentication": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.AuthenticationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"binaryscanner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.BinaryScannerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"cfssl": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.CfsslContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"documentation": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.DocumentationContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"jobrunner": {
				Replica: 4,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.JobrunnerContainerName: {
						MinCPU: util.IntToInt32(4),
						MaxCPU: util.IntToInt32(4),
						MinMem: util.IntToInt32(7168),
						MaxMem: util.IntToInt32(7168),
					},
				},
			},
			"rabbitmq": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RabbitMQContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"registration": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RegistrationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"scan": {
				Replica: 2,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ScanContainerName: {
						MinMem: util.IntToInt32(5120),
						MaxMem: util.IntToInt32(5120),
					},
				},
			},
			"solr": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.SolrContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"uploadcache": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.UploadCacheContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"webapp-logstash": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebappContainerName: {
						MinCPU: util.IntToInt32(2),
						MaxCPU: util.IntToInt32(2),
						MinMem: util.IntToInt32(5120),
						MaxMem: util.IntToInt32(5120),
					},
					blackduck.LogstashContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"webserver": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebserverContainerName: {
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"zookeeper": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ZookeeperContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"postgres": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.PostgresContainerName: {
						MinCPU: util.IntToInt32(2),
						MaxCPU: util.IntToInt32(2),
						MinMem: util.IntToInt32(8192),
						MaxMem: util.IntToInt32(8192),
					},
				},
			},
		}
	case "LARGE":
		return map[string]*types.Size{
			"authentication": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.AuthenticationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"binaryscanner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.BinaryScannerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"cfssl": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.CfsslContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"documentation": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.DocumentationContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"jobrunner": {
				Replica: 6,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.JobrunnerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(13824),
						MaxMem: util.IntToInt32(13824),
					},
				},
			},
			"rabbitmq": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RabbitMQContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"registration": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RegistrationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"scan": {
				Replica: 3,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ScanContainerName: {
						MinMem: util.IntToInt32(9728),
						MaxMem: util.IntToInt32(9728),
					},
				},
			},
			"solr": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.SolrContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"uploadcache": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.UploadCacheContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"webapp-logstash": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebappContainerName: {
						MinCPU: util.IntToInt32(2),
						MaxCPU: util.IntToInt32(2),
						MinMem: util.IntToInt32(9728),
						MaxMem: util.IntToInt32(9728),
					},
					blackduck.LogstashContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"webserver": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebserverContainerName: {
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"zookeeper": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ZookeeperContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"postgres": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.PostgresContainerName: {
						MinCPU: util.IntToInt32(2),
						MaxCPU: util.IntToInt32(2),
						MinMem: util.IntToInt32(12288),
						MaxMem: util.IntToInt32(12288),
					},
				},
			},
		}
	case "X-LARGE":
		return map[string]*types.Size{
			"authentication": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.AuthenticationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"binaryscanner": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.BinaryScannerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"cfssl": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.CfsslContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"documentation": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.DocumentationContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"jobrunner": {
				Replica: 10,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.JobrunnerContainerName: {
						MinCPU: util.IntToInt32(1),
						MaxCPU: util.IntToInt32(1),
						MinMem: util.IntToInt32(13824),
						MaxMem: util.IntToInt32(13824),
					},
				},
			},
			"rabbitmq": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RabbitMQContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"registration": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.RegistrationContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"scan": {
				Replica: 5,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ScanContainerName: {
						MinMem: util.IntToInt32(9728),
						MaxMem: util.IntToInt32(9728),
					},
				},
			},
			"solr": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.SolrContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"uploadcache": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.UploadCacheContainerName: {
						MinMem: util.IntToInt32(512),
						MaxMem: util.IntToInt32(512),
					},
				},
			},
			"webapp-logstash": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebappContainerName: {
						MinCPU: util.IntToInt32(3),
						MaxCPU: util.IntToInt32(3),
						MinMem: util.IntToInt32(9728),
						MaxMem: util.IntToInt32(9728),
					},
					blackduck.LogstashContainerName: {
						MinMem: util.IntToInt32(1024),
						MaxMem: util.IntToInt32(1024),
					},
				},
			},
			"webserver": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.WebserverContainerName: {
						MinMem: util.IntToInt32(2048),
						MaxMem: util.IntToInt32(2048),
					},
				},
			},
			"zookeeper": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.ZookeeperContainerName: {
						MinMem: util.IntToInt32(640),
						MaxMem: util.IntToInt32(640),
					},
				},
			},
			"postgres": {
				Replica: 1,
				Containers: map[types.ContainerName]types.ContainerSize{
					blackduck.PostgresContainerName: {
						MinCPU: util.IntToInt32(3),
						MaxCPU: util.IntToInt32(3),
						MinMem: util.IntToInt32(12288),
						MaxMem: util.IntToInt32(12288),
					},
				},
			},
		}
	}
	return nil
}

func init() {
	store.Register(blackduck.BlackDuckSizeV1, &size{})
}
