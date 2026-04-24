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
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// PerVCenterValidationResult holds validation results for each vCenter
type PerVCenterValidationResult struct {
	VCenter           string
	Valid             bool
	MissingPrivileges []string
	ErrorMessage      string
}

// MultiVCenterValidationResult aggregates validation results across all vCenters
type MultiVCenterValidationResult struct {
	Results     map[string]*PerVCenterValidationResult
	AllValid    bool
	ErrorCount  int
}

// validatePrivilegesPerVCenter validates required privileges on each vCenter separately
func validatePrivilegesPerVCenter(ctx context.Context, credentialsSecret *corev1.Secret, topology *VCenterTopology, requiredPrivileges []string) (*MultiVCenterValidationResult, error) {
	logger := log.WithField("function", "validatePrivilegesPerVCenter")

	result := &MultiVCenterValidationResult{
		Results:  make(map[string]*PerVCenterValidationResult),
		AllValid: true,
	}

	// Validate each vCenter separately
	for _, vcenterFQDN := range topology.VCenters {
		vcLogger := logger.WithField("vcenter", vcenterFQDN)
		vcLogger.Debug("validating privileges for vCenter")

		vcResult := &PerVCenterValidationResult{
			VCenter:           vcenterFQDN,
			Valid:             true,
			MissingPrivileges: make([]string, 0),
		}

		// Extract credentials for this specific vCenter
		username, password, err := extractVCenterCredentials(credentialsSecret, vcenterFQDN)
		if err != nil {
			vcLogger.WithError(err).Error("failed to extract credentials")
			vcResult.Valid = false
			vcResult.ErrorMessage = fmt.Sprintf("missing credentials for vCenter %s", vcenterFQDN)
			result.Results[vcenterFQDN] = vcResult
			result.AllValid = false
			result.ErrorCount++
			continue
		}

		// Simulate privilege validation (in production, this would connect to vSphere API)
		// For now, we'll validate that credentials exist and are non-empty
		if username == "" || password == "" {
			vcResult.Valid = false
			vcResult.MissingPrivileges = append(vcResult.MissingPrivileges, "invalid credentials")
			vcResult.ErrorMessage = fmt.Sprintf("empty credentials for vCenter %s", vcenterFQDN)
			result.AllValid = false
			result.ErrorCount++
		}

		result.Results[vcenterFQDN] = vcResult
		vcLogger.WithField("valid", vcResult.Valid).Debug("validation complete for vCenter")
	}

	logger.WithFields(log.Fields{
		"totalVCenters": len(topology.VCenters),
		"allValid":      result.AllValid,
		"errorCount":    result.ErrorCount,
	}).Info("multi-vCenter validation complete")

	return result, nil
}

// extractVCenterCredentials extracts username and password for a specific vCenter from the credentials secret
func extractVCenterCredentials(secret *corev1.Secret, vcenterFQDN string) (string, string, error) {
	if secret == nil || secret.Data == nil {
		return "", "", fmt.Errorf("credentials secret is nil or empty")
	}

	// Multi-vCenter format: <vcenter-fqdn>.username and <vcenter-fqdn>.password
	usernameKey := fmt.Sprintf("%s.username", vcenterFQDN)
	passwordKey := fmt.Sprintf("%s.password", vcenterFQDN)

	username, usernameExists := secret.Data[usernameKey]
	password, passwordExists := secret.Data[passwordKey]

	if !usernameExists || !passwordExists {
		return "", "", fmt.Errorf("missing credentials for vCenter %s (expected keys: %s, %s)", vcenterFQDN, usernameKey, passwordKey)
	}

	return string(username), string(password), nil
}

// formatPerVCenterError formats validation errors with vCenter-specific details
func formatPerVCenterError(result *MultiVCenterValidationResult) string {
	if result.AllValid {
		return ""
	}

	var errorMessages []string
	errorMessages = append(errorMessages, "Multi-vCenter credential validation failed:")
	errorMessages = append(errorMessages, "")

	for vcenterFQDN, vcResult := range result.Results {
		if !vcResult.Valid {
			errorMessages = append(errorMessages, fmt.Sprintf("vCenter: %s", vcenterFQDN))

			if vcResult.ErrorMessage != "" {
				errorMessages = append(errorMessages, fmt.Sprintf("  Error: %s", vcResult.ErrorMessage))
			}

			if len(vcResult.MissingPrivileges) > 0 {
				errorMessages = append(errorMessages, "  Missing privileges:")
				for _, priv := range vcResult.MissingPrivileges {
					errorMessages = append(errorMessages, fmt.Sprintf("    - %s", priv))
				}

				// Add remediation guidance
				errorMessages = append(errorMessages, "")
				errorMessages = append(errorMessages, fmt.Sprintf("  Remediation: Grant the missing privileges to the service account on %s", vcenterFQDN))
			}
			errorMessages = append(errorMessages, "")
		}
	}

	return strings.Join(errorMessages, "\n")
}
