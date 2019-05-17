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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	rgpv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRgps implements RgpInterface
type FakeRgps struct {
	Fake *FakeSynopsysV1
	ns   string
}

var rgpsResource = schema.GroupVersionResource{Group: "synopsys", Version: "v1", Resource: "rgps"}

var rgpsKind = schema.GroupVersionKind{Group: "synopsys", Version: "v1", Kind: "Rgp"}

// Get takes name of the rgp, and returns the corresponding rgp object, and an error if there is any.
func (c *FakeRgps) Get(name string, options v1.GetOptions) (result *rgpv1.Rgp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(rgpsResource, c.ns, name), &rgpv1.Rgp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*rgpv1.Rgp), err
}

// List takes label and field selectors, and returns the list of Rgps that match those selectors.
func (c *FakeRgps) List(opts v1.ListOptions) (result *rgpv1.RgpList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(rgpsResource, rgpsKind, c.ns, opts), &rgpv1.RgpList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &rgpv1.RgpList{ListMeta: obj.(*rgpv1.RgpList).ListMeta}
	for _, item := range obj.(*rgpv1.RgpList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested rgps.
func (c *FakeRgps) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(rgpsResource, c.ns, opts))

}

// Create takes the representation of a rgp and creates it.  Returns the server's representation of the rgp, and an error, if there is any.
func (c *FakeRgps) Create(rgp *rgpv1.Rgp) (result *rgpv1.Rgp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(rgpsResource, c.ns, rgp), &rgpv1.Rgp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*rgpv1.Rgp), err
}

// Update takes the representation of a rgp and updates it. Returns the server's representation of the rgp, and an error, if there is any.
func (c *FakeRgps) Update(rgp *rgpv1.Rgp) (result *rgpv1.Rgp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(rgpsResource, c.ns, rgp), &rgpv1.Rgp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*rgpv1.Rgp), err
}

// Delete takes name of the rgp and deletes it. Returns an error if one occurs.
func (c *FakeRgps) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(rgpsResource, c.ns, name), &rgpv1.Rgp{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRgps) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(rgpsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &rgpv1.RgpList{})
	return err
}

// Patch applies the patch and returns the patched rgp.
func (c *FakeRgps) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *rgpv1.Rgp, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(rgpsResource, c.ns, name, pt, data, subresources...), &rgpv1.Rgp{})

	if obj == nil {
		return nil, err
	}
	return obj.(*rgpv1.Rgp), err
}
