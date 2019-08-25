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
	"fmt"
	"io/ioutil"
	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/controllers_utils"
	flying_dutchman "github.com/blackducksoftware/synopsys-operator/meta-builder/flying-dutchman"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OpsSightReconciler reconciles a OpsSight object
type OpsSightReconciler struct {
	client.Client
	Log logr.Logger
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
	fmt.Printf("get OpsSight: %+v\n", opsSight)
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
	FinalYamlPath := "config/samples/opssight_runtime_objects.yaml"
	byteArrayContentFromFile, err := ioutil.ReadFile(FinalYamlPath)
	if err != nil {
		return nil, err
	}

	content := string(byteArrayContentFromFile)
	content = strings.ReplaceAll(content, "${NAME}", opsSight.Name)
	content = strings.ReplaceAll(content, "${NAMESPACE}", opsSight.Spec.Namespace)

	mapOfUniqueIDToDesiredRuntimeObject := controllers_utils.ConvertYamlFileToRuntimeObjects([]byte(content))
	// TODO: [yash commented this] figure out why this didn't work for black duck, probably just a namespace thing
	//for _, desiredRuntimeObject := range mapOfUniqueIdToDesiredRuntimeObject {
	//	// set an owner reference
	//	if err := ctrl.SetControllerReference(blackduck, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
	//		// requeue if we cannot set owner on the object
	//		// TODO: change this to requeue, and only not requeue when we get "newAlreadyOwnedError", i.e: if it's already owned by our CR
	//		//return ctrl.Result{}, err
	//		return mapOfUniqueIdToDesiredRuntimeObject, nil
	//	}
	//}
	fmt.Printf("Before - Num mapOfUniqueIdToDesiredRuntimeObject: %+v\n", len(mapOfUniqueIDToDesiredRuntimeObject))
	objs := patchOpsSight(r.Client, opsSight, mapOfUniqueIDToDesiredRuntimeObject)
	fmt.Printf("After - Num mapOfUniqueIdToDesiredRuntimeObject: %+v\n", len(objs))

	return objs, nil
}

// GetInstructionManual creates the instruction manual on the fly based on labels and annotations
func (r *OpsSightReconciler) GetInstructionManual(obj map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	// 2. Create Instruction Manual From Runtime Objects
	instructionManual, err := controllers_utils.CreateInstructionManual(obj)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

// +kubebuilder:rbac:groups=synopsys.com,resources=opssights,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=opssights/status,verbs=get;update;patch

// Reconcile reconcile the OpsSight custom resources
func (r *OpsSightReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("opsSight", req.NamespacedName)

	// your logic here

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
	r.setIndexingForChildrenObjects(mgr, &corev1.Secret{})

	// TODO add jobs and depployment?
	builder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.OpsSight{})
	builder = builder.Owns(&corev1.ConfigMap{})
	builder = builder.Owns(&corev1.Service{})
	builder = builder.Owns(&corev1.ReplicationController{})
	builder = builder.Owns(&corev1.Secret{})

	return builder.Complete(r)
}
