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
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1 "github.com/openshift/api/config/v1"
)

// VCenterTopology holds the discovered vCenter topology for multi-vCenter deployments
type VCenterTopology struct {
	VCenters []string // List of vCenter FQDNs
}

// getVCenterTopology discovers the vCenter topology from the cluster infrastructure
func getVCenterTopology(ctx context.Context, c client.Client) (*VCenterTopology, error) {
	logger := log.WithField("function", "getVCenterTopology")

	infra := &configv1.Infrastructure{}
	if err := c.Get(ctx, types.NamespacedName{Name: "cluster"}, infra); err != nil {
		logger.WithError(err).Error("failed to get infrastructure")
		return nil, fmt.Errorf("failed to get infrastructure: %w", err)
	}

	if infra.Spec.PlatformSpec.Type != configv1.VSpherePlatformType {
		logger.WithField("platform", infra.Spec.PlatformSpec.Type).Debug("not a vSphere platform")
		return nil, fmt.Errorf("not a vSphere platform: %s", infra.Spec.PlatformSpec.Type)
	}

	if infra.Spec.PlatformSpec.VSphere == nil {
		logger.Debug("vSphere platform spec is nil")
		return nil, fmt.Errorf("vSphere platform spec is nil")
	}

	topology := &VCenterTopology{
		VCenters: make([]string, 0),
	}

	// Extract vCenter FQDNs from the infrastructure spec
	for _, vcenter := range infra.Spec.PlatformSpec.VSphere.VCenters {
		if vcenter.Server != "" {
			topology.VCenters = append(topology.VCenters, vcenter.Server)
			logger.WithField("vcenter", vcenter.Server).Debug("discovered vCenter")
		}
	}

	logger.WithField("vcenterCount", len(topology.VCenters)).Info("discovered vCenter topology")
	return topology, nil
}

// isMultiVCenter returns true if the cluster is configured with multiple vCenters
func (t *VCenterTopology) isMultiVCenter() bool {
	return len(t.VCenters) > 1
}
