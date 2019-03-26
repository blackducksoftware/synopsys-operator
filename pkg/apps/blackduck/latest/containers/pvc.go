package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GetPVCs will return the PVCs
func (c *Creater) GetPVCs() []*components.PersistentVolumeClaim {
	var pvcs []*components.PersistentVolumeClaim

	if c.hubSpec.PersistentStorage {
		for _, claim := range c.hubSpec.PVC {
			var defaultsize string
			// Set default value if size isn't specified
			switch claim.Name {
			case "blackduck-postgres":
				defaultsize = "150Gi"
			case "blackduck-authentication":
				defaultsize = "2Gi"
			case "blackduck-cfssl":
				defaultsize = "2Gi"
			case "blackduck-registration":
				defaultsize = "2Gi"
			case "blackduck-solr":
				defaultsize = "2Gi"
			case "blackduck-webapp":
				defaultsize = "2Gi"
			case "blackduck-logstash":
				defaultsize = "20Gi"
			case "blackduck-zookeeper-data":
				defaultsize = "2Gi"
			case "blackduck-zookeeper-datalog":
				defaultsize = "2Gi"
			case "blackduck-rabbitmq":
				defaultsize = "5Gi"
			case "blackduck-uploadcache-data":
				defaultsize = "100Gi"
			case "blackduck-uploadcache-key":
				defaultsize = "2Gi"
			default:
				defaultsize = claim.Size
			}
			pvcs = append(pvcs, c.createPVC(claim.Name, claim.Size, defaultsize, claim.StorageClass, horizonapi.ReadWriteOnce))
		}
	}

	return pvcs
}

func (c *Creater) createPVC(name string, requestedSize string, defaultSize string, storageclass string, accessMode horizonapi.PVCAccessModeType) *components.PersistentVolumeClaim {
	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string
	if len(storageclass) > 0 {
		class = &storageclass
	} else if len(c.hubSpec.PVCStorageClass) > 0 {
		class = &c.hubSpec.PVCStorageClass
	} else {
		class = nil
	}

	var size string
	_, err := resource.ParseQuantity(requestedSize)
	if err != nil {
		size = defaultSize
	} else {
		size = requestedSize
	}

	pvc, _ := components.NewPersistentVolumeClaim(horizonapi.PVCConfig{
		Name:      name,
		Namespace: c.hubSpec.Namespace,
		Size:      size,
		Class:     class,
	})

	pvc.AddAccessMode(accessMode)

	return pvc
}
