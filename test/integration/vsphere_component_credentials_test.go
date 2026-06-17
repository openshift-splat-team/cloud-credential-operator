/*
Copyright 2024 The OpenShift Authors.

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
package integration

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	minterv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	"github.com/openshift/cloud-credential-operator/pkg/operator/constants"
)

var _ = Describe("VSphere Component Credentials", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	Context("Component Secret Detection", func() {
		It("should detect machine-api component secret in kube-system", func() {
			ctx := context.Background()

			// Create component secret in kube-system
			componentSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereMachineAPICredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
				},
				Data: map[string][]byte{
					"username": []byte("machine-api-user"),
					"password": []byte("machine-api-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, componentSecret)).Should(Succeed())

			// Verify secret exists
			fetchedSecret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      constants.VSphereMachineAPICredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
				}, fetchedSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(string(fetchedSecret.Data["username"])).To(Equal("machine-api-user"))
		})

		It("should detect all four component secrets", func() {
			ctx := context.Background()

			componentSecrets := map[string]string{
				constants.VSphereMachineAPICredSecretName:      "machine-api-user",
				constants.VSphereStorageCredSecretName:         "storage-user",
				constants.VSphereCloudControllerCredSecretName: "cloud-controller-user",
				constants.VSphereDiagnosticsCredSecretName:     "diagnostics-user",
			}

			for secretName, username := range componentSecrets {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretName,
						Namespace: constants.CloudCredSecretNamespace,
					},
					Data: map[string][]byte{
						"username": []byte(username),
						"password": []byte("test-pass"),
					},
				}
				Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
			}

			// Verify all secrets exist
			for secretName, expectedUsername := range componentSecrets {
				fetchedSecret := &corev1.Secret{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      secretName,
						Namespace: constants.CloudCredSecretNamespace,
					}, fetchedSecret)
					return err == nil
				}, timeout, interval).Should(BeTrue())

				Expect(string(fetchedSecret.Data["username"])).To(Equal(expectedUsername))
			}
		})
	})

	Context("Credential Provisioning", func() {
		It("should provision machine-api credentials to openshift-machine-api namespace", func() {
			ctx := context.Background()

			// Create component secret
			componentSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereMachineAPICredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
				},
				Data: map[string][]byte{
					"username": []byte("machine-api-user"),
					"password": []byte("machine-api-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, componentSecret)).Should(Succeed())

			// Create target namespace
			targetNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "openshift-machine-api",
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).Should(Succeed())

			// Create CredentialsRequest
			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine-api-creds",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Name:      "vsphere-machine-api-credentials",
						Namespace: "openshift-machine-api",
					},
				},
			}
			Expect(k8sClient.Create(ctx, cr)).Should(Succeed())

			// Verify target secret is provisioned with correct data
			targetSecret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "vsphere-machine-api-credentials",
					Namespace: "openshift-machine-api",
				}, targetSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(string(targetSecret.Data["username"])).To(Equal("machine-api-user"))
		})

		It("should provision credentials to all operator namespaces", func() {
			ctx := context.Background()

			testCases := []struct {
				componentSecret string
				targetNamespace string
				targetSecret    string
				username        string
			}{
				{
					componentSecret: constants.VSphereMachineAPICredSecretName,
					targetNamespace: "openshift-machine-api",
					targetSecret:    "machine-api-creds",
					username:        "machine-api-user",
				},
				{
					componentSecret: constants.VSphereStorageCredSecretName,
					targetNamespace: "openshift-cluster-csi-drivers",
					targetSecret:    "storage-creds",
					username:        "storage-user",
				},
				{
					componentSecret: constants.VSphereCloudControllerCredSecretName,
					targetNamespace: "openshift-cloud-controller-manager",
					targetSecret:    "cloud-controller-creds",
					username:        "cloud-controller-user",
				},
				{
					componentSecret: constants.VSphereDiagnosticsCredSecretName,
					targetNamespace: "openshift-config",
					targetSecret:    "diagnostics-creds",
					username:        "diagnostics-user",
				},
			}

			for _, tc := range testCases {
				// Create component secret
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.componentSecret,
						Namespace: constants.CloudCredSecretNamespace,
					},
					Data: map[string][]byte{
						"username": []byte(tc.username),
						"password": []byte("test-pass"),
					},
				}
				Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

				// Create target namespace
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: tc.targetNamespace,
					},
				}
				Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

				// Create CredentialsRequest
				cr := &minterv1.CredentialsRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.targetSecret + "-cr",
						Namespace: "openshift-cloud-credential-operator",
					},
					Spec: minterv1.CredentialsRequestSpec{
						SecretRef: corev1.ObjectReference{
							Name:      tc.targetSecret,
							Namespace: tc.targetNamespace,
						},
					},
				}
				Expect(k8sClient.Create(ctx, cr)).Should(Succeed())

				// Verify provisioning
				targetSecret := &corev1.Secret{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      tc.targetSecret,
						Namespace: tc.targetNamespace,
					}, targetSecret)
					return err == nil
				}, timeout, interval).Should(BeTrue())

				Expect(string(targetSecret.Data["username"])).To(Equal(tc.username))
			}
		})
	})

	Context("Fallback to Shared Credential", func() {
		It("should use shared credential when component secret is missing", func() {
			ctx := context.Background()

			// Create only shared credential
			sharedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereCloudCredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
					Annotations: map[string]string{
						constants.AnnotationKey: constants.PassthroughAnnotation,
					},
				},
				Data: map[string][]byte{
					"username": []byte("shared-user"),
					"password": []byte("shared-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, sharedSecret)).Should(Succeed())

			// Create target namespace
			targetNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "openshift-cluster-csi-drivers",
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).Should(Succeed())

			// Create CredentialsRequest for storage (component secret doesn't exist)
			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "storage-fallback-cr",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Name:      "storage-fallback-creds",
						Namespace: "openshift-cluster-csi-drivers",
					},
				},
			}
			Expect(k8sClient.Create(ctx, cr)).Should(Succeed())

			// Verify target secret uses shared credential data
			targetSecret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "storage-fallback-creds",
					Namespace: "openshift-cluster-csi-drivers",
				}, targetSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(string(targetSecret.Data["username"])).To(Equal("shared-user"))
		})
	})

	Context("Migration Scenario", func() {
		It("should auto-reconcile when component secrets are added to existing cluster", func() {
			ctx := context.Background()

			// Setup: Start with shared credential
			sharedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereCloudCredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
					Annotations: map[string]string{
						constants.AnnotationKey: constants.PassthroughAnnotation,
					},
				},
				Data: map[string][]byte{
					"username": []byte("shared-user"),
					"password": []byte("shared-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, sharedSecret)).Should(Succeed())

			targetNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "openshift-machine-api",
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).Should(Succeed())

			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "migration-cr",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Name:      "migration-target-creds",
						Namespace: "openshift-machine-api",
					},
				},
			}
			Expect(k8sClient.Create(ctx, cr)).Should(Succeed())

			// Verify using shared credential
			targetSecret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "migration-target-creds",
					Namespace: "openshift-machine-api",
				}, targetSecret)
				return err == nil && string(targetSecret.Data["username"]) == "shared-user"
			}, timeout, interval).Should(BeTrue())

			// Migration: Add component secret
			componentSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereMachineAPICredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
				},
				Data: map[string][]byte{
					"username": []byte("machine-api-user"),
					"password": []byte("machine-api-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, componentSecret)).Should(Succeed())

			// Trigger reconciliation by updating CR
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(cr), cr)).Should(Succeed())
			cr.Annotations = map[string]string{"migration": "true"}
			Expect(k8sClient.Update(ctx, cr)).Should(Succeed())

			// Verify target secret now uses component credential
			Eventually(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "migration-target-creds",
					Namespace: "openshift-machine-api",
				}, targetSecret)
				if err != nil {
					return ""
				}
				return string(targetSecret.Data["username"])
			}, timeout, interval).Should(Equal("machine-api-user"))
		})
	})

	Context("CredentialsRequest Status", func() {
		It("should update status to reflect provisioning success", func() {
			ctx := context.Background()

			componentSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.VSphereMachineAPICredSecretName,
					Namespace: constants.CloudCredSecretNamespace,
				},
				Data: map[string][]byte{
					"username": []byte("machine-api-user"),
					"password": []byte("machine-api-pass"),
				},
			}
			Expect(k8sClient.Create(ctx, componentSecret)).Should(Succeed())

			targetNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "openshift-machine-api",
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).Should(Succeed())

			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "status-test-cr",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Name:      "status-test-creds",
						Namespace: "openshift-machine-api",
					},
				},
			}
			Expect(k8sClient.Create(ctx, cr)).Should(Succeed())

			// Verify status is updated
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cr), cr)
				if err != nil {
					return false
				}
				// Check if Provisioned condition exists
				for _, condition := range cr.Status.Conditions {
					if condition.Type == minterv1.CredentialsProvisionedConditionType {
						return condition.Status == corev1.ConditionTrue
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})
})
