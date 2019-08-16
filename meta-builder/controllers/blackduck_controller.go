/*

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

package controllers

import (
	"context"
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/controllers_utils"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/flying-dutchman"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
)

// BlackduckReconciler reconciles a Black Duck object
type BlackduckReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *BlackduckReconciler) GetClient() client.Client {
	return r.Client
}

func (r *BlackduckReconciler) GetCustomResource(req ctrl.Request) (v1.Object, error) {
	var blackduck synopsysv1.Blackduck
	if err := r.Get(context.Background(), req.NamespacedName, &blackduck); err != nil {
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
	fmt.Printf("Get Blackduck: %+v\n", blackduck)
	return &blackduck, nil
}

func (r *BlackduckReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	blackduck := cr.(*synopsysv1.Blackduck)
	// 1. Get List of Runtime Objects (Base Yamls)
	// TODO: either read contents of yaml from locally mounted file
	// read content of full desired yaml from externally hosted file
	// FinalYamlUrl := "https://raw.githubusercontent.com/mphammer/customer-on-prem-alert-final-yaml/master/base-on-prem-alert-final.yaml"
	// byteArrayContentFromFile, err := controllers_utils.HttpGet(FinalYamlUrl)
	// if err != nil {
	// 	return nil, err
	// }
	FinalYamlPath := "config/samples/blackduck_runtime_objects.yaml"
	byteArrayContentFromFile, err := ioutil.ReadFile(FinalYamlPath)
	if err != nil {
		return nil, err
	}

	mapOfUniqueIdToDesiredRuntimeObject := controllers_utils.ConvertYamlFileToRuntimeObjects(byteArrayContentFromFile)
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
	fmt.Printf("Before - Num mapOfUniqueIdToDesiredRuntimeObject: %+v\n", len(mapOfUniqueIdToDesiredRuntimeObject))
	objs := patchBlackduck(blackduck, mapOfUniqueIdToDesiredRuntimeObject)
	fmt.Printf("After - Num mapOfUniqueIdToDesiredRuntimeObject: %+v\n", len(objs))

	return objs, nil
}

func (r *BlackduckReconciler) GetInstructionManual(obj map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	// 2. Create Instruction Manual From Runtime Objects
	instructionManual, err := controllers_utils.CreateInstructionManual(obj)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

// +kubebuilder:rbac:groups=synopsys.com,resources=blackducks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=blackducks/status,verbs=get;update;patch

func (r *BlackduckReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("blackduck", req.NamespacedName)

	// your logic here

	return flying_dutchman.MetaReconcile(req, r)
}

func (r *BlackduckReconciler) SetIndexingForChildrenObjects(mgr ctrl.Manager, ro runtime.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ro, flying_dutchman.JobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		owner := metav1.GetControllerOf(ro.(metav1.Object))
		if owner == nil {
			return nil
		}
		// ...make sure it's a Blackduck...
		if owner.APIVersion != synopsysv1.GroupVersion.String() || owner.Kind != "Blackduck" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return nil
}
func (r *BlackduckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Code here allows to kick off a reconciliation when objects our controller manages are changed somehow
	r.SetIndexingForChildrenObjects(mgr, &corev1.ConfigMap{})
	r.SetIndexingForChildrenObjects(mgr, &corev1.Service{})
	r.SetIndexingForChildrenObjects(mgr, &corev1.ReplicationController{})
	r.SetIndexingForChildrenObjects(mgr, &corev1.Secret{})

	// TODO add jobs and depployment?
	builder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.Blackduck{})
	builder = builder.Owns(&corev1.ConfigMap{})
	builder = builder.Owns(&corev1.Service{})
	builder = builder.Owns(&corev1.ReplicationController{})
	builder = builder.Owns(&corev1.Secret{})

	return builder.Complete(r)
}
