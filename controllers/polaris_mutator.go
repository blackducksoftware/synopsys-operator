package controllers

import (
	"fmt"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchPolaris(client client.Client, polarisCr *synopsysv1.Polaris, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object) map[string]runtime.Object {
	patcher := PolarisPatcher{
		Client:                           client,
		polarisCr:                        polarisCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
	}
	return patcher.patch()
}

type PolarisPatcher struct {
	polarisCr                        *synopsysv1.Polaris
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	client.Client
}

func (p *PolarisPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *PolarisPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polarisCr.Spec.Namespace)
	}
	return nil
}
