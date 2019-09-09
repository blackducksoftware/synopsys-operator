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

// AuthServerReconciler reconciles a AuthServer object
type AuthServerReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Log         logr.Logger
	IsOpenShift bool
	IsDryRun    bool
}

func (r *AuthServerReconciler) GetClient() client.Client {
	return r.Client
}

func (r *AuthServerReconciler) GetCustomResource(req ctrl.Request) (metav1.Object, error) {
	var authServer synopsysv1.AuthServer
	if err := r.Get(context.Background(), req.NamespacedName, &authServer); err != nil {
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
	return &authServer, nil
}

func (r *AuthServerReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	authServerCr := cr.(*synopsysv1.AuthServer)
	// TODO: either read contents of yaml from locally mounted file
	// read content of full desired yaml from externally hosted file
	// FinalYamlUrl := "https://raw.githubusercontent.com/mphammer/customer-on-prem-alert-final-yaml/master/base-on-prem-alert-final.yaml"
	// byteArrayContentFromFile, err := controllers_utils.HttpGet(FinalYamlUrl)
	// if err != nil {
	// 	return nil, err
	// }
	content, err := controllers_utils.GetBaseYaml(controllers_utils.AUTHSERVICE, authServerCr.Spec.Version, "polaris")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", authServerCr.Spec.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", authServerCr.Spec.EnvironmentName)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", authServerCr.Spec.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", authServerCr.Spec.ImagePullSecrets)

	mapOfUniqueIdToBaseRuntimeObject := controllers_utils.GetAuthServerRuntimeObjects(controllers_utils.ConvertYamlFileToRuntimeObjects(content, r.IsOpenShift))

	// filter auth-server runtimeobjects

	if !r.IsDryRun {
		for _, desiredRuntimeObject := range mapOfUniqueIdToBaseRuntimeObject {
			// set an owner reference
			if err := ctrl.SetControllerReference(authServerCr, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
				// requeue if we cannot set owner on the object
				// TODO: change this to requeue, and only not requeue when we get "newAlreadyOwnedError", i.e: if it's already owned by our CR
				//return ctrl.Result{}, err
				return mapOfUniqueIdToBaseRuntimeObject, nil
			}
		}
	}
	mapOfUniqueIdToDesiredRuntimeObject := patchAuthServer(authServerCr, mapOfUniqueIdToBaseRuntimeObject, meta.NewAccessor())

	return mapOfUniqueIdToDesiredRuntimeObject, nil
}

func (r *AuthServerReconciler) GetInstructionManual(mapOfUniqueIdToDesiredRuntimeObject map[string]runtime.Object) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	instructionManualLocation := "config/samples/dependency_manual_authserver.yaml"
	instructionManual, err := controllers_utils.CreateInstructionManualFromYaml(instructionManualLocation)
	if err != nil {
		return nil, err
	}
	return instructionManual, nil
}

// +kubebuilder:rbac:groups=synopsys.com,resources=authServer,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=authServer/status,verbs=get;update;patch

func (r *AuthServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// _ = context.Background()
	// _ = r.Log.WithValues("authServer", req.NamespacedName)
	// your logic here

	return flying_dutchman.MetaReconcile(req, r)
}

func (r *AuthServerReconciler) SetIndexingForChildrenObjects(mgr ctrl.Manager, ro runtime.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ro, flying_dutchman.JobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		owner := metav1.GetControllerOf(ro.(metav1.Object))
		if owner == nil {
			return nil
		}
		// ...make sure it's a Alert...
		if owner.APIVersion != synopsysv1.GroupVersion.String() || owner.Kind != "AuthServer" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return nil
}

func (r *AuthServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.SetIndexingForChildrenObjects(mgr, &corev1.Service{})
	r.SetIndexingForChildrenObjects(mgr, &appsv1.Deployment{})
	r.SetIndexingForChildrenObjects(mgr, &corev1.ServiceAccount{})
	// Add HorizontalPodAutoscaler to the list

	authServerBuilder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.AuthServer{})
	authServerBuilder = authServerBuilder.Owns(&corev1.ConfigMap{})
	authServerBuilder = authServerBuilder.Owns(&corev1.Service{})
	authServerBuilder = authServerBuilder.Owns(&corev1.ReplicationController{})
	authServerBuilder = authServerBuilder.Owns(&corev1.Secret{})

	return authServerBuilder.Complete(r)
}
