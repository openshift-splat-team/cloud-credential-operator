package actuator

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// machineAPIPrivileges lists the 19 vCenter privileges required by the Machine API operator.
var machineAPIPrivileges = []string{
	"Datastore.AllocateSpace",
	"Network.Assign",
	"Resource.AssignVMToPool",
	"VirtualMachine.Config.AddExistingDisk",
	"VirtualMachine.Config.AddNewDisk",
	"VirtualMachine.Config.AddRemoveDevice",
	"VirtualMachine.Config.AdvancedConfig",
	"VirtualMachine.Config.CPUCount",
	"VirtualMachine.Config.DiskExtend",
	"VirtualMachine.Config.EditDevice",
	"VirtualMachine.Config.Memory",
	"VirtualMachine.Config.RemoveDisk",
	"VirtualMachine.Config.Resource",
	"VirtualMachine.Config.Settings",
	"VirtualMachine.Interact.PowerOff",
	"VirtualMachine.Interact.PowerOn",
	"VirtualMachine.Interact.Reset",
	"VirtualMachine.Inventory.Create",
	"VirtualMachine.Inventory.Delete",
}

// storagePrivileges lists the 6 vCenter privileges required by the CSI driver.
var storagePrivileges = []string{
	"Datastore.AllocateSpace",
	"Datastore.FileManagement",
	"StoragePod.Config",
	"VirtualMachine.Config.AddExistingDisk",
	"VirtualMachine.Config.AddNewDisk",
	"VirtualMachine.Config.RemoveDisk",
}

// cloudControllerPrivileges lists the 3 vCenter privileges required by the Cloud Controller Manager.
var cloudControllerPrivileges = []string{
	"System.Read",
	"System.View",
	"VirtualMachine.Inventory.Create",
}

// vsphereProblemDetectorPrivileges lists the 2 vCenter privileges required by the vSphere Problem Detector.
var vsphereProblemDetectorPrivileges = []string{
	"Sessions.ValidateSession",
	"StorageProfile.View",
}

// privilegeRequirementsFor returns the required vCenter privileges for the given component.
// Returns nil for unknown components (no requirement → validation skipped).
func privilegeRequirementsFor(component string) []string {
	switch component {
	case "machine-api":
		return machineAPIPrivileges
	case "storage":
		return storagePrivileges
	case "cloud-controller":
		return cloudControllerPrivileges
	case "vsphere-problem-detector":
		return vsphereProblemDetectorPrivileges
	default:
		return nil
	}
}

// PrivilegeChecker retrieves the vCenter privileges granted to a user account.
type PrivilegeChecker interface {
	// FetchUserPrivileges returns the privileges the named user holds on the given vCenter.
	// Returns an authentication error (containing "authentication", "invalidlogin", or
	// "unauthenticated", case-insensitive) when the credentials are rejected by vCenter.
	FetchUserPrivileges(ctx context.Context, username, vcenterFQDN string) ([]string, error)
}

// ComponentCredential holds the vCenter login information for one component on one vCenter.
type ComponentCredential struct {
	Component   string
	Username    string
	Password    string
	VCenterFQDN string
}

// ValidateComponentPrivileges validates that each credential satisfies the privilege
// requirements for its component role. Returns the first validation error encountered
// (fail-fast semantics across multiple credentials / vCenters).
func ValidateComponentPrivileges(ctx context.Context, checker PrivilegeChecker, creds []ComponentCredential) error {
	for _, cred := range creds {
		required := privilegeRequirementsFor(cred.Component)
		if len(required) == 0 {
			continue
		}

		actual, err := checker.FetchUserPrivileges(ctx, cred.Username, cred.VCenterFQDN)
		if err != nil {
			if isAuthError(err) {
				return fmt.Errorf("Authentication failed for '%s' on vCenter '%s'", cred.Username, cred.VCenterFQDN)
			}
			return err
		}

		missing := missingPrivileges(required, actual)
		if len(missing) > 0 {
			sort.Strings(missing)
			return fmt.Errorf("Account '%s' on vCenter '%s' missing privileges: %v", cred.Username, cred.VCenterFQDN, missing)
		}
	}
	return nil
}

func isAuthError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "authentication") ||
		strings.Contains(msg, "invalidlogin") ||
		strings.Contains(msg, "unauthenticated")
}

func missingPrivileges(required, actual []string) []string {
	have := make(map[string]struct{}, len(actual))
	for _, p := range actual {
		have[p] = struct{}{}
	}
	var missing []string
	for _, p := range required {
		if _, ok := have[p]; !ok {
			missing = append(missing, p)
		}
	}
	return missing
}
