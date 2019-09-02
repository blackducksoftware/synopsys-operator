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

	"k8s.io/apimachinery/pkg/api/meta"

	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/util"
	flying_dutchman "github.com/blackducksoftware/synopsys-operator/meta-builder/flying-dutchman"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReportingReconciler reconciles a Reporting object
type ReportingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *ReportingReconciler) GetClient() client.Client {
	return r.Client
}

func (r *ReportingReconciler) GetCustomResource(req ctrl.Request) (metav1.Object, error) {
	var reporting synopsysv1.Reporting
	if err := r.Get(context.Background(), req.NamespacedName, &reporting); err != nil {
		if !apierrs.IsNotFound(err) {
			return nil, err
		}
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
	}
	return &reporting, nil
}

func (r *ReportingReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	reportingCr := cr.(*synopsysv1.Reporting)
	content, err := controllers_utils.GetBaseYaml(controllers_utils.REPORTING, reportingCr.Spec.Version, "")
	if err != nil {
		return nil, err
	}

	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", reportingCr.Spec.EnvironmentName)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", reportingCr.Spec.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${POSTGRES_HOST}", reportingCr.Spec.PostgresDetails.Hostname)
	content = strings.ReplaceAll(content, "${POSTGRES_PORT}", fmt.Sprint(reportingCr.Spec.PostgresDetails.Port))
	content = strings.ReplaceAll(content, "${POSTGRES_USERNAME}", controllers_utils.EncodeStringToBase64(reportingCr.Spec.PostgresDetails.Username))
	content = strings.ReplaceAll(content, "${POSTGRES_PASSWORD}", controllers_utils.EncodeStringToBase64(reportingCr.Spec.PostgresDetails.Password))
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", reportingCr.Spec.ImagePullSecrets)

	fmt.Println("---------------------------------")
	fmt.Println(controllers_utils.EncodeStringToBase64(reportingCr.Spec.PostgresDetails.Username))
	fmt.Println(controllers_utils.EncodeStringToBase64(reportingCr.Spec.PostgresDetails.Password))
	fmt.Println("---------------------------------")

	mapOfUniqueIdToBaseRuntimeObject := controllers_utils.ConvertYamlFileToRuntimeObjects(content)
	for _, desiredRuntimeObject := range mapOfUniqueIdToBaseRuntimeObject {
		if err := ctrl.SetControllerReference(reportingCr, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
			return mapOfUniqueIdToBaseRuntimeObject, nil
		}
	}
	mapOfUniqueIdToDesiredRuntimeObject := patchReporting(reportingCr, mapOfUniqueIdToBaseRuntimeObject, meta.NewAccessor())

	return mapOfUniqueIdToDesiredRuntimeObject, nil
}

func (r *ReportingReconciler) GetInstructionManual(mapOfUniqueIdToDesiredRuntimeObject map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	instructionManualLocation := "config/samples/dependency_manual_reporting.yaml"
	instructionManual, err := controllers_utils.CreateInstructionManualFromYaml(instructionManualLocation, mapOfUniqueIdToDesiredRuntimeObject)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

func (r *ReportingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return flying_dutchman.MetaReconcile(req, r)
}

func (r *ReportingReconciler) SetIndexingForChildrenObjects(mgr ctrl.Manager, ro runtime.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ro, flying_dutchman.JobOwnerKey, func(rawObj runtime.Object) []string {
		owner := metav1.GetControllerOf(ro.(metav1.Object))
		if owner == nil {
			return nil
		}
		if owner.APIVersion != synopsysv1.GroupVersion.String() || owner.Kind != "Reporting" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return nil
}

func (r *ReportingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.SetIndexingForChildrenObjects(mgr, &corev1.Service{})
	r.SetIndexingForChildrenObjects(mgr, &appsv1.Deployment{})
	r.SetIndexingForChildrenObjects(mgr, &corev1.ServiceAccount{})
	// Add HorizontalPodAutoscaler to the list

	reportingBuilder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.Reporting{})
	reportingBuilder = reportingBuilder.Owns(&corev1.ConfigMap{})
	reportingBuilder = reportingBuilder.Owns(&corev1.Service{})
	reportingBuilder = reportingBuilder.Owns(&corev1.ReplicationController{})
	reportingBuilder = reportingBuilder.Owns(&corev1.Secret{})

	return reportingBuilder.Complete(r)
}
