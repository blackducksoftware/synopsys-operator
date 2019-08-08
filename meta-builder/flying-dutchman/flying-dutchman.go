package flying_dutchman

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	scheduler "github.com/yashbhutwala/go-scheduler"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	JobOwnerKey = ".metadata.controller"
)

type MetaReconcilerInterface interface {
	GetClient() client.Client
	GetCustomResource(ctrl.Request) (metav1.Object, error)
	GetRuntimeObjects(interface{}) (map[string]runtime.Object, error)
	CreateInstructionManual() (*RuntimeObjectDepencyYaml, error)
}

func MetaReconciler(req ctrl.Request, mri MetaReconcilerInterface) (ctrl.Result, error) {
	myClient := mri.GetClient()
	cr, err := mri.GetCustomResource(req)
	if err != nil {
		return ctrl.Result{}, err
	}
	mapOfUniqueIdToDesiredRuntimeObject, err := mri.GetRuntimeObjects(cr)
	if err != nil {
		return ctrl.Result{}, err
	}
	instructionManual, err := mri.CreateInstructionManual()
	if err != nil {
		return ctrl.Result{}, err
	}
	a := cr.(runtime.Object).DeepCopyObject()
	err = ScheduleResources(myClient, a.(metav1.Object), mapOfUniqueIdToDesiredRuntimeObject, instructionManual)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func ScheduleResources(myClient client.Client, cr metav1.Object, mapOfUniqueIdToDesiredRuntimeObject map[string]runtime.Object, instructionManual *RuntimeObjectDepencyYaml) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("ScheduleResources")
	// Get current runtime objects "owned" by Alert CR
	fmt.Printf("Creating Tasks for RuntimeObjects...\n")
	var listOfCurrentRuntimeObjectsOwnedByAlertCr metav1.List
	if err := myClient.List(ctx, &listOfCurrentRuntimeObjectsOwnedByAlertCr, client.InNamespace(cr.GetNamespace()), client.MatchingField(JobOwnerKey, cr.GetName())); err != nil {
		log.Error(err, "unable to list currentRuntimeObjectsOwnedByAlertCr")
		//TODO: redo
		//return ctrl.Result{}, nil
	}

	// If any of the current objects are not in the desired objects, delete them
	fmt.Printf("Creating Task Dependencies...\n")
	accessor := meta.NewAccessor()
	for _, currentRuntimeRawExtensionOwnedByAlertCr := range listOfCurrentRuntimeObjectsOwnedByAlertCr.Items {
		currentRuntimeObjectOwnedByAlertCr := currentRuntimeRawExtensionOwnedByAlertCr.Object.(runtime.Object)
		currentRuntimeObjectKind, _ := accessor.Kind(currentRuntimeObjectOwnedByAlertCr)
		currentRuntimeObjectName, _ := accessor.Name(currentRuntimeObjectOwnedByAlertCr)
		currentRuntimeObjectNamespace, _ := accessor.Namespace(currentRuntimeObjectOwnedByAlertCr)
		uniqueId := fmt.Sprintf("%s.%s.%s", currentRuntimeObjectKind, currentRuntimeObjectNamespace, currentRuntimeObjectName)
		_, ok := mapOfUniqueIdToDesiredRuntimeObject[uniqueId]
		if !ok {
			err := myClient.Delete(ctx, currentRuntimeObjectOwnedByAlertCr)
			if err != nil {
				// if any error in deleting, just continue
				continue
			}
		}
	}

	alertScheduler := scheduler.New(scheduler.ConcurrentTasks(5))
	taskMap := make(map[string]*scheduler.Task)
	for uniqueId, desiredRuntimeObject := range mapOfUniqueIdToDesiredRuntimeObject {
		rto := desiredRuntimeObject.DeepCopyObject()
		taskFunc := func(ctx context.Context) error {
			_, err := EnsureRuntimeObjects(myClient, ctx, log, rto)
			return err
		}
		task := alertScheduler.AddTask(taskFunc)
		taskMap[uniqueId] = task
	}

	for _, dependency := range instructionManual.Dependencies {
		depTail := dependency.Obj
		depHead := dependency.IsDependentOn // depTail --> depHead
		fmt.Printf("Creating Task Dependency: %s -> %s\n", depTail, depHead)
		// Get all RuntimeObjects for the Tail
		tailRuntimeObjectIDs, ok := instructionManual.Groups[depTail]
		if !ok { // no group due to single object name
			tailRuntimeObjectIDs = []string{depTail}
		}
		// Get all RuntimeObjects for the Head
		headRuntimeObjectIDs, ok := instructionManual.Groups[depHead]
		if !ok { // no group due to single object name
			headRuntimeObjectIDs = []string{depHead}
		}
		// Create dependencies from each tail to each head
		for _, tailRuntimeObjectName := range tailRuntimeObjectIDs {
			for _, headRuntimeObjectName := range headRuntimeObjectIDs {
				taskMap[tailRuntimeObjectName].DependsOn(taskMap[headRuntimeObjectName])
				fmt.Printf("   -  %s -> %s\n", tailRuntimeObjectName, headRuntimeObjectName)
			}
		}
	}

	if err := alertScheduler.Run(context.Background()); err != nil {
		return err
	}
	return nil
}

func EnsureRuntimeObjects(myClient client.Client, ctx context.Context, log logr.Logger, desiredRuntimeObject runtime.Object) (ctrl.Result, error) {
	// TODO: either get this working or wait for server side apply
	// TODO: https://github.com/kubernetes-sigs/controller-runtime/issues/347
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

	var opResult controllerutil.OperationResult
	key, err := client.ObjectKeyFromObject(desiredRuntimeObject)
	if err != nil {
		opResult = controllerutil.OperationResultNone
	}

	currentRuntimeObject := desiredRuntimeObject.DeepCopyObject()
	if err := myClient.Get(ctx, key, currentRuntimeObject); err != nil {
		if !apierrs.IsNotFound(err) {
			opResult = controllerutil.OperationResultNone
		}
		if err := myClient.Create(ctx, desiredRuntimeObject); err != nil {
			opResult = controllerutil.OperationResultNone
		}
		opResult = controllerutil.OperationResultCreated
	}

	existing := currentRuntimeObject
	if reflect.DeepEqual(existing, desiredRuntimeObject) {
		opResult = controllerutil.OperationResultNone
	}

	if err := myClient.Update(ctx, desiredRuntimeObject); err != nil {
		opResult = controllerutil.OperationResultNone
	}

	log.V(1).Info("Result of CreateOrUpdate on CFSSL desiredRuntimeObject", "desiredRuntimeObject", desiredRuntimeObject, "opResult", opResult)

	// TODO: Case 1: we needed to update the configMap and now we should delete and redploy objects in STAGE 3, 4 ...
	// TODO: Case 2: we failed to update the configMap...TODO
	if err != nil {
		// TODO: delete everything in stages 3, 4 ... and requeue
		log.Error(err, "unable to create or update STAGE 2 objects, deleting all child Objects", "desiredRuntimeObject", desiredRuntimeObject)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
