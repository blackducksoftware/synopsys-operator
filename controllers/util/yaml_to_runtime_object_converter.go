package util

import (
	"fmt"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ConvertYamlFileToRuntimeObjects converts the yaml file string to map of runtime object
func ConvertYamlFileToRuntimeObjects(stringContent string, isOpenShift bool) map[string]runtime.Object {
	routev1.AddToScheme(scheme.Scheme)
	securityv1.AddToScheme(scheme.Scheme)
	// TODO: use logr.Logr
	log := ctrl.Log.WithName("ConvertYamlFileToRuntimeObjects")

	listOfSingleK8sResourceYaml := strings.Split(stringContent, "---")
	mapOfUniqueIDToDesiredRuntimeObject := make(map[string]runtime.Object, 0)

	log.V(1).Info("listOfYamlsToGoThrough", "Len Yamls", len(listOfSingleK8sResourceYaml))

	for _, singleYaml := range listOfSingleK8sResourceYaml {
		if singleYaml == "\n" || singleYaml == "" {
			// ignore empty cases
			log.V(1).Info("Got empty", "here", singleYaml)
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		runtimeObject, groupVersionKind, err := decode([]byte(singleYaml), nil, nil)
		if err != nil {
			log.V(1).Info("unable to decode a single yaml object, skipping", "singleYaml", singleYaml, "error", err)
			continue
		}

		accessor := meta.NewAccessor()
		runtimeObjectKind := groupVersionKind.Kind
		runtimeObjectName, err := accessor.Name(runtimeObject)
		if err != nil {
			log.V(1).Info("Failed to get runtimeObject's name", "err", err)
			continue
		}
		uniqueID := fmt.Sprintf("%s.%s", runtimeObjectKind, runtimeObjectName)
		log.V(1).Info("creating runtime object label", "uniqueId", uniqueID)
		mapOfUniqueIDToDesiredRuntimeObject[uniqueID] = runtimeObject
	}
	return mapOfUniqueIDToDesiredRuntimeObject
}
