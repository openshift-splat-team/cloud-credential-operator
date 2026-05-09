package actuator

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// mockPrivilegeChecker implements PrivilegeChecker for unit tests.
type mockPrivilegeChecker struct {
	privileges  map[string][]string // "username@vcenter" → granted privileges
	authErrors  map[string]bool     // "username@vcenter" → trigger auth failure
	otherErrors map[string]error    // "username@vcenter" → trigger non-auth error
}

func newMock() *mockPrivilegeChecker {
	return &mockPrivilegeChecker{
		privileges:  make(map[string][]string),
		authErrors:  make(map[string]bool),
		otherErrors: make(map[string]error),
	}
}

func (m *mockPrivilegeChecker) FetchUserPrivileges(_ context.Context, username, vcenterFQDN string) ([]string, error) {
	key := username + "@" + vcenterFQDN
	if m.authErrors[key] {
		return nil, fmt.Errorf("authentication failed: InvalidLogin for %s", username)
	}
	if err, ok := m.otherErrors[key]; ok {
		return nil, err
	}
	return m.privileges[key], nil
}

func setPrivileges(m *mockPrivilegeChecker, username, vcenter string, privs []string) {
	m.privileges[username+"@"+vcenter] = privs
}

func setAuthError(m *mockPrivilegeChecker, username, vcenter string) {
	m.authErrors[username+"@"+vcenter] = true
}

func setOtherError(m *mockPrivilegeChecker, username, vcenter string, err error) {
	m.otherErrors[username+"@"+vcenter] = err
}

// allExcept returns privs with the named elements removed.
func allExcept(privs []string, exclude ...string) []string {
	excSet := make(map[string]struct{}, len(exclude))
	for _, e := range exclude {
		excSet[e] = struct{}{}
	}
	out := make([]string, 0, len(privs))
	for _, p := range privs {
		if _, skip := excSet[p]; !skip {
			out = append(out, p)
		}
	}
	return out
}

const (
	vcenter1 = "vcenter1.example.com"
	vcenter2 = "vcenter2.example.com"
)

// ---------------------------------------------------------------------------
// AC1: Missing privilege → CredentialsProvisionFailed condition
// ---------------------------------------------------------------------------

func TestValidatePrivileges_MachineAPI_MissingOnePrivilege(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1,
		allExcept(machineAPIPrivileges, "VirtualMachine.Config.AddNewDisk"))

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	const wantMsg = "Account 'ocp-machine-api@vsphere.local' on vCenter 'vcenter1.example.com' missing privileges: [VirtualMachine.Config.AddNewDisk]"
	if err.Error() != wantMsg {
		t.Errorf("got:\n  %s\nwant:\n  %s", err.Error(), wantMsg)
	}
}

func TestValidatePrivileges_MachineAPI_MissingMultiplePrivileges(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1,
		allExcept(machineAPIPrivileges, "VirtualMachine.Config.AddNewDisk", "VirtualMachine.Interact.PowerOn", "Network.Assign"))

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	for _, priv := range []string{"Network.Assign", "VirtualMachine.Config.AddNewDisk", "VirtualMachine.Interact.PowerOn"} {
		if !strings.Contains(err.Error(), priv) {
			t.Errorf("expected %q in error message %q", priv, err.Error())
		}
	}
}

func TestValidatePrivileges_CSIDriver_MissingPrivilege(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-csi@vsphere.local", vcenter1,
		allExcept(storagePrivileges, "Datastore.FileManagement"))

	creds := []ComponentCredential{
		{Component: "storage", Username: "ocp-csi@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "ocp-csi@vsphere.local") ||
		!strings.Contains(err.Error(), vcenter1) ||
		!strings.Contains(err.Error(), "Datastore.FileManagement") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestValidatePrivileges_CloudController_MissingPrivilege(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-cloud-controller@vsphere.local", vcenter1,
		allExcept(cloudControllerPrivileges, "System.Read"))

	creds := []ComponentCredential{
		{Component: "cloud-controller", Username: "ocp-cloud-controller@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "System.Read") {
		t.Errorf("expected System.Read in error: %s", err.Error())
	}
}

func TestValidatePrivileges_VSphereProblemDetector_MissingPrivilege(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-vpd@vsphere.local", vcenter1,
		allExcept(vsphereProblemDetectorPrivileges, "Sessions.ValidateSession"))

	creds := []ComponentCredential{
		{Component: "vsphere-problem-detector", Username: "ocp-vpd@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Sessions.ValidateSession") {
		t.Errorf("expected Sessions.ValidateSession in error: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// AC2: All privileges present → nil error, provisioning proceeds
// ---------------------------------------------------------------------------

func TestValidatePrivileges_MachineAPI_FullPrivilegeSet(t *testing.T) {
	privs := privilegeRequirementsFor("machine-api")
	if len(privs) != 19 {
		t.Errorf("machine-api privilege set: got %d; want 19", len(privs))
	}

	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1, machineAPIPrivileges)

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidatePrivileges_CSIDriver_FullPrivilegeSet(t *testing.T) {
	privs := privilegeRequirementsFor("storage")
	if len(privs) != 6 {
		t.Errorf("csi-driver privilege set: got %d; want 6", len(privs))
	}

	m := newMock()
	setPrivileges(m, "ocp-csi@vsphere.local", vcenter1, storagePrivileges)

	creds := []ComponentCredential{
		{Component: "storage", Username: "ocp-csi@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidatePrivileges_CloudController_FullPrivilegeSet(t *testing.T) {
	privs := privilegeRequirementsFor("cloud-controller")
	if len(privs) != 3 {
		t.Errorf("cloud-controller privilege set: got %d; want 3", len(privs))
	}

	m := newMock()
	setPrivileges(m, "ocp-cc@vsphere.local", vcenter1, cloudControllerPrivileges)

	creds := []ComponentCredential{
		{Component: "cloud-controller", Username: "ocp-cc@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidatePrivileges_VSphereProblemDetector_FullPrivilegeSet(t *testing.T) {
	privs := privilegeRequirementsFor("vsphere-problem-detector")
	if len(privs) != 2 {
		t.Errorf("vsphere-problem-detector privilege set: got %d; want 2", len(privs))
	}

	m := newMock()
	setPrivileges(m, "ocp-vpd@vsphere.local", vcenter1, vsphereProblemDetectorPrivileges)

	creds := []ComponentCredential{
		{Component: "vsphere-problem-detector", Username: "ocp-vpd@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidatePrivileges_AllComponents_AllPrivilegesPresent(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1, machineAPIPrivileges)
	setPrivileges(m, "ocp-csi@vsphere.local", vcenter1, storagePrivileges)
	setPrivileges(m, "ocp-cc@vsphere.local", vcenter1, cloudControllerPrivileges)
	setPrivileges(m, "ocp-vpd@vsphere.local", vcenter1, vsphereProblemDetectorPrivileges)

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
		{Component: "storage", Username: "ocp-csi@vsphere.local", VCenterFQDN: vcenter1},
		{Component: "cloud-controller", Username: "ocp-cc@vsphere.local", VCenterFQDN: vcenter1},
		{Component: "vsphere-problem-detector", Username: "ocp-vpd@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// AC3: Authentication failure → "Authentication failed for" message
// ---------------------------------------------------------------------------

func TestValidatePrivileges_CSIDriver_AuthenticationFailure(t *testing.T) {
	m := newMock()
	setAuthError(m, "ocp-csi@vsphere.local", vcenter1)

	creds := []ComponentCredential{
		{Component: "storage", Username: "ocp-csi@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	const wantMsg = "Authentication failed for 'ocp-csi@vsphere.local' on vCenter 'vcenter1.example.com'"
	if err.Error() != wantMsg {
		t.Errorf("got:\n  %s\nwant:\n  %s", err.Error(), wantMsg)
	}
}

func TestValidatePrivileges_MachineAPI_AuthenticationFailure(t *testing.T) {
	m := newMock()
	setAuthError(m, "ocp-machine-api@vsphere.local", vcenter1)

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Authentication failed for") ||
		!strings.Contains(err.Error(), "ocp-machine-api@vsphere.local") ||
		!strings.Contains(err.Error(), vcenter1) {
		t.Errorf("unexpected auth failure message: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Adversarial cases
// ---------------------------------------------------------------------------

func TestValidatePrivileges_AccountWithNoPrivileges(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1, []string{})

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error when account has no privileges, got nil")
	}
	for _, priv := range machineAPIPrivileges {
		if !strings.Contains(err.Error(), priv) {
			t.Errorf("expected %q in error message", priv)
		}
	}
}

func TestValidatePrivileges_UnknownComponent_NoRequirements(t *testing.T) {
	m := newMock()
	creds := []ComponentCredential{
		{Component: "unknown-component", Username: "ocp-unknown@vsphere.local", VCenterFQDN: vcenter1},
	}
	if err := ValidateComponentPrivileges(context.Background(), m, creds); err != nil {
		t.Errorf("expected nil for unknown component, got %v", err)
	}
}

func TestValidatePrivileges_MultipleVCenters_OneVCenterFails(t *testing.T) {
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1, machineAPIPrivileges)
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter2,
		allExcept(machineAPIPrivileges, "VirtualMachine.Config.AddNewDisk"))

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter2},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error for vcenter2 missing privilege, got nil")
	}
	if !strings.Contains(err.Error(), vcenter2) {
		t.Errorf("error should identify vcenter2 as the failing vCenter: %s", err.Error())
	}
}

func TestValidatePrivileges_MultipleVCenters_BothFail(t *testing.T) {
	// ValidateComponentPrivileges uses fail-fast: returns on the first failure.
	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1,
		allExcept(machineAPIPrivileges, "Network.Assign"))
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter2,
		allExcept(machineAPIPrivileges, "VirtualMachine.Config.AddNewDisk"))

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter2},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error when both vCenters fail, got nil")
	}
	if !strings.Contains(err.Error(), vcenter1) {
		t.Errorf("fail-fast: expected error about vcenter1 (first), got: %s", err.Error())
	}
}

func TestValidatePrivileges_FetchPrivileges_NonAuthError(t *testing.T) {
	m := newMock()
	networkErr := fmt.Errorf("connection refused: timeout connecting to vcenter1.example.com:443")
	setOtherError(m, "ocp-machine-api@vsphere.local", vcenter1, networkErr)

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if strings.Contains(err.Error(), "Authentication failed") || strings.Contains(err.Error(), "missing privileges") {
		t.Errorf("non-auth error was misclassified: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected original error propagated, got: %s", err.Error())
	}
}

func TestValidatePrivileges_ConditionMessage_DoesNotContainPassword(t *testing.T) {
	const sensitivePassword = "s3cr3t-p@ssw0rd-unique"

	m := newMock()
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1,
		allExcept(machineAPIPrivileges, "Network.Assign"))

	creds := []ComponentCredential{
		{
			Component:   "machine-api",
			Username:    "ocp-machine-api@vsphere.local",
			Password:    sensitivePassword,
			VCenterFQDN: vcenter1,
		},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if strings.Contains(err.Error(), sensitivePassword) {
		t.Errorf("condition message must not contain the password; got: %s", err.Error())
	}
}

func TestValidatePrivileges_MissingPrivileges_SortedInMessage(t *testing.T) {
	m := newMock()
	// Remove three privileges that sort as: Datastore.AllocateSpace < Network.Assign < VirtualMachine.Inventory.Delete
	setPrivileges(m, "ocp-machine-api@vsphere.local", vcenter1,
		allExcept(machineAPIPrivileges, "VirtualMachine.Inventory.Delete", "Datastore.AllocateSpace", "Network.Assign"))

	creds := []ComponentCredential{
		{Component: "machine-api", Username: "ocp-machine-api@vsphere.local", VCenterFQDN: vcenter1},
	}
	err := ValidateComponentPrivileges(context.Background(), m, creds)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	msg := err.Error()
	iDA := strings.Index(msg, "Datastore.AllocateSpace")
	iNA := strings.Index(msg, "Network.Assign")
	iVID := strings.Index(msg, "VirtualMachine.Inventory.Delete")
	if !(iDA < iNA && iNA < iVID) {
		t.Errorf("missing privileges not in sorted order in message: %s", msg)
	}
}
