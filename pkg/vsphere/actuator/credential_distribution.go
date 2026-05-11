package actuator

import (
	"context"
	"fmt"

	minterv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	corev1 "k8s.io/api/core/v1"
)

const vsphereComponentAnnotationKey = "cloudcredential.openshift.io/vsphere-component"

// ComponentSecretReader looks up a named secret from kube-system.
type ComponentSecretReader interface {
	GetComponentSecret(ctx context.Context, name string) (*corev1.Secret, error)
}

// componentSecretName maps an annotation value to the kube-system secret name.
// Returns "" for unrecognized components.
func componentSecretName(component string) string {
	switch component {
	case "machineAPI":
		return "vsphere-machine-api-creds"
	case "csiDriver":
		return "vsphere-storage-creds"
	case "cloudController":
		return "vsphere-cloud-controller-creds"
	case "vsphereProblemDetector":
		return "vsphere-problem-detector-creds"
	default:
		return ""
	}
}

// resolveVSphereCredentials routes a CredentialsRequest to per-component or shared credentials.
//
// If the CR has a vsphere-component annotation and a matching kube-system secret with non-empty
// data, the per-component credential is returned. If the component secret is absent or empty,
// shared credentials are returned and a warning is set. A CR without the annotation is a
// passthrough — shared credentials, no warning.
func resolveVSphereCredentials(
	ctx context.Context,
	cr *minterv1.CredentialsRequest,
	componentSecretReader ComponentSecretReader,
	sharedSecretData map[string][]byte,
	vcenters []string,
) (secretData map[string][]byte, warning string, err error) {
	component := cr.Annotations[vsphereComponentAnnotationKey]
	if component == "" {
		return sharedSecretData, "", nil
	}

	secretName := componentSecretName(component)
	if secretName != "" {
		secret, err := componentSecretReader.GetComponentSecret(ctx, secretName)
		if err != nil {
			return nil, "", err
		}

		if secret != nil && len(secret.Data) > 0 {
			// Verify credentials exist for every cluster vCenter before returning.
			for _, vcenter := range vcenters {
				if _, ok := secret.Data[vcenter+".username"]; !ok {
					return nil, "", fmt.Errorf("component secret %q missing credentials for vCenter %q", secretName, vcenter)
				}
			}

			return secret.Data, "", nil
		}
	}

	// No component secret (or unrecognized component): fall back to shared credentials.
	return sharedSecretData, fmt.Sprintf("Component '%s' using fallback shared credentials", component), nil
}
