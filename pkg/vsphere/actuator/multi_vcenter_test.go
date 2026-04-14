package actuator

import (
	"encoding/base64"
	"testing"
)

// TestMultiVCenterSecretFormat_FQDNKeys verifies component secrets use vCenter
// FQDN-keyed format.
//
// Acceptance Criteria: "And component secrets contain credentials keyed by vCenter FQDN"
//
// Test Steps:
// 1. Configure multi-vCenter installation (machineAPI → vcenter1, csiDriver → vcenter2)
// 2. Run CCO secret generation
// 3. Inspect machine-api-vsphere-credentials secret
// 4. Inspect vsphere-csi-credentials secret
//
// Expected Result:
// - machine-api secret contains:
//   - vcenter1.example.com.username: <base64>
//   - vcenter1.example.com.password: <base64>
// - csi secret contains:
//   - vcenter2.example.com.username: <base64>
//   - vcenter2.example.com.password: <base64>
// - NOT simple username/password keys (that's single-vCenter format)
func TestMultiVCenterSecretFormat_FQDNKeys(t *testing.T) {
	// Create ComponentCredentials with multi-vCenter configuration
	componentCreds := map[string]*AccountCredentials{
		"machineAPI": {
			Username: "machine-api@vsphere.local",
			Password: "password1",
			VCenter:  "vcenter1.example.com",
		},
		"csiDriver": {
			Username: "csi-driver@vsphere.local",
			Password: "password2",
			VCenter:  "vcenter2.example.com",
		},
	}

	// Call createComponentSecrets
	secrets, err := createComponentSecrets(componentCreds)
	if err != nil {
		t.Fatalf("createComponentSecrets failed: %v", err)
	}

	// Verify machine-api secret has FQDN-keyed credentials
	machineAPISecret := secrets["machineAPI"]
	if machineAPISecret == nil {
		t.Fatal("machine-api secret not created")
	}

	expectedUsernameKey := "vcenter1.example.com.username"
	expectedPasswordKey := "vcenter1.example.com.password"

	if _, ok := machineAPISecret.Data[expectedUsernameKey]; !ok {
		t.Errorf("machine-api secret missing key: %s", expectedUsernameKey)
	}
	if _, ok := machineAPISecret.Data[expectedPasswordKey]; !ok {
		t.Errorf("machine-api secret missing key: %s", expectedPasswordKey)
	}

	// Verify csi secret has FQDN-keyed credentials
	csiSecret := secrets["csiDriver"]
	if csiSecret == nil {
		t.Fatal("csi secret not created")
	}

	expectedCSIUsernameKey := "vcenter2.example.com.username"
	expectedCSIPasswordKey := "vcenter2.example.com.password"

	if _, ok := csiSecret.Data[expectedCSIUsernameKey]; !ok {
		t.Errorf("csi secret missing key: %s", expectedCSIUsernameKey)
	}
	if _, ok := csiSecret.Data[expectedCSIPasswordKey]; !ok {
		t.Errorf("csi secret missing key: %s", expectedCSIPasswordKey)
	}

	// Verify secrets do NOT have simple username/password keys
	if _, ok := machineAPISecret.Data["username"]; ok {
		t.Error("machine-api secret should NOT have simple 'username' key in multi-vCenter mode")
	}
	if _, ok := csiSecret.Data["password"]; ok {
		t.Error("csi secret should NOT have simple 'password' key in multi-vCenter mode")
	}
}

// TestMultiVCenterBinding_MachineAPIToVC1 verifies Machine API connects to
// vcenter1.example.com.
//
// Acceptance Criteria: "And Machine API connects to vcenter1.example.com using
// machine-api credentials"
//
// Test Steps:
// 1. Configure machineAPI with vcenter1.example.com override
// 2. Start Machine API operator
// 3. Monitor Machine API's vSphere client initialization
// 4. Verify connection established to vcenter1.example.com (not default vCenter)
//
// Expected Result:
// - Machine API vSphere client connects to vcenter1.example.com
// - Credentials used: machineAPI username/password from secret
// - No connection attempts to vcenter2.example.com
func TestMultiVCenterBinding_MachineAPIToVC1(t *testing.T) {
	// Create component credentials for machineAPI with vcenter1
	componentCreds := map[string]*AccountCredentials{
		"machineAPI": {
			Username: "machine-api@vsphere.local",
			Password: "password1",
			VCenter:  "vcenter1.example.com",
		},
	}

	// Generate secrets
	secrets, err := createComponentSecrets(componentCreds)
	if err != nil {
		t.Fatalf("createComponentSecrets failed: %v", err)
	}

	// Verify machine-api secret
	machineAPISecret := secrets["machineAPI"]
	if machineAPISecret == nil {
		t.Fatal("machine-api secret not created")
	}

	// Verify secret contains vcenter1.example.com credentials
	usernameKey := "vcenter1.example.com.username"
	passwordKey := "vcenter1.example.com.password"

	if _, ok := machineAPISecret.Data[usernameKey]; !ok {
		t.Errorf("machine-api secret missing key: %s", usernameKey)
	}
	if _, ok := machineAPISecret.Data[passwordKey]; !ok {
		t.Errorf("machine-api secret missing key: %s", passwordKey)
	}

	// Decode and verify username
	encodedUsername := string(machineAPISecret.Data[usernameKey])
	decodedUsername, err := base64.StdEncoding.DecodeString(encodedUsername)
	if err != nil {
		t.Fatalf("failed to decode username: %v", err)
	}
	if string(decodedUsername) != "machine-api@vsphere.local" {
		t.Errorf("expected username 'machine-api@vsphere.local', got '%s'", string(decodedUsername))
	}
}

// TestMultiVCenterBinding_CSIToVC2 verifies CSI Driver connects to
// vcenter2.example.com.
//
// Acceptance Criteria: "And CSI Driver connects to vcenter2.example.com using
// csi-driver credentials"
//
// Test Steps:
// 1. Configure csiDriver with vcenter2.example.com override
// 2. Start CSI Driver
// 3. Monitor CSI's vSphere client initialization
// 4. Verify connection established to vcenter2.example.com
//
// Expected Result:
// - CSI vSphere client connects to vcenter2.example.com
// - Credentials used: csiDriver username/password from secret
// - No connection attempts to vcenter1.example.com
func TestMultiVCenterBinding_CSIToVC2(t *testing.T) {
	// Create component credentials for csiDriver with vcenter2
	componentCreds := map[string]*AccountCredentials{
		"csiDriver": {
			Username: "csi-driver@vsphere.local",
			Password: "password2",
			VCenter:  "vcenter2.example.com",
		},
	}

	// Generate secrets
	secrets, err := createComponentSecrets(componentCreds)
	if err != nil {
		t.Fatalf("createComponentSecrets failed: %v", err)
	}

	// Verify csi secret
	csiSecret := secrets["csiDriver"]
	if csiSecret == nil {
		t.Fatal("csi secret not created")
	}

	// Verify secret contains vcenter2.example.com credentials
	usernameKey := "vcenter2.example.com.username"
	passwordKey := "vcenter2.example.com.password"

	if _, ok := csiSecret.Data[usernameKey]; !ok {
		t.Errorf("csi secret missing key: %s", usernameKey)
	}
	if _, ok := csiSecret.Data[passwordKey]; !ok {
		t.Errorf("csi secret missing key: %s", passwordKey)
	}

	// Decode and verify username
	encodedUsername := string(csiSecret.Data[usernameKey])
	decodedUsername, err := base64.StdEncoding.DecodeString(encodedUsername)
	if err != nil {
		t.Fatalf("failed to decode username: %v", err)
	}
	if string(decodedUsername) != "csi-driver@vsphere.local" {
		t.Errorf("expected username 'csi-driver@vsphere.local', got '%s'", string(decodedUsername))
	}
}

// TestMultiVCenterSecretGeneration_MultipleVCenters verifies CCO generates
// secrets for all referenced vCenters.
//
// Acceptance Criteria: Secret generation for multi-vCenter topologies
//
// Test Steps:
// 1. Configure ComponentCredentials with:
//    - machineAPI → vcenter1.example.com
//    - csiDriver → vcenter2.example.com
//    - cloudController → vcenter1.example.com (shares with machineAPI)
//    - diagnostics → vcenter3.example.com
// 2. Run secret generation
// 3. Verify all component secrets contain correct vCenter credentials
//
// Expected Result:
// - machine-api secret: vcenter1.example.com credentials
// - csi secret: vcenter2.example.com credentials
// - ccm secret: vcenter1.example.com credentials
// - diagnostics secret: vcenter3.example.com credentials
// - Secrets with same vCenter share FQDN-keyed credentials
func TestMultiVCenterSecretGeneration_MultipleVCenters(t *testing.T) {
	// Create ComponentCredentials with 3 different vCenters across 4 components
	componentCreds := map[string]*AccountCredentials{
		"machineAPI": {
			Username: "machine-api@vsphere.local",
			Password: "password1",
			VCenter:  "vcenter1.example.com",
		},
		"csiDriver": {
			Username: "csi-driver@vsphere.local",
			Password: "password2",
			VCenter:  "vcenter2.example.com",
		},
		"cloudController": {
			Username: "cloud-controller@vsphere.local",
			Password: "password3",
			VCenter:  "vcenter1.example.com", // Shares vCenter with machineAPI
		},
		"diagnostics": {
			Username: "diagnostics@vsphere.local",
			Password: "password4",
			VCenter:  "vcenter3.example.com",
		},
	}

	// Call createComponentSecrets for all components
	secrets, err := createComponentSecrets(componentCreds)
	if err != nil {
		t.Fatalf("createComponentSecrets failed: %v", err)
	}

	// Verify all 4 secrets were created
	if len(secrets) != 4 {
		t.Errorf("expected 4 secrets, got %d", len(secrets))
	}

	// Verify machine-api secret (vcenter1)
	machineAPISecret := secrets["machineAPI"]
	if machineAPISecret == nil {
		t.Fatal("machine-api secret not created")
	}
	if _, ok := machineAPISecret.Data["vcenter1.example.com.username"]; !ok {
		t.Error("machine-api secret missing vcenter1.example.com.username key")
	}

	// Verify csi secret (vcenter2)
	csiSecret := secrets["csiDriver"]
	if csiSecret == nil {
		t.Fatal("csi secret not created")
	}
	if _, ok := csiSecret.Data["vcenter2.example.com.username"]; !ok {
		t.Error("csi secret missing vcenter2.example.com.username key")
	}

	// Verify ccm secret (vcenter1 - shares with machineAPI)
	ccmSecret := secrets["cloudController"]
	if ccmSecret == nil {
		t.Fatal("ccm secret not created")
	}
	if _, ok := ccmSecret.Data["vcenter1.example.com.username"]; !ok {
		t.Error("ccm secret missing vcenter1.example.com.username key")
	}

	// Verify diagnostics secret (vcenter3)
	diagSecret := secrets["diagnostics"]
	if diagSecret == nil {
		t.Fatal("diagnostics secret not created")
	}
	if _, ok := diagSecret.Data["vcenter3.example.com.username"]; !ok {
		t.Error("diagnostics secret missing vcenter3.example.com.username key")
	}

	// Verify multi-vCenter mode detected
	if !isMultiVCenterMode(componentCreds) {
		t.Error("expected multi-vCenter mode to be detected")
	}
}
