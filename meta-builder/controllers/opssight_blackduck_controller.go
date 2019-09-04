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

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"

	"github.com/go-logr/logr"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// OpsSightBlackDuckReconciler reconciles an OpsSight and Black Duck object
type OpsSightBlackDuckReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// NewOpsSightBlackDuckReconciler ...
func NewOpsSightBlackDuckReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger) *OpsSightBlackDuckReconciler {
	return &OpsSightBlackDuckReconciler{
		Client: client,
		Scheme: scheme,
		Log:    log,
	}
}

// +kubebuilder:rbac:groups=synopsys.com,resources=blackducks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=synopsys.com,resources=blackducks/status,verbs=get;update;patch

// Reconcile reconcile the Black Duck custom resources for all OpsSight instances
func (p *OpsSightBlackDuckReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	// Fetch the Black Duck instance
	blackDuck := &synopsysv1.Blackduck{}
	err := p.Client.Get(context.TODO(), request.NamespacedName, blackDuck)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			p.Log.V(1).Info("updater - Black Duck deleted event", "namespacedName", request.NamespacedName)
			p.sync()
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	p.Log.V(1).Info("updater - Black Duck added event", "obj", blackDuck)
	running := p.isBlackDuckRunning(blackDuck)
	if !running {
		p.sync()
	}
	return reconcile.Result{}, nil
}

// SetupWithManager setup the Controller runtime manager
func (p *OpsSightBlackDuckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&synopsysv1.Blackduck{})
	return builder.Complete(p)
}
