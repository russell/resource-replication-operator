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

package replicator

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	utilsv1alpha1 "github.com/russell/resource-replication-operator/api/v1alpha1"
	"github.com/russell/resource-replication-operator/replicator/common"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

type SecretReplicator struct {
	client.Client
	Log logr.Logger
}

func (r *SecretReplicator) ReplicateSecret(ctx context.Context, rep *utilsv1alpha1.ReplicatedResource) (controllerutil.OperationResult, *corev1.Secret, error) {
	sourceNamespacedName := types.NamespacedName{Namespace: rep.Spec.Source.Namespace, Name: rep.Spec.Source.Name}
	destNamespacedName := types.NamespacedName{Namespace: rep.Namespace, Name: rep.Name}

	log := r.Log.WithValues(
		"type", "secret",
		"source", fmt.Sprintf("%s/%s", sourceNamespacedName.Namespace, sourceNamespacedName.Name),
		"destination", fmt.Sprintf("%s/%s", destNamespacedName.Namespace, destNamespacedName.Name))
	log.Info("Replicating")
	source := &corev1.Secret{}
	if err := r.Get(ctx, sourceNamespacedName, source); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Info("Error reading source")
			return controllerutil.OperationResultNone, nil, err
		} else {
			log.Info("Could not find source secret")
			return controllerutil.OperationResultNone, nil, errors.New(fmt.Sprintf("Could not find source secret %s/%s",
				rep.Spec.Source.Namespace, rep.Spec.Source.Name))
		}
	}

	t := true
	owners := []metav1.OwnerReference{
		{
			Name:               rep.Name,
			Kind:               rep.Kind,
			APIVersion:         rep.APIVersion,
			UID:                rep.UID,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		},
	}

	dest := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            rep.ObjectMeta.Name,
			Namespace:       rep.ObjectMeta.Namespace,
			OwnerReferences: owners,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dest, func() error {
		// Only update if there is a new version
		if dest.Annotations != nil && dest.Annotations[common.ReplicatedFromVersionAnnotation] == source.ResourceVersion {
			return nil
		}

		if dest.Annotations == nil {
			dest.Annotations = make(map[string]string)
		}
		dest.Annotations[common.ReplicatedAtAnnotation] = time.Now().Format(time.RFC3339)
		dest.Annotations[common.ReplicatedFromVersionAnnotation] = source.ResourceVersion
		dest.Type = source.Type
		dest.Data = source.Data
		return nil
	})
	log.Info(fmt.Sprintf("Updated Secret %s", op))
	log.Info(fmt.Sprintf("Updated err %s", err))
	return op, dest, err
}

func SecretNeedsUpdating(rep utilsv1alpha1.ReplicatedResource, secret corev1.Secret) (bool, error) {
	return false, nil
}
