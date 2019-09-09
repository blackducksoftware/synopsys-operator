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
	"strconv"
	"strings"
	"time"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/controllers/util"
	flying_dutchman "github.com/blackducksoftware/synopsys-operator/flying-dutchman"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PolarisDBReconciler reconciles a PolarisDB object
type PolarisDBReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Log         logr.Logger
	IsOpenShift bool
	IsDryRun    bool
}

func (r *PolarisDBReconciler) GetClient() client.Client {
	return r.Client
}

func (r *PolarisDBReconciler) GetCustomResource(req ctrl.Request) (metav1.Object, error) {
	var polarisDb synopsysv1.PolarisDB
	if err := r.Get(context.Background(), req.NamespacedName, &polarisDb); err != nil {
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
	return &polarisDb, nil
}

func (r *PolarisDBReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	polarisDbCr := cr.(*synopsysv1.PolarisDB)
	// TODO: either read contents of yaml from locally mounted file
	// read content of full desired yaml from externally hosted file
	// FinalYamlUrl := "https://raw.githubusercontent.com/mphammer/customer-on-prem-alert-final-yaml/master/base-on-prem-alert-final.yaml"
	// byteArrayContentFromFile, err := controllers_utils.HttpGet(FinalYamlUrl)
	// if err != nil {
	// 	return nil, err
	// }
	content, err := controllers_utils.GetBaseYaml(controllers_utils.POLARISDB, polarisDbCr.Spec.Version, "polarisdb")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", polarisDbCr.Spec.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polarisDbCr.Spec.EnvironmentName)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polarisDbCr.Spec.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${POSTGRES_USERNAME}", polarisDbCr.Spec.PostgresDetails.Username)
	content = strings.ReplaceAll(content, "${POSTGRES_PASSWORD}", polarisDbCr.Spec.PostgresDetails.Password)
	content = strings.ReplaceAll(content, "${SMTP_HOST}", polarisDbCr.Spec.SMTPDetails.Host)
	if polarisDbCr.Spec.SMTPDetails.Port != 2525 {
		content = strings.ReplaceAll(content, "2525", strconv.Itoa(polarisDbCr.Spec.SMTPDetails.Port))
	}
	if len(polarisDbCr.Spec.SMTPDetails.Username) != 0 {
		content = strings.ReplaceAll(content, "${SMTP_USERNAME}", controllers_utils.EncodeStringToBase64(polarisDbCr.Spec.SMTPDetails.Username))
	} else {
		content = strings.ReplaceAll(content, "${SMTP_USERNAME}", "Cg==")
	}
	if len(polarisDbCr.Spec.SMTPDetails.Password) != 0 {
		content = strings.ReplaceAll(content, "${SMTP_PASSWORD}", fmt.Sprintf("\"%s\"", controllers_utils.EncodeStringToBase64(polarisDbCr.Spec.SMTPDetails.Password)))
	} else {
		content = strings.ReplaceAll(content, "${SMTP_PASSWORD}", "Cg==")
	}
	content = strings.ReplaceAll(content, "${POSTGRES_HOST}", polarisDbCr.Spec.PostgresDetails.Host)
	if polarisDbCr.Spec.PostgresDetails.Port != 5432 {
		content = strings.ReplaceAll(content, "5432", strconv.Itoa(polarisDbCr.Spec.PostgresDetails.Port))
	}
	if polarisDbCr.Spec.PostgresInstanceType == "internal" {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "internal")
	} else {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "external")
	}

	mapOfUniqueIdToBaseRuntimeObject := controllers_utils.ConvertYamlFileToRuntimeObjects(content, r.IsOpenShift)
	mapOfUniqueIdToBaseRuntimeObject = removeTestManifests(mapOfUniqueIdToBaseRuntimeObject)

	if !r.IsDryRun {
		for _, desiredRuntimeObject := range mapOfUniqueIdToBaseRuntimeObject {
			// set an owner reference
			if err := ctrl.SetControllerReference(polarisDbCr, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
				// requeue if we cannot set owner on the object
				// TODO: change this to requeue, and only not requeue when we get "newAlreadyOwnedError", i.e: if it's already owned by our CR
				//return ctrl.Result{}, err
				return mapOfUniqueIdToBaseRuntimeObject, nil
			}
		}
	}
	mapOfUniqueIdToDesiredRuntimeObject := patchPolarisDB(r.Client, polarisDbCr, mapOfUniqueIdToBaseRuntimeObject)

	return mapOfUniqueIdToDesiredRuntimeObject, nil
}

func removeTestManifests(objects map[string]runtime.Object) map[string]runtime.Object {
	objectsToBeRemoved := []string{
		"Pod.swip-db-ui-test-3gfbs",
		"Pod.swip-db-vault-status-test",
		"ConfigMap.polaris-db-consul-tests",
	}
	for _, object := range objectsToBeRemoved {
		delete(objects, object)
	}
	return objects
}

func (r *PolarisDBReconciler) GetInstructionManual(mapOfUniqueIdToDesiredRuntimeObject map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	instructionManualLocation := "config/samples/dependency_manual_polarisdb.yaml"
	instructionManual, err := controllers_utils.CreateInstructionManualFromYaml(instructionManualLocation)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

// +kubebuilder:rbac:groups=synopsys.com,resources=polarisdb,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=polarisdb/status,verbs=get;update;patch

func (r *PolarisDBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// _ = context.Background()
	// _ = r.Log.WithValues("polarisdb", req.NamespacedName)
	// your logic here

	res, err := flying_dutchman.MetaReconcile(req, r)
	if strings.Contains(fmt.Sprint(err), "is not ready") {
		res = ctrl.Result{RequeueAfter: 10 * time.Second}
	}
	return res, err
}

func (r *PolarisDBReconciler) SetIndexingForChildrenObjects(mgr ctrl.Manager, ro runtime.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ro, flying_dutchman.JobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		owner := metav1.GetControllerOf(ro.(metav1.Object))
		if owner == nil {
			return nil
		}
		// ...make sure it's a Alert...
		if owner.APIVersion != synopsysv1.GroupVersion.String() || owner.Kind != "PolarisDB" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return nil
}

func (r *PolarisDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.SetIndexingForChildrenObjects(mgr, &corev1.Service{})
	r.SetIndexingForChildrenObjects(mgr, &appsv1.Deployment{})
	// Add HorizontalPodAutoscaler to the list

	polarisDbBuilder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.PolarisDB{})
	polarisDbBuilder = polarisDbBuilder.Owns(&corev1.ConfigMap{})
	polarisDbBuilder = polarisDbBuilder.Owns(&corev1.Service{})
	polarisDbBuilder = polarisDbBuilder.Owns(&corev1.Secret{})

	return polarisDbBuilder.Complete(r)
}
