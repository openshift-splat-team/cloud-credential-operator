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
package actuator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	minterv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	"github.com/openshift/cloud-credential-operator/pkg/operator/constants"
)

func TestGetComponentSecretName(t *testing.T) {
	tests := []struct {
		name              string
		targetNamespace   string
		expectedSecretName string
	}{
		{
			name:              "machine-api namespace",
			targetNamespace:   "openshift-machine-api",
			expectedSecretName: constants.VSphereMachineAPICredSecretName,
		},
		{
			name:              "storage namespace",
			targetNamespace:   "openshift-cluster-csi-drivers",
			expectedSecretName: constants.VSphereStorageCredSecretName,
		},
		{
			name:              "cloud-controller namespace",
			targetNamespace:   "openshift-cloud-controller-manager",
			expectedSecretName: constants.VSphereCloudControllerCredSecretName,
		},
		{
			name:              "diagnostics namespace",
			targetNamespace:   "openshift-config",
			expectedSecretName: constants.VSphereDiagnosticsCredSecretName,
		},
		{
			name:              "unknown namespace",
			targetNamespace:   "openshift-other",
			expectedSecretName: "",
		},
	}

	actuator := &VSphereActuator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &minterv1.CredentialsRequest{
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Namespace: tt.targetNamespace,
						Name:      "test-secret",
					},
				},
			}

			result := actuator.getComponentSecretName(cr)
			assert.Equal(t, tt.expectedSecretName, result)
		})
	}
}

func TestGetCredentialsRootSecret_ComponentSpecific(t *testing.T) {
	minterv1.AddToScheme(scheme.Scheme)

	componentSecretData := map[string][]byte{
		"username": []byte("component-user"),
		"password": []byte("component-pass"),
	}

	sharedSecretData := map[string][]byte{
		"username": []byte("shared-user"),
		"password": []byte("shared-pass"),
	}

	tests := []struct {
		name                    string
		targetNamespace         string
		componentSecretExists   bool
		sharedSecretExists      bool
		expectedSecretData      map[string][]byte
		expectError             bool
	}{
		{
			name:                  "component secret exists - machine-api",
			targetNamespace:       "openshift-machine-api",
			componentSecretExists: true,
			sharedSecretExists:    true,
			expectedSecretData:    componentSecretData,
			expectError:           false,
		},
		{
			name:                  "component secret missing - fallback to shared",
			targetNamespace:       "openshift-machine-api",
			componentSecretExists: false,
			sharedSecretExists:    true,
			expectedSecretData:    sharedSecretData,
			expectError:           false,
		},
		{
			name:                  "both secrets missing - error",
			targetNamespace:       "openshift-machine-api",
			componentSecretExists: false,
			sharedSecretExists:    false,
			expectedSecretData:    nil,
			expectError:           true,
		},
		{
			name:                  "unknown namespace - uses shared secret",
			targetNamespace:       "openshift-other",
			componentSecretExists: false,
			sharedSecretExists:    true,
			expectedSecretData:    sharedSecretData,
			expectError:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}

			// Create component-specific secret if needed
			if tt.componentSecretExists {
				componentSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      constants.VSphereMachineAPICredSecretName,
						Namespace: constants.CloudCredSecretNamespace,
					},
					Data: componentSecretData,
				}
				objects = append(objects, componentSecret)
			}

			// Create shared secret if needed
			if tt.sharedSecretExists {
				sharedSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      constants.VSphereCloudCredSecretName,
						Namespace: constants.CloudCredSecretNamespace,
						Annotations: map[string]string{
							constants.AnnotationKey: constants.PassthroughAnnotation,
						},
					},
					Data: sharedSecretData,
				}
				objects = append(objects, sharedSecret)
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()

			actuator := &VSphereActuator{
				Client:         fakeClient,
				RootCredClient: fakeClient,
			}

			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cr",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Namespace: tt.targetNamespace,
						Name:      "test-target-secret",
					},
				},
			}

			secret, err := actuator.GetCredentialsRootSecret(context.TODO(), cr)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, secret)
			assert.Equal(t, tt.expectedSecretData, secret.Data)
		})
	}
}

func TestGetCredentialsRootSecret_MultiComponent(t *testing.T) {
	minterv1.AddToScheme(scheme.Scheme)

	// Create all component-specific secrets
	machineAPISecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.VSphereMachineAPICredSecretName,
			Namespace: constants.CloudCredSecretNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("machine-api-user"),
			"password": []byte("machine-api-pass"),
		},
	}

	storageSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.VSphereStorageCredSecretName,
			Namespace: constants.CloudCredSecretNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("storage-user"),
			"password": []byte("storage-pass"),
		},
	}

	cloudControllerSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.VSphereCloudControllerCredSecretName,
			Namespace: constants.CloudCredSecretNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("cloud-controller-user"),
			"password": []byte("cloud-controller-pass"),
		},
	}

	diagnosticsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.VSphereDiagnosticsCredSecretName,
			Namespace: constants.CloudCredSecretNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("diagnostics-user"),
			"password": []byte("diagnostics-pass"),
		},
	}

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

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(machineAPISecret, storageSecret, cloudControllerSecret, diagnosticsSecret, sharedSecret).
		Build()

	actuator := &VSphereActuator{
		Client:         fakeClient,
		RootCredClient: fakeClient,
	}

	tests := []struct {
		targetNamespace    string
		expectedUsername   string
	}{
		{
			targetNamespace:  "openshift-machine-api",
			expectedUsername: "machine-api-user",
		},
		{
			targetNamespace:  "openshift-cluster-csi-drivers",
			expectedUsername: "storage-user",
		},
		{
			targetNamespace:  "openshift-cloud-controller-manager",
			expectedUsername: "cloud-controller-user",
		},
		{
			targetNamespace:  "openshift-config",
			expectedUsername: "diagnostics-user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.targetNamespace, func(t *testing.T) {
			cr := &minterv1.CredentialsRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cr",
					Namespace: "openshift-cloud-credential-operator",
				},
				Spec: minterv1.CredentialsRequestSpec{
					SecretRef: corev1.ObjectReference{
						Namespace: tt.targetNamespace,
						Name:      "test-target-secret",
					},
				},
			}

			secret, err := actuator.GetCredentialsRootSecret(context.TODO(), cr)
			require.NoError(t, err)
			require.NotNil(t, secret)
			assert.Equal(t, tt.expectedUsername, string(secret.Data["username"]))
		})
	}
}
