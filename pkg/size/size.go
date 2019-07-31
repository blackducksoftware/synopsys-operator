package size

import (
	sizev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/size/v1"
	types2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strings"
)

// GetDefaultSize returns the default size. This will be used ny synopsysctl to create the Size custom resources during the deployment
func GetDefaultSize(name string) *sizev1.Size {
	switch strings.ToUpper(name) {
	case "SMALL":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				Rc: map[string]sizev1.RCSize{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(4608),
								MaxMem: util.IntToInt32(4608),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ScanContainerName): {
								MinMem: util.IntToInt32(2560),
								MaxMem: util.IntToInt32(2560),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebappContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2560),
								MaxMem: util.IntToInt32(2560),
							},
							string(types2.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebserverContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.PostgresContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(3072),
								MaxMem: util.IntToInt32(3072),
							},
						},
					},
				}},
		}
	case "MEDIUM":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				Rc: map[string]sizev1.RCSize{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 4,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(4),
								MaxCPU: util.IntToInt32(4),
								MinMem: util.IntToInt32(7168),
								MaxMem: util.IntToInt32(7168),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 2,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ScanContainerName): {
								MinMem: util.IntToInt32(5120),
								MaxMem: util.IntToInt32(5120),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebappContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(5120),
								MaxMem: util.IntToInt32(5120),
							},
							string(types2.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebserverContainerName): {
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.PostgresContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(8192),
								MaxMem: util.IntToInt32(8192),
							},
						},
					},
				},
			},
		}
	case "LARGE":
		return &sizev1.Size{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: sizev1.SizeSpec{
				Rc: map[string]sizev1.RCSize{
					"authentication": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.AuthenticationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"binaryscanner": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.BinaryScannerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"cfssl": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.CfsslContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"documentation": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.DocumentationContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"jobrunner": {
						Replica: 6,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.JobrunnerContainerName): {
								MinCPU: util.IntToInt32(1),
								MaxCPU: util.IntToInt32(1),
								MinMem: util.IntToInt32(13824),
								MaxMem: util.IntToInt32(13824),
							},
						},
					},
					"rabbitmq": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RabbitMQContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"registration": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.RegistrationContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"scan": {
						Replica: 3,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ScanContainerName): {
								MinMem: util.IntToInt32(9728),
								MaxMem: util.IntToInt32(9728),
							},
						},
					},
					"solr": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.SolrContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"uploadcache": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.UploadCacheContainerName): {
								MinMem: util.IntToInt32(512),
								MaxMem: util.IntToInt32(512),
							},
						},
					},
					"webapp-logstash": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebappContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(9728),
								MaxMem: util.IntToInt32(9728),
							},
							string(types2.LogstashContainerName): {
								MinMem: util.IntToInt32(1024),
								MaxMem: util.IntToInt32(1024),
							},
						},
					},
					"webserver": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.WebserverContainerName): {
								MinMem: util.IntToInt32(2048),
								MaxMem: util.IntToInt32(2048),
							},
						},
					},
					"zookeeper": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.ZookeeperContainerName): {
								MinMem: util.IntToInt32(640),
								MaxMem: util.IntToInt32(640),
							},
						},
					},
					"postgres": {
						Replica: 1,
						ContainerLimit: map[string]sizev1.ContainerSize{
							string(types2.PostgresContainerName): {
								MinCPU: util.IntToInt32(2),
								MaxCPU: util.IntToInt32(2),
								MinMem: util.IntToInt32(12288),
								MaxMem: util.IntToInt32(12288),
							},
						},
					},
				},
			},
		}
	}
	return nil
}
