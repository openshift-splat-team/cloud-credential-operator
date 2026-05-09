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

// componentPrivilegeSet maps an annotation value to the privilege-set name used by ValidateComponentPrivileges.
func componentPrivilegeSet(component string) string {
	switch component {
	case "machineAPI":
		return "machine-api"
	case "csiDriver":
		return "storage"
	case "cloudController":
		return "cloud-controller"
	case "vsphereProblemDetector":
		return "vsphere-problem-detector"
	default:
		return ""
	}
}

// resolveVSphereCredentials routes a CredentialsRequest to per-component or shared credentials.
//
// If the CR has a vsphere-component annotation and a matching kube-system secret with non-empty
// data, the per-component credential is validated (via Story #37 ValidateComponentPrivileges) and
// returned. If the component secret is absent or empty, shared credentials are returned and a
// warning is set. A CR without the annotation is a passthrough — shared credentials, no warning.
func resolveVSphereCredentials(
	ctx context.Context,
	cr *minterv1.CredentialsRequest,
	componentSecretReader ComponentSecretReader,
	sharedSecretData map[string][]byte,
	checker PrivilegeChecker,
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
			// Verify credentials exist for every cluster vCenter before validating.
			for _, vcenter := range vcenters {
				if _, ok := secret.Data[vcenter+".username"]; !ok {
					return nil, "", fmt.Errorf("component secret %q missing credentials for vCenter %q", secretName, vcenter)
				}
			}

			// Build a ComponentCredential per vCenter for privilege validation.
			creds := make([]ComponentCredential, 0, len(vcenters))
			privSet := componentPrivilegeSet(component)
			for _, vcenter := range vcenters {
				creds = append(creds, ComponentCredential{
					Component:   privSet,
					Username:    string(secret.Data[vcenter+".username"]),
					Password:    string(secret.Data[vcenter+".password"]),
					VCenterFQDN: vcenter,
				})
			}

			if err := ValidateComponentPrivileges(ctx, checker, creds); err != nil {
				return nil, "", err
			}

			return secret.Data, "", nil
		}
	}

	// No component secret (or unrecognized component): fall back to shared credentials.
	return sharedSecretData, fmt.Sprintf("Component '%s' using fallback shared credentials", component), nil
}
