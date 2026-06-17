/*
Copyright 2026 The OpenShift Authors.

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
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractVCenterCredentials_Success(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-ns",
		},
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("user1@vsphere.local"),
			"vcenter1.example.com.password": []byte("password1"),
			"vcenter2.example.com.username": []byte("user2@vsphere.local"),
			"vcenter2.example.com.password": []byte("password2"),
		},
	}

	username, password, err := extractVCenterCredentials(secret, "vcenter1.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if username != "user1@vsphere.local" {
		t.Errorf("expected username user1@vsphere.local, got %s", username)
	}

	if password != "password1" {
		t.Errorf("expected password password1, got %s", password)
	}
}

func TestExtractVCenterCredentials_MissingCredentials(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-ns",
		},
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("user1@vsphere.local"),
			"vcenter1.example.com.password": []byte("password1"),
		},
	}

	_, _, err := extractVCenterCredentials(secret, "vcenter2.example.com")
	if err == nil {
		t.Error("expected error for missing credentials")
	}

	if !strings.Contains(err.Error(), "vcenter2.example.com") {
		t.Errorf("error message should mention vcenter2.example.com: %v", err)
	}
}

func TestValidatePrivilegesPerVCenter_AllValid(t *testing.T) {
	secret := &corev1.Secret{
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("user1@vsphere.local"),
			"vcenter1.example.com.password": []byte("password1"),
			"vcenter2.example.com.username": []byte("user2@vsphere.local"),
			"vcenter2.example.com.password": []byte("password2"),
		},
	}

	topology := &VCenterTopology{
		VCenters: []string{"vcenter1.example.com", "vcenter2.example.com"},
	}

	requiredPrivileges := []string{
		"VirtualMachine.Inventory.Create",
		"Datastore.AllocateSpace",
	}

	result, err := validatePrivilegesPerVCenter(context.TODO(), secret, topology, requiredPrivileges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.AllValid {
		t.Error("expected all vCenters to be valid")
	}

	if result.ErrorCount != 0 {
		t.Errorf("expected 0 errors, got %d", result.ErrorCount)
	}

	if len(result.Results) != 2 {
		t.Errorf("expected results for 2 vCenters, got %d", len(result.Results))
	}
}

func TestValidatePrivilegesPerVCenter_MissingCredentials(t *testing.T) {
	secret := &corev1.Secret{
		Data: map[string][]byte{
			"vcenter1.example.com.username": []byte("user1@vsphere.local"),
			"vcenter1.example.com.password": []byte("password1"),
		},
	}

	topology := &VCenterTopology{
		VCenters: []string{"vcenter1.example.com", "vcenter2.example.com"},
	}

	requiredPrivileges := []string{"VirtualMachine.Inventory.Create"}

	result, err := validatePrivilegesPerVCenter(context.TODO(), secret, topology, requiredPrivileges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AllValid {
		t.Error("expected validation to fail for vcenter2")
	}

	if result.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount)
	}

	vcenter2Result, exists := result.Results["vcenter2.example.com"]
	if !exists {
		t.Fatal("expected result for vcenter2.example.com")
	}

	if vcenter2Result.Valid {
		t.Error("expected vcenter2 to be invalid")
	}

	if !strings.Contains(vcenter2Result.ErrorMessage, "vcenter2.example.com") {
		t.Errorf("error message should mention vcenter2.example.com: %s", vcenter2Result.ErrorMessage)
	}
}

func TestFormatPerVCenterError_AllValid(t *testing.T) {
	result := &MultiVCenterValidationResult{
		Results: map[string]*PerVCenterValidationResult{
			"vcenter1.example.com": {
				VCenter: "vcenter1.example.com",
				Valid:   true,
			},
		},
		AllValid:   true,
		ErrorCount: 0,
	}

	errorMsg := formatPerVCenterError(result)
	if errorMsg != "" {
		t.Errorf("expected empty error message for valid result, got: %s", errorMsg)
	}
}

func TestFormatPerVCenterError_WithErrors(t *testing.T) {
	result := &MultiVCenterValidationResult{
		Results: map[string]*PerVCenterValidationResult{
			"vcenter1.example.com": {
				VCenter: "vcenter1.example.com",
				Valid:   true,
			},
			"vcenter2.example.com": {
				VCenter:           "vcenter2.example.com",
				Valid:             false,
				MissingPrivileges: []string{"VirtualMachine.Inventory.Create", "Datastore.AllocateSpace"},
				ErrorMessage:      "insufficient privileges",
			},
		},
		AllValid:   false,
		ErrorCount: 1,
	}

	errorMsg := formatPerVCenterError(result)

	// Verify error message contains expected information
	expectedStrings := []string{
		"Multi-vCenter credential validation failed",
		"vcenter2.example.com",
		"VirtualMachine.Inventory.Create",
		"Datastore.AllocateSpace",
		"Remediation",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(errorMsg, expected) {
			t.Errorf("error message should contain '%s', got: %s", expected, errorMsg)
		}
	}

	// Verify vcenter1 is NOT mentioned (it's valid)
	if strings.Contains(errorMsg, "vcenter1.example.com") {
		t.Error("error message should not mention valid vCenter vcenter1.example.com")
	}
}

func TestFormatPerVCenterError_MultipleVCentersWithErrors(t *testing.T) {
	result := &MultiVCenterValidationResult{
		Results: map[string]*PerVCenterValidationResult{
			"vcenter1.example.com": {
				VCenter:           "vcenter1.example.com",
				Valid:             false,
				MissingPrivileges: []string{"VirtualMachine.Inventory.Create"},
				ErrorMessage:      "missing privilege",
			},
			"vcenter2.example.com": {
				VCenter:           "vcenter2.example.com",
				Valid:             false,
				MissingPrivileges: []string{"Datastore.AllocateSpace"},
				ErrorMessage:      "missing privilege",
			},
		},
		AllValid:   false,
		ErrorCount: 2,
	}

	errorMsg := formatPerVCenterError(result)

	// Verify both vCenters are mentioned
	if !strings.Contains(errorMsg, "vcenter1.example.com") {
		t.Error("error message should mention vcenter1.example.com")
	}

	if !strings.Contains(errorMsg, "vcenter2.example.com") {
		t.Error("error message should mention vcenter2.example.com")
	}

	// Verify specific privileges are mentioned for each
	if !strings.Contains(errorMsg, "VirtualMachine.Inventory.Create") {
		t.Error("error message should mention VirtualMachine.Inventory.Create")
	}

	if !strings.Contains(errorMsg, "Datastore.AllocateSpace") {
		t.Error("error message should mention Datastore.AllocateSpace")
	}
}
