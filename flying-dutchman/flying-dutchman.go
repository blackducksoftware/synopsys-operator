package flyingdutchman

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"

	scheduler "github.com/blackducksoftware/synopsys-operator/go-scheduler"

	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apijson "k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	JobOwnerKey = ".metadata.controller"
)

// MetaReconcilerInterface is a generic interface for running a Reconcile process
type MetaReconcilerInterface interface {
	// GetClient expects that the implementer returns a "controller-runtime" client
	GetClient() client.Client
	// GetCustomResource expects that the implementer returns the custom resource to watch against
	GetCustomResource(ctrl.Request) (metav1.Object, error)
	// GetRuntimeObjects expects that the implementer returns a map of uniqueID to runtime.Object to schedule to the api-server
	GetRuntimeObjects(interface{}) (map[string]runtime.Object, error)
	// GetInstructionManual expects that the implementer returns a pointer to the instruction manual
	GetInstructionManual(obj map[string]runtime.Object) (*RuntimeObjectDependencyYaml, error)
}

// MetaReconcile takes as input a request and an implementer of MetaReconcilerInterface
// It's to be used inside of a Reconcile loop
func MetaReconcile(req ctrl.Request, mri MetaReconcilerInterface, crScheme *runtime.Scheme) (ctrl.Result, error) {

	// get the client
	givenClient := mri.GetClient()

	// get the specific custom resource
	cr, err := mri.GetCustomResource(req)
	if err != nil {
		return ctrl.Result{}, err
	}

	// if cr is not found, currently we will not reconcile
	if cr == nil {
		// TODO: rethink what to do when copyOfCr CR isn't found (requeue after ??)
		return ctrl.Result{}, nil
	}

	// get copyOfCr map of unique id to runtime.Object
	mapOfuniqueIDToDesiredRuntimeObject, err := mri.GetRuntimeObjects(cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	// get the instruction manual
	instructionManual, err := mri.GetInstructionManual(mapOfuniqueIDToDesiredRuntimeObject)
	if err != nil {
		return ctrl.Result{}, err
	}

	// make a copy of the cr
	copyOfCr := cr.(runtime.Object).DeepCopyObject()
	metav1CopyOfCr := copyOfCr.(metav1.Object)

	// hand-off to the scheduler
	err = scheduleResources(givenClient, metav1CopyOfCr, mapOfuniqueIDToDesiredRuntimeObject, instructionManual, crScheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// scheduleResources takes as input a controller-runtime client, a custom resource, a map of runtime.Object, and a instruction manual
// It creates a task graph and schedules the runtime objects to the Kubernetes api-server
func scheduleResources(myClient client.Client, cr metav1.Object, mapOfuniqueIDToDesiredRuntimeObject map[string]runtime.Object, instructionManual *RuntimeObjectDependencyYaml, crScheme *runtime.Scheme) error {

	// create a context for ScheduleResources
	ctx := context.Background()

	// create a log
	log := ctrl.Log.WithName("ScheduleResources")

	// get current runtime objects "owned" by CR
	log.V(1).Info("Getting a list of existing runtime objects owned by", "Custom Resource: ", cr.GetName())
	var listOfCurrentRuntimeObjectsOwnedByCr metav1.List
	if err := myClient.List(ctx, &listOfCurrentRuntimeObjectsOwnedByCr, client.InNamespace(cr.GetNamespace()), client.MatchingField(JobOwnerKey, cr.GetName())); err != nil {
		// TODO: this is not working
		//log.Error(err, "unable to list currentRuntimeObjectsOwnedByAlertCr")
		// TODO: rethink what to do when we cannot list currentRuntimeObjectsOwnedByAlertCr
		//return ctrl.Result{}, nil
	}

	// if any of the existing objects are not in the desired objects, delete them
	accessor := meta.NewAccessor()
	for _, currentRuntimeRawExtensionOwnedByCr := range listOfCurrentRuntimeObjectsOwnedByCr.Items {
		// TODO: change all of this once we use labels
		currentRuntimeObjectOwnedByCr := currentRuntimeRawExtensionOwnedByCr.Object.(runtime.Object)
		currentRuntimeObjectKind, _ := accessor.Kind(currentRuntimeObjectOwnedByCr)
		currentRuntimeObjectName, _ := accessor.Name(currentRuntimeObjectOwnedByCr)
		currentRuntimeObjectNamespace, _ := accessor.Namespace(currentRuntimeObjectOwnedByCr)
		uniqueID := fmt.Sprintf("%s.%s.%s", currentRuntimeObjectKind, currentRuntimeObjectNamespace, currentRuntimeObjectName)
		_, ok := mapOfuniqueIDToDesiredRuntimeObject[uniqueID]
		if !ok {
			log.V(1).Info("Deleting runtime objects owned by cr, but no longer desired", "Custom Resource: ", cr, "Object being deleted", currentRuntimeObjectOwnedByCr)
			// TODO: third parameter in Delete: not sure if we should propagate
			err := myClient.Delete(ctx, currentRuntimeObjectOwnedByCr, client.PropagationPolicy(metav1.DeletePropagationForeground))
			if err != nil {
				// TODO: rethink what to do when we cannot delete a non-desired item
				// if any error in deleting, just continue
				continue
			}
		}
	}

	// create a scheduler
	crScheduler := scheduler.New()
	// create a task map to use later to draw all the dependencies
	taskMap := make(map[string]*scheduler.Task)
	for uniqueID, desiredRuntimeObject := range mapOfuniqueIDToDesiredRuntimeObject {
		// pass a copy of the runtime object to scheduler to avoid concurrency issues
		copyOfDesiredRuntimeObject := desiredRuntimeObject.DeepCopyObject()
		// create a task function
		taskFunc := func(ctx context.Context) error {
			// TODO: rethink, currently we only use the error, maybe EnsureRuntimeObject should just return an error
			_, err := EnsureRuntimeObject(ctx, myClient, log, copyOfDesiredRuntimeObject, crScheme, cr)
			return err
		}
		//log.V(1).Info("Adding a task for the runtime object", "Runtime Object", copyOfDesiredRuntimeObject)
		// add the task function to the scheduler
		task := crScheduler.AddTask(taskFunc)
		// add the task to the task map
		taskMap[uniqueID] = task
	}

	// iterate through the given dependencies in the instruction manual and add the dependency edge for the tasks
	for _, dependency := range instructionManual.Dependencies {
		child := dependency.Obj
		parents := dependency.IsDependentOn
		// get all RuntimeObjects for the Tail
		listOfChildRuntimeObjects, ok := instructionManual.Groups[child]
		if !ok {
			// no group due to single object name
			listOfChildRuntimeObjects = []string{child}
		}

		var listOfParentRuntimeObjects []string
		// get all RuntimeObjects for the Head
		for _, parent := range parents {
			parentRuntimeObjects, ok := instructionManual.Groups[parent]
			if ok {
				listOfParentRuntimeObjects = append(listOfParentRuntimeObjects, parentRuntimeObjects...)
			}
		}
		// Create dependencies from each tail to each head
		for _, childRuntimeObjectuniqueID := range listOfChildRuntimeObjects {
			for _, parentRuntimeObjectuniqueID := range listOfParentRuntimeObjects {
				taskMap[childRuntimeObjectuniqueID].DependsOn(taskMap[parentRuntimeObjectuniqueID])
				log.V(1).Info("Creating Task Dependency", "Child", childRuntimeObjectuniqueID, "depends on Parent", parentRuntimeObjectuniqueID)
			}
		}
	}

	// finally run the scheduler with a new context
	schedulerCtx := context.Background()
	if err := crScheduler.Run(schedulerCtx); err != nil {
		return err
	}

	// if everything runs successfully, return nil to caller
	return nil
}

// EnsureRuntimeObject ensures the run time object
func EnsureRuntimeObject(ctx context.Context, myClient client.Client, log logr.Logger, desiredRuntimeObject runtime.Object, crScheme *runtime.Scheme, cr metav1.Object) (ctrl.Result, error) {
	// TODO: either get this working or wait for server side apply
	// TODO: https://github.com/kubernetes-sigs/controller-runtime/issues/347
	// TODO: https://github.com/kubernetes/kubernetes/issues/73723
	// TODO: https://github.com/kubernetes-sigs/structured-merge-diff
	// TODO: https://github.com/kubernetes/apimachinery/tree/master/pkg/util/strategicpatch
	// TODO: https://github.com/kubernetes-sigs/controller-runtime/issues/464
	// TODO: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate

	//pointerToDesiredRuntimeObject := &desiredRuntimeObject
	//copyOfDesiredRuntimeObject := desiredRuntimeObject.DeepCopyObject()
	//pointerToCopyOfDesiredRuntimeObject := &copyOfDesiredRuntimeObject
	//opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, *pointerToDesiredRuntimeObject, func() error {
	//	*pointerToDesiredRuntimeObject = *pointerToCopyOfDesiredRuntimeObject
	//	// Set an owner reference
	//	if err := ctrl.SetControllerReference(cr, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
	//		// Requeue if we cannot set owner on the object
	//		//return err
	//		return nil
	//	}
	//	return nil
	//})

	// BEGIN BORROWED CODE FROM controllerutil.CreateOrUpdate
	opResult, err := CreateOrUpdate(ctx, myClient, desiredRuntimeObject, crScheme, cr)
	// END BORROWED CODE FROM controllerutil.CreateOrUpdate
	if err != nil {
		// TODO: Case 1: we needed to update the configMap and now we should delete and redeploy objects in STAGE 3, 4 ...
		// TODO: Case 2: we failed to update the configMap...TODO
		// TODO: delete everything in stages 3, 4 ... and requeue
		log.Error(err, "Unable to create or update", "desiredRuntimeObject", desiredRuntimeObject, "error", err)
		return ctrl.Result{}, err
	}
	log.V(1).Info("Result of create or update for", "desiredRuntimeObject", desiredRuntimeObject.GetObjectKind(), "opResult", opResult)

	if err := CheckForReadiness(myClient, desiredRuntimeObject); err != nil {
		// TODO: requeue after here, think about logic here [jeremy / aditya]
		log.Error(err, "CheckForReadiness failed", "desiredRuntimeObject", desiredRuntimeObject)
		return ctrl.Result{}, err
	}

	// finally return nil if ensured successfully
	return ctrl.Result{}, nil
}

// CheckForReadiness checks the readiness of the run time object
func CheckForReadiness(myClient client.Client, desiredRuntimeObject runtime.Object) error {
	// TODO: Check for readiness/completeness
	// TODO: This will probably be a complex topic, good place to start is this upstream issue and doc:
	// TODO: https://github.com/kubernetes/kubernetes/issues/1899
	// TODO: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/

	// get the key from desired rto to look up the live object in cluster
	key, err := client.ObjectKeyFromObject(desiredRuntimeObject)
	if err != nil {
		return err
	}

	// switch on the type
	switch desiredRuntimeObject.(type) {

	case *corev1.Pod:
		livePod := &corev1.Pod{}
		_ = myClient.Get(context.TODO(), key, livePod)
		return isPodReady(livePod)

	case *corev1.Service:
		liveService := &corev1.Service{}
		_ = myClient.Get(context.TODO(), key, liveService)
		return isServiceReady(liveService)

	case *corev1.ReplicationController:
		liveReplicationController := &corev1.ReplicationController{}
		_ = myClient.Get(context.TODO(), key, liveReplicationController)
		return isReplicationControllerReady(liveReplicationController)

	case *appsv1.Deployment:
		liveDeployment := &appsv1.Deployment{}
		_ = myClient.Get(context.TODO(), key, liveDeployment)
		return isDeploymentReady(liveDeployment)

	case *batchv1.Job:
		liveJob := &batchv1.Job{}
		_ = myClient.Get(context.TODO(), key, liveJob)
		return isJobCompleted(liveJob)

	case *appsv1.StatefulSet:
		liveStatefulSet := &appsv1.StatefulSet{}
		_ = myClient.Get(context.TODO(), key, liveStatefulSet)
		return isStatefulSetReady(liveStatefulSet)

	default:
		fmt.Printf("Unknown type: %s is marked healthy by default\n", desiredRuntimeObject.GetObjectKind())
		return nil

	}
}

func isPodReady(pod *corev1.Pod) error {
	if &pod.Status != nil && len(pod.Status.Conditions) > 0 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady &&
				condition.Status == corev1.ConditionTrue {
				return nil
			}
		}
	}
	return fmt.Errorf("pod is not ready: %s/%s", pod.GetNamespace(), pod.GetName())
}

func isServiceReady(svc *corev1.Service) error {
	// Make sure the service is not explicitly set to "None" before checking the IP
	if svc.Spec.ClusterIP != corev1.ClusterIPNone && svc.Spec.ClusterIP == "" {
		return fmt.Errorf("service is not ready: %s/%s", svc.GetNamespace(), svc.GetName())
	}
	// This checks if the service has a load-balancer and that lb has an Ingress defined
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && svc.Status.LoadBalancer.Ingress == nil {
		return fmt.Errorf("service is not ready: %s/%s", svc.GetNamespace(), svc.GetName())
	}
	return nil
}

func isReplicationControllerReady(rc *corev1.ReplicationController) error {
	if rc.Spec.Replicas == nil || rc.Status.ReadyReplicas != *rc.Spec.Replicas {
		return fmt.Errorf("replication controller is not ready: %s/%s", rc.GetNamespace(), rc.GetName())
	}
	return nil
}

func isDeploymentReady(deployment *appsv1.Deployment) error {
	// TODO: Check for "Avaialbility" of the deployment
	if deployment.Spec.Replicas == nil || deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
		return fmt.Errorf("deployment is not ready: %s/%s", deployment.GetNamespace(), deployment.GetName())
	}
	return nil
}

func isStatefulSetReady(statefulSet *appsv1.StatefulSet) error {
	if statefulSet.Spec.Replicas == nil || statefulSet.Status.ReadyReplicas != *statefulSet.Spec.Replicas {
		return fmt.Errorf("statefulSet is not ready: %s/%s", statefulSet.GetNamespace(), statefulSet.GetName())
	}
	return nil
}

func isJobCompleted(job *batchv1.Job) error {
	// TODO: https://github.com/kubernetes/kubernetes/issues/68712
	if &job.Status != nil && len(job.Status.Conditions) > 0 {
		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
				return nil
			}
		}
	}
	return fmt.Errorf("job is not ready: %s/%s", job.GetNamespace(), job.GetName())
}

// TODO: Borrowed and modified slightly from https://godoc.org/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
// Changes include removing the mutateFn, using semantic.DeepEqual, doing a deepcopy before create cause create will muck with it

// CreateOrUpdate creates or updates the run time object
func CreateOrUpdate(ctx context.Context, c client.Client, desiredRuntimeObject runtime.Object, crScheme *runtime.Scheme, cr metav1.Object) (controllerutil.OperationResult, error) {
	key, err := client.ObjectKeyFromObject(desiredRuntimeObject)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	// CHANGE #1
	currentRuntimeObject := desiredRuntimeObject.DeepCopyObject()
	if err := c.Get(ctx, key, currentRuntimeObject); err != nil {
		if !apierrs.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}

		// set an owner reference only when we create
		if err := ctrl.SetControllerReference(cr, desiredRuntimeObject.(metav1.Object), crScheme); err != nil {
			// requeue if we cannot set owner on the object
			return controllerutil.OperationResultNone, err
		}

		if err := c.Create(ctx, desiredRuntimeObject); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	// CHANGE #2
	// TODO: need more than this cause server puts some default
	// TODO: good info in issue here: https://github.com/kubernetes-sigs/controller-runtime/issues/464
	if equality.Semantic.DeepEqual(currentRuntimeObject, desiredRuntimeObject) {
		return controllerutil.OperationResultNone, nil
	}
	//TODO: maybe use https://godoc.org/k8s.io/apimachinery/pkg/util/strategicpatch or https://godoc.org/sigs.k8s.io/structured-merge-diff
	//strategicpatch.CreateTwoWayMergePatch(existing, desiredRuntimeObject, )
	rawDesiredRuntimeObjectInBytes, _ := apijson.Marshal(desiredRuntimeObject)

	if err := c.Patch(ctx, currentRuntimeObject, client.ConstantPatch(types.ApplyPatchType, rawDesiredRuntimeObjectInBytes)); err != nil {
		// CHANGE #3
		// TODO:
		return controllerutil.OperationResultNone, nil
		//return controllerutil.OperationResultNone, err
	}
	return controllerutil.OperationResultUpdated, nil
}

// TODO: original that the above code was modified from on 2019-08-12
//func CreateOrUpdate(ctx context.Context, c client.Client, obj runtime.Object, f MutateFn) (OperationResult, error) {
//	key, err := client.ObjectKeyFromObject(obj)
//	if err != nil {
//		return OperationResultNone, err
//	}
//
//	if err := c.Get(ctx, key, obj); err != nil {
//		if !errors.IsNotFound(err) {
//			return OperationResultNone, err
//		}
//		if err := mutate(f, key, obj); err != nil {
//			return OperationResultNone, err
//		}
//		if err := c.Create(ctx, obj); err != nil {
//			return OperationResultNone, err
//		}
//		return OperationResultCreated, nil
//	}
//
//	existing := obj.DeepCopyObject()
//	if err := mutate(f, key, obj); err != nil {
//		return OperationResultNone, err
//	}
//
//	if reflect.DeepEqual(existing, obj) {
//		return OperationResultNone, nil
//	}
//
//	if err := c.Update(ctx, obj); err != nil {
//		return OperationResultNone, err
//	}
//	return OperationResultUpdated, nil
//}
