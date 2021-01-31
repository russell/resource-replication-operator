/*
Copyright 2021.

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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	utilsv1 "github.com/russell/resource-replication-operator/api/v1"
	"github.com/russell/resource-replication-operator/replicator/common"
)

// ReplicatedResourceReconciler reconciles a ReplicatedResource object
type ReplicatedResourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=utils.simopolis.xyz,resources=replicatedresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReplicatedResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ReplicatedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("replicatedresource", req.NamespacedName)
	rr := &utilsv1.ReplicatedResource{}
	if err := r.Get(ctx, req.NamespacedName, rr); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		} else {
			log.Info("Could not find ReplicatedResource")
			return ctrl.Result{}, nil
		}
	}
	log.Info("processing")

	sourceNamespacedName := types.NamespacedName{Namespace: rr.Spec.Source.Namespace, Name: rr.Spec.Source.Name}
	sourceKind := rr.Spec.Source.Kind

	if req.NamespacedName == sourceNamespacedName {
		log.Info("Can't replicate when the source matches the source")
		return ctrl.Result{}, nil
	}

	if sourceKind == "Secret" {
		log.Info("Replicating secret")
		source := &corev1.Secret{}
		if err := r.Get(ctx, sourceNamespacedName, source); err != nil {
			if !errors.IsNotFound(err) {
				log.Info("Error reading source")
				return ctrl.Result{}, err
			} else {
				log.Info("Could not find source secret")
				return ctrl.Result{}, nil
			}
		}

		destNamespacedName := types.NamespacedName{Namespace: rr.Namespace, Name: rr.Name}

		dest := &corev1.Secret{}
		if err := r.Get(ctx, destNamespacedName, dest); err != nil {
			if !errors.IsNotFound(err) {
				log.Info("Error reading dest")
				return ctrl.Result{}, err
			} else {
				dest.Type = source.Type
			}
		}

		dest.ObjectMeta.Namespace = rr.ObjectMeta.Namespace
		dest.ObjectMeta.Name = rr.ObjectMeta.Name

		op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dest, func() error {
			if dest.Annotations == nil {
				dest.Annotations = make(map[string]string)
			}
			dest.Annotations[common.ReplicatedAtAnnotation] = time.Now().Format(time.RFC3339)
			dest.Annotations[common.ReplicatedFromVersionAnnotation] = source.ResourceVersion
			dest.Data = source.Data
			t := true
			dest.SetOwnerReferences(
				[]metav1.OwnerReference{
					{
						Name:               rr.Name,
						Kind:               rr.Kind,
						APIVersion:         rr.APIVersion,
						UID:                rr.UID,
						Controller:         &t,
						BlockOwnerDeletion: &t,
					},
				},
			)
			return nil
		})
		log.Info(fmt.Sprintf("Updated Secret %s", op))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Unsupported kind %s", sourceKind))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReplicatedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&utilsv1.ReplicatedResource{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
