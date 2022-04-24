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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"

	utilsv1alpha1 "github.com/russell/resource-replication-operator/api/v1alpha1"
)

var _ = Describe("CronJob controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ReplicatedResourceName      = "test-replicated-secret"
		ReplicatedResourceNamespace = "default"
		SecretName                  = "test-secret"
		SecretNamespace             = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating ReplicatedResource Status", func() {
		It("Should increase ReplicatedResource Status.Active count when new Jobs are created", func() {
			By("By creating a new ReplicatedResource")
			ctx := context.Background()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretName,
					Namespace: SecretNamespace,
				},
				Data: map[string][]byte{
					"test": []byte("dGhpcyBpcyBhIHRlc3Qu"),
				},
			}

			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			// Wait for secret to be created
			secretLookupKey := types.NamespacedName{Name: SecretName, Namespace: SecretNamespace}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, &corev1.Secret{})
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			replicatedResource := &utilsv1alpha1.ReplicatedResource{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "utils.simopolis.xyz/v1alpha1",
					Kind:       "ReplicatedResource",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ReplicatedResourceName,
					Namespace: ReplicatedResourceNamespace,
				},
				Spec: utilsv1alpha1.ReplicatedResourceSpec{
					Source: utilsv1alpha1.ReplicatedResourceSource{
						Namespace: SecretNamespace,
						Name:      SecretName,
						Kind:      "Secret",
					},
				},
			}
			Expect(k8sClient.Create(ctx, replicatedResource)).Should(Succeed())
			replicatedResourceLookupKey := types.NamespacedName{Name: ReplicatedResourceName, Namespace: ReplicatedResourceNamespace}
			createdReplicatedResource := &utilsv1alpha1.ReplicatedResource{}

			// We'll need to retry getting this newly created
			// ReplicatedResource, given that creation may not
			// immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, replicatedResourceLookupKey, createdReplicatedResource)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			replicatedSecretLookupKey := types.NamespacedName{Name: ReplicatedResourceName, Namespace: ReplicatedResourceNamespace}
			createdSecret := &corev1.Secret{}

			// We'll need to retry getting this newly created
			// ReplicatedResource, given that creation may not
			// immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, replicatedSecretLookupKey, createdSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdSecret.Data["test"]).Should(Equal([]byte("dGhpcyBpcyBhIHRlc3Qu")))
		})
	})
})
