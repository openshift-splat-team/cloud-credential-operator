package actuator

import (
	"encoding/base64"
	"fmt"
)

// AccountCredentials represents credentials for a single component account with optional vCenter override.
// This type is used for both the installer and CCO to support multi-vCenter topologies.
type AccountCredentials struct {
	Username string
	Password string
	VCenter  string // Optional: override default vCenter
}

// isMultiVCenterMode determines if the configuration uses multi-vCenter topology.
// Multi-vCenter mode is detected when any component specifies a vCenter override.
// In multi-vCenter mode, secrets use FQDN-keyed format (vcenter1.example.com.username)
// instead of simple keys (username/password).
//
// Note: AccountCredentials type is defined in actuator_test.go from Story #5.
func isMultiVCenterMode(componentCreds map[string]*AccountCredentials) bool {
	for _, cred := range componentCreds {
		if cred != nil && cred.VCenter != "" {
			return true
		}
	}
	return false
}

// createComponentSecrets generates all component-specific secrets with appropriate
// credential format based on multi-vCenter detection.
//
// This function implements the standalone secret generation logic. It can be used
// independently or wrapped by the VSphereActuator's CreateComponentSecrets method.
func createComponentSecrets(componentCreds map[string]*AccountCredentials) (map[string]*Secret, error) {
	secrets := make(map[string]*Secret)

	// Detect multi-vCenter mode
	multiVCenter := isMultiVCenterMode(componentCreds)

	// Define component secret mappings
	componentSecretMap := map[string]struct {
		namespace string
		name      string
	}{
		"machineAPI": {
			namespace: "openshift-machine-api",
			name:      "machine-api-vsphere-credentials",
		},
		"csiDriver": {
			namespace: "openshift-cluster-csi-drivers",
			name:      "vsphere-csi-credentials",
		},
		"cloudController": {
			namespace: "openshift-cloud-controller-manager",
			name:      "vsphere-ccm-credentials",
		},
		"diagnostics": {
			namespace: "openshift-config",
			name:      "vsphere-diagnostics-credentials",
		},
	}

	// Create secret for each component
	for componentName, secretConfig := range componentSecretMap {
		cred := componentCreds[componentName]
		if cred == nil {
			continue
		}

		secretData := make(map[string][]byte)

		if multiVCenter && cred.VCenter != "" {
			// Multi-vCenter mode: use FQDN-keyed credentials
			usernameKey := fmt.Sprintf("%s.username", cred.VCenter)
			passwordKey := fmt.Sprintf("%s.password", cred.VCenter)
			secretData[usernameKey] = []byte(base64.StdEncoding.EncodeToString([]byte(cred.Username)))
			secretData[passwordKey] = []byte(base64.StdEncoding.EncodeToString([]byte(cred.Password)))
		} else {
			// Single-vCenter mode: use simple keys
			secretData["username"] = []byte(base64.StdEncoding.EncodeToString([]byte(cred.Username)))
			secretData["password"] = []byte(base64.StdEncoding.EncodeToString([]byte(cred.Password)))
		}

		secret := &Secret{
			Data: secretData,
		}
		secret.Name = secretConfig.name
		secret.Namespace = secretConfig.namespace

		secrets[componentName] = secret
	}

	return secrets, nil
}

// Secret represents a simplified Kubernetes secret for testing.
// In production code, this would be corev1.Secret from k8s.io/api/core/v1.
type Secret struct {
	Name      string
	Namespace string
	Data      map[string][]byte
}

// getSecretKeyFormat returns the expected secret key format for a given vCenter.
// In multi-vCenter mode, returns FQDN-prefixed keys. Otherwise, returns simple keys.
func getSecretKeyFormat(vcenterFQDN string, multiVCenterMode bool) (usernameKey, passwordKey string) {
	if multiVCenterMode && vcenterFQDN != "" {
		return fmt.Sprintf("%s.username", vcenterFQDN), fmt.Sprintf("%s.password", vcenterFQDN)
	}
	return "username", "password"
}
