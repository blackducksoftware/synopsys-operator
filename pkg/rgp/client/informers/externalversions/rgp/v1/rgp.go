/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	rgpv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	versioned "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	internalinterfaces "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/informers/externalversions/internalinterfaces"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/listers/rgp/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// RgpInformer provides access to a shared informer and lister for
// Rgps.
type RgpInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.RgpLister
}

type rgpInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewRgpInformer constructs a new informer for Rgp type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRgpInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredRgpInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredRgpInformer constructs a new informer for Rgp type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredRgpInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SynopsysV1().Rgps(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SynopsysV1().Rgps(namespace).Watch(options)
			},
		},
		&rgpv1.Rgp{},
		resyncPeriod,
		indexers,
	)
}

func (f *rgpInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredRgpInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *rgpInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&rgpv1.Rgp{}, f.defaultInformer)
}

func (f *rgpInformer) Lister() v1.RgpLister {
	return v1.NewRgpLister(f.Informer().GetIndexer())
}
