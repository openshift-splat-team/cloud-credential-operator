// credential_distribution_test.go tests Story #38:
// CCO Per-Component Credential Distribution with Graceful Fallback.
//
// Build guard prevents compile errors until the implementation exists.
// Remove //go:build ignore once credential_distribution.go is implemented.
//
// Implementation contract expected by these tests:
//
//	func resolveVSphereCredentials(
//	    ctx context.Context,
//	    cr *minterv1.CredentialsRequest,
//	    componentSecretReader ComponentSecretReader,
//	    sharedSecretData map[string][]byte,
//	    checker PrivilegeChecker,
//	    vcenters []string,
//	) (secretData map[string][]byte, warning string, err error)
//
// ComponentSecretReader is an interface (or func) that looks up a named secret from kube-system.
// The component annotation key is: cloudcredential.openshift.io/vsphere-component
// Annotation values: machineAPI, csiDriver, cloudController, vsphereProblemDetector
// Component secret names: vsphere-machine-api-creds, vsphere-storage-creds,
//
//	vsphere-cloud-controller-creds, vsphere-problem-detector-creds
//
// Per-component secret key format: {vcenter.fqdn}.username / {vcenter.fqdn}.password
// Warning format: Component '{component}' using fallback shared credentials
package actuator

import (
	"context"
	"fmt"
	"testing"

	minterv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// vsphereComponentAnnotation is the annotation key used to identify per-component CredentialsRequests.
const vsphereComponentAnnotation = "cloudcredential.openshift.io/vsphere-component"

// stubComponentSecretReader is a test double for looking up kube-system component secrets.
type stubComponentSecretReader struct {
	secrets map[string]*corev1.Secret // secretName → secret (nil means NotFound)
}

func (s *stubComponentSecretReader) GetComponentSecret(_ context.Context, name string) (*corev1.Secret, error) {
	if secret, ok := s.secrets[name]; ok {
		return secret, nil // nil value means secret was explicitly absent
	}
	return nil, nil // not configured = absent
}

// makeCredCR builds a minimal CredentialsRequest with the given component annotation value.
func makeCredCR(component string) *minterv1.CredentialsRequest {
	cr := &minterv1.CredentialsRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cr",
			Namespace: "openshift-cloud-credential-operator",
		},
	}
	if component != "" {
		cr.Annotations = map[string]string{vsphereComponentAnnotation: component}
	}
	return cr
}

// makeComponentSecret builds a kube-system Secret with the given vCenter credentials.
func makeComponentSecret(name string, creds map[string]string) *corev1.Secret {
	data := make(map[string][]byte, len(creds))
	for k, v := range creds {
		data[k] = []byte(v)
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "kube-system"},
		Data:       data,
	}
}

var sharedSecretData = map[string][]byte{
	"vcenter.example.com.username": []byte("shared@vsphere.local"),
	"vcenter.example.com.password": []byte("shared-password"),
}

// ── AC1: Per-component routing and shared-credential fallback ─────────────────

// TestResolveComponentCredential_AnnotatedCRWithSecret_UsesPerComponentCred verifies AC1 (first
// branch): a CredentialsRequest annotated machineAPI, whose component secret exists, receives the
// per-component credential — not the shared credential.
func TestResolveComponentCredential_AnnotatedCRWithSecret_UsesPerComponentCred(t *testing.T) {
	cr := makeCredCR("machineAPI")

	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter.example.com.username": "machine-api@vsphere.local",
		"vcenter.example.com.password": "machine-api-password",
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{
			"vsphere-machine-api-creds": componentSecret,
		},
	}

	checker := newMock()
	setPrivileges(checker, "machine-api@vsphere.local", "vcenter.example.com", machineAPIPrivileges)

	data, warning, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if warning != "" {
		t.Errorf("expected no warning for annotated CR with existing component secret, got: %q", warning)
	}
	if string(data["vcenter.example.com.username"]) != "machine-api@vsphere.local" {
		t.Errorf("expected per-component username, got: %q", data["vcenter.example.com.username"])
	}
}

// TestResolveComponentCredential_AnnotatedCRWithoutSecret_FallsBackToShared verifies AC1 (fallback
// branch): a CredentialsRequest annotated csiDriver, with no component secret present, receives the
// shared credential and a warning is returned.
func TestResolveComponentCredential_AnnotatedCRWithoutSecret_FallsBackToShared(t *testing.T) {
	cr := makeCredCR("csiDriver")

	// No component secret registered for csiDriver
	reader := &stubComponentSecretReader{secrets: map[string]*corev1.Secret{}}

	checker := newMock()
	// Shared credential checker — should be queried for shared credential validation (if any)

	data, warning, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("expected no error during fallback, got: %v", err)
	}
	expectedWarning := "Component 'csiDriver' using fallback shared credentials"
	if warning != expectedWarning {
		t.Errorf("expected warning %q, got %q", expectedWarning, warning)
	}
	if string(data["vcenter.example.com.username"]) != "shared@vsphere.local" {
		t.Errorf("expected shared username in fallback data, got: %q", data["vcenter.example.com.username"])
	}
}

// ── AC2: Multi-vCenter credential distribution ────────────────────────────────

// TestResolveComponentCredential_MultiVCenter_BothVCentersPopulated verifies AC2: when a
// component secret contains credentials for two vCenters, both sets of keyed entries appear
// in the provisioned secret data.
func TestResolveComponentCredential_MultiVCenter_BothVCentersPopulated(t *testing.T) {
	cr := makeCredCR("machineAPI")

	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter1.example.com.username": "machine-api@vc1.local",
		"vcenter1.example.com.password": "password1",
		"vcenter2.example.com.username": "machine-api@vc2.local",
		"vcenter2.example.com.password": "password2",
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": componentSecret},
	}

	checker := newMock()
	setPrivileges(checker, "machine-api@vc1.local", "vcenter1.example.com", machineAPIPrivileges)
	setPrivileges(checker, "machine-api@vc2.local", "vcenter2.example.com", machineAPIPrivileges)

	data, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter1.example.com", "vcenter2.example.com"},
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(data["vcenter1.example.com.username"]) != "machine-api@vc1.local" {
		t.Errorf("vcenter1 username missing or wrong: %q", data["vcenter1.example.com.username"])
	}
	if string(data["vcenter2.example.com.username"]) != "machine-api@vc2.local" {
		t.Errorf("vcenter2 username missing or wrong: %q", data["vcenter2.example.com.username"])
	}
}

// ── AC3: Privilege validation invoked before provisioning ────────────────────

// TestResolveComponentCredential_PrivilegeValidationInvoked verifies AC3: the Story #37
// ValidateComponentPrivileges call is made before the secret data is returned. A mock checker
// that records calls is used; provisioning succeeds only when the checker is exercised.
func TestResolveComponentCredential_PrivilegeValidationInvoked(t *testing.T) {
	cr := makeCredCR("machineAPI")

	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter.example.com.username": "machine-api@vsphere.local",
		"vcenter.example.com.password": "machine-api-password",
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": componentSecret},
	}

	called := false
	recordingChecker := &recordingPrivilegeChecker{
		inner:   newMock(),
		onCalled: func() { called = true },
	}
	setPrivileges(recordingChecker.inner, "machine-api@vsphere.local", "vcenter.example.com", machineAPIPrivileges)

	_, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, recordingChecker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !called {
		t.Error("expected PrivilegeChecker.FetchUserPrivileges to be called (AC3 requires validation before provisioning)")
	}
}

// TestResolveComponentCredential_PrivilegeValidationFails_SecretNotProvisioned verifies AC3 (failure
// path): if privilege validation fails, an error is returned and no secret data is returned.
func TestResolveComponentCredential_PrivilegeValidationFails_SecretNotProvisioned(t *testing.T) {
	cr := makeCredCR("machineAPI")

	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter.example.com.username": "machine-api@vsphere.local",
		"vcenter.example.com.password": "machine-api-password",
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": componentSecret},
	}

	// Checker grants only a subset of required privileges → validation fails
	checker := newMock()
	setPrivileges(checker, "machine-api@vsphere.local", "vcenter.example.com",
		allExcept(machineAPIPrivileges, "VirtualMachine.Inventory.Create", "VirtualMachine.Inventory.Delete"))

	data, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if len(data) > 0 {
		t.Errorf("expected no secret data when validation fails, got: %v", data)
	}
}

// ── Adversarial cases ─────────────────────────────────────────────────────────

// TestResolveComponentCredential_NoAnnotation_Passthrough verifies that a CredentialsRequest
// with no vsphere-component annotation is not routed to per-component logic (passthrough).
func TestResolveComponentCredential_NoAnnotation_Passthrough(t *testing.T) {
	cr := makeCredCR("") // no annotation

	reader := &stubComponentSecretReader{secrets: map[string]*corev1.Secret{}}
	checker := newMock()

	data, warning, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("expected no error for passthrough CR, got: %v", err)
	}
	if warning != "" {
		t.Errorf("expected no warning for passthrough CR, got: %q", warning)
	}
	// Passthrough returns the shared secret data unchanged
	if string(data["vcenter.example.com.username"]) != "shared@vsphere.local" {
		t.Errorf("expected shared credential in passthrough, got: %q", data["vcenter.example.com.username"])
	}
}

// TestResolveComponentCredential_NilAnnotations_Passthrough is an edge case ensuring nil
// annotations map (not just absent key) is handled without panic.
func TestResolveComponentCredential_NilAnnotations_Passthrough(t *testing.T) {
	cr := &minterv1.CredentialsRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-cr",
			Namespace:   "openshift-cloud-credential-operator",
			Annotations: nil, // explicitly nil
		},
	}

	reader := &stubComponentSecretReader{secrets: map[string]*corev1.Secret{}}
	checker := newMock()

	_, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("nil annotations must not panic or error, got: %v", err)
	}
}

// TestResolveComponentCredential_UnknownComponentAnnotation_FallsBackToShared verifies that an
// unrecognized annotation value (e.g. "diagnostics") does not panic and falls back to shared.
func TestResolveComponentCredential_UnknownComponentAnnotation_FallsBackToShared(t *testing.T) {
	cr := makeCredCR("diagnostics") // not a known component

	reader := &stubComponentSecretReader{secrets: map[string]*corev1.Secret{}}
	checker := newMock()

	_, warning, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("unknown component should fall back gracefully, got error: %v", err)
	}
	if warning == "" {
		t.Error("expected a fallback warning for unknown component")
	}
}

// TestResolveComponentCredential_MultiVCenter_MissingOneVCenterEntry_Error verifies that when a
// component secret is present but lacks credentials for one of the cluster's vCenters, an error
// is returned (not silent omission — the operator would fail to connect to that vCenter).
func TestResolveComponentCredential_MultiVCenter_MissingOneVCenterEntry_Error(t *testing.T) {
	cr := makeCredCR("machineAPI")

	// Secret only has vcenter1; cluster has vcenter1 + vcenter2
	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter1.example.com.username": "machine-api@vc1.local",
		"vcenter1.example.com.password": "password1",
		// vcenter2 entries absent
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": componentSecret},
	}

	checker := newMock()
	setPrivileges(checker, "machine-api@vc1.local", "vcenter1.example.com", machineAPIPrivileges)

	_, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter1.example.com", "vcenter2.example.com"},
	)

	if err == nil {
		t.Error("expected error when component secret is missing credentials for a cluster vCenter")
	}
}

// TestResolveComponentCredential_AuthFailureDuringValidation_Error verifies that an authentication
// failure from the PrivilegeChecker is surfaced as an error (and not silently treated as
// insufficient privileges — the error type matters for operator status messages).
func TestResolveComponentCredential_AuthFailureDuringValidation_Error(t *testing.T) {
	cr := makeCredCR("machineAPI")

	componentSecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{
		"vcenter.example.com.username": "machine-api@vsphere.local",
		"vcenter.example.com.password": "wrong-password",
	})

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": componentSecret},
	}

	checker := newMock()
	setAuthError(checker, "machine-api@vsphere.local", "vcenter.example.com")

	_, _, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err == nil {
		t.Fatal("expected error for authentication failure")
	}
	if !containsInsensitive(err.Error(), "authentication") {
		t.Errorf("expected authentication error message, got: %v", err)
	}
}

// TestResolveComponentCredential_EmptyComponentSecret_FallsBackToShared verifies that a component
// secret with no keys (empty Data) is treated equivalently to a missing secret — fallback occurs.
func TestResolveComponentCredential_EmptyComponentSecret_FallsBackToShared(t *testing.T) {
	cr := makeCredCR("machineAPI")

	emptySecret := makeComponentSecret("vsphere-machine-api-creds", map[string]string{}) // no keys

	reader := &stubComponentSecretReader{
		secrets: map[string]*corev1.Secret{"vsphere-machine-api-creds": emptySecret},
	}

	checker := newMock()

	_, warning, err := resolveVSphereCredentials(
		context.Background(), cr, reader, sharedSecretData, checker,
		[]string{"vcenter.example.com"},
	)

	if err != nil {
		t.Fatalf("empty component secret should fall back, not error: %v", err)
	}
	if warning == "" {
		t.Error("expected fallback warning for empty component secret")
	}
}

// ── Helpers used by the stubs above ──────────────────────────────────────────

// recordingPrivilegeChecker wraps an inner checker and calls onCalled on the first invocation.
type recordingPrivilegeChecker struct {
	inner    *mockPrivilegeChecker
	onCalled func()
}

func (r *recordingPrivilegeChecker) FetchUserPrivileges(ctx context.Context, username, vcenterFQDN string) ([]string, error) {
	if r.onCalled != nil {
		r.onCalled()
	}
	return r.inner.FetchUserPrivileges(ctx, username, vcenterFQDN)
}

func containsInsensitive(s, substr string) bool {
	return fmt.Sprintf("%s", s) != "" && len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsBytes([]byte(s), []byte(substr)))
}

func containsBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j, b := range needle {
			hb := haystack[i+j]
			if hb != b && hb != b-32 && hb != b+32 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
