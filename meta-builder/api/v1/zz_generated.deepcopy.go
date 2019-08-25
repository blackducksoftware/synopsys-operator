// +build !ignore_autogenerated

/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

// autogenerated by controller-gen object, do not modify manually

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Alert) DeepCopyInto(out *Alert) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Alert.
func (in *Alert) DeepCopy() *Alert {
	if in == nil {
		return nil
	}
	out := new(Alert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Alert) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertList) DeepCopyInto(out *AlertList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Alert, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertList.
func (in *AlertList) DeepCopy() *AlertList {
	if in == nil {
		return nil
	}
	out := new(AlertList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AlertList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertSpec) DeepCopyInto(out *AlertSpec) {
	*out = *in
	if in.StandAlone != nil {
		in, out := &in.StandAlone, &out.StandAlone
		*out = new(bool)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]*string, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(string)
				**out = **in
			}
		}
	}
	if in.Environs != nil {
		in, out := &in.Environs, &out.Environs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ImageRegistries != nil {
		in, out := &in.ImageRegistries, &out.ImageRegistries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RegistryConfiguration != nil {
		in, out := &in.RegistryConfiguration, &out.RegistryConfiguration
		*out = new(RegistryConfiguration)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertSpec.
func (in *AlertSpec) DeepCopy() *AlertSpec {
	if in == nil {
		return nil
	}
	out := new(AlertSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertStatus) DeepCopyInto(out *AlertStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertStatus.
func (in *AlertStatus) DeepCopy() *AlertStatus {
	if in == nil {
		return nil
	}
	out := new(AlertStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthServerSpec) DeepCopyInto(out *AuthServerSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	out.ResourcesSpec = in.ResourcesSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthServerSpec.
func (in *AuthServerSpec) DeepCopy() *AuthServerSpec {
	if in == nil {
		return nil
	}
	out := new(AuthServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Blackduck) DeepCopyInto(out *Blackduck) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Blackduck.
func (in *Blackduck) DeepCopy() *Blackduck {
	if in == nil {
		return nil
	}
	out := new(Blackduck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Blackduck) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BlackduckList) DeepCopyInto(out *BlackduckList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Blackduck, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BlackduckList.
func (in *BlackduckList) DeepCopy() *BlackduckList {
	if in == nil {
		return nil
	}
	out := new(BlackduckList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BlackduckList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BlackduckSpec) DeepCopyInto(out *BlackduckSpec) {
	*out = *in
	if in.ExternalPostgres != nil {
		in, out := &in.ExternalPostgres, &out.ExternalPostgres
		*out = new(PostgresExternalDBConfig)
		**out = **in
	}
	if in.PVC != nil {
		in, out := &in.PVC, &out.PVC
		*out = make([]PVC, len(*in))
		copy(*out, *in)
	}
	if in.Environs != nil {
		in, out := &in.Environs, &out.Environs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ImageRegistries != nil {
		in, out := &in.ImageRegistries, &out.ImageRegistries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.RegistryConfiguration.DeepCopyInto(&out.RegistryConfiguration)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BlackduckSpec.
func (in *BlackduckSpec) DeepCopy() *BlackduckSpec {
	if in == nil {
		return nil
	}
	out := new(BlackduckSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BlackduckStatus) DeepCopyInto(out *BlackduckStatus) {
	*out = *in
	if in.PVCVolumeName != nil {
		in, out := &in.PVCVolumeName, &out.PVCVolumeName
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BlackduckStatus.
func (in *BlackduckStatus) DeepCopy() *BlackduckStatus {
	if in == nil {
		return nil
	}
	out := new(BlackduckStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Blackducks) DeepCopyInto(out *Blackducks) {
	*out = *in
	if in.ExternalHosts != nil {
		in, out := &in.ExternalHosts, &out.ExternalHosts
		*out = make([]*Host, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Host)
				**out = **in
			}
		}
	}
	if in.BlackduckSpec != nil {
		in, out := &in.BlackduckSpec, &out.BlackduckSpec
		*out = new(BlackduckSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Blackducks.
func (in *Blackducks) DeepCopy() *Blackducks {
	if in == nil {
		return nil
	}
	out := new(Blackducks)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerSize) DeepCopyInto(out *ContainerSize) {
	*out = *in
	if in.MinCPU != nil {
		in, out := &in.MinCPU, &out.MinCPU
		*out = new(int32)
		**out = **in
	}
	if in.MaxCPU != nil {
		in, out := &in.MaxCPU, &out.MaxCPU
		*out = new(int32)
		**out = **in
	}
	if in.MinMem != nil {
		in, out := &in.MinMem, &out.MinMem
		*out = new(int32)
		**out = **in
	}
	if in.MaxMem != nil {
		in, out := &in.MaxMem, &out.MaxMem
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerSize.
func (in *ContainerSize) DeepCopy() *ContainerSize {
	if in == nil {
		return nil
	}
	out := new(ContainerSize)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Environs) DeepCopyInto(out *Environs) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Environs.
func (in *Environs) DeepCopy() *Environs {
	if in == nil {
		return nil
	}
	out := new(Environs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventstoreDetails) DeepCopyInto(out *EventstoreDetails) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventstoreDetails.
func (in *EventstoreDetails) DeepCopy() *EventstoreDetails {
	if in == nil {
		return nil
	}
	out := new(EventstoreDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Host) DeepCopyInto(out *Host) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Host.
func (in *Host) DeepCopy() *Host {
	if in == nil {
		return nil
	}
	out := new(Host)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageFacade) DeepCopyInto(out *ImageFacade) {
	*out = *in
	if in.InternalRegistries != nil {
		in, out := &in.InternalRegistries, &out.InternalRegistries
		*out = make([]*RegistryAuth, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(RegistryAuth)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageFacade.
func (in *ImageFacade) DeepCopy() *ImageFacade {
	if in == nil {
		return nil
	}
	out := new(ImageFacade)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LimitsSpec) DeepCopyInto(out *LimitsSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LimitsSpec.
func (in *LimitsSpec) DeepCopy() *LimitsSpec {
	if in == nil {
		return nil
	}
	out := new(LimitsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeAffinity) DeepCopyInto(out *NodeAffinity) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeAffinity.
func (in *NodeAffinity) DeepCopy() *NodeAffinity {
	if in == nil {
		return nil
	}
	out := new(NodeAffinity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpsSight) DeepCopyInto(out *OpsSight) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpsSight.
func (in *OpsSight) DeepCopy() *OpsSight {
	if in == nil {
		return nil
	}
	out := new(OpsSight)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpsSight) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpsSightList) DeepCopyInto(out *OpsSightList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpsSight, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpsSightList.
func (in *OpsSightList) DeepCopy() *OpsSightList {
	if in == nil {
		return nil
	}
	out := new(OpsSightList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpsSightList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpsSightSpec) DeepCopyInto(out *OpsSightSpec) {
	*out = *in
	if in.Perceptor != nil {
		in, out := &in.Perceptor, &out.Perceptor
		*out = new(Perceptor)
		**out = **in
	}
	if in.ScannerPod != nil {
		in, out := &in.ScannerPod, &out.ScannerPod
		*out = new(ScannerPod)
		(*in).DeepCopyInto(*out)
	}
	if in.Perceiver != nil {
		in, out := &in.Perceiver, &out.Perceiver
		*out = new(Perceiver)
		(*in).DeepCopyInto(*out)
	}
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(Prometheus)
		**out = **in
	}
	if in.Blackduck != nil {
		in, out := &in.Blackduck, &out.Blackduck
		*out = new(Blackducks)
		(*in).DeepCopyInto(*out)
	}
	if in.ImageRegistries != nil {
		in, out := &in.ImageRegistries, &out.ImageRegistries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.RegistryConfiguration.DeepCopyInto(&out.RegistryConfiguration)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpsSightSpec.
func (in *OpsSightSpec) DeepCopy() *OpsSightSpec {
	if in == nil {
		return nil
	}
	out := new(OpsSightSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpsSightStatus) DeepCopyInto(out *OpsSightStatus) {
	*out = *in
	if in.InternalHosts != nil {
		in, out := &in.InternalHosts, &out.InternalHosts
		*out = make([]*Host, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Host)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpsSightStatus.
func (in *OpsSightStatus) DeepCopy() *OpsSightStatus {
	if in == nil {
		return nil
	}
	out := new(OpsSightStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PVC) DeepCopyInto(out *PVC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PVC.
func (in *PVC) DeepCopy() *PVC {
	if in == nil {
		return nil
	}
	out := new(PVC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Perceiver) DeepCopyInto(out *Perceiver) {
	*out = *in
	if in.PodPerceiver != nil {
		in, out := &in.PodPerceiver, &out.PodPerceiver
		*out = new(PodPerceiver)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Perceiver.
func (in *Perceiver) DeepCopy() *Perceiver {
	if in == nil {
		return nil
	}
	out := new(Perceiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Perceptor) DeepCopyInto(out *Perceptor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Perceptor.
func (in *Perceptor) DeepCopy() *Perceptor {
	if in == nil {
		return nil
	}
	out := new(Perceptor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodPerceiver) DeepCopyInto(out *PodPerceiver) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodPerceiver.
func (in *PodPerceiver) DeepCopy() *PodPerceiver {
	if in == nil {
		return nil
	}
	out := new(PodPerceiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodResource) DeepCopyInto(out *PodResource) {
	*out = *in
	if in.ContainerLimit != nil {
		in, out := &in.ContainerLimit, &out.ContainerLimit
		*out = make(map[string]ContainerSize, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodResource.
func (in *PodResource) DeepCopy() *PodResource {
	if in == nil {
		return nil
	}
	out := new(PodResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Polaris) DeepCopyInto(out *Polaris) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Polaris.
func (in *Polaris) DeepCopy() *Polaris {
	if in == nil {
		return nil
	}
	out := new(Polaris)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Polaris) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisDB) DeepCopyInto(out *PolarisDB) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisDB.
func (in *PolarisDB) DeepCopy() *PolarisDB {
	if in == nil {
		return nil
	}
	out := new(PolarisDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisDB) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisDBList) DeepCopyInto(out *PolarisDBList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolarisDB, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisDBList.
func (in *PolarisDBList) DeepCopy() *PolarisDBList {
	if in == nil {
		return nil
	}
	out := new(PolarisDBList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisDBList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisDBSpec) DeepCopyInto(out *PolarisDBSpec) {
	*out = *in
	in.SMTPDetails.DeepCopyInto(&out.SMTPDetails)
	in.PostgresStorageDetails.DeepCopyInto(&out.PostgresStorageDetails)
	in.PostgresDetails.DeepCopyInto(&out.PostgresDetails)
	in.EventstoreDetails.DeepCopyInto(&out.EventstoreDetails)
	in.UploadServerDetails.DeepCopyInto(&out.UploadServerDetails)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisDBSpec.
func (in *PolarisDBSpec) DeepCopy() *PolarisDBSpec {
	if in == nil {
		return nil
	}
	out := new(PolarisDBSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisDBStatus) DeepCopyInto(out *PolarisDBStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisDBStatus.
func (in *PolarisDBStatus) DeepCopy() *PolarisDBStatus {
	if in == nil {
		return nil
	}
	out := new(PolarisDBStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisList) DeepCopyInto(out *PolarisList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Polaris, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisList.
func (in *PolarisList) DeepCopy() *PolarisList {
	if in == nil {
		return nil
	}
	out := new(PolarisList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisSpec) DeepCopyInto(out *PolarisSpec) {
	*out = *in
	in.AuthServerSpec.DeepCopyInto(&out.AuthServerSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisSpec.
func (in *PolarisSpec) DeepCopy() *PolarisSpec {
	if in == nil {
		return nil
	}
	out := new(PolarisSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisStatus) DeepCopyInto(out *PolarisStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisStatus.
func (in *PolarisStatus) DeepCopy() *PolarisStatus {
	if in == nil {
		return nil
	}
	out := new(PolarisStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresDetails) DeepCopyInto(out *PostgresDetails) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresDetails.
func (in *PostgresDetails) DeepCopy() *PostgresDetails {
	if in == nil {
		return nil
	}
	out := new(PostgresDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresExternalDBConfig) DeepCopyInto(out *PostgresExternalDBConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresExternalDBConfig.
func (in *PostgresExternalDBConfig) DeepCopy() *PostgresExternalDBConfig {
	if in == nil {
		return nil
	}
	out := new(PostgresExternalDBConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresStorageDetails) DeepCopyInto(out *PostgresStorageDetails) {
	*out = *in
	if in.StorageClass != nil {
		in, out := &in.StorageClass, &out.StorageClass
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresStorageDetails.
func (in *PostgresStorageDetails) DeepCopy() *PostgresStorageDetails {
	if in == nil {
		return nil
	}
	out := new(PostgresStorageDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Prometheus) DeepCopyInto(out *Prometheus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Prometheus.
func (in *Prometheus) DeepCopy() *Prometheus {
	if in == nil {
		return nil
	}
	out := new(Prometheus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryAuth) DeepCopyInto(out *RegistryAuth) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryAuth.
func (in *RegistryAuth) DeepCopy() *RegistryAuth {
	if in == nil {
		return nil
	}
	out := new(RegistryAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryConfiguration) DeepCopyInto(out *RegistryConfiguration) {
	*out = *in
	if in.PullSecrets != nil {
		in, out := &in.PullSecrets, &out.PullSecrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryConfiguration.
func (in *RegistryConfiguration) DeepCopy() *RegistryConfiguration {
	if in == nil {
		return nil
	}
	out := new(RegistryConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportStorageSpec) DeepCopyInto(out *ReportStorageSpec) {
	*out = *in
	out.Volume = in.Volume
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportStorageSpec.
func (in *ReportStorageSpec) DeepCopy() *ReportStorageSpec {
	if in == nil {
		return nil
	}
	out := new(ReportStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Reporting) DeepCopyInto(out *Reporting) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Reporting.
func (in *Reporting) DeepCopy() *Reporting {
	if in == nil {
		return nil
	}
	out := new(Reporting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Reporting) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportingList) DeepCopyInto(out *ReportingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Reporting, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportingList.
func (in *ReportingList) DeepCopy() *ReportingList {
	if in == nil {
		return nil
	}
	out := new(ReportingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReportingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportingPostgresDetails) DeepCopyInto(out *ReportingPostgresDetails) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportingPostgresDetails.
func (in *ReportingPostgresDetails) DeepCopy() *ReportingPostgresDetails {
	if in == nil {
		return nil
	}
	out := new(ReportingPostgresDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportingSpec) DeepCopyInto(out *ReportingSpec) {
	*out = *in
	in.PostgresDetails.DeepCopyInto(&out.PostgresDetails)
	out.ReportServiceSpec = in.ReportServiceSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportingSpec.
func (in *ReportingSpec) DeepCopy() *ReportingSpec {
	if in == nil {
		return nil
	}
	out := new(ReportingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReportingStatus) DeepCopyInto(out *ReportingStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReportingStatus.
func (in *ReportingStatus) DeepCopy() *ReportingStatus {
	if in == nil {
		return nil
	}
	out := new(ReportingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RequestsSpec) DeepCopyInto(out *RequestsSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RequestsSpec.
func (in *RequestsSpec) DeepCopy() *RequestsSpec {
	if in == nil {
		return nil
	}
	out := new(RequestsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcesSpec) DeepCopyInto(out *ResourcesSpec) {
	*out = *in
	out.RequestsSpec = in.RequestsSpec
	out.LimitsSpec = in.LimitsSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcesSpec.
func (in *ResourcesSpec) DeepCopy() *ResourcesSpec {
	if in == nil {
		return nil
	}
	out := new(ResourcesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SMTPDetails) DeepCopyInto(out *SMTPDetails) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SMTPDetails.
func (in *SMTPDetails) DeepCopy() *SMTPDetails {
	if in == nil {
		return nil
	}
	out := new(SMTPDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scanner) DeepCopyInto(out *Scanner) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scanner.
func (in *Scanner) DeepCopy() *Scanner {
	if in == nil {
		return nil
	}
	out := new(Scanner)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScannerPod) DeepCopyInto(out *ScannerPod) {
	*out = *in
	if in.Scanner != nil {
		in, out := &in.Scanner, &out.Scanner
		*out = new(Scanner)
		**out = **in
	}
	if in.ImageFacade != nil {
		in, out := &in.ImageFacade, &out.ImageFacade
		*out = new(ImageFacade)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScannerPod.
func (in *ScannerPod) DeepCopy() *ScannerPod {
	if in == nil {
		return nil
	}
	out := new(ScannerPod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Size) DeepCopyInto(out *Size) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Size.
func (in *Size) DeepCopy() *Size {
	if in == nil {
		return nil
	}
	out := new(Size)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Size) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SizeList) DeepCopyInto(out *SizeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Size, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SizeList.
func (in *SizeList) DeepCopy() *SizeList {
	if in == nil {
		return nil
	}
	out := new(SizeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SizeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SizeSpec) DeepCopyInto(out *SizeSpec) {
	*out = *in
	if in.PodResources != nil {
		in, out := &in.PodResources, &out.PodResources
		*out = make(map[string]PodResource, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SizeSpec.
func (in *SizeSpec) DeepCopy() *SizeSpec {
	if in == nil {
		return nil
	}
	out := new(SizeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SizeStatus) DeepCopyInto(out *SizeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SizeStatus.
func (in *SizeStatus) DeepCopy() *SizeStatus {
	if in == nil {
		return nil
	}
	out := new(SizeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Storage) DeepCopyInto(out *Storage) {
	*out = *in
	if in.StorageClass != nil {
		in, out := &in.StorageClass, &out.StorageClass
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Storage.
func (in *Storage) DeepCopy() *Storage {
	if in == nil {
		return nil
	}
	out := new(Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UploadServerDetails) DeepCopyInto(out *UploadServerDetails) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	out.ResourcesSpec = in.ResourcesSpec
	in.Storage.DeepCopyInto(&out.Storage)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UploadServerDetails.
func (in *UploadServerDetails) DeepCopy() *UploadServerDetails {
	if in == nil {
		return nil
	}
	out := new(UploadServerDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSpec) DeepCopyInto(out *VolumeSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSpec.
func (in *VolumeSpec) DeepCopy() *VolumeSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeSpec)
	in.DeepCopyInto(out)
	return out
}
