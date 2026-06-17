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
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetVCenterTopology_SingleVCenter(t *testing.T) {
	scheme := runtime.NewScheme()
	configv1.Install(scheme)

	infra := &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.InfrastructureSpec{
			PlatformSpec: configv1.PlatformSpec{
				Type: configv1.VSpherePlatformType,
				VSphere: &configv1.VSpherePlatformSpec{
					VCenters: []configv1.VSpherePlatformVCenterSpec{
						{
							Server:      "vcenter1.example.com",
							Port:        443,
							Datacenters: []string{"dc1"},
						},
					},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(infra).Build()
	topology, err := getVCenterTopology(context.TODO(), client)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(topology.VCenters) != 1 {
		t.Errorf("expected 1 vCenter, got %d", len(topology.VCenters))
	}

	if topology.VCenters[0] != "vcenter1.example.com" {
		t.Errorf("expected vcenter1.example.com, got %s", topology.VCenters[0])
	}

	if topology.isMultiVCenter() {
		t.Error("single vCenter should not be reported as multi-vCenter")
	}
}

func TestGetVCenterTopology_MultiVCenter(t *testing.T) {
	scheme := runtime.NewScheme()
	configv1.Install(scheme)

	infra := &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.InfrastructureSpec{
			PlatformSpec: configv1.PlatformSpec{
				Type: configv1.VSpherePlatformType,
				VSphere: &configv1.VSpherePlatformSpec{
					VCenters: []configv1.VSpherePlatformVCenterSpec{
						{
							Server:      "vcenter1.example.com",
							Port:        443,
							Datacenters: []string{"dc1"},
						},
						{
							Server:      "vcenter2.example.com",
							Port:        443,
							Datacenters: []string{"dc2"},
						},
					},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(infra).Build()
	topology, err := getVCenterTopology(context.TODO(), client)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(topology.VCenters) != 2 {
		t.Errorf("expected 2 vCenters, got %d", len(topology.VCenters))
	}

	if !topology.isMultiVCenter() {
		t.Error("expected multi-vCenter deployment to be detected")
	}

	expectedVCenters := map[string]bool{
		"vcenter1.example.com": false,
		"vcenter2.example.com": false,
	}

	for _, vcenter := range topology.VCenters {
		if _, ok := expectedVCenters[vcenter]; ok {
			expectedVCenters[vcenter] = true
		} else {
			t.Errorf("unexpected vCenter: %s", vcenter)
		}
	}

	for vcenter, found := range expectedVCenters {
		if !found {
			t.Errorf("expected vCenter not found: %s", vcenter)
		}
	}
}

func TestGetVCenterTopology_NotVSpherePlatform(t *testing.T) {
	scheme := runtime.NewScheme()
	configv1.Install(scheme)

	infra := &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.InfrastructureSpec{
			PlatformSpec: configv1.PlatformSpec{
				Type: configv1.AWSPlatformType,
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(infra).Build()
	_, err := getVCenterTopology(context.TODO(), client)

	if err == nil {
		t.Error("expected error for non-vSphere platform")
	}
}
