package controllers_utils

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	//"github.com/go-logr/logr"
)

func ConvertYamlFileToRuntimeObjects(stringContent string) map[string]runtime.Object {
	// TODO: use logr.Logr
	log := ctrl.Log.WithName("ConvertYamlFileToRuntimeObjects")

	listOfSingleK8sResourceYaml := strings.Split(stringContent, "---")
	mapOfUniqueIdToDesiredRuntimeObject := make(map[string]runtime.Object, 0)

	for _, singleYaml := range listOfSingleK8sResourceYaml {
		if singleYaml == "\n" || singleYaml == "" {
			// ignore empty cases
			continue
		}
		decode := scheme.Codecs.UniversalDeserializer().Decode
		runtimeObject, groupVersionKind, err := decode([]byte(singleYaml), nil, nil)
		if err != nil {
			log.V(1).Info("unable to decode a single yaml object, skipping", "singleYaml", singleYaml)
			continue
		}

		accessor := meta.NewAccessor()
		runtimeObjectKind := groupVersionKind.Kind
		runtimeObjectName, err := accessor.Name(runtimeObject)
		if err != nil {
			log.V(1).Info("Failed to get runtimeObject's name", "err", err)
			continue
		}
		uniqueId := fmt.Sprintf("%s.%s", runtimeObjectKind, runtimeObjectName)
		log.V(1).Info("creating runtime object label", "uniqueId", uniqueId)
		mapOfUniqueIdToDesiredRuntimeObject[uniqueId] = runtimeObject
	}
	return mapOfUniqueIdToDesiredRuntimeObject
}
