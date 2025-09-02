/*
Copyright 2021-2022 Russell Sim <russell.sim@gmail.com>

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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	utilsv1alpha1 "github.com/russell/resource-replication-operator/api/v1alpha1"
	"github.com/russell/resource-replication-operator/replicator"
)

// ReplicatedResourceReconciler reconciles a ReplicatedResource object
type ReplicatedResourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	nameField      = ".spec.source.name"
	namespaceField = ".spec.source.namespace"
	kindField      = ".spec.source.kind"
)

// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources/finalizers,verbs=update
func (r *ReplicatedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("replicatedresource", req.NamespacedName)

	rr := &utilsv1alpha1.ReplicatedResource{}
	if err := r.Get(ctx, req.NamespacedName, rr); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		} else {
			log.Info("Could not find ReplicatedResource. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
	}
	log.Info("Started Processing")

	sourceNamespacedName := types.NamespacedName{Namespace: rr.Spec.Source.Namespace, Name: rr.Spec.Source.Name}
	sourceKind := rr.Spec.Source.Kind
	var replicateError error = nil
	requeueAfter := time.Duration(0)

	if req.NamespacedName == sourceNamespacedName {
		log.Info("Can't replicate when the source matches the source")
		return ctrl.Result{}, nil
	}
	var op controllerutil.OperationResult
	if sourceKind == "Secret" {
		sr := replicator.SecretReplicator{Client: r.Client, Log: r.Log, Scheme: r.Scheme}
		operation, _, err := sr.ReplicateSecret(ctx, rr)
		op = operation
		replicateError = err
	} else {
		replicateError = fmt.Errorf("Unsupported kind %s", sourceKind)
	}

	if replicateError != nil {
		rr.Status.Phase = "Failed"
		rr.Status.Conditions = []utilsv1alpha1.ReplicatedResourceCondition{{
			Type:               utilsv1alpha1.ReplicatedResourceComplete,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "Error",
			Message:            replicateError.Error(),
		}}

	} else if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		rr.Status.Phase = "Completed"
		rr.Status.Conditions = []utilsv1alpha1.ReplicatedResourceCondition{{
			Type:               utilsv1alpha1.ReplicatedResourceComplete,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "Replicated",
			Message:            "Successfully Replicated",
		}}
	} else {
		// No operation was performed - this could be due to cache consistency issues
		// Requeue after a short delay to allow cache to sync
		log.Info("No operation performed, requeuing to check for cache updates")
		return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
	}

	if err := r.Status().Update(ctx, rr); err != nil {
		log.Info(fmt.Sprintf("Error updating ReplicatedResource: %s", err))
		return ctrl.Result{RequeueAfter: requeueAfter}, err
	}

	log.Info("Successfully Replicated")

	return ctrl.Result{}, nil
}

func (r *ReplicatedResourceReconciler) findObjectsForSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	return r.findObjectsForReplicatedResource(obj, "Secret")
}

func (r *ReplicatedResourceReconciler) findObjectsForReplicatedResource(obj client.Object, objKind string) []reconcile.Request {
	attachedReplicatedResource := &utilsv1alpha1.ReplicatedResourceList{}

	objName := obj.GetName()
	objNamespace := obj.GetNamespace()

	r.Log.Info(fmt.Sprintf("Dependent %s/%s of kind %s updated triggering a refresh", objNamespace, objName, objKind))

	// Filter the list of replicated resources by the ones that
	// target this object by name and namespace
	listOps := &client.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector(kindField, objKind),
			fields.OneTermEqualSelector(nameField, objName),
			fields.OneTermEqualSelector(namespaceField, objNamespace)),
	}
	err := r.List(context.TODO(), attachedReplicatedResource, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedReplicatedResource.Items))
	for i, item := range attachedReplicatedResource.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReplicatedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &utilsv1alpha1.ReplicatedResource{}, nameField, func(rawObj client.Object) []string {
		// Extract the name from the ReplicatedResource Spec, if one is provided
		replicatedResource := rawObj.(*utilsv1alpha1.ReplicatedResource)
		if replicatedResource.Spec.Source.Name == "" {
			return nil
		}
		return []string{replicatedResource.Spec.Source.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &utilsv1alpha1.ReplicatedResource{}, namespaceField, func(rawObj client.Object) []string {
		// Extract the namespace from the ReplicatedResource Spec, if one is provided
		replicatedResource := rawObj.(*utilsv1alpha1.ReplicatedResource)
		if replicatedResource.Spec.Source.Namespace == "" {
			return nil
		}
		return []string{replicatedResource.Spec.Source.Namespace}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &utilsv1alpha1.ReplicatedResource{}, kindField, func(rawObj client.Object) []string {
		// Extract the namespace from the ReplicatedResource Spec, if one is provided
		replicatedResource := rawObj.(*utilsv1alpha1.ReplicatedResource)
		if replicatedResource.Spec.Source.Kind == "" {
			return nil
		}
		return []string{replicatedResource.Spec.Source.Kind}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&utilsv1alpha1.ReplicatedResource{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
