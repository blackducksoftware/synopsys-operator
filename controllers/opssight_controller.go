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

package controllers

import (
	"context"
	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/controllers/util"
	flying_dutchman "github.com/blackducksoftware/synopsys-operator/flying-dutchman"

	"github.com/go-logr/logr"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OpsSightReconciler reconciles a OpsSight object
type OpsSightReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Log         logr.Logger
	IsOpenShift bool
}

// GetClient returns the controller runtime client
func (r *OpsSightReconciler) GetClient() client.Client {
	return r.Client
}

// GetCustomResource returns the custom resource that need to be processed
func (r *OpsSightReconciler) GetCustomResource(req ctrl.Request) (metav1.Object, error) {
	opsSight := &synopsysv1.OpsSight{}
	if err := r.Get(context.Background(), req.NamespacedName, opsSight); err != nil {
		//log.Error(err, "unable to fetch Alert")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		// TODO:
		// We generally want to ignore (not requeue) NotFound errors, since we’ll get a reconciliation request once the object exists, and requeuing in the meantime won’t help.
		if !apierrs.IsNotFound(err) {
			return nil, err
		}
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
	}
	return opsSight, nil
}

// GetRuntimeObjects returns the runtime objects of OpsSight
func (r *OpsSightReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	opsSight := cr.(*synopsysv1.OpsSight)
	// 1. Get List of Runtime Objects (Base Yamls)
	// TODO: either read contents of yaml from locally mounted file
	// read content of full desired yaml from externally hosted file
	// FinalYamlUrl := "https://raw.githubusercontent.com/mphammer/customer-on-prem-alert-final-yaml/master/base-on-prem-alert-final.yaml"
	// byteArrayContentFromFile, err := controllers_utils.HttpGet(FinalYamlUrl)
	// if err != nil {
	// 	return nil, err
	// }
	// get the base yaml for the app
	latestBaseYamlAsString, err := controllers_utils.GetBaseYaml(controllers_utils.OPSSIGHT, opsSight.Spec.Version, "")
	if err != nil {
		return nil, err
	}

	latestBaseYamlAsString = strings.ReplaceAll(latestBaseYamlAsString, "${NAME}", opsSight.Name)
	latestBaseYamlAsString = strings.ReplaceAll(latestBaseYamlAsString, "${NAMESPACE}", opsSight.Spec.Namespace)

	mapOfUniqueIDToDesiredRuntimeObject := controllers_utils.ConvertYamlFileToRuntimeObjects(latestBaseYamlAsString, r.IsOpenShift)
	for _, desiredRuntimeObject := range mapOfUniqueIDToDesiredRuntimeObject {
		// set an owner reference
		if err := ctrl.SetControllerReference(opsSight, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
			// requeue if we cannot set owner on the object
			// TODO: change this to requeue, and only not requeue when we get "newAlreadyOwnedError", i.e: if it's already owned by our CR
			//return ctrl.Result{}, err
			return mapOfUniqueIDToDesiredRuntimeObject, nil
		}
	}
	objs := patchOpsSight(r.Client, r.Scheme, opsSight, mapOfUniqueIDToDesiredRuntimeObject, r.Log, r.IsOpenShift)

	return objs, nil
}

// GetInstructionManual creates the instruction manual on the fly based on labels and annotations
func (r *OpsSightReconciler) GetInstructionManual(mapOfUniqueIDToDesiredRuntimeObject map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	// 2. Create Instruction Manual From Runtime Objects
	instructionManual, err := controllers_utils.CreateInstructionManual(mapOfUniqueIDToDesiredRuntimeObject)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

// +kubebuilder:rbac:groups=synopsys.com,resources=opssights,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=opssights/status,verbs=get;update;patch

// Reconcile reconcile the OpsSight custom resources
func (r *OpsSightReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return flying_dutchman.MetaReconcile(req, r)
}

// setIndexingForChildrenObjects set the indexing for child objects
func (r *OpsSightReconciler) setIndexingForChildrenObjects(mgr ctrl.Manager, ro runtime.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ro, flying_dutchman.JobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		owner := metav1.GetControllerOf(ro.(metav1.Object))
		if owner == nil {
			return nil
		}
		// ...make sure it's a OpsSight...
		if owner.APIVersion != synopsysv1.GroupVersion.String() || owner.Kind != "OpsSight" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return nil
}

// SetupWithManager setup the Controller runtime manager
func (r *OpsSightReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Code here allows to kick off a reconciliation when objects our controller manages are changed somehow
	r.setIndexingForChildrenObjects(mgr, &corev1.ConfigMap{})
	r.setIndexingForChildrenObjects(mgr, &corev1.Service{})
	r.setIndexingForChildrenObjects(mgr, &corev1.ReplicationController{})
	r.setIndexingForChildrenObjects(mgr, &appsv1.Deployment{})
	r.setIndexingForChildrenObjects(mgr, &corev1.Secret{})

	builder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.OpsSight{})
	builder = builder.Owns(&corev1.ConfigMap{})
	builder = builder.Owns(&corev1.Service{})
	builder = builder.Owns(&corev1.ReplicationController{})
	builder = builder.Owns(&appsv1.Deployment{})
	builder = builder.Owns(&corev1.Secret{})
	builder = builder.Owns(&corev1.ServiceAccount{})
	builder = builder.Owns(&rbacv1.ClusterRole{})
	builder = builder.Owns(&rbacv1.ClusterRoleBinding{})
	if r.IsOpenShift {
		builder = builder.Owns(&routev1.Route{})
	}

	return builder.Complete(r)
}
