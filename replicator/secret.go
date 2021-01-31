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
	//	"github.com/go-logr/logr"
	utilsv1 "github.com/russell/resource-replication-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func ReplicateConfigMap(rep utilsv1.ReplicatedResource, secret corev1.Secret) (corev1.Secret, error) {
	return corev1.Secret{}, nil
}

func SecretNeedsUpdating(rep utilsv1.ReplicatedResource, secret corev1.Secret) (bool, error) {
	return false, nil
}
